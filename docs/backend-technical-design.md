# Backend Technical Design
## gollery backend v3

Language: English  
Implementation language: Go  
Storage model: filesystem-first content + optional PostgreSQL popularity analytics  
License: MIT

---

## 1. Summary

The backend is a filesystem-first image gallery API.

It discovers albums and assets from a filesystem tree, applies publication and policy rules from `album.json`, loads mutable sidecar state from `.gallery/*.state.json`, generates derivatives such as thumbnails, enforces ACLs, and exposes a REST API.

Optional discussion providers allow albums or assets to be associated with social discussion threads such as Mastodon or Bluesky.

Optional popularity analytics are stored in PostgreSQL and designed to be GDPR-safe for Europe.

---

## 2. Core design decisions

1. Filesystem is the canonical source of truth for content.
2. `album.json` is declarative only.
3. `.gallery/*.state.json` stores mutable editorial state.
4. Only folders inside a published subtree are visible.
5. Stable object IDs are persisted in sidecar state files.
6. ACLs support `public`, `authenticated`, and `restricted`.
7. Discussions are provider-pluggable.
8. Popularity analytics are optional and stored in PostgreSQL.
9. Analytics must be privacy-preserving and must not be required for gallery correctness.
10. Backend and frontend live in the same monorepo but remain logically separated.

---

## 3. Filesystem-first model

The filesystem contains:
- albums
- subalbums
- image assets

Each published folder may contain `album.json`, which defines:
- title
- description
- visibility defaults
- discussion policy
- analytics policy
- derivative defaults

Mutable sidecar state lives in:
- `.gallery/album.state.json`
- `.gallery/assets/<filename>.json`

This state contains:
- stable object IDs
- discussion bindings
- per-asset ACL overrides

Generated artifacts live outside the content tree:
- `/gallery-cache/thumbs`
- `/gallery-cache/previews`
- `/gallery-cache/snapshots`
- `/gallery-cache/meta`

---

## 4. Publication rules

A folder is visible if:
- it has a valid `album.json`, or
- one of its ancestors has a valid `album.json`

Folders outside a published subtree are invisible to the API.

Local config may be invalid. Recommended default behavior:
- log/record the config error
- ignore the invalid local config
- continue using inherited config when safe

---

## 5. Inheritance rules

Config is inherited from the nearest published ancestor down to the folder being resolved.

Merge semantics:
- scalar values: child overrides parent
- objects/maps: merge by key
- arrays/lists: child replaces parent

If `"inherit": false`, parent inheritance stops, but global server defaults still apply.

---

## 6. Access control

Modes:
- `public`
- `authenticated`
- `restricted`

Example:

```json
{
  "access": {
    "view": "restricted",
    "allowed_users": ["horacio", "maria"],
    "allowed_groups": ["family"],
    "admins": ["horacio"]
  }
}
```

Asset-level overrides may be stored in asset sidecars.

Global admin always overrides object-level restrictions for administrative operations.

---

## 7. Identity model

IDs must not be derived only from paths because discussions and future dynamic features need stable object identity.

Use persisted IDs in sidecar state:
- album IDs: `alb_<ulid-like>`
- asset IDs: `ast_<ulid-like>`

Scanner/index behavior:
1. load sidecar state
2. reuse `object_id` if present
3. create one if absent
4. persist it back atomically

---

## 8. Discussions

Initial providers:
- Mastodon
- Bluesky

Bindings are stored in sidecar state and may include:
- provider
- remote ID
- URL
- created_at
- created_by
- provider metadata

Discussion publication is editorial state. It belongs in sidecars, not in `album.json`.

---

## 9. Optional popularity analytics

Track, as precisely as practical, events such as:
- album views
- asset detail views
- original asset hits
- discussion clicks

Do not track thumbnails by default.

Use PostgreSQL for analytics.

Reasons:
- operationally natural for this deployment style
- simpler in Kubernetes / Ansible environments
- strong support for aggregation and concurrency
- cleaner separation from filesystem editorial state

Analytics must not become a dependency for rendering the gallery correctly.

---

## 10. GDPR-safe analytics requirements

Popularity tracking must be useful but privacy-preserving.

Must not store by default:
- raw IP addresses
- precise geolocation
- full user agent strings
- unnecessary personal identifiers

Recommended privacy pattern:
- hash IP with a server-side salt if deduplication is needed
- use a session identifier if available
- keep retention bounded
- make analytics optional and configurable
- expose only aggregate data publicly
- allow disabling analytics by subtree

Example `album.json` policy:

```json
{
  "analytics": {
    "enabled": true,
    "track_album_views": true,
    "track_asset_views": true,
    "track_original_hits": true,
    "expose_popularity": false
  }
}
```

Example global config:

```json
{
  "analytics": {
    "enabled": true,
    "backend": "postgres",
    "postgres_dsn_env": "GOLLERY_POSTGRES_DSN",
    "hash_ip": true,
    "dedup_window_seconds": 1800,
    "retain_events_days": 90
  }
}
```

Suggested PostgreSQL tables:
- `analytics_events`
- `popularity_daily`

Admin analytics may expose:
- total views
- views_7d
- views_30d
- original hits
- discussion clicks
- trending score

---

## 11. API shape

Albums:
- `GET /api/v1/albums/root`
- `GET /api/v1/albums/{id}`
- `GET /api/v1/albums?path=/relative/path`

Assets:
- `GET /api/v1/assets/{id}`
- `GET /api/v1/assets/{id}/original`
- `GET /api/v1/assets/{id}/thumbnail?size=400`
- `GET /api/v1/assets/{id}/preview?size=1600`

Discussions:
- `GET /api/v1/albums/{id}/discussion-threads`
- `POST /api/v1/albums/{id}/discussion-threads`
- `GET /api/v1/assets/{id}/discussion-threads`
- `POST /api/v1/assets/{id}/discussion-threads`

Access:
- `GET /api/v1/albums/{id}/access`
- `GET /api/v1/assets/{id}/access`
- `PATCH /api/v1/assets/{id}/access`

Auth:
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout`

Admin:
- `POST /api/v1/admin/reindex`
- `GET /api/v1/admin/status`
- `GET /api/v1/admin/diagnostics`

Analytics:
- `GET /api/v1/albums/{id}/stats`
- `GET /api/v1/assets/{id}/stats`
- `GET /api/v1/albums/{id}/popular-assets`
- `GET /api/v1/admin/analytics/overview`

---

## 12. Recommended package layout

```text
backend/
  cmd/galleryd
  internal/app
  internal/config
  internal/domain
  internal/fswalk
  internal/state
  internal/index
  internal/meta
  internal/derive
  internal/access
  internal/auth
  internal/discussion
  internal/discussion/providers/mastodon
  internal/discussion/providers/bluesky
  internal/analytics
  internal/analytics/postgres
  internal/api
  internal/watch
  internal/cache
  migrations/
```

---

## 13. Implementation order

1. repo skeleton
2. config/domain/merge logic
3. scanner + publication rules
4. sidecar state + stable IDs
5. snapshot builder
6. ACL engine
7. auth abstraction
8. baseline REST API
9. derivatives
10. discussion provider abstraction
11. Mastodon and Bluesky providers
12. analytics schema and service (PostgreSQL)
13. watcher and reconciliation
14. hardening

---

## 14. Final summary

Filesystem owns:
- content
- publication rules
- editorial state
- stable IDs
- discussion bindings
- derivative caches

PostgreSQL owns:
- optional analytics events
- popularity aggregates
