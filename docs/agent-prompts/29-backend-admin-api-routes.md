# Prompt 29 — Backend admin API routes

Implement admin-only management endpoints.

Implement:
- `POST /api/v1/admin/reindex` — trigger an immediate filesystem re-scan and snapshot rebuild
- `GET /api/v1/admin/status` — return server status (uptime, snapshot age, watcher state, content root)
- `GET /api/v1/admin/diagnostics` — return last scan errors and warnings
- admin-only enforcement on all endpoints
- tests

Do not change the watcher or scanner implementations.
