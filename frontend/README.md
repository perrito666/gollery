# frontend

Lightweight static frontend for gollery.

## Layers

- `src/core/` — API access, auth, router, state, controllers
- `src/ui-contract/` — view models, component contracts, events
- `src/ui-default/` — default layouts, views, components, styles
- `src/site/` — site-specific overrides (branding, CSS, components)

## Build

```bash
make install
make build
```

## Development

```bash
make dev
```

## Layout

See `docs/frontend-technical-design.md` for the full design.
