# Architecture Diagrams

## 1. Backend Subsystem Diagram

```mermaid
graph TB
    subgraph "Entry Point"
        MAIN["cmd/galleryd<br/>main.go"]
    end

    subgraph "Orchestration"
        APP["app<br/>Run(), wiring,<br/>graceful shutdown"]
    end

    subgraph "HTTP Layer"
        API["api<br/>Server, 27 routes,<br/>Handler()"]
        MW["api/middleware<br/>security headers,<br/>CORS, analytics"]
        RL["api/ratelimit<br/>per-IP token bucket"]
    end

    subgraph "Domain Services"
        ACCESS["access<br/>CheckView,<br/>EffectiveAssetACL"]
        AUTH["auth<br/>FileUserStore,<br/>HMACSessionStore"]
        DISC["discussion<br/>Service,<br/>Provider interface"]
        DERIVE["derive<br/>GenerateThumbnail,<br/>GeneratePreview"]
        META["meta<br/>EXIF extraction"]
    end

    subgraph "Discussion Providers"
        MASTO["mastodon<br/>Poster interface,<br/>API v1 statuses"]
        BSKY["bluesky<br/>Poster interface,<br/>AT Protocol"]
    end

    subgraph "Infrastructure"
        CONFIG["config<br/>ServerConfig,<br/>AlbumConfig, merge"]
        DOMAIN["domain<br/>Album, Asset,<br/>Snapshot, Principal"]
        FSWALK["fswalk<br/>Scan(), discover<br/>albums + assets"]
        STATE["state<br/>sidecar .gallery/,<br/>atomic writes"]
        INDEX["index<br/>BuildSnapshot()"]
        CACHE["cache<br/>Layout, PurgeOrphans"]
        WATCH["watch<br/>polling watcher,<br/>debounce"]
        LOG["logging<br/>slog setup"]
    end

    subgraph "Analytics"
        ANALYTICS["analytics<br/>Store interface,<br/>Event, HashVisitorID"]
        PGSTORE["analytics/postgres<br/>pgx pool, tern<br/>migrations"]
    end

    subgraph "Storage"
        FS[("Filesystem<br/>content + .gallery/")]
        PG[("PostgreSQL<br/>analytics tables")]
    end

    MAIN --> APP
    APP --> API
    APP --> AUTH
    APP --> CONFIG
    APP --> FSWALK
    APP --> INDEX
    APP --> CACHE
    APP --> WATCH
    APP --> LOG
    APP --> PGSTORE

    API --> MW
    API --> RL
    API --> ACCESS
    API --> AUTH
    API --> DISC
    API --> DERIVE
    API --> META
    API --> ANALYTICS
    API --> DOMAIN
    API --> CONFIG
    API --> CACHE

    ACCESS --> CONFIG
    ACCESS --> DOMAIN
    AUTH --> DOMAIN
    FSWALK --> CONFIG
    DERIVE --> CACHE
    META --> DOMAIN
    DISC --> STATE
    DISC --> MASTO
    DISC --> BSKY
    INDEX --> DOMAIN
    INDEX --> FSWALK
    INDEX --> STATE
    PGSTORE --> ANALYTICS

    FSWALK --> FS
    STATE --> FS
    CACHE --> FS
    WATCH --> FS
    PGSTORE --> PG

    style FS fill:#e8f5e9,stroke:#2e7d32
    style PG fill:#e3f2fd,stroke:#1565c0
    style API fill:#fff3e0,stroke:#e65100
    style APP fill:#fce4ec,stroke:#c62828
    style DOMAIN fill:#f3e5f5,stroke:#6a1b9a
```

## 2. Frontend Architecture Diagram

