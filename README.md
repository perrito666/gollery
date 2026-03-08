# gollery

A filesystem-first image gallery monorepo written in Go, with a lightweight customizable frontend and optional PostgreSQL-backed popularity analytics.

## What this is

- **Backend**: Go REST API, filesystem-first catalog, ACLs, pluggable discussion providers (Mastodon, Bluesky)
- **Frontend**: lightweight static JavaScript app, customizable without touching core functionality
- **Optional analytics**: PostgreSQL-backed popularity tracking designed to be GDPR-safe
- **Agent-first workflow**: 52 short prompts across 12 phases, narrow tasks, ADRs

## Current status

Phase 1 (foundation, prompts 01–18) is complete. The backend compiles and passes all tests. The frontend bundles at ~12kb. The server does not run end-to-end yet — that comes in Phase 7 (prompts 37–38).

See `docs/agent-workflow.md` for the full implementation plan.

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
```

## Implementation plan

### Completed
- Filesystem scanner with publication rules and config inheritance
- Sidecar state management with stable object IDs
- Snapshot builder (scan → domain model)
- ACL engine (public / authenticated / restricted)
- Auth abstraction (interfaces)
- REST API baseline (6 endpoints: albums, assets, derivatives)
- Image derivative generation (thumbnails, previews)
- Discussion providers (Mastodon, Bluesky)
- PostgreSQL popularity analytics with tern migrations
- Filesystem watcher with debounced reconciliation
- Frontend core (API client, router, store, session, controllers)
- Frontend default UI (7 views, classic album grid)
- Frontend build/override system (site layer)
- Frontend optional popularity components
- First hardening pass

### Remaining (prompts 19–52)
- Frontend refactors (shared utilities, error handling)
- Production infrastructure (logging, config, auth, CSRF, rate limiting)
- Missing API routes (discussions, access, admin, analytics)
- Analytics wiring (event recording, retention jobs)
- API middleware chain and route group refactor
- App wiring and server startup
- Derivative quality, cache eviction, asset ACLs, prev/next nav
- Frontend tests
- Pagination, EXIF metadata
- Deployment configs (Dockerfile, docker-compose)
- Final hardening and production readiness audit

## For coding agents

Read `CLAUDE.md` (for Claude Code) or `docs/using-this-with-an-llm-cli.md` (for other agents).

## License

MIT
