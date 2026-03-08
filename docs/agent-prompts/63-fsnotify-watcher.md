# Prompt 63 — Replace polling watcher with fsnotify

Replace the polling-based filesystem watcher with `fsnotify` for efficient change detection.

## Current state

`backend/internal/watch/watch.go` polls the content directory on a timer (default 5s), walks the entire tree, and compares file state against the last scan. This works but is O(n) per tick for large directories.

The existing `ReconcileFunc` callback and debounce logic should be preserved — only the change detection mechanism changes.

## Implement

1. Add dependency: `go get github.com/fsnotify/fsnotify`

2. Rewrite `watch.go`:
   - Replace the polling loop with `fsnotify.NewWatcher()`.
   - Recursively add all directories under `contentRoot` at startup.
   - On `fsnotify.Create` events for directories, add the new directory to the watcher.
   - On `fsnotify.Remove` events for directories, remove from watcher.
   - Map events to dirty paths (parent directory of the changed file).
   - Keep the existing debounce logic: accumulate dirty paths, call `ReconcileFunc` after the debounce period elapses with no new events.

3. Keep the `Config` struct and `ReconcileFunc` type unchanged.

4. Keep `DirtyPaths()`, `PendingDir()`, `MarkClean()` for test observability.

5. Remove `ScanOnce()` and `ForceReconcile()` test helpers if they no longer make sense — or reimplement them for the new backend.

6. Update tests:
   - Test that creating a file in a watched directory marks it dirty.
   - Test that debounce groups rapid changes.
   - Test that new subdirectories are automatically watched.
   - Test graceful shutdown (context cancellation stops the watcher).

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Change the `ReconcileFunc` signature
- Change how `app.go` creates and starts the watcher
- Change the debounce delay default
