# Prompt 71 — App lifecycle phases

Refactor `app.Run()` from a single 150-line function into discrete lifecycle phases.

## Current state

`app.Run()` does everything sequentially: config → logging → scan → API → auth → analytics → watcher → HTTP → shutdown. As more features are added, this function will grow unwieldy.

## Implement

1. Define phase types in `app.go`:

```go
type App struct {
    cfg        *config.ServerConfig
    server     *api.Server
    watcher    *watch.Watcher
    httpServer *http.Server
    analytics  *pganalytics.Store
}

func (a *App) Init(configPath string) error    // Load config, set up logging
func (a *App) Wire(ctx context.Context) error  // Build snapshot, create server, auth, analytics
func (a *App) Serve(ctx context.Context) error // Start HTTP server and watcher
func (a *App) Shutdown(ctx context.Context) error // Graceful shutdown with deadline
```

2. Rewrite `Run()` to:
```go
func Run(ctx context.Context, configPath string) error {
    a := &App{}
    if err := a.Init(configPath); err != nil { return err }
    if err := a.Wire(ctx); err != nil { return err }
    return a.Serve(ctx) // blocks until shutdown
}
```

3. Move the `doReindex`, `extractConfigs`, `setupAuth`, `setupAnalytics` helpers into methods on `App`.

4. Update `app_test.go` — existing tests should work via `Run()`, but add tests for individual phases:
   - `TestApp_Init_InvalidConfig` — bad config path
   - `TestApp_Wire_BadContentRoot` — non-existent content dir

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Add dependency injection frameworks (fx, wire)
- Change the public `Run()` API — it should still work exactly as before
- Change HTTP behavior
