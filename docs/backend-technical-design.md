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

## 9. Derivatives and cache

### What derivatives are

Derivatives are resized versions of source images, generated on demand by the server. There are two kinds:

- **Thumbnails** — small images (default 400px longest edge) used in album grid views.
- **Previews** — larger images (default 1600px longest edge) used in the asset detail/lightbox view.

The original source file is never modified. Derivatives are always JPEG regardless of the source format (JPEG or PNG).

### Cache directory layout

Derivatives are stored in a cache directory outside the content tree:

```text
<cache-root>/
├── thumbs/
│   ├── ast_a1b2c3_400.jpg
│   └── ast_d4e5f6_200.jpg
└── previews/
    ├── ast_a1b2c3_1600.jpg
    └── ast_d4e5f6_1200.jpg
```

The cache root is configured via `derivative_cache_dir` in the server config and defaults to `.gallery-cache` relative to the content root.

### Filename convention

Each cached file is named `<assetID>_<size>.jpg`, where:

- `assetID` is the stable sidecar ID (e.g. `ast_a1b2c3`)
- `size` is the longest-edge pixel count

A single asset can have multiple cached sizes if clients request different dimensions.

### Generation flow

1. API handler receives request (e.g. `GET /api/v1/assets/{id}/thumbnail?size=400`).
2. Handler looks up the asset by ID in the in-memory index, checks ACL.
3. Handler calls `derive.GenerateThumbnail(layout, assetID, sourcePath, size)`.
4. Derive function computes the expected cache path and checks if it exists (cache hit → return immediately).
5. On cache miss: decode source image, scale with CatmullRom interpolation (aspect-ratio preserving, no upscaling), encode as JPEG quality 85, write to cache path.
6. Handler serves the resulting file via `http.ServeFile`.

### Scaling algorithm

CatmullRom interpolation from `golang.org/x/image/draw` is used for high-quality downscaling. Images are never upscaled — if both dimensions are already within the requested size, the original dimensions are preserved.

### Cache eviction

There is no TTL-based expiration. Source images are treated as immutable — if a file changes, it gets a new sidecar ID, so old derivatives become orphans.

`cache.PurgeOrphans(layout, knownAssetIDs)` scans both subdirectories and removes any file whose asset ID prefix is not in the known set. This runs after re-indexing.

### Path safety

All cache paths are constructed by `cache.Layout` methods using `filepath.Join` on the configured root plus a filename built from the asset ID and size integer. Asset IDs come from the sidecar state layer (`ast_<hex>` format) and are never derived from user input. The API layer resolves IDs from an in-memory map; raw URL parameters never reach path construction.

### Concurrency

Multiple requests can generate derivatives concurrently. File creation for distinct paths is naturally safe. Two concurrent requests for the same asset+size may both generate, but the output is identical so the last writer wins harmlessly.

---

## 10. In-memory data model and scaling

### Everything is in memory

At runtime, the server holds a complete in-memory representation of the gallery:

- **Snapshot** (`domain.Snapshot`) — the full album tree with all assets, built by `index.BuildSnapshot` from a filesystem scan plus sidecar state.
- **Index maps** — `albumsByID`, `albumsByPath`, and `assetsByID` for O(1) lookups.
- **Album configs** — the merged `config.AlbumConfig` for each album path, including ACL rules.
- **Sessions** — user sessions in `CookieSessionStore` (in-memory map).
- **Users** — loaded from `users.json` once at startup into `FileUserStore` (in-memory map).
- **Watcher state** — last-seen modtime/size for every file in the content tree.

There are **no database queries, no file reads, and no network calls** on the API request hot path (except derivative generation on cache miss, which reads the source image file).

### ACL evaluation at request time

ACL checks are pure functions over in-memory data:

1. Handler looks up album/asset by ID from the index maps.
2. Retrieves the merged `AccessConfig` from the configs map.
3. For assets, merges album ACL with per-asset override via `access.EffectiveAssetACL`.
4. Calls `access.CheckView(acl, principal)` — a simple switch on the view mode.

Album listing responses are ACL-filtered: `albumToResponse` iterates assets and children, calling `CheckView` on each, so the response never leaks restricted content.

### Concurrency

The snapshot and indexes are protected by a `sync.RWMutex`:
- All API handlers take the **read lock** (concurrent).
- `SetSnapshot` takes the **write lock** to swap in a new snapshot (brief exclusive lock).

This means API requests are fully concurrent with each other and only block momentarily during re-index.

### Memory usage estimates

| Content size | Albums | Assets | Approx. index RAM |
|---|---|---|---|
| Small personal gallery | 50 | 2,000 | ~1 MB |
| Medium gallery | 500 | 50,000 | ~12 MB |
| Large gallery | 2,000 | 200,000 | ~50 MB |
| Very large | 5,000 | 1,000,000 | ~250 MB |

Per-object overhead: ~500 bytes/album, ~200 bytes/asset, ~300 bytes/config. The watcher maintains a separate map (~100 bytes per filesystem entry). Sessions and users are negligible.

### Re-index cost

Every re-index (triggered by the watcher or admin endpoint) rebuilds the full snapshot:

1. `fswalk.Scan` walks the content tree (reads directory listings + album.json files).
2. `index.BuildSnapshot` loads/creates sidecar state for each album and asset.
3. `Server.SetSnapshot` builds new index maps and swaps the snapshot.

There is no incremental update. Rebuild time scales with total content:
- 10,000 assets: ~1-2 seconds
- 100,000 assets: ~10-15 seconds (dominated by sidecar I/O on first scan)

During rebuild, the old snapshot continues serving requests. The swap is atomic from the API's perspective.

### Scaling limitations

This architecture is designed for **single-instance deployments**:

- **Sessions** are in-memory — not shared across instances. Multi-instance would need Redis, PostgreSQL, or JWT-based sessions.
- **Snapshot** is per-process — each instance would need its own scan, leading to redundant I/O and inconsistent views during re-index windows.
- **User store** is loaded once at startup — adding a user requires restart.
- **Watcher** state is per-process.

For the intended use case (personal or small-team photo galleries), single-instance with the full index in RAM is the right tradeoff: zero-latency ACL checks, no external dependencies for core functionality, and simple deployment.

---

## 11. Optional popularity analytics

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

## 12. GDPR-safe analytics requirements

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

## 13. API shape

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

## 14. Recommended package layout

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

## 15. Implementation order

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

## 16. Final summary

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
