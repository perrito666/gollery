# Developer Guide

## Project Overview

Gollery is a filesystem-first image gallery. You point it at a directory tree of images and it serves a browsable web gallery with albums, thumbnails, access control, discussion links, and optional popularity analytics.

The key design principle is that **the filesystem is the source of truth**. Albums are directories. Assets are image files. Configuration lives in `album.json` files alongside the images. The server discovers content by walking the directory tree — it never writes to content directories.

Mutable editorial state (stable IDs, discussion bindings, access overrides) lives in `.gallery/` sidecar directories. Optional popularity analytics live in PostgreSQL, completely separate from the filesystem.

The project is a monorepo with two top-level components:

- **`backend/`** — Go HTTP server (`galleryd`)
- **`frontend/`** — Vanilla JavaScript single-page application (no framework)

## Architecture

### Layered Design

```
┌─────────────────────────────────────────────────┐
│                   HTTP Client                    │
└────────────────────────┬────────────────────────┘
                         │
┌────────────────────────▼────────────────────────┐
│              API Layer (api package)             │
│  routes, middleware, auth, rate limiting, CSRF   │
└────────────────────────┬────────────────────────┘
                         │
┌────────────────────────▼────────────────────────┐
│            Domain + Services Layer               │
│  access, auth, discussion, analytics, derive     │
└────────────────────────┬────────────────────────┘
                         │
┌────────────────────────▼────────────────────────┐
│           Infrastructure Layer                   │
│  fswalk, state, index, cache, watch, meta, config│
└────────────────────────┬────────────────────────┘
                         │
┌────────────────────────▼────────────────────────┐
│         Filesystem + PostgreSQL (optional)        │
└─────────────────────────────────────────────────┘
```

### Backend Package Dependency Graph

No package cycles exist. Dependencies flow downward:

```
domain, config, cache, state, watch, analytics, logging  (leaf packages — no internal deps)
    │
    ├── access    → config, domain
    ├── auth      → domain
    ├── fswalk    → config
    ├── derive    → cache
    ├── meta      → domain
    ├── discussion → state
    ├── index     → domain, fswalk, state
    ├── analytics/postgres → analytics
    │
    └── api       → access, auth, cache, config, derive, discussion, domain, analytics, meta
        │
        └── app   → api, access, auth, cache, config, derive, fswalk, index,
                     logging, meta, state, watch, analytics/postgres
            │
            └── cmd/galleryd/main.go → app
```

### Frontend Layer Architecture

The frontend enforces strict layer separation to allow UI replacement without rewriting logic:

```
┌────────────────────────────────────────────────┐
│  site/           User customizations            │
│  (overrides any view, adds CSS, static assets)  │
├────────────────────────────────────────────────┤
│  ui-default/     Default view implementations   │
│  (views, layouts, components, CSS)              │
├────────────────────────────────────────────────┤
│  ui-contract/    View interface definitions      │
│  (ComponentRegistry, view model types, events)  │
├────────────────────────────────────────────────┤
│  core/           Application logic              │
│  (router, store, controllers, API client, auth) │
└────────────────────────────────────────────────┘
```

Views never import from `core/` directly. Data arrives via `render(container, viewModel, ctx)` where `ctx` provides access to services.

## Backend Subsystem Walkthrough

### config — Configuration Loading

Loads `gollery.json` (server config) and per-album `album.json` files. Server config supports environment variable overrides for secrets:

- `GOLLERY_SESSION_SECRET` overrides `auth.session_secret`
- `GOLLERY_POSTGRES_DSN` overrides `analytics.postgres_dsn_env`

Album configs inherit from parent directories. Child values override parent scalars, merge objects, and replace arrays.

```go
merged := config.MergeAlbumConfigs(parentConfig, childConfig)
```

Validation collects all errors and returns them joined:

```go
if err := cfg.Validate(); err != nil {
    // err may contain multiple joined errors
}
```

### domain — Core Types

Pure data types with no behavior and no dependencies on other packages:

