# Prompt 73 — Derivative worker pool: API integration

Wire the derivative worker pool from prompt 72 into the API server.

## Implement

1. Add the `Pool` to the `Deps` struct in `api.go`:
   ```go
   DerivativePool *derive.Pool // optional — nil means synchronous generation
   ```

2. In asset handlers (`handleAssetThumbnail`, `handleAssetPreview`):
   - If the cached file exists, serve it immediately (no change).
   - If the cached file does NOT exist and `DerivativePool` is non-nil:
     - Submit a generation job to the pool.
     - Return `202 Accepted` with a `Retry-After: 2` header.
   - If `DerivativePool` is nil, fall back to synchronous generation (current behavior).

3. In `app.go`, create and start the pool:
   ```go
   pool := derive.NewPool(cacheLayout, runtime.NumCPU())
   pool.Start(ctx)
   ```
   Pass it into `Deps`.

4. Add tests:
   - Request for uncached thumbnail with pool returns 202.
   - Subsequent request after generation completes returns 200 with image.
   - Request for cached thumbnail returns 200 immediately regardless of pool.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Change the original synchronous `GenerateThumbnail`/`GeneratePreview` functions
- Remove the synchronous fallback (pool is optional)
