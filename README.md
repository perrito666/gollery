# gollery

A filesystem-first image gallery monorepo written in Go, with a lightweight customizable frontend and optional PostgreSQL-backed popularity analytics.

This is the **full starter kit** intended to replace the existing `github.com/perrito666/gollery` repository with a clean monorepo structure optimized for incremental development through coding agents.

## What this repository is for

- **Backend**: Go REST API, filesystem-first catalog, ACLs, pluggable discussion providers
- **Frontend**: lightweight static JavaScript app, customizable without touching core functionality
- **Optional analytics**: PostgreSQL-backed popularity tracking designed to be GDPR-safe for Europe
- **Agent-first workflow**: short prompts, narrow tasks, ADRs, and explicit contracts to avoid giant context windows

## Monorepo structure

```text
backend/
frontend/
docs/
  backend-technical-design.md
  frontend-technical-design.md
  monorepo-layout.md
  agent-workflow.md
  agent-prompts/
  adrs/
.github/
LICENSE
Makefile
AGENTS.md
CONTRIBUTING.md
```

## Recommended implementation order

### Backend
1. repo skeleton
2. config + domain model
3. scanner + publication rules
4. sidecar state + stable IDs
5. snapshot builder
6. ACL engine
7. auth abstraction
8. baseline REST API
9. derivative generation
10. discussion abstraction
11. Mastodon + Bluesky providers
12. PostgreSQL popularity analytics
13. watcher and reconciliation
14. hardening

### Frontend
1. skeleton
2. core router/state/api
3. ui-contract
4. default pages
5. build + override system
6. optional popularity UI
7. hardening

## Quick start

```bash
make tree
make docs
make prompts
make package
```

## License

MIT
