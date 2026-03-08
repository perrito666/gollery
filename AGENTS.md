# AGENTS.md

This repository is designed for development with coding agents.

## Core rules

- Keep tasks **small and local**
- Prefer **one subsystem per prompt**
- Do not redesign architecture unless the relevant ADR or design document is updated
- Do not move content/publication rules out of the filesystem
- Do not place popularity analytics in the filesystem state layer
- Keep frontend **core functionality** separate from **UI implementation**
- Preserve the MIT license

## Source-of-truth documents

Read these before changing the codebase:

1. `docs/backend-technical-design.md`
2. `docs/frontend-technical-design.md`
3. `docs/monorepo-layout.md`
4. `docs/agent-workflow.md`
5. relevant ADRs in `docs/adrs/`

## Backend invariants

- Filesystem is the source of truth for content
- `album.json` is declarative only
- `.gallery/*.state.json` stores mutable editorial state
- access control supports `public`, `authenticated`, and `restricted`
- discussions are provider-pluggable
- popularity analytics are optional and stored in PostgreSQL
- analytics must not be required for gallery correctness
- analytics must be privacy-preserving by default

## Frontend invariants

- frontend lives in the same monorepo
- keep `core`, `ui-contract`, `ui-default`, and `site` separated
- UI can be replaced without rewriting functionality
- build should resolve defaults + site overrides
- avoid framework lock-in unless explicitly chosen later

## Working style

When implementing features:

1. restate the task briefly
2. identify affected files
3. implement only the requested scope
4. add tests
5. document tradeoffs
6. stop without expanding scope

## Prompt discipline

Use the short prompts in `docs/agent-prompts/`.
Do not merge several prompts into one giant request unless necessary.
