# Using This Repository With an LLM CLI

This repository is designed to be the **control plane** for a coding agent.

## Quick start

If using Claude Code, the project includes a `CLAUDE.md` with full instructions. Claude Code will read it automatically.

For other agents, use the standard prompt template below.

## Basic workflow

1. Clone or open `github.com/perrito666/gollery`
2. Check the current state in `AGENTS.md` (look for "Next prompt to execute")
3. Feed the agent **one prompt file at a time** from `docs/agent-prompts/`
4. The agent follows the plan → implement → verify → self-review → correct cycle
5. Review the diff, test, commit
6. Move to the next prompt

## Why one prompt at a time

Do not ask the agent to build the whole app in one shot.

This repository is intentionally split into:
- ADRs (architectural decisions)
- design docs (backend + frontend)
- 52 short prompts across 12 phases

This helps the agent stay aligned without using a huge context window.

## Standard agent prompt template

```text
You are working inside the gollery monorepo.

Before making changes, read:
- AGENTS.md
- CLAUDE.md (if present)
- docs/agent-workflow.md
- The specific prompt file: docs/agent-prompts/NN-*.md

Follow this work process for the prompt:

1. PLAN: Re-read the prompt and relevant architecture docs. Read existing files
   before modifying them. State the plan briefly.
2. IMPLEMENT: One subsystem, 1–8 files changed. Add tests. Do not expand scope.
3. VERIFY: Run build and tests (see CLAUDE.md for commands).
4. SELF-REVIEW: Check for architecture violations, missing tests, scope creep.
5. CORRECT: Fix issues found, re-verify.
6. REPORT: State files changed, tests added, concerns, next prompt.

Rules:
- Do not expand scope beyond what the prompt requests
- Do not redesign architecture
- Keep filesystem as source of truth
- Keep analytics optional and PostgreSQL-backed
- Keep frontend layers separated (core/ui-contract/ui-default/site)
- Add or update tests if applicable
- Stop when the prompt is complete
```

## Prompt sequence

See `docs/agent-workflow.md` for the full 52-prompt, 12-phase plan.

### Phase 1 — Foundation (01–18) ✅ COMPLETE
Backend skeleton through first hardening pass, frontend core through build system.

### Phase 2 — Refactors (19–20)
Frontend code deduplication.

### Phase 3 — Production infrastructure (21–26)
Logging, config, auth, CSRF, rate limiting.

### Phase 4 — Missing API routes (27–31)
Discussions, access, admin, analytics, path query.

### Phase 5 — Analytics wiring (32–33)
Event recording and retention jobs.

### Phase 6 — API architecture (34–36)
Middleware chain, route groups, timeouts, health.

### Phase 7 — Server startup (37–38)
App wiring and main.go — makes the server actually run.

### Phase 8 — Quality and features (39–44)
Derivative quality, cache eviction, asset ACLs, prev/next nav, login UX.

### Phase 9 — Testing (45–46)
Frontend test setup and controller tests.

### Phase 10 — Scale and metadata (47–48)
Pagination and EXIF extraction.

### Phase 11 — Deployment (49–50)
Dockerfile and docker-compose.

### Phase 12 — Final hardening (51–52)
Full review and production readiness audit.

## Semi-autonomous mode

If you want the agent to execute multiple prompts in sequence, use:

```text
Execute prompts 19 through 26 in order. For each prompt:
1. Re-read architecture docs
2. Follow the plan/implement/verify/self-review/correct/report cycle
3. Stop after each prompt's self-review before continuing to the next
```

## Best practice

- One prompt = one branch or one commit
- Review diff after every prompt
- Do a second "self-review" prompt after implementation
- Do not let the agent redesign the architecture unless you explicitly want that
- Check `go test ./...` and the frontend build after every prompt
