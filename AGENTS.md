# AGENTS.md

This repository is designed for development with coding agents.

## Current state

**All 52 prompts across 12 phases are complete.** The development sequence is finished.
See `docs/agent-workflow.md` for the full plan. Backend: 219 unit tests + 4 integration tests. Frontend: 38 tests, 11.4kb bundle.

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
6. `CLAUDE.md` for project-level instructions

## Backend invariants

- Filesystem is the source of truth for content
- `album.json` is declarative only — the server never writes to it
- `.gallery/*.state.json` stores mutable editorial state
- access control supports `public`, `authenticated`, and `restricted`
- discussions are provider-pluggable via the `Provider` interface
- popularity analytics are optional and stored in PostgreSQL
- analytics must not be required for gallery correctness
- analytics must be privacy-preserving by default

## Frontend invariants

- frontend lives in the same monorepo
- keep `core`, `ui-contract`, `ui-default`, and `site` separated
- views never import from `core/` directly — data arrives via `render(container, viewModel, ctx)`
- UI can be replaced without rewriting functionality
- build resolves defaults + site overrides via `resolve-theme.js`
- avoid framework lock-in unless explicitly chosen later

## Work process — every prompt follows this cycle

### 1. Plan
- Re-read the prompt file in `docs/agent-prompts/`
- Re-read relevant architecture docs
- Read existing files before modifying
- State the plan briefly

### 2. Implement
- One subsystem per prompt, 1–8 files changed
- Write only the requested scope — do not expand
- Add tests for new functionality
- Follow existing code patterns

### 3. Verify
- Backend: `go build ./...`, `go vet ./...`, `go test ./...`
- Frontend: `node scripts/resolve-theme.js && npx esbuild src/main.js --bundle --outfile=/dev/null --format=esm`
- Full: `make backend-build && make backend-test && make frontend-build`

### 4. Self-review
Check for: architecture violations, layer boundary violations, missing tests, scope creep, XSS safety, error handling completeness.

### 5. Correct
Fix issues found in self-review, re-verify.

### 6. Report and advance
State: files changed, tests added, concerns, which prompt comes next.

## Code patterns

### Backend
- Tests: `t.TempDir()` for filesystem fixtures
- Discussion providers: `Poster` interface for testability
- Integration tests: `//go:build integration`, `GOLLERY_TEST_POSTGRES_DSN`
- Sidecar writes: temp file + `os.Rename` for atomicity
- IDs: `alb_<hex>` / `ast_<hex>` via `crypto/rand`
- PostgreSQL: pgx v5, tern v2 for migrations
- Logging: use `log/slog` (once prompt 21 is implemented)

### Frontend
- Views export `render(container, viewModel, ctx)` and `destroy()`
- ctx: `{ store, router, session, permissions, features, popularity }`
- All dynamic HTML: use `esc()` for XSS safety
- Build: `resolve-theme.js` → `_resolved/registry.js` → esbuild
- ESM throughout: `"type": "module"` in package.json

## Prompt discipline

Use the short prompts in `docs/agent-prompts/`.
Do not merge several prompts into one giant request unless necessary.