- **`Album`** — ID, path, title, description, parent/children, assets
- **`Asset`** — ID, filename, title, description, album path, optional access override, optional metadata
- **`Snapshot`** — Point-in-time view of the entire gallery (map of path → album)
- **`Principal`** — Authenticated user (username, groups, admin flag)
- **`DiscussionBinding`** — Link to an external discussion thread
- **`AccessOverride`** — Per-asset ACL override
- **`ImageMetadata`** — EXIF data (camera, dimensions, GPS, date taken)

IDs are stable across restarts. They use the format `alb_<hex>` and `ast_<hex>`, generated via `crypto/rand`.

### fswalk — Filesystem Discovery

Walks a content directory tree and discovers albums and assets:

```go
result := fswalk.Scan("/data/content")
// result.Albums: map[relativePath]*ScannedAlbum
// result.Errors: []error (non-fatal)
```

Key behaviors:
- Skips hidden directories (`.gallery`, `.git`)
- Loads `album.json` at each level and merges with parent config
- Collects image files by extension (`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`)
- Non-fatal errors (unreadable files, bad JSON) are collected, not returned as failures

### state — Sidecar State

Manages `.gallery/` directories that sit alongside content:

```
content/
  vacation/
    .gallery/
      album.state.json        ← album ID + discussion bindings
      assets/
        photo1.jpg.json       ← asset ID + title + description + discussion bindings + access override
    photo1.jpg
    album.json                ← declarative config (admin metadata PATCH can update title/description)
```

State files are written atomically (temp file + `os.Rename`) to prevent corruption on crash.

ID assignment is idempotent:

```go
albumState, isNew, err := state.EnsureAlbumID(galleryDir)
// If the album already has an ID, returns it unchanged.
// If not, generates a new alb_<hex> ID and saves.
```

### index — Snapshot Builder

Combines filesystem discovery with sidecar state to produce a `domain.Snapshot`:

```go
snapshot, err := index.BuildSnapshot(contentRoot, scanResult)
```

This is the central assembly point. It:
1. Takes `fswalk.ScanResult` as input
2. Calls `state.EnsureAlbumID/EnsureAssetID` to attach stable IDs
3. Reads access overrides from sidecar state
4. Produces a `domain.Snapshot` that the API server uses

### access — ACL Engine

Stateless functions that evaluate access control:

```go
decision := access.CheckView(albumACL, principal)
// decision is access.Allow or access.Deny
```

Access modes:
- **`public`** — anyone can view (including anonymous)
- **`authenticated`** — any logged-in user
- **`restricted`** — only users in the allowed users/groups list

Asset-level overrides replace album ACLs entirely:

```go
effectiveACL := access.EffectiveAssetACL(albumACL, asset.Access)
```

Nil principal means anonymous. Nil ACL means public. Admins always get access.

### auth — Authentication

Interface-based design with one concrete implementation:

- **`FileUserStore`** — loads users from a JSON file, authenticates with bcrypt
- **`HMACSessionStore`** — cookie-based sessions signed with HMAC-SHA256

```go
// Authenticate
principal, err := authenticator.Authenticate(ctx, "admin", "password")

// Create session
token, err := sessions.Create(ctx, principal)

// Retrieve from context (set by middleware)
principal := auth.PrincipalFromContext(r.Context())
// Returns nil for anonymous requests
```

### derive — Image Derivatives

Generates thumbnails and previews on demand, caching results:

```go
err := derive.GenerateThumbnail(sourcePath, cacheLayout, assetID, maxSize)
```

Uses `draw.CatmullRom` from `golang.org/x/image/draw` for high-quality scaling. Output is JPEG at quality 85. If a cached file already exists, generation is skipped.

### cache — Cache Layout

Manages the directory structure for cached derivatives:

```
/data/cache/
  thumbs/ast_a1b2c3_200.jpg
  previews/ast_a1b2c3_1200.jpg
```

`PurgeOrphans` removes cached files for assets that no longer exist in the snapshot.

### watch — Filesystem Watcher

Polls the content directory for changes and triggers reindexing:

```go
watcher := watch.New(watch.Config{
    ContentRoot:   "/data/content",
    PollInterval:  5 * time.Second,
    Debounce:      2 * time.Second,
    ReconcileFunc: func(ctx context.Context) error { return doReindex() },
})
go watcher.Run(ctx)
```

