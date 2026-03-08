//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/perrito666/gollery/backend/internal/analytics"
)

func testDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("GOLLERY_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("GOLLERY_TEST_POSTGRES_DSN not set; skipping integration test")
	}
	return dsn
}

func setupStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()
	dsn := testDSN(t)

	pool, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	store := New(pool)
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrating: %v", err)
	}

	// Clean tables for a fresh test run.
	pool.Exec(ctx, "DELETE FROM popularity_daily")
	pool.Exec(ctx, "DELETE FROM analytics_events")

	return store
}

func TestRecordAndQuery(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	now := time.Now().UTC()

	// Record some events.
	events := []analytics.Event{
		{Type: analytics.EventAlbumView, ObjectID: "alb_test1", VisitorHash: "h1", CreatedAt: now},
		{Type: analytics.EventAlbumView, ObjectID: "alb_test1", VisitorHash: "h2", CreatedAt: now},
		{Type: analytics.EventAssetView, ObjectID: "alb_test1", VisitorHash: "h1", CreatedAt: now},
		{Type: analytics.EventOriginalHit, ObjectID: "alb_test1", VisitorHash: "h1", CreatedAt: now},
		{Type: analytics.EventDiscussionClick, ObjectID: "alb_test1", VisitorHash: "h1", CreatedAt: now},
	}
	for _, e := range events {
		if err := store.RecordEvent(ctx, e); err != nil {
			t.Fatalf("recording event: %v", err)
		}
	}

	// Aggregate.
	if err := store.AggregateDailyPopularity(ctx, now); err != nil {
		t.Fatalf("aggregating: %v", err)
	}

	// Query.
	summary, err := store.QueryPopularity(ctx, "alb_test1")
	if err != nil {
		t.Fatalf("querying: %v", err)
	}

	if summary.TotalViews != 3 {
		t.Errorf("total_views = %d, want 3", summary.TotalViews)
	}
	if summary.Views7d != 3 {
		t.Errorf("views_7d = %d, want 3", summary.Views7d)
	}
	if summary.Views30d != 3 {
		t.Errorf("views_30d = %d, want 3", summary.Views30d)
	}
	if summary.OriginalHits != 1 {
		t.Errorf("original_hits = %d, want 1", summary.OriginalHits)
	}
	if summary.DiscussionClicks != 1 {
		t.Errorf("discussion_clicks = %d, want 1", summary.DiscussionClicks)
	}
}

func TestPurgeOldEvents(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	old := time.Now().UTC().AddDate(0, 0, -100)
	recent := time.Now().UTC()

	store.RecordEvent(ctx, analytics.Event{
		Type: analytics.EventAlbumView, ObjectID: "alb_old", VisitorHash: "h", CreatedAt: old,
	})
	store.RecordEvent(ctx, analytics.Event{
		Type: analytics.EventAlbumView, ObjectID: "alb_new", VisitorHash: "h", CreatedAt: recent,
	})

	cutoff := time.Now().UTC().AddDate(0, 0, -90)
	if err := store.PurgeOldEvents(ctx, cutoff); err != nil {
		t.Fatalf("purging: %v", err)
	}

	// Old event should be gone, new should remain.
	var count int
	store.pool.QueryRow(ctx, "SELECT COUNT(*) FROM analytics_events").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 event after purge, got %d", count)
	}
}

func TestQueryPopularity_NoData(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	summary, err := store.QueryPopularity(ctx, "alb_nonexistent")
	if err != nil {
		t.Fatalf("querying: %v", err)
	}
	if summary.TotalViews != 0 {
		t.Errorf("total_views = %d, want 0", summary.TotalViews)
	}
}

func TestAggregateIdempotent(t *testing.T) {
	store := setupStore(t)
	ctx := context.Background()

	now := time.Now().UTC()
	store.RecordEvent(ctx, analytics.Event{
		Type: analytics.EventAlbumView, ObjectID: "alb_idem", VisitorHash: "h", CreatedAt: now,
	})

	// Aggregate twice — should produce the same result (ON CONFLICT UPDATE).
	store.AggregateDailyPopularity(ctx, now)
	store.AggregateDailyPopularity(ctx, now)

	summary, _ := store.QueryPopularity(ctx, "alb_idem")
	if summary.TotalViews != 1 {
		t.Errorf("total_views = %d, want 1 (idempotent)", summary.TotalViews)
	}
}
