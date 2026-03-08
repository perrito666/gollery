# Monorepo Layout

This repository hosts both the backend and the frontend in the same monorepo.

## Top-level layout

```text
backend/
frontend/
docs/
.github/
scripts/
Makefile
AGENTS.md
LICENSE
README.md
CONTRIBUTING.md
```

## Backend future layout

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

## Frontend future layout

```text
frontend/
  src/
    core/
      api/
      auth/
      router/
      state/
      controllers/
      services/
      models/
    ui-contract/
      view-models.js
      interfaces.js
      events.js
    ui-default/
      layouts/
      views/
      components/
      styles/
    site/
      assets/
      components/
      views/
      styles/
      site.config.json
  public/
  dist/
  Makefile
```

## Ownership boundaries

### backend/
Owns:
- catalog
- publication
- ACLs
- auth
- API
- derivatives
- discussion providers
- analytics collection and popularity queries

### frontend/
Owns:
- view composition
- navigation UI
- styling
- customization layer
- build pipeline for static assets

### docs/
Owns:
- architecture
- ADRs
- agent prompts
- contribution guidance
