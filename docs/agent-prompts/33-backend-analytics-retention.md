# Prompt 33 — Backend analytics retention job

Implement periodic analytics maintenance.

Implement:
- a background goroutine that runs `PurgeOldEvents` at a configurable interval (default: daily)
- a background goroutine that runs `AggregateDailyPopularity` for the previous day
- respect `retain_events_days` from server config
- graceful shutdown via context cancellation
- logging of maintenance actions via slog
- tests (use short intervals and verify purge/aggregate are called)

Do not change the analytics store interface.
