# Prompt 52 — Final production readiness audit

Perform a read-only audit of the full repository.

Verify against the production readiness checklist:

Must have:
- [ ] app wiring connects all subsystems
- [ ] main.go starts a real HTTP server
- [ ] concrete auth with sessions and middleware
- [ ] auth API routes (login, logout, me)
- [ ] CSRF protection on state-changing endpoints
- [ ] request logging with structured output
- [ ] error logging with context
- [ ] configuration loading from file and environment
- [ ] security headers (CSP, X-Content-Type-Options, X-Frame-Options)
- [ ] CORS configuration

Should have:
- [ ] rate limiting on auth endpoints
- [ ] request timeouts configured
- [ ] health check endpoint
- [ ] graceful shutdown
- [ ] analytics retention job
- [ ] analytics dedup window
- [ ] cache eviction for orphaned derivatives
- [ ] pagination for album asset listings
- [ ] discussion API routes
- [ ] admin API routes
- [ ] frontend test runner with passing tests

Nice to have:
- [ ] deployment configs (Dockerfile, docker-compose)
- [ ] image metadata extraction
- [ ] improved derivative quality
- [ ] asset-level ACL overrides
- [ ] previous/next asset navigation

Report any items that are incomplete or missing.
Do not modify the repository.