Uses polling (not inotify/fsnotify) for portability. Debouncing prevents thrashing when many files change at once. The first scan establishes a baseline without triggering reconciliation.

### meta — EXIF Extraction

Extracts camera metadata from JPEG files using `github.com/rwcarlsen/goexif`:

```go
metadata, err := meta.Extract("/path/to/photo.jpg")
// metadata.CameraMake, metadata.CameraModel, metadata.DateTaken, etc.
```

Missing EXIF data is not an error — it returns a zero-value struct. Only file-open failures return errors.

### discussion — Discussion Providers

Pluggable system for linking gallery items to external discussion threads:

```go
svc := discussion.NewService(mastodonProvider, blueskyProvider)

// Create a thread on Mastodon and store the binding
err := svc.CreateBinding(ctx, "mastodon", albumPath, "album", "", "My Album", "Check it out", "admin")

// List bindings for an album
bindings, err := svc.ListBindings(albumPath, "album", "")
```

Bindings are persisted to sidecar state files. See [Adding Discussion Providers](#adding-a-new-discussion-provider) below.

### analytics — Popularity Tracking

Optional PostgreSQL-backed analytics with privacy-preserving visitor hashing. See [How Analytics Works](#how-analytics-works) below.

### api — HTTP API Server

29 routes organized into groups:

| Group | Routes | Auth Required |
|-------|--------|---------------|
| Public content | `/albums/root`, `/albums/{id}`, `/assets/{id}`, thumbnails, previews, originals | No (ACL checked) |
| Auth | `/auth/login`, `/auth/me`, `/auth/logout`, `/auth/csrf-token` | Varies |
| Admin | `/admin/reindex`, `/admin/status`, `/admin/diagnostics` | Admin only |
| Metadata | `PATCH /assets/{id}/metadata`, `PATCH /albums/{id}/metadata` | Admin only |
| Analytics | `/albums/{id}/stats`, `/assets/{id}/stats`, popular assets, overview | Admin only |
| Access | `/albums/{id}/access`, `/assets/{id}/access`, asset access PATCH | ACL checked |
| Discussions | album/asset discussion list and create | Auth required for create |
| Health | `/healthz` | No |

Middleware chain (outermost first): Rate Limiting → Security Headers → CORS → Auth → CSRF → Analytics Recording → Route Handler.

The server holds an immutable `Snapshot` protected by `sync.RWMutex`. When the watcher detects changes, it rebuilds the snapshot atomically via `SetSnapshot()`.

### app — Application Wiring

`app.Run()` orchestrates the entire startup sequence in numbered phases:

1. Load and validate config
2. Configure structured logging
3. Scan filesystem and build initial snapshot
4. Create API server with snapshot
5. Set up auth (file user store + HMAC sessions)
6. Set up analytics (connect PostgreSQL, run migrations) — if enabled
7. Start filesystem watcher
8. Start HTTP server
9. Wait for shutdown signal
10. Graceful shutdown (10-second deadline)

### cmd/galleryd — Entry Point

Minimal main function: parse flags, set up signal context, call `app.Run()`:

```go
galleryd --config /etc/gollery/gollery.json
galleryd --version
```

### cmd/gollery-users — User Management CLI

Standalone tool for managing `users.json` and album configs. Commands:

- **User management**: `list`, `add`, `remove`, `passwd`, `set-admin`, `set-groups`, `add-groups`, `remove-groups`
- **Album config**: `init-album` — creates `album.json` with flags for title, access mode, allowed users/groups
- **Validated editing**: `edit users` / `edit album -file path` — visudo-style editing using `$EDITOR` with validation before saving

Handles bcrypt hashing automatically. See [docs/user-management.md](docs/user-management.md).

## Frontend Subsystem Walkthrough

### Data Flow

```
URL hash change
  → Router matches pattern (e.g., #/albums/:id)
  → Controller fetches data from API, builds view model
  → store.set({ currentView: 'album', viewModel: {...} })
  → Store notifies subscriber in core/index.js
  → registry.get('album') → view renderer
  → renderer.render(container, viewModel, ctx)
  → DOM updated
```

### core/api/client.js — API Client

Centralized fetch wrapper. Views and controllers never call `fetch` directly:

```javascript
const api = new ApiClient();
const album = await api.getAlbum('alb_abc123');
const url = api.thumbnailURL('ast_def456', 200);
```

Throws typed `ApiError` with HTTP status and message on failure. Handles CSRF tokens automatically — fetches a token after login/session restore and includes it as `X-CSRF-Token` on all POST and PATCH requests. Supports mutation methods via `_mutate(method, path, body)` used by both `_post` and `_patch`. Additional methods: `patchAssetMetadata(id, {title, description})`, `patchAlbumMetadata(id, {title, description})`, `getAssetDiscussions(id)`, `getAlbumDiscussions(id)`.

### core/state/store.js — Reactive Store

Single source of truth for UI state:

```javascript
const store = new Store();
store.set({ currentView: 'album', viewModel: albumData });

const unsubscribe = store.subscribe((state) => {
    // Called on every state change
});
```

State shape: `{ currentView, viewModel, principal, loading, error }`.

### core/router/router.js — Hash Router

Client-side routing via URL hash fragments:

```javascript
const router = new Router();
router.on('/', () => albumController.showRoot());
router.on('/albums/:id', ({ id }) => albumController.showAlbum(id));
router.on('/assets/:id', ({ id }) => assetController.showAsset(id));
router.on('/login', () => store.set({ currentView: 'login', ... }));
router.start();
```

Pattern parameters (`:id`) are extracted and passed to handlers. Album IDs use stable `alb_<hex>` identifiers, not filesystem paths.

### core/controllers — Business Logic

Controllers mediate between API and views. They follow a consistent pattern:

```javascript
async showAlbum(id) {
    this.store.set({ loading: true, error: null });
    try {
        const album = await this.api.getAlbum(id);
        const viewModel = this._toViewModel(album);
        this.store.set({ currentView: 'album', viewModel, loading: false });
    } catch (err) {
        handleApiError(this.store, err);
    }
}
```

Error handling maps HTTP status codes to views: 401 → login, 403 → forbidden, 404 → not-found.

### core/auth/session.js — Session Manager

Wraps auth API calls and syncs the principal to the store:

```javascript
await session.restore();   // Check if already logged in (silent failure if not)
await session.login(username, password);
await session.logout();
```

### core/services — Permissions and Feature Flags

**PermissionService** reads the principal from the store for optimistic UI decisions. The server enforces actual access — this is cosmetic only.

**FeatureFlags** controls optional UI sections (e.g., popularity badges):

```javascript
if (features.isEnabled('popularity')) {
    // Render popularity badge
}
```

### ui-contract — View Interface

Every view renderer must implement:

```javascript
export function render(container, viewModel, ctx) {
    // Build HTML, attach to container
}

export function destroy() {
    // Clean up event listeners
}
```

The `ctx` object provides: `{ store, router, session, permissions, features, popularity }`.

Views are registered by name in a `ComponentRegistry`. The core layer calls `registry.get(viewName)` to load the appropriate renderer.

### ui-default — Default Views

Seven views: `home`, `album`, `asset`, `login`, `not-found`, `forbidden`, `error`.

All views:
- Build HTML strings (no virtual DOM, no framework)
- Escape all dynamic content with `esc()` for XSS safety
- Use hash links (`#/albums/alb_abc123`) for navigation — album children include `id`, `path`, and `title`
- Receive data exclusively through `viewModel` and `ctx`
- `home` and `album` views include a shared nav bar (`ui-default/util/nav.js`) with login/logout controls
- `home`, `album`, and `asset` views include admin-only edit forms for title/description (toggle show/hide)
- `asset` view includes a Mastodon share button (prompts for instance, opens share URL) and discussion links

### Build System

The frontend build has two steps:

1. **`resolve-theme.js`** — Scans `src/site/views/` for overrides, falls back to `src/ui-default/views/`, generates `src/_resolved/registry.js` with imports and registration code
2. **esbuild** — Bundles everything into a single JS file (~11kb minified)

```bash
cd frontend
node scripts/resolve-theme.js   # Generate resolved registry
npx esbuild src/main.js --bundle --outfile=dist/bundle.js --format=esm --minify
```

## How to Run Locally

### Prerequisites

- Go 1.25+
- Node.js 22+
- Docker and Docker Compose (for PostgreSQL and nginx)

### Option 1: Docker Compose (recommended)

```bash
# Build frontend first
make frontend-build

# Create a sample content directory with some images and album.json files
mkdir -p sample-content/vacation
cp /path/to/some/photos/* sample-content/vacation/
# Each directory needs an album.json — see docs/running-locally.md

# Create users.json with the gollery-users tool
cd backend && go build -o gollery-users ./cmd/gollery-users
./gollery-users -file ../users.json add -username admin -password admin -admin -groups admins

# Start everything
docker compose up --build

# Open http://localhost:8090 (nginx serves frontend + proxies API)
# API also available at http://localhost:8080
```

The `docker-compose.yml` starts three services: `galleryd` (API), `nginx` (frontend + proxy), and `postgres` (analytics). It mounts:
- `./sample-content` → `/data/content` (your images)
- `./gollery.json` → `/etc/gollery/gollery.json` (server config)
- `./users.json` → `/etc/gollery/users.json` (user database)
- `./frontend/dist` → nginx html root (frontend files)
- `./nginx.conf` → nginx config

See [docs/running-locally.md](docs/running-locally.md) for a detailed guide.

### Option 2: Run directly

```bash
# Build backend
cd backend
go build -o galleryd ./cmd/galleryd

# Build frontend
cd ../frontend
npm ci
make build

# Start PostgreSQL for analytics (optional)
cd ../backend
./scripts/start-test-db.sh

# Run the server
./galleryd --config ../gollery.json
```

### Running Tests

```bash
# Backend unit tests (219 tests)
cd backend && go test ./...

# Backend with verbose output
cd backend && go test -v ./...

# Frontend tests (38 tests)
cd frontend && npm test

# Integration tests (requires PostgreSQL)
cd backend
GOLLERY_TEST_POSTGRES_DSN="postgres://gollery:gollery@localhost:5432/gollery_test?sslmode=disable" \
  go test -tags integration ./internal/analytics/postgres/

# Full verification
make verify
```

### Configuration

**`gollery.json`** — Server configuration:

```json
{
  "content_root": "/data/content",
  "cache_dir": "/data/cache",
  "listen_addr": ":8080",
  "auth": {
    "provider": "static",
    "session_secret": "override-via-env"
  },
  "analytics": {
    "enabled": true,
    "backend": "postgres",
    "postgres_dsn_env": "override-via-env",
    "hash_ip": true,
    "dedup_window_seconds": 300,
    "retain_events_days": 90
  }
}
```

Sensitive values should be set via environment variables:
- `GOLLERY_SESSION_SECRET` — HMAC signing key for sessions
- `GOLLERY_POSTGRES_DSN` — PostgreSQL connection string

**`users.json`** — User database:

```json
[
  {
    "username": "admin",
    "password": "$2a$10$...",
    "groups": ["admins"],
    "is_admin": true
  }
]
```

Passwords are bcrypt hashes. Use the `gollery-users` CLI tool to manage users:

```bash
cd backend && go build -o gollery-users ./cmd/gollery-users

# Add a user (auto-hashes password)
./gollery-users add -username alice -password secret -groups editors

# Change password
./gollery-users passwd -username alice -password newpass

# List users
./gollery-users list
```

See [docs/user-management.md](docs/user-management.md) for the full reference.

**`album.json`** — Per-album configuration (placed alongside images):

```json
{
  "title": "Summer Vacation",
  "description": "Photos from our trip",
  "access": {
    "mode": "authenticated"
  },
  "discussion": {
    "enabled": true,
    "providers": ["mastodon"]
  }
}
```

Album configs inherit from parent directories. A child album only needs to specify values it wants to override.

## Adding a New Discussion Provider

Discussion providers implement a simple interface. Here's how to add one:

### 1. Create the provider package

```
backend/internal/discussion/providers/yourplatform/
  yourplatform.go
  yourplatform_test.go
```

### 2. Implement the Provider interface

```go
package yourplatform

import (
    "context"
    "github.com/perrito666/gollery/backend/internal/discussion"
)

// Poster abstracts HTTP calls for testability.
type Poster interface {
    CreatePost(ctx context.Context, apiURL, token, text string) (id, url string, err error)
}

type Config struct {
    APIURL string `json:"api_url"`
    Token  string `json:"token"`
}

type Provider struct {
    cfg    Config
    poster Poster
}

func New(cfg Config) *Provider {
    return &Provider{cfg: cfg, poster: &httpPoster{}}
}

// NewWithPoster allows injecting a mock for tests.
func NewWithPoster(cfg Config, p Poster) *Provider {
    return &Provider{cfg: cfg, poster: p}
}

func (p *Provider) Name() string {
    return "yourplatform"
}

func (p *Provider) CreateThread(ctx context.Context, title, body string) (discussion.Thread, error) {
    text := title + "\n\n" + body
    id, url, err := p.poster.CreatePost(ctx, p.cfg.APIURL, p.cfg.Token, text)
    if err != nil {
        return discussion.Thread{}, fmt.Errorf("yourplatform: %w", err)
    }
    return discussion.Thread{
        Provider: "yourplatform",
        RemoteID: id,
        URL:      url,
    }, nil
}
```

### 3. Write tests with a mock Poster

```go
type mockPoster struct {
    id, url string
    err     error
}

func (m *mockPoster) CreatePost(ctx context.Context, apiURL, token, text string) (string, string, error) {
    return m.id, m.url, m.err
}

func TestCreateThread(t *testing.T) {
    p := NewWithPoster(Config{APIURL: "https://example.com"}, &mockPoster{
        id:  "post-123",
        url: "https://example.com/post/123",
    })
    thread, err := p.CreateThread(context.Background(), "Title", "Body")
    if err != nil {
        t.Fatal(err)
    }
    if thread.URL != "https://example.com/post/123" {
        t.Errorf("got URL %s", thread.URL)
    }
}
```

### 4. Register in app wiring

In `backend/internal/app/app.go`, instantiate your provider and pass it to the discussion service:

```go
yourProvider := yourplatform.New(yourplatform.Config{...})
svc := discussion.NewService(mastodonProvider, blueskyProvider, yourProvider)
srv.SetDiscussions(svc)
```

### Key patterns to follow

- **Always use a `Poster` interface** — never make real HTTP calls in unit tests
- **Use `NewWithPoster()` constructor** for dependency injection in tests
- **Prefix errors** with the provider name: `fmt.Errorf("yourplatform: %w", err)`
- **Return `discussion.Thread`** with Provider, RemoteID, and URL filled in
- **Store extra data** in `Thread.ProviderMeta` (a `map[string]string`) if needed

## How Analytics Works

Analytics is an optional subsystem that tracks page views and asset popularity. It is entirely separate from the filesystem — all data lives in PostgreSQL.

### Architecture

```
HTTP Request
  → Analytics Recording Middleware (records view events)
  → Route Handler (serves content)

Background:
  → Daily Aggregation Job (rolls up raw events into daily summaries)
  → Retention Purge Job (deletes events older than configured threshold)
```

### Event Recording

The analytics middleware in the API layer records events for album views, asset views, original file downloads, and discussion link clicks:

```go
type Event struct {
    Type        EventType  // "album_view", "asset_view", "original_hit", "discussion_click"
    ObjectID    string     // album or asset ID
    VisitorHash string     // privacy-safe hash of visitor IP
    CreatedAt   time.Time
}
```

Visitor IPs are never stored raw. They are hashed with a configurable salt:

```go
hash := analytics.HashVisitorID(clientIP, salt)
// SHA256(salt + IP), truncated to 16 hex characters
```

A dedup window (default 300 seconds) prevents the same visitor from inflating counts by refreshing.

### Storage Schema

Two tables in PostgreSQL:

**`analytics_events`** — Raw event log:
```sql
CREATE TABLE analytics_events (
    id          BIGSERIAL PRIMARY KEY,
    event_type  TEXT NOT NULL,
    object_id   TEXT NOT NULL,
    visitor_hash TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**`popularity_daily`** — Aggregated daily summaries:
```sql
CREATE TABLE popularity_daily (
    object_id   TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    day         DATE NOT NULL,
    count       INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (object_id, event_type, day)
);
```

### Aggregation

A daily job groups raw events by (object_id, event_type, date) and upserts into `popularity_daily`:

```go
store.AggregateDailyPopularity(ctx, time.Now().UTC().Truncate(24*time.Hour))
```

### Retention

Raw events are purged after a configurable retention period (default 90 days):

```go
store.PurgeOldEvents(ctx, time.Now().Add(-90 * 24 * time.Hour))
```

Aggregated daily summaries are kept indefinitely for historical trends.

### Querying

The API exposes popularity data through admin-only endpoints:

```
GET /api/v1/albums/{id}/stats         → PopularitySummary for an album
GET /api/v1/assets/{id}/stats         → PopularitySummary for an asset
GET /api/v1/albums/{id}/popular-assets → Top assets by views in an album
GET /api/v1/admin/analytics/overview   → System-wide analytics summary
```

`PopularitySummary` contains: total views, 7-day views, 30-day views, original download count, and discussion click count.

### Disabling Analytics

Set `analytics.enabled: false` in `gollery.json`. The gallery works identically without analytics — no PostgreSQL connection is needed, and popularity-related API endpoints return empty responses.

## How to Extend the UI

The frontend is designed for UI replacement without rewriting application logic.

### Overriding a Single View

To replace the album view with your own:

1. Create `frontend/src/site/views/album.js`
2. Export `render` and `destroy`:

```javascript
import { esc } from '../../core/utils/esc.js';

export function render(container, viewModel, ctx) {
    container.innerHTML = `
        <div class="my-custom-album">
            <h1>${esc(viewModel.title)}</h1>
            <div class="my-grid">
                ${viewModel.assets.map(a => `
                    <a href="#/assets/${esc(a.id)}">
                        <img src="${esc(a.thumbnailUrl)}" alt="${esc(a.filename)}">
                    </a>
                `).join('')}
            </div>
        </div>
    `;
}

export function destroy() {
    // Clean up event listeners if you attached any
}
```

3. Run `node scripts/resolve-theme.js` to regenerate the registry
4. Build: `make build` or `npx esbuild src/main.js --bundle --outfile=dist/bundle.js --format=esm`

The resolve-theme script automatically detects your override and uses it instead of the default.

### Available View Names

Override any of these by creating `frontend/src/site/views/<name>.js`:

| View | viewModel | When shown |
|------|-----------|------------|
| `home` | Root album data | Landing page |
| `album` | Album with children and assets | Browsing an album |
| `asset` | Single asset with metadata | Viewing an image |
| `login` | None | Authentication required |
| `not-found` | `{ message }` | 404 |
| `forbidden` | `{ message }` | 403 |
| `error` | `{ message }` | Server errors |

### The Context Object

Every view receives `ctx` as the third argument to `render()`:

```javascript
ctx.store        // Reactive state store
ctx.router       // Hash router (navigate, currentPath)
ctx.session      // Auth (login, logout, restore)
ctx.permissions  // PermissionService (isAuthenticated, isAdmin, canView)
ctx.features     // FeatureFlags (isEnabled)
ctx.popularity   // PopularityClient (getPopularity, getPopularInAlbum) — may be null
```

### Adding Custom CSS

Place CSS files in `frontend/src/site/styles/`. They are appended after the default styles during the build, so your rules take precedence.

### Adding Static Assets

Place files in `frontend/src/site/assets/`. They are copied to the build output directory.

### XSS Safety

All dynamic content in HTML strings **must** be escaped:

```javascript
import { esc } from '../../core/utils/esc.js';

// Correct
container.innerHTML = `<h1>${esc(userProvidedTitle)}</h1>`;

// WRONG — XSS vulnerability
container.innerHTML = `<h1>${userProvidedTitle}</h1>`;
```

### Adding a New Component

Components are reusable render functions for UI fragments:

```javascript
// frontend/src/site/components/my-badge.js
import { esc } from '../../core/utils/esc.js';

export function renderMyBadge(container, data) {
    container.innerHTML = `<span class="badge">${esc(data.label)}</span>`;
}
```

Components are called directly from views — they are not registered in the ComponentRegistry (that's only for full-page views).
