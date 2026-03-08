# CONTRIBUTING.md

## Workflow

This repository is prepared for incremental implementation through coding agents and human review.

### Recommended process

1. Pick one prompt from `docs/agent-prompts/`
2. Implement only that scope
3. Keep changes small
4. Add tests
5. Update docs or ADRs if the design changed
6. Open a focused pull request

## Scope discipline

Good:
- scanner only
- state store only
- one provider only
- one frontend view only
- analytics schema only

Bad:
- implement the whole backend
- redesign auth, ACLs, API, and frontend in one pass

## Documentation

Any meaningful architectural change should update:
- a technical design doc, or
- an ADR

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
