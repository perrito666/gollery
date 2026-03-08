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

## Prompt sequence

### Phase 1 — Foundation (01–18, complete)

1. backend repo skeleton
2. config/domain
3. scanner
4. sidecar state
5. snapshot builder
6. ACL and auth abstraction
7. REST baseline
8. derivatives
9. discussion abstraction
10. mastodon/bluesky providers
11. popularity analytics (PostgreSQL)
12. watcher
13. frontend skeleton
14. frontend core and contracts
15. frontend default UI
16. frontend build overrides
17. frontend popularity UI (optional)
18. first hardening pass

### Phase 2 — Refactors (19–20)

19. frontend shared utilities (deduplicate `esc()`)
20. frontend shared controller error handling

### Phase 3 — Production infrastructure (21–26)

21. backend structured logging (slog)
22. backend config loading (file + env)
23. backend concrete auth (user store + sessions)
24. backend auth middleware and routes
25. backend CSRF protection
26. backend rate limiting

### Phase 4 — Missing API routes (27–31)

27. discussion API routes
28. access info API routes
29. admin API routes
30. analytics API routes
31. album path-based lookup

### Phase 5 — Analytics wiring (32–33)

32. analytics event recording middleware
33. analytics retention job

### Phase 6 — API architecture (34–36)

34. API middleware chain refactor
35. API route group split
36. server timeouts and health check

### Phase 7 — Server startup (37–38)

37. app package wiring
38. main.go entrypoint

### Phase 8 — Quality and features (39–44)

39. derivative quality improvement (CatmullRom)
40. cache eviction for orphaned derivatives
41. asset-level ACL overrides
42. previous/next asset navigation (backend)
43. frontend previous/next wiring
44. frontend login loading state

### Phase 9 — Testing (45–46)

45. frontend test setup
46. frontend controller tests

### Phase 10 — Scale and metadata (47–48)

47. backend API pagination
48. backend image metadata extraction (EXIF)

### Phase 11 — Deployment (49–50)

49. Dockerfile
50. docker-compose

### Phase 12 — Final hardening (51–52)

51. second hardening pass
52. final production readiness audit

---

**Phases 1–12 (prompts 01–52) are complete.** The prompts below are the roadmap for long-term maintainability.

---

### Phase 13 — API cleanup (53–54)

53. API server dependency struct (replace Set* methods)
54. update tests for dependency struct

### Phase 14 — Observability (55–56)

55. request ID middleware
56. wire request ID into structured logging

### Phase 15 — Unified migrations (57–58)

57. create shared migration package
58. wire unified migrations into analytics and app

### Phase 16 — Session store upgrade (59–62, conditional)

59. evaluate scs session library replacement (read-only)
60. add scs session store adapter (if 59 recommends)
61. wire scs session store into app (if 60 done)
62. evaluate CSRF with scs (read-only, if 61 done)

### Phase 17 — Filesystem watcher upgrade (63–64)

63. replace polling watcher with fsnotify
64. update app wiring for fsnotify watcher

### Phase 18 — OpenAPI specification (65–67)

65. OpenAPI spec: content routes
66. OpenAPI spec: auth and admin routes
67. OpenAPI spec: analytics, discussion, and access routes

### Phase 19 — Frontend build modernization (68–70)

68. initialize Vite project
69. Vite CSS handling and esbuild removal
70. Vite plugin for theme resolution

### Phase 20 — App lifecycle (71)

71. app lifecycle phases (Init/Wire/Serve/Shutdown)

### Phase 21 — Background processing (72–73)

72. derivative worker pool: types and queue
73. derivative worker pool: API integration

### Phase 22 — Cache improvements (74)

74. content-addressable cache keys

### Phase 23 — Search (75–76)

75. search index: types and interface
76. search API route and app wiring

## Deliverable expectations per prompt

Every agent step should return:
1. files changed
2. rationale
3. tests added
4. tradeoffs
5. remaining TODOs
