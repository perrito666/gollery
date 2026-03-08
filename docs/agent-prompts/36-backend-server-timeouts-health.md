# Prompt 36 — Backend server timeouts and health check

Add production-safe HTTP server settings and a health endpoint.

Implement:
- `GET /healthz` endpoint returning `200 OK` with JSON body `{"status": "ok"}` — no auth required
- configure `http.Server` with `ReadTimeout`, `WriteTimeout`, `ReadHeaderTimeout`, `IdleTimeout` from `ServerConfig` (with sensible defaults)
- `MaxHeaderBytes` set to 1MB
- tests for health endpoint and timeout configuration

Do not change existing handler behavior.
