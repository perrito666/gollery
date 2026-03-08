package analytics

import (
	"context"
	"log/slog"
	"time"
)

// RetentionConfig configures the background maintenance jobs.
type RetentionConfig struct {
	// RetainEventsDays is the number of days to keep raw events.
	RetainEventsDays int
	// RunInterval is how often maintenance runs (default: 24h).
	RunInterval time.Duration
}

// StartRetentionJobs starts background goroutines for analytics maintenance.
// The goroutines run PurgeOldEvents and AggregateDailyPopularity at the
// configured interval. They stop when ctx is cancelled.
func StartRetentionJobs(ctx context.Context, store Store, cfg RetentionConfig) {
	interval := cfg.RunInterval
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	retainDays := cfg.RetainEventsDays
	if retainDays <= 0 {
		retainDays = 90
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("analytics retention job stopped")
				return
			case <-ticker.C:
				runMaintenance(ctx, store, retainDays)
			}
		}
	}()
}

func runMaintenance(ctx context.Context, store Store, retainDays int) {
	// Purge old events.
	cutoff := time.Now().AddDate(0, 0, -retainDays)
	if err := store.PurgeOldEvents(ctx, cutoff); err != nil {
		slog.Error("purge old events failed", "error", err)
	} else {
		slog.Info("purged old events", "older_than", cutoff.Format(time.DateOnly))
	}

	// Aggregate yesterday's events.
	yesterday := time.Now().AddDate(0, 0, -1)
	if err := store.AggregateDailyPopularity(ctx, yesterday); err != nil {
		slog.Error("aggregate daily popularity failed", "error", err)
	} else {
		slog.Info("aggregated daily popularity", "day", yesterday.Format(time.DateOnly))
	}
}
