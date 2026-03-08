# Prompt 27 — Backend discussion API routes

Expose discussion bindings through the REST API.

Implement:
- `GET /api/v1/albums/{id}/discussion-threads` — list bindings for an album
- `POST /api/v1/albums/{id}/discussion-threads` — create a binding (admin only)
- `GET /api/v1/assets/{id}/discussion-threads` — list bindings for an asset
- `POST /api/v1/assets/{id}/discussion-threads` — create a binding (admin only)
- ACL checks on all endpoints
- admin-only enforcement on POST endpoints
- tests

Do not change the discussion service or provider implementations.
