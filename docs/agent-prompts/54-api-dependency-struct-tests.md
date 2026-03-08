# Prompt 54 — Update tests for dependency struct

After prompt 53 changed `NewServer()` to accept a `Deps` struct, verify and fix all test files.

## Files to check

- `backend/internal/api/api_test.go`
- `backend/internal/api/health_test.go`
- `backend/internal/api/navigation_test.go`
- `backend/internal/api/pagination_test.go`
- `backend/internal/api/access_test.go`
- `backend/internal/api/admin_test.go`
- `backend/internal/api/analytics_test.go`
- `backend/internal/api/analytics_middleware_test.go`
- `backend/internal/api/discussion_test.go`
- `backend/internal/api/albums_query_test.go`
- `backend/internal/api/middleware_test.go`
- `backend/internal/api/ratelimit_test.go`
- `backend/internal/app/app_test.go`

## Implement

1. Read each test file. Replace `NewServer(snap, configs)` + `Set*()` calls with `NewServer(Deps{...})`.
2. For tests that only need a snapshot, pass a `Deps` with just `Snapshot` and `Configs` filled in.
3. For tests that need auth, discussions, or analytics, fill the corresponding `Deps` field.
4. Add one new test: `TestNewServerNilOptionalDeps` — construct with only required fields, call `Handler()`, confirm it doesn't panic.

## Verify

```bash
cd backend && go test -count=1 ./...
```

## Do not

- Change any test assertions or expected behavior
- Add new test coverage beyond what's listed above
