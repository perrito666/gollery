# Prompt 74 — Content-addressable cache keys

Change cache key format from `{assetID}_{size}.jpg` to a content-addressed scheme that includes generation parameters.

## Current state

Cache paths: `thumbs/ast_abc123_200.jpg`. If the source file changes or quality settings change, the stale cached file is served until explicitly purged.

## Implement

1. In `cache.go`, add a function to compute a content-addressed key:

```go
// CacheKey computes a deterministic key from source file identity and generation params.
// Uses SHA256(assetID + ":" + sourceModTime + ":" + size + ":" + quality).
// Returns first 16 hex characters.
func CacheKey(assetID string, sourceModTime time.Time, size, quality int) string
```

2. Update `ThumbPath` and `PreviewPath` to accept the new key:
   ```go
   func (l *Layout) ThumbPath(key string) string  // thumbs/{key}.jpg
   func (l *Layout) PreviewPath(key string) string // previews/{key}.jpg
   ```

3. Update `derive.go` to compute the cache key before checking existence:
   - `GenerateThumbnail` and `GeneratePreview` now accept source mod time.
   - Compute key, check cache, generate if missing.

4. Update `PurgeOrphans` — it no longer extracts asset IDs from filenames. Instead, collect all valid cache keys from the current snapshot and remove files whose names don't match.

5. Update all callers in `api/assets.go` to pass source mod time when calling derive functions.

6. Update all tests.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Add WebP/AVIF support (separate concern)
- Change the JPEG quality setting
- Change the API response format
