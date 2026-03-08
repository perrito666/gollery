# Prompt 35 — Backend API route group split

Split the single `api.go` file into route group files.

Implement:
- `api/albums.go` — album handlers (root, by-id, by-path)
- `api/assets.go` — asset handlers (by-id, thumbnail, preview, original)
- `api/auth_handlers.go` — auth handlers (login, me, logout, csrf-token)
- `api/discussions.go` — discussion handlers
- `api/access_handlers.go` — access handlers
- `api/admin.go` — admin handlers
- `api/analytics_handlers.go` — analytics handlers
- keep `api.go` for `Server` struct, `Handler()` route registration, `writeJSON`, `writeError`

Do not change any handler behavior.
Verify all existing tests still pass.