```mermaid
graph TB
    subgraph "Browser"
        URL["URL Hash<br/>#/albums/alb_abc"]
    end

    subgraph "Entry"
        MAINJS["main.js<br/>bootstrap"]
    end

    subgraph "core/ — Application Logic"
        ROUTER["router<br/>hash routing,<br/>pattern matching"]
        STORE["store<br/>reactive state,<br/>subscribe/notify"]
        SESSION["session<br/>login, logout,<br/>restore"]
        APICLIENT["api/client<br/>fetch wrapper,<br/>ApiError"]
        POPULARITY["api/popularity<br/>optional analytics<br/>client"]
        ALBUM_CTRL["controllers/album<br/>showRoot,<br/>showAlbum"]
        ASSET_CTRL["controllers/asset<br/>showAsset,<br/>prev/next"]
        PERMS["services/permissions<br/>isAuthenticated,<br/>canView"]
        FLAGS["services/features<br/>isEnabled,<br/>enable/disable"]
        ESC["utils/esc<br/>XSS escaping"]
        ERRH["utils/error-handler<br/>handleApiError"]
    end

    subgraph "ui-contract/ — Interfaces"
        REGISTRY["ComponentRegistry<br/>register, get,<br/>override"]
        VIEWMODELS["view-models<br/>AlbumViewModel,<br/>AssetViewModel"]
        EVENTS["events<br/>ALBUM_SELECTED,<br/>LOGIN_REQUESTED"]
    end

    subgraph "ui-default/ — Default Views"
        V_HOME["views/home"]
        V_ALBUM["views/album"]
        V_ASSET["views/asset"]
        V_LOGIN["views/login"]
        V_NOTFOUND["views/not-found"]
        V_FORBIDDEN["views/forbidden"]
        V_ERROR["views/error"]
        C_DISC["components/<br/>discussion-links"]
        C_POP["components/<br/>popularity"]
        LAYOUT["layouts/main"]
        CSS["styles/main.css"]
    end

    subgraph "site/ — User Overrides"
        SCONFIG["site.config.json"]
        SVIEWS["views/ (overrides)"]
        SSTYLES["styles/ (additions)"]
    end

    subgraph "Build"
        RESOLVE["resolve-theme.js<br/>generates registry"]
        ESBUILD["esbuild<br/>bundle + minify"]
    end

    URL --> ROUTER
    MAINJS --> REGISTRY
    MAINJS --> RESOLVE

    ROUTER --> ALBUM_CTRL
    ROUTER --> ASSET_CTRL

    ALBUM_CTRL --> APICLIENT
    ALBUM_CTRL --> STORE
    ALBUM_CTRL --> ERRH
    ASSET_CTRL --> APICLIENT
    ASSET_CTRL --> STORE
    ASSET_CTRL --> ERRH

    SESSION --> APICLIENT
    SESSION --> STORE

    STORE -->|"state change"| REGISTRY
    REGISTRY -->|"get(viewName)"| V_HOME
    REGISTRY -->|"get(viewName)"| V_ALBUM
    REGISTRY -->|"get(viewName)"| V_ASSET
    REGISTRY -->|"get(viewName)"| V_LOGIN

    V_ALBUM --> ESC
    V_ASSET --> ESC
    V_ASSET --> C_DISC
    V_ALBUM --> C_POP

    RESOLVE --> SVIEWS
    RESOLVE --> SCONFIG
    RESOLVE --> V_HOME
    RESOLVE --> ESBUILD

    style STORE fill:#fff3e0,stroke:#e65100
    style REGISTRY fill:#e3f2fd,stroke:#1565c0
    style APICLIENT fill:#fce4ec,stroke:#c62828
    style ROUTER fill:#e8f5e9,stroke:#2e7d32
```

## 3. Data Flow: Serving an Asset

```mermaid
sequenceDiagram
    participant B as Browser
    participant R as Router (frontend)
    participant C as AssetController
    participant S as Store
    participant V as Asset View
    participant API as API Server
    participant MW as Middleware Chain
    participant Auth as Auth Middleware
    participant CSRF as CSRF Middleware
    participant ACL as Access Checker
    participant Snap as Snapshot (RLock)
    participant Der as Derive
    participant Cache as Cache
    participant FS as Filesystem

    Note over B,FS: Phase 1: Load asset metadata
    B->>R: Navigate to #/assets/ast_abc123
    R->>C: showAsset("ast_abc123")
    C->>S: set({ loading: true })
    C->>API: GET /api/v1/assets/ast_abc123
    API->>MW: Security headers, CORS
    MW->>Auth: Extract session cookie
    Auth->>Auth: PrincipalFromContext (may be nil)
    Auth->>API: handleAssetByID()
    API->>Snap: RLock, lookup asset by ID
    Snap-->>API: asset + album
    API->>ACL: CheckView(albumACL, principal)
    ACL-->>API: Allow
    API->>API: findAdjacentAssets (prev/next)
    API-->>B: 200 JSON { id, filename, prev, next, metadata }
    C->>C: Build viewModel with preview/original URLs
    C->>S: set({ currentView: 'asset', viewModel })
    S->>V: Notify subscriber
    V->>V: render(container, viewModel, ctx)
    V-->>B: DOM updated with <img src="preview URL">

    Note over B,FS: Phase 2: Load preview image
    B->>API: GET /api/v1/assets/ast_abc123/preview?size=1200
    API->>MW: Middleware chain
    MW->>Auth: Auth check
    Auth->>API: handleAssetPreview()
    API->>Snap: RLock, lookup asset
    API->>ACL: CheckView
    ACL-->>API: Allow
    API->>Cache: Exists(previewPath)?
    alt Cached
        Cache-->>API: true
        API->>FS: Open cached file
    else Not cached
        Cache-->>API: false
        API->>Der: GeneratePreview(source, cache, id, 1200)
        Der->>FS: Read original image
        Der->>Der: CatmullRom scale, JPEG q85
        Der->>Cache: Write to cache dir
        Der-->>API: done
        API->>FS: Open newly cached file
    end
    FS-->>API: File reader
    API-->>B: 200 image/jpeg (streamed)
```

