# Prompt 66 — OpenAPI spec: auth and admin routes

Continue the OpenAPI spec from prompt 65 with auth and admin route documentation.

## Implement

Add to `docs/openapi.yaml`:

1. Auth routes:
   - `POST /api/v1/auth/login` — body: `LoginRequest {username, password}`, 200/401
   - `GET /api/v1/auth/me` — 200 with `MeResponse`, 401
   - `POST /api/v1/auth/logout` — 200, requires CSRF token header
   - `GET /api/v1/auth/csrf-token` — 200 with `{token: string}`

2. Admin routes (require admin role):
   - `POST /api/v1/admin/reindex` — 200, requires CSRF token
   - `GET /api/v1/admin/status` — 200 with status object
   - `GET /api/v1/admin/diagnostics` — 200 with diagnostics object

3. Add component schemas: `LoginRequest`, `MeResponse`, `StatusResponse`, `DiagnosticsResponse`.

4. Add security requirement annotations (admin routes require `gollery_session` with admin role).

## Verify

```bash
npx @redocly/cli lint docs/openapi.yaml
```

## Do not

- Document analytics, discussion, or access routes (next prompts)
- Modify backend code
