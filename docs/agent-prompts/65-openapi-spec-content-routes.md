# Prompt 65 — OpenAPI spec: content routes

Write the first section of an OpenAPI 3.1 specification covering the public content routes.

## Implement

1. Create `docs/openapi.yaml` with:
   - OpenAPI 3.1 header, info block (title: Gollery API, version: 1.0.0)
   - Server entry for `http://localhost:8080`
   - Security scheme: cookie-based session (`gollery_session`)

2. Document these routes with request/response schemas:
   - `GET /healthz` — 200 with `{"status": "ok"}`
   - `GET /api/v1/albums/root` — 200 with AlbumResponse schema
   - `GET /api/v1/albums/{id}` — 200/403/404, query params: `offset`, `limit`
   - `GET /api/v1/albums?path={path}` — 200/403/404
   - `GET /api/v1/assets/{id}` — 200/403/404 with AssetResponse schema
   - `GET /api/v1/assets/{id}/thumbnail?size={n}` — 200 image/jpeg, 403/404
   - `GET /api/v1/assets/{id}/preview?size={n}` — 200 image/jpeg, 403/404
   - `GET /api/v1/assets/{id}/original` — 200 (file), 403/404

3. Define component schemas: `AlbumResponse`, `AssetResponse`, `AssetSummary`, `APIError`.

Match the actual response shapes from `api.go` type definitions.

## Verify

Validate the spec:
```bash
npx @redocly/cli lint docs/openapi.yaml
```

If the linter isn't available, at minimum verify the YAML is valid.

## Do not

- Document auth, admin, analytics, discussion, or access routes (subsequent prompts)
- Add code generation or validation middleware
