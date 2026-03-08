# Prompt 67 — OpenAPI spec: analytics, discussion, and access routes

Complete the OpenAPI spec with the remaining route groups.

## Implement

Add to `docs/openapi.yaml`:

1. Analytics routes:
   - `GET /api/v1/albums/{id}/stats` — 200 with `PopularitySummary`
   - `GET /api/v1/assets/{id}/stats` — 200 with `PopularitySummary`
   - `GET /api/v1/albums/{id}/popular-assets` — 200 with array
   - `GET /api/v1/admin/analytics/overview` — 200 with overview object
   - `GET /api/v1/popularity/status` — 200 with `{available: boolean}`

2. Discussion routes:
   - `GET /api/v1/albums/{id}/discussion-threads` — 200 with array of `DiscussionBindingResponse`
   - `POST /api/v1/albums/{id}/discussion-threads` — 201, requires auth + CSRF
   - `GET /api/v1/assets/{id}/discussion-threads` — 200
   - `POST /api/v1/assets/{id}/discussion-threads` — 201

3. Access routes:
   - `GET /api/v1/albums/{id}/access` — 200 with `AccessResponse`
   - `GET /api/v1/assets/{id}/access` — 200 with `AccessResponse`
   - `PATCH /api/v1/assets/{id}/access` — 200, requires admin + CSRF

4. Add remaining component schemas: `PopularitySummary`, `DiscussionBindingResponse`, `AccessResponse`.

## Verify

```bash
npx @redocly/cli lint docs/openapi.yaml
```

All 27 routes should now be documented.

## Do not

- Modify backend code
- Add code generation