## 4. Popularity Analytics Pipeline

```mermaid
graph TB
    subgraph "Request Path"
        REQ["HTTP Request<br/>GET /api/v1/albums/alb_123"]
        AMWPRE["Analytics Middleware<br/>(before handler)"]
        HANDLER["Route Handler<br/>(serves response)"]
        AMWPOST["Analytics Middleware<br/>(after handler — records event)"]
    end

    subgraph "Event Recording"
        IP["Client IP<br/>r.RemoteAddr"]
        HASH["HashVisitorID<br/>SHA256(salt + IP)<br/>→ 16 hex chars"]
        DEDUP["Dedup Check<br/>same visitor + object<br/>within 300s window?"]
        EVENT["Event{<br/>  type: album_view<br/>  object_id: alb_123<br/>  visitor_hash: a1b2c3...<br/>  created_at: now()<br/>}"]
    end

    subgraph "PostgreSQL"
        RAW["analytics_events<br/>───────────────<br/>id BIGSERIAL<br/>event_type TEXT<br/>object_id TEXT<br/>visitor_hash TEXT<br/>created_at TIMESTAMPTZ"]
        DAILY["popularity_daily<br/>───────────────<br/>object_id TEXT<br/>event_type TEXT<br/>day DATE<br/>count INTEGER<br/>PK(object_id, event_type, day)"]
    end

    subgraph "Background Jobs"
        AGG["AggregateDailyPopularity<br/>GROUP BY object_id, event_type, date<br/>UPSERT into popularity_daily"]
        PURGE["PurgeOldEvents<br/>DELETE WHERE created_at < (now - 90 days)"]
    end

    subgraph "Query Path"
        Q_STATS["GET /api/v1/albums/{id}/stats"]
        Q_POP["GET /api/v1/albums/{id}/popular-assets"]
        Q_OVER["GET /api/v1/admin/analytics/overview"]
        RESULT["PopularitySummary{<br/>  total_views<br/>  views_7d<br/>  views_30d<br/>  original_hits<br/>  discussion_clicks<br/>}"]
    end

    REQ --> AMWPRE
    AMWPRE --> HANDLER
    HANDLER --> AMWPOST

    AMWPOST --> IP
    IP --> HASH
    HASH --> DEDUP

    DEDUP -->|"new visitor"| EVENT
    DEDUP -->|"duplicate"| SKIP["Skip (no insert)"]
    EVENT --> RAW

    RAW --> AGG
    AGG --> DAILY
    RAW --> PURGE

    Q_STATS --> DAILY
    Q_POP --> DAILY
    Q_OVER --> DAILY
    DAILY --> RESULT

    style RAW fill:#e3f2fd,stroke:#1565c0
    style DAILY fill:#e3f2fd,stroke:#1565c0
    style HASH fill:#fff3e0,stroke:#e65100
    style DEDUP fill:#fce4ec,stroke:#c62828
    style RESULT fill:#e8f5e9,stroke:#2e7d32
```

### Event Types

| Event Type | Recorded When | Object ID |
|------------|---------------|-----------|
| `album_view` | Album page loaded | Album ID (`alb_*`) |
| `asset_view` | Asset detail page loaded | Asset ID (`ast_*`) |
| `original_hit` | Original file downloaded | Asset ID (`ast_*`) |
| `discussion_click` | Discussion link clicked | Asset or Album ID |

### Privacy Model

```mermaid
graph LR
    A["Client IP<br/>192.168.1.42"] --> B["Salt<br/>(from config)"]
    B --> C["SHA256(salt + IP)"]
    C --> D["Truncate to<br/>16 hex chars"]
    D --> E["Stored hash<br/>a1b2c3d4e5f6g7h8"]

    style A fill:#fce4ec,stroke:#c62828
    style E fill:#e8f5e9,stroke:#2e7d32
```

- Raw IPs are **never stored**
- Hashing is one-way — IPs cannot be recovered
- Salt rotation invalidates old hashes (new "visitors")
- `hash_ip: true` in config enables this (default)

### Retention Lifecycle

```mermaid
graph LR
    A["Raw event<br/>recorded"] -->|"daily"| B["Aggregated into<br/>popularity_daily"]
    B -->|"kept forever"| C["Historical<br/>trends"]
    A -->|"after 90 days"| D["Purged by<br/>PurgeOldEvents"]

    style A fill:#fff3e0,stroke:#e65100
    style B fill:#e3f2fd,stroke:#1565c0
    style C fill:#e8f5e9,stroke:#2e7d32
    style D fill:#fce4ec,stroke:#c62828
```
