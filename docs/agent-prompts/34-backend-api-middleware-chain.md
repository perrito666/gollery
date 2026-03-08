# Prompt 34 — Backend API middleware chain

Refactor API handler registration into a middleware chain.

Implement:
- a middleware composition helper (`internal/api/middleware.go`) that chains handlers
- extract existing logging, auth, CSRF, and rate limiting into composable middleware functions
- apply middleware in order: logging → rate-limit (where needed) → auth → CSRF (where needed) → handler
- split route registration so read-only routes skip CSRF and rate limiting
- add CORS middleware with configurable allowed origins from `ServerConfig`
- add `Content-Security-Policy`, `X-Content-Type-Options`, `X-Frame-Options` response headers
- tests for middleware ordering and header presence

Do not change individual handler implementations.
