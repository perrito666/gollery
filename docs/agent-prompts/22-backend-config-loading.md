# Prompt 22 — Backend configuration loading

Implement server configuration loading for `galleryd`.

Implement:
- extend `ServerConfig` with missing fields: `analytics.dedup_window_seconds`, `analytics.retain_events_days`, `auth` section (provider, session_secret)
- `LoadServerConfig(path string) (*ServerConfig, error)` that reads a JSON config file
- environment variable overrides for sensitive fields (`GOLLERY_POSTGRES_DSN`, `GOLLERY_SESSION_SECRET`, `GOLLERY_LISTEN_ADDR`)
- `Validate()` for the extended config
- tests for loading, env overrides, and validation

Do not implement the auth provider or session store yet.
Do not wire into `main.go` yet.
