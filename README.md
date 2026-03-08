# gollery

A filesystem-first image gallery monorepo written in Go, with a lightweight customizable frontend and optional PostgreSQL-backed popularity analytics.

## A Note from the partial author

I had previously abandonned this project due to lack of time yet I kept wanting the functionality.
Over one rainy weekend I tried my hand at talking to an LLM into building what I wanted.

The result is ok-ish, I would not use this if I were you, I have not yet reviewed all of the code
but I thought it would be an interesting exercise to publish the result as it is with all the prompting
and feedback loop as used by the AI. I polished the architecture based on my original idea talking to
ChatGPT, mostly to cleanup my architecture design into wording understandable by an LLM then used the
initial prompt to get Claude Code going.

## Table of Contents

- [What this is](#what-this-is)
- [Current status](#current-status)
- [Features](#features)
- [Monorepo structure](#monorepo-structure)
- [Quick start](#quick-start)
- [Building](#building)
- [Running with Docker](#running-with-docker)
- [Configuration](#configuration)
- [User management](#user-management)
- [Content structure](#content-structure)
- [Documentation](#documentation)
- [For coding agents](#for-coding-agents)
- [License](#license)

## What this is

- **Backend**: Go REST API, filesystem-first catalog, ACLs, pluggable discussion providers (Mastodon, Bluesky)
- **Frontend**: lightweight static JavaScript app, customizable without touching core functionality
- **Optional analytics**: PostgreSQL-backed popularity tracking designed to be GDPR-safe
- **Agent-first workflow**: 52 short prompts across 12 phases, narrow tasks, ADRs

## Current status

All 12 phases (prompts 01–52) are complete. The backend compiles, vets, and passes all 219 unit tests (+ 4 integration tests) across 19 packages. The frontend bundles at ~11.4kb minified and passes 38 tests across 8 suites. The server runs end-to-end via `app.Run()` with signal-aware graceful shutdown.

See `docs/agent-workflow.md` for the full implementation plan and future roadmap (phases 13–23).

## Features

- **Filesystem scanner** with publication rules, config inheritance, and debounced watcher
- **Sidecar state** management with stable object IDs (`alb_<hex>` / `ast_<hex>`)
- **Snapshot builder** (scan → domain model)
- **ACL engine** (public / authenticated / restricted) with asset-level overrides
- **Concrete auth** — file-based user store, bcrypt passwords, HMAC cookie sessions
- **CSRF protection** and **rate limiting**
- **REST API** — albums, assets, derivatives, discussions, access, admin, analytics, pagination, prev/next navigation
- **Image derivatives** — CatmullRom quality scaling, cache eviction for orphans
- **EXIF metadata** extraction
- **Discussion providers** — Mastodon, Bluesky (pluggable via `Provider` interface)
- **PostgreSQL popularity analytics** — tern migrations, event recording, retention jobs
- **Structured logging** (slog) throughout
- **Frontend** — 7 default views, classic album grid, login state, optional popularity components
- **Frontend build/override system** — site layer for customization without touching core
- **Deployment** — multi-stage Dockerfile, docker-compose with PostgreSQL and nginx
- **Signal-aware graceful shutdown**

## Monorepo structure

```text
backend/          Go REST API server
frontend/         Lightweight static JavaScript app
docs/             Architecture docs, ADRs, agent prompts
scripts/          Helper scripts (test DB, etc.)
Makefile          Top-level build targets
AGENTS.md         Rules for coding agents
CLAUDE.md         Claude Code project instructions
CONTRIBUTING.md   How to contribute
LICENSE           MIT
```

## Quick start

```bash
# 1. Build frontend
make frontend-build

# 2. Start everything with Docker
docker-compose up --build

# 3. Open http://localhost:8090 in your browser
```

The default `docker-compose.yml` expects:
- `sample-content/` — a directory of images with `album.json` files (see [Content structure](#content-structure))
- `users.json` — user credentials with bcrypt-hashed passwords (see [Configuration](#configuration))
- `gollery.json` — server configuration

## Building

```bash
# Backend only
make backend-build    # compile galleryd binary
make backend-test     # run all Go tests

# Frontend only
make frontend-build   # resolve theme + esbuild bundle
make frontend-test    # run frontend tests

# Both
make build            # compile backend + bundle frontend

# Full verification
make verify           # build + test + vet
```

## Running with Docker

The project uses Docker Compose with three services:

| Service     | Purpose                              | Port |
|-------------|--------------------------------------|------|
| `galleryd`  | Go backend API server                | 8080 (API only) |
| `nginx`     | Frontend static files + API proxy    | 8090 (browse here) |
| `postgres`  | PostgreSQL for optional analytics    | 5432 (internal) |

```bash
# Build and start
docker-compose up --build

# Stop
docker-compose down

# Stop and remove all data (cache, database)
docker-compose down -v
```

Browse the gallery at **http://localhost:8090**. The API is also directly accessible at `http://localhost:8080/api/v1/...`.

For a detailed guide including how to set up content, users, and configuration from scratch, see [docs/running-locally.md](docs/running-locally.md).

## Configuration

### Server config (`gollery.json`)

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

### Environment variable overrides

| Variable | Overrides |
|----------|-----------|
| `GOLLERY_LISTEN_ADDR` | `listen_addr` |
| `GOLLERY_POSTGRES_DSN` | `analytics.postgres_dsn_env` |
| `GOLLERY_SESSION_SECRET` | `auth.session_secret` |

### User credentials (`users.json`)

```json
[
  {
    "username": "admin",
    "password": "$2a$10$BCRYPT_HASH_HERE",
    "groups": ["admins"],
    "is_admin": true
  }
]
```

Passwords must be bcrypt-hashed (`$2a$` prefix). Use the `gollery-users` tool (see below) or see [docs/running-locally.md](docs/running-locally.md) for alternatives.

## User management

The `gollery-users` CLI tool manages `users.json` — adding/removing users, changing passwords, and setting roles. It handles bcrypt hashing automatically.

```bash
# Build the tool
make backend-build

# Add a user
./backend/gollery-users add -username alice -password secret -admin -groups admins

# List users
./backend/gollery-users list

# Change password
./backend/gollery-users passwd -username alice -password newpass

# Remove a user
./backend/gollery-users remove -username alice
```

For the full command reference, see [docs/user-management.md](docs/user-management.md).

**Note:** The server does not hot-reload `users.json` — restart after changes.

## Content structure

```
sample-content/
├── album.json              ← root album config (required)
├── Vacation/
│   ├── album.json          ← sub-album config (inherits from root)
│   ├── beach.jpg
│   └── sunset.png
└── Family/
    ├── album.json
    └── portrait.jpg
```

- **Albums** are directories with an `album.json` file
- **Assets** are image files (`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`)
- Child albums inherit parent config unless `"inherit": false` is set
- The server never writes to content directories — all state goes to `.gallery/` sidecars and the cache directory

### album.json

```json
{
  "title": "Album Title",
  "description": "Album description",
  "access": {
    "view": "public"
  },
  "derivatives": {
    "thumbnail_sizes": [200, 400],
    "preview_sizes": [800, 1600]
  }
}
```

Access modes: `"public"` (anyone), `"authenticated"` (logged-in users), `"restricted"` (specific users/groups).

## Documentation

| Document | Description |
|----------|-------------|
| [Running Locally](docs/running-locally.md) | Step-by-step guide to run with Docker Compose |
| [User Management](docs/user-management.md) | Managing users, passwords, groups, and ACLs |
| [Backend Technical Design](docs/backend-technical-design.md) | Backend architecture and package design |
| [Frontend Technical Design](docs/frontend-technical-design.md) | Frontend architecture, views, and build system |
| [Monorepo Layout](docs/monorepo-layout.md) | Directory structure and conventions |
| [Architecture Diagrams](docs/architecture-diagrams.md) | Visual architecture overview |
| [Agent Workflow](docs/agent-workflow.md) | 52-prompt development plan and phase status |
| [Using with an LLM CLI](docs/using-this-with-an-llm-cli.md) | Guide for AI coding agents |
| [ADRs](docs/adrs/) | Architecture Decision Records |
| [Developer Guide](DEVELOPER_GUIDE.md) | Detailed subsystem walkthrough for contributors |
| [Contributing](CONTRIBUTING.md) | How to contribute |

## For coding agents

Read `CLAUDE.md` (for Claude Code) or `docs/using-this-with-an-llm-cli.md` (for other agents).

## License

MIT
