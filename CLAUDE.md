# Claude Code Project Instructions

## What this project is

Gollery is a filesystem-first image gallery monorepo with a Go backend and lightweight JavaScript frontend. It discovers albums from a directory tree, serves a REST API, and optionally tracks popularity analytics in PostgreSQL.

## Before you start any work

1. Read `AGENTS.md` for invariants and rules
2. Read the relevant prompt in `docs/agent-prompts/NN-*.md`
3. Check `docs/agent-workflow.md` for the current phase and prompt sequence
4. Read the architecture docs relevant to your task:
   - Backend: `docs/backend-technical-design.md`
   - Frontend: `docs/frontend-technical-design.md`
   - Layout: `docs/monorepo-layout.md`
   - ADRs: `docs/adrs/`

## Current project state

**All 52 prompts across 12 phases are complete.** The development sequence is finished.

The backend compiles, vets, and passes all 219 unit tests (+ 4 integration tests) across 19 packages. The frontend bundles at 11.4kb minified and passes 38 tests across 8 suites. The server starts via `app.Run()` with signal-aware graceful shutdown, concrete auth (bcrypt + HMAC sessions), CSRF protection, rate limiting, structured logging, all API routes, pagination, EXIF extraction, and deployment configs (Dockerfile + docker-compose).

See `docs/agent-workflow.md` for the full 52-prompt, 12-phase plan (all complete).

## Work process — follow this cycle for every prompt

### 1. Plan
- Re-read the prompt file and relevant docs
- Read existing files before modifying them
- State the plan briefly

### 2. Implement
- One subsystem per prompt, 1–8 files changed
- Do not expand scope beyond what the prompt requests
- Add tests for new functionality
- Follow existing patterns (see below)

### 3. Verify
```bash
# Backend
cd backend && go build ./... && go vet ./... && go test ./...

# Frontend
cd frontend && node scripts/resolve-theme.js && npx esbuild src/main.js --bundle --outfile=/dev/null --format=esm

# Full build
make backend-build && make backend-test && make frontend-build
```

### 4. Self-review
Check for: architecture violations, layer boundary violations, missing tests, scope creep, XSS safety, error handling completeness.

### 5. Correct
Fix issues found, re-verify.

### 6. Report
State: files changed, tests added, any concerns, which prompt comes next.

## Architecture invariants — never violate these

1. **Filesystem is source of truth** for content structure
2. **`album.json` is declarative only** — server never writes to it
3. **`.gallery/*.state.json`** stores mutable editorial state (IDs, discussion bindings)
4. **Analytics in PostgreSQL only**, never in filesystem state
5. **Analytics are optional** — gallery works without them
6. **Frontend layers**: core / ui-contract / ui-default / site — views never import from core directly
7. **Discussions are provider-pluggable** via the Provider interface

## Code patterns

### Backend
- Tests use `t.TempDir()` for filesystem fixtures
- Discussion providers use `Poster` interface for testability
- Integration tests: `//go:build integration` tag, `GOLLERY_TEST_POSTGRES_DSN` env var
- Atomic sidecar writes: temp file + `os.Rename`
- IDs: `alb_<hex>` / `ast_<hex>` via `crypto/rand`
- PostgreSQL: pgx v5 driver, tern v2 migrations

### Frontend
- Views export `render(container, viewModel, ctx)` and `destroy()`
- ctx: `{ store, router, session, permissions, features, popularity }`
- All dynamic HTML content must use `esc()` for XSS safety
- Build: `resolve-theme.js` generates `_resolved/registry.js`, then esbuild bundles
- Site overrides: place files in `src/site/views/` to replace default views
- `package.json` has `"type": "module"` — all scripts are ESM

## What NOT to do

- Do not redesign architecture without updating docs and ADRs
- Do not move content/publication rules out of the filesystem
- Do not place analytics data in filesystem state
- Do not add framework dependencies to the frontend
- Do not merge multiple prompts into one giant change
- Do not commit unless explicitly asked
- Do not add features beyond what the current prompt requests
