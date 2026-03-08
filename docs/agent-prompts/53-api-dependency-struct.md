# Prompt 53 — API server dependency struct

Replace the scattered `Set*()` setter methods on `api.Server` with a single `Dependencies` struct passed to `NewServer()`.

## Current state

`api.Server` has seven setter methods called in sequence by `app.Run()`:

- `SetSnapshot(snap, configs)`
- `SetContentRoot(contentRoot, cacheLayout)`
- `SetAuth(authenticator, sessions, csrfSecret, rateLimitCfg)`
- `SetDiscussions(svc)`
- `SetAnalytics(store)`
- `SetAdmin(reindexFunc)`
- `SetScanErrors(errs)`

The problem: callers can forget a setter, order is implicit, and the struct has many fields that are nil until set.

## Implement

1. Define a `Deps` struct in `api/api.go` grouping all external dependencies:

```go
type Deps struct {
    Snapshot    *domain.Snapshot
    Configs     map[string]*config.AlbumConfig
    ContentRoot string
    CacheLayout *cache.Layout

    // Optional — nil means feature disabled
    Authenticator auth.Authenticator
    Sessions      *auth.CookieSessionStore
    CSRFSecret    string
    RateLimitCfg  *RateLimitConfig
    Discussions   *discussion.Service
    Analytics     AnalyticsStore
    ReindexFunc   func() error
}
```

2. Change `NewServer(deps Deps) *Server` to accept the struct. Build all internal indexes in the constructor.

3. Keep `SetSnapshot()` as the only remaining setter — it is called at runtime by the watcher to hot-reload content. Also keep `SetScanErrors()` since scan errors update at runtime.

4. Update `app.Run()` in `app.go` to build the `Deps` struct and pass it to `NewServer()`.

5. Remove all other `Set*()` methods.

6. Update all existing tests that call `NewServer()` or `Set*()` methods.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Change the HTTP API behavior or routes
- Change the middleware chain
- Modify the `SetSnapshot()` hot-reload path
