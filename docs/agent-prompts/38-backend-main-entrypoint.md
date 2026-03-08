# Prompt 38 — Backend main entrypoint

Replace the placeholder `main.go` with a real server entrypoint.

Implement:
- parse command-line flags: `--config` (path to config file, default `gollery.json`), `--version`
- call `app.Run(ctx, configPath)` with a signal-aware context (SIGINT, SIGTERM)
- print version info if `--version` is passed
- log startup and shutdown events
- exit with code 1 on fatal errors

Do not add daemon/background mode.
Keep it simple — under 50 lines.
