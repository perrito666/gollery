# CONTRIBUTING.md

## Workflow

This repository is prepared for incremental implementation through coding agents and human review.

### Recommended process

1. Check `AGENTS.md` for the current state and next prompt number
2. Pick the next prompt from `docs/agent-prompts/`
3. Follow the work cycle: plan → implement → verify → self-review → correct → report
4. Keep changes small (1–8 files per prompt)
5. Add tests
6. Update docs or ADRs if the design changed
7. Open a focused pull request

### Work cycle detail

Every prompt follows this cycle:

1. **Plan**: Read the prompt file and relevant architecture docs. Read existing files before modifying.
2. **Implement**: Write only the requested scope. Add tests.
3. **Verify**: Run `go build`, `go vet`, `go test` (backend), or `make frontend-build` (frontend).
4. **Self-review**: Check for architecture violations, missing tests, scope creep.
5. **Correct**: Fix issues found, re-verify.
6. **Report**: State files changed, tests added, concerns, next prompt.

## Scope discipline

Good:
- scanner only
- state store only
- one provider only
- one frontend view only
- analytics schema only
- one middleware concern only

Bad:
- implement the whole backend
- redesign auth, ACLs, API, and frontend in one pass
- add features not requested by the current prompt

## Architecture rules

Read `AGENTS.md` for the full list of invariants. Key rules:
- Filesystem is source of truth for content
- `album.json` is declarative only
- Analytics in PostgreSQL only, optional
- Frontend layers (core/ui-contract/ui-default/site) must stay separated
- Views never import from core directly

## Documentation

Any meaningful architectural change should update:
- a technical design doc, or
- an ADR in `docs/adrs/`

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
