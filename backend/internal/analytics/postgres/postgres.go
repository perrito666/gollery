// Package postgres implements analytics storage using PostgreSQL.
package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"

	"github.com/perrito666/gollery/backend/internal/analytics"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Store implements analytics.Store using PostgreSQL via pgx.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new PostgreSQL analytics store.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Migrate runs all pending database migrations.
func (s *Store) Migrate(ctx context.Context) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}
	defer conn.Release()

	m, err := migrate.NewMigrator(ctx, conn.Conn(), "public.schema_version")
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}

	// fs.Sub strips the "migrations" prefix so tern sees files at root.
	subFS, err := fs.Sub(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating sub FS: %w", err)
	}

	if err := m.LoadMigrations(subFS); err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	if err := m.Migrate(ctx); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}

// RecordEvent inserts a single analytics event.
func (s *Store) RecordEvent(ctx context.Context, e analytics.Event) error {
	createdAt := e.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO analytics_events (event_type, object_id, visitor_hash, created_at)
		 VALUES ($1, $2, $3, $4)`,
		string(e.Type), e.ObjectID, e.VisitorHash, createdAt,
	)
	if err != nil {
		return fmt.Errorf("recording event: %w", err)
	}
	return nil
}

// QueryPopularity returns the popularity summary for an object.
func (s *Store) QueryPopularity(ctx context.Context, objectID string) (*analytics.PopularitySummary, error) {
	now := time.Now().UTC()
	day7 := now.AddDate(0, 0, -7)
	day30 := now.AddDate(0, 0, -30)

	summary := &analytics.PopularitySummary{ObjectID: objectID}

	// Total views (from daily aggregates).
	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(view_count), 0) FROM popularity_daily
		 WHERE object_id = $1 AND event_type IN ('album_view', 'asset_view')`,
		objectID,
	).Scan(&summary.TotalViews)
	if err != nil {
		return nil, fmt.Errorf("querying total views: %w", err)
	}

	// Views 7d.
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(view_count), 0) FROM popularity_daily
		 WHERE object_id = $1 AND event_type IN ('album_view', 'asset_view') AND day >= $2`,
		objectID, day7,
	).Scan(&summary.Views7d)
	if err != nil {
		return nil, fmt.Errorf("querying views_7d: %w", err)
	}

	// Views 30d.
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(view_count), 0) FROM popularity_daily
		 WHERE object_id = $1 AND event_type IN ('album_view', 'asset_view') AND day >= $2`,
		objectID, day30,
	).Scan(&summary.Views30d)
	if err != nil {
		return nil, fmt.Errorf("querying views_30d: %w", err)
	}

	// Original hits.
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(view_count), 0) FROM popularity_daily
		 WHERE object_id = $1 AND event_type = 'original_hit'`,
		objectID,
	).Scan(&summary.OriginalHits)
	if err != nil {
		return nil, fmt.Errorf("querying original_hits: %w", err)
	}

	// Discussion clicks.
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(view_count), 0) FROM popularity_daily
		 WHERE object_id = $1 AND event_type = 'discussion_click'`,
		objectID,
	).Scan(&summary.DiscussionClicks)
	if err != nil {
		return nil, fmt.Errorf("querying discussion_clicks: %w", err)
	}

	return summary, nil
}

// AggregateDailyPopularity rolls up raw events into the daily aggregates table.
func (s *Store) AggregateDailyPopularity(ctx context.Context, day time.Time) error {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	_, err := s.pool.Exec(ctx,
		`INSERT INTO popularity_daily (object_id, event_type, day, view_count)
		 SELECT object_id, event_type, $1::date, COUNT(*)
		 FROM analytics_events
		 WHERE created_at >= $2 AND created_at < $3
		 GROUP BY object_id, event_type
		 ON CONFLICT (object_id, event_type, day)
		 DO UPDATE SET view_count = EXCLUDED.view_count`,
		dayStart, dayStart, dayEnd,
	)
	if err != nil {
		return fmt.Errorf("aggregating daily popularity: %w", err)
	}
	return nil
}

// PurgeOldEvents deletes events older than the given time.
func (s *Store) PurgeOldEvents(ctx context.Context, olderThan time.Time) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM analytics_events WHERE created_at < $1`,
		olderThan,
	)
	if err != nil {
		return fmt.Errorf("purging old events: %w", err)
	}
	return nil
}

// Close releases the connection pool.
func (s *Store) Close() error {
	s.pool.Close()
	return nil
}

// Connect creates a new pgx connection pool from a DSN string.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing DSN: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return pool, nil
}

// ConnectSingle creates a single pgx connection (useful for migrations in tests).
func ConnectSingle(ctx context.Context, dsn string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	return conn, nil
}
