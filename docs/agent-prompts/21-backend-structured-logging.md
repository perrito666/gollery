# Prompt 21 — Backend structured logging

Replace `log.Printf` calls with `log/slog` structured logging.

Implement:
- a `Logger` setup helper in `internal/app` (or a small `internal/logging` package) that configures `slog` with JSON output
- update `api.go` to use structured logging with fields: method, path, status, duration, request_id
- add a logging middleware that wraps `http.Handler` and logs every request
- update derivative error logs to use slog with asset ID context
- tests for the middleware

Do not add external logging dependencies.
Use only `log/slog` from the standard library (Go 1.21+).
