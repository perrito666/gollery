# Prompt 28 — Backend access info API routes

Expose access configuration through the REST API.

Implement:
- `GET /api/v1/albums/{id}/access` — return effective ACL for an album (admin only)
- `GET /api/v1/assets/{id}/access` — return effective ACL for an asset (admin only)
- `PATCH /api/v1/assets/{id}/access` — update per-asset ACL override in sidecar state (admin only)
- request body validation for PATCH
- tests

Do not change the ACL evaluation engine.
