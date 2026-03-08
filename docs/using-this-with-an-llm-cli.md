# Using This Starter Kit With an LLM CLI

This repository is designed to be the **control plane** for a coding agent.

## Basic workflow

1. Unzip or copy this starter kit into your local clone of `github.com/perrito666/gollery`
2. Commit it as the new baseline
3. Run your coding agent against the repo
4. Feed it **one prompt file at a time**
5. Review, test, commit
6. Move to the next prompt

## Why one prompt at a time

Do not ask the agent to build the whole app in one shot.

This repository is intentionally split into:
- ADRs
- design docs
- short prompts

This helps the agent stay aligned without using a huge context window.

## Standard agent prompt template

Use a prompt like this:

```text
You are working inside the gollery monorepo.

Before making changes, read:
- AGENTS.md
- docs/backend-technical-design.md
- docs/frontend-technical-design.md
- docs/monorepo-layout.md
- docs/agent-workflow.md
- docs/agent-prompts/01-backend-repo-skeleton.md

Task:
Implement only the scope described in docs/agent-prompts/01-backend-repo-skeleton.md.

Rules:
- Do not expand scope
- Do not redesign architecture
- Keep filesystem as source of truth
- Keep analytics optional and PostgreSQL-backed
- Keep frontend separate from backend
- Add or update tests if applicable
- Stop when the prompt is complete

At the end, provide:
1. files changed
2. rationale
3. tests added
4. follow-up TODOs
```

## Recommended sequence

Backend:
1. 01-backend-repo-skeleton
2. 02-backend-config-domain
3. 03-backend-scanner
4. 04-backend-sidecar-state
5. 05-backend-snapshot-builder
6. 06-backend-acl-and-auth-abstraction
7. 07-backend-rest-baseline
8. 08-backend-derivatives
9. 09-backend-discussion-abstraction
10. 10-backend-mastodon-bluesky
11. 11-backend-popularity-postgres
12. 12-backend-watcher

Frontend:
13. 13-frontend-skeleton
14. 14-frontend-core-and-contracts
15. 15-frontend-default-ui
16. 16-frontend-build-overrides
17. 17-frontend-popularity-ui-optional

Then:
18. 18-hardening-pass

## Best practice

- one prompt = one branch or one commit
- review diff after every prompt
- do a second “self-review” prompt after implementation
- do not let the agent redesign the architecture unless you explicitly want that
