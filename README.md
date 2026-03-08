# gollery

A filesystem-first image gallery monorepo written in Go, with a lightweight customizable frontend and optional PostgreSQL-backed popularity analytics.

## What this is

- **Backend**: Go REST API, filesystem-first catalog, ACLs, pluggable discussion providers (Mastodon, Bluesky)
- **Frontend**: lightweight static JavaScript app, customizable without touching core functionality
- **Optional analytics**: PostgreSQL-backed popularity tracking designed to be GDPR-safe
- **Agent-first workflow**: 52 short prompts across 12 phases, narrow tasks, ADRs

## Current status

All 12 phases (prompts 01–52) are complete. The backend compiles, vets, and passes all 219 unit tests (+ 4 integration tests) across 19 packages. The frontend bundles at ~11.4kb minified and passes 38 tests across 8 suites. The server runs end-to-end via `app.Run()` with signal-aware graceful shutdown.

See `docs/agent-workflow.md` for the full implementation plan and future roadmap (phases 13–23).

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

## Building

```bash
# Backend
make backend-build    # compile galleryd binary
make backend-test     # run all Go tests

# Frontend
make frontend-build   # resolve theme + esbuild bundle

# Both
make backend-build && make frontend-build

# Docker
docker-compose up     # runs gallery + PostgreSQL
```

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
- **Deployment** — multi-stage Dockerfile, docker-compose with PostgreSQL
- **Signal-aware graceful shutdown**

## For coding agents

Read `CLAUDE.md` (for Claude Code) or `docs/using-this-with-an-llm-cli.md` (for other agents).

## License

MIT
