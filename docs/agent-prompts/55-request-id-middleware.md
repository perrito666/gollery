# Prompt 55 — Request ID middleware

Add a middleware that assigns a unique ID to every HTTP request for log correlation.

## Implement

1. In `backend/internal/api/middleware.go`:
   - Define a context key type and `RequestIDFromContext(ctx) string` helper.
   - Write `RequestIDMiddleware(next http.Handler) http.Handler` that:
     - Reads `X-Request-ID` from the incoming request header. If present and non-empty, use it.
     - Otherwise, generate a new UUID v4 using `crypto/rand` (no external deps — format 16 random bytes as `xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`).
     - Store the ID in the request context.
     - Set the `X-Request-ID` response header.

2. In `backend/internal/api/api.go`, add `RequestIDMiddleware` to the middleware chain as the **outermost** wrapper (before security headers) so all subsequent middleware and handlers have access to the ID.

3. In `backend/internal/api/middleware.go` or a new `request_id_test.go`, add tests:
   - Request without `X-Request-ID` header gets one generated and returned.
   - Request with `X-Request-ID` header has the same value echoed back.
   - `RequestIDFromContext` returns the correct ID.
   - Generated IDs are valid UUID v4 format.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Add external UUID libraries — use `crypto/rand` directly
- Change existing middleware behavior
- Modify logging yet (that's the next prompt)
