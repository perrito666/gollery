# Agent Workflow

This repository is optimized for short-context coding agent sessions.

## Why

Large architecture prompts cause:
- drift
- accidental redesigns
- incomplete tests
- broken package boundaries

Short, focused prompts work better.

## Rules

- one prompt = one subsystem or one vertical slice
- prefer 1 to 8 files changed, not 60
- require tests on each step
- stop after the requested scope
- update docs only when necessary

## Recommended backend sequence

1. skeleton
2. config/domain
3. scanner
4. state store
5. snapshot builder
6. ACL
7. auth
8. API baseline
9. derivatives
10. discussions
11. analytics
12. watcher
13. hardening

## Recommended frontend sequence

1. skeleton
2. core router/state/api
3. ui-contract
4. default home/album pages
5. asset page
6. login
7. build/override system
8. admin actions
9. optional popularity UI

## Deliverable expectations per prompt

Every agent step should return:
1. files changed
2. rationale
3. tests added
4. tradeoffs
5. remaining TODOs
