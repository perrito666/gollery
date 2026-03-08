# Prompt 56 — Wire request ID into structured logging

Now that request IDs exist in context (prompt 55), include them in all log output.

## Implement

1. In `backend/internal/api/api.go`, update `Handler()` to wrap the `slog.Default()` logger with the request ID as an attribute. Add a small middleware (or extend `RequestIDMiddleware`) that replaces the logger in context:

```go
// Inside middleware, after request ID is set:
logger := slog.Default().With("request_id", requestID)
ctx = context.WithValue(ctx, slogKey, logger)
```

Or use `slog.NewLogLogger` / handler approach — whichever is simpler.

2. Update the request-scoped log calls in handlers (e.g., error logging in `writeError`, admin endpoints) to pull the logger from context rather than using `slog.Default()`.

3. Add one test: verify that a handler's log output includes the `request_id` attribute. Use `slog.NewJSONHandler` writing to a `bytes.Buffer` to capture output.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Refactor all logging in a single pass — only add request ID enrichment
- Change the logging package in `internal/logging/`
- Add any external logging libraries
