# Prompt 30 — Backend analytics API routes

Expose popularity analytics through the REST API.

Implement:
- `GET /api/v1/albums/{id}/stats` — return popularity summary for an album
- `GET /api/v1/assets/{id}/stats` — return popularity summary for an asset
- `GET /api/v1/albums/{id}/popular-assets` — return top assets by views within an album
- `GET /api/v1/admin/analytics/overview` — admin-only aggregate overview
- `GET /api/v1/popularity/status` — probe endpoint for frontend feature detection
- return empty/404 gracefully when analytics are not enabled
- ACL checks: stats endpoints respect album access, admin endpoint is admin-only
- tests

Do not change the analytics store implementation.
