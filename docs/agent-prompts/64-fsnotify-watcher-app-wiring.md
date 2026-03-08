# Prompt 64 — Update app wiring for fsnotify watcher

After prompt 63 replaced the polling watcher, update `app.go` if needed and verify integration.

## Implement

1. Read `app.go` and check if the watcher creation/startup code needs changes. If the `Config` struct and `Run()` signature are unchanged, this may be a no-op.

2. If the watcher now requires initial directory setup (e.g., calling `watcher.AddRecursive(contentRoot)`), add that call in `app.Run()` before starting the watcher.

3. Update `app_test.go` if any watcher-related test setup changed.

4. Run the full test suite to catch any breakage.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test -count=1 ./...
```

## Do not

- Change any other subsystem
- Modify the reindex logic
