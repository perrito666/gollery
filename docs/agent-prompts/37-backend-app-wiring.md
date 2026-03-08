# Prompt 37 — Backend app wiring

Wire all subsystems together in the `app` package.

Implement:
- `app.Run(ctx context.Context, configPath string) error` that:
  1. loads `ServerConfig` from the config file
  2. sets up slog logger
  3. runs initial filesystem scan and builds the first snapshot
  4. initializes the API server with the snapshot
  5. initializes the auth backend (user store + session store)
  6. optionally connects to PostgreSQL and runs migrations if analytics enabled
  7. starts the filesystem watcher with a reconcile callback that rebuilds the snapshot and calls `SetSnapshot`
  8. optionally starts analytics retention background jobs
  9. starts the HTTP server
  10. waits for context cancellation and performs graceful shutdown
- error handling: log and return errors, do not panic
- tests for startup sequence with mocked components

Do not implement TLS directly — assume a reverse proxy handles it.
