# Frontend Technical Design
## gollery frontend v2

Language: English  
Implementation style: lightweight static JavaScript  
License: MIT

---

## 1. Summary

The frontend is a lightweight static application that consumes the backend REST API.

It is designed to be customizable without modifying core functionality.

The frontend is split into four layers:

1. `core`
2. `ui-contract`
3. `ui-default`
4. `site`

This allows deployers to change the UI without understanding routing, auth, ACLs, or backend semantics.

---

## 2. Goals

- very lightweight
- framework-independent by default
- easy to self-host
- classic album UI
- customizable without touching core logic
- supports admin actions later
- built in the same monorepo as the backend

---

## 3. Layer model

### `core`
Owns functionality:
- API access
- auth/session
- router
- state
- controllers
- action execution
- permission interpretation

### `ui-contract`
Defines:
- view models
- component contracts
- events
- expected action hooks

### `ui-default`
Implements the default visual UI:
- layouts
- pages
- components
- styles

### `site`
Contains site-specific overrides:
- branding
- CSS overrides
- component overrides
- view overrides
- site config

---

## 4. Render flow

```text
route -> controller -> view model -> renderer -> DOM
```

Views should not fetch backend data directly.
Views should not contain ACL logic directly.
Views should not know backend endpoint details.

---

## 5. Main views

- HomePage
- AlbumPage
- AssetPage
- LoginPage
- ForbiddenPage
- NotFoundPage

---

## 6. Build model

The frontend build should resolve:
1. core
2. ui-contract
3. ui-default
4. site overrides

If a site override exists, use it. Otherwise use the default implementation.

A lightweight bundler such as `esbuild` may be used internally, but the user-facing workflow should remain simple, for example `make build`.

---

## 7. Suggested layout

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

---

## 8. Customization model

Supported customization levels:
1. branding only
2. CSS/layout overrides
3. component replacement
4. full view replacement

Things that must remain in core:
- API semantics
- auth/session logic
- route parsing
- ACL interpretation
- action execution
- error normalization

---

## 9. Minimal UX target

Start with a classic album UI:
- root album list
- album page with child albums and photo grid
- asset page with large image and previous/next navigation
- login page
- discussion links displayed simply
- admin actions progressively enabled

---

## 10. Optional popularity UI

The frontend may optionally consume popularity data from the backend, but:
- it must not assume analytics exist
- all popularity UI should be feature-flagged
- restricted content must never leak popularity data publicly

Possible optional UI:
- popular assets in album
- “trending” badges
- admin-only popularity panel
