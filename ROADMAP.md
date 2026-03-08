# Roadmap

Architectural recommendations for long-term maintainability. Each item has agent prompts in `docs/agent-prompts/` ready for execution.

See `docs/agent-workflow.md` for the full prompt sequence (phases 13–23).

## Priority 1 — Do soon

### Explicit dependency struct for API server
**Prompts: 53–54**

Replace the `Set*()` setter methods on `api.Server` with a single `Dependencies` struct passed to `NewServer()`. This makes the dependency graph visible at construction time rather than scattered across setter calls that can be forgotten or called in the wrong order.

**Files:** `backend/internal/api/api.go`, `backend/internal/app/app.go`

### Structured request IDs
**Prompts: 55–56**

Every HTTP request should get a UUID in the context, logged with slog, and returned in a response header (`X-Request-ID`). Without this, debugging production issues across logs becomes guesswork.

**Files:** `backend/internal/api/middleware.go`

### Unified migration framework
**Prompts: 57–58**

Only `analytics/postgres` has migrations (tern). If user persistence, session storage, or any other stateful feature is added, there will be competing migration directories. Consolidate into a single top-level migration runner at app startup covering all tables.

**Files:** `backend/internal/migrate/`, `backend/internal/app/app.go`

### Replace hand-rolled session/CSRF with battle-tested middleware
**Prompts: 59–62 (conditional)**

The HMAC session tokens and CSRF implementation work, but over time edge cases accumulate: token rotation, cookie SameSite evolution, clock skew. Consider `alexedwards/scs` — it handles sharp edges that haven't been hit yet. Prompt 59 is a read-only evaluation; 60–62 only execute if the evaluation recommends proceeding.

**Files:** `backend/internal/auth/auth.go`, `backend/internal/api/api.go`

## Priority 2 — Plan for

### Replace polling watcher with fsnotify
**Prompts: 63–64**

The polling watcher was a reasonable zero-dependency choice, but for large content directories inotify/kqueue is significantly more efficient. The `ReconcileFunc` abstraction means this is a drop-in replacement.

**Files:** `backend/internal/watch/watch.go`

### OpenAPI/Swagger spec
**Prompts: 65–67**

With 27 routes the API surface is large enough to justify a machine-readable spec: client generation, documentation, contract testing. Write the spec, then validate handlers against it.

**Files:** `docs/openapi.yaml`

### Frontend: adopt Vite
**Prompts: 68–70**

The current `resolve-theme.js` + esbuild pipeline works but is bespoke. Vite provides HMR for development, CSS modules, and a standard that contributors will recognize. The site-override pattern can be preserved as a Vite plugin.

**Files:** `frontend/`

### App lifecycle phases
**Prompt: 71**

As the system grows, splitting `app.Run()` into discrete lifecycle phases (Init → Wire → Serve → Shutdown) prevents the wiring function from becoming unmanageable. Lightweight refactor without adding DI frameworks.

**Files:** `backend/internal/app/app.go`

## Priority 3 — Worth thinking about

### Background image processing worker pool
**Prompts: 72–73**

Derivative generation is synchronous in the request path. For large galleries this blocks HTTP workers. A bounded goroutine pool would let the server return 202 Accepted and serve derivatives when ready.

**Files:** `backend/internal/derive/`, `backend/internal/api/assets.go`

### Content-addressable cache keys
**Prompt: 74**

Current cache keys use asset IDs + dimensions. If re-processing is ever needed (quality changes, format changes like WebP/AVIF), content-addressed keys (hash of source + params) avoid stale cache issues without a full purge.

**Files:** `backend/internal/cache/cache.go`

### Read-only search index
**Prompts: 75–76**

The filesystem-as-source-of-truth invariant is sound for content, but querying across albums (search, filtering, sorting by date) will eventually need an index. A read-only in-memory index that rebuilds from snapshots enables search without violating the invariant.

**Files:** `backend/internal/search/`, `backend/internal/api/search.go`
