# Prompt 51 — Second hardening pass

Review the full monorepo after all feature work is complete.

Check:
- all API routes have ACL enforcement
- all POST/PATCH/DELETE routes have CSRF protection
- auth middleware is applied to all routes except healthz and login
- rate limiting is applied to auth endpoints
- structured logging covers all error paths
- no TODO/FIXME comments remain without tracked issues
- all existing tests pass
- `go vet` and `go build` are clean
- frontend bundle compiles
- `make build` succeeds at root level

Fix any issues found.
Do not introduce new features.
