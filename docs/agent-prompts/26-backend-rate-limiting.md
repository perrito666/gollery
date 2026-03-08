# Prompt 26 — Backend rate limiting

Add rate limiting to auth endpoints.

Implement:
- an in-memory rate limiter using `golang.org/x/time/rate` (token bucket per IP)
- rate limit middleware applied to `/api/v1/auth/login` and `/api/v1/auth/logout`
- configurable burst and rate from `ServerConfig`
- return `429 Too Many Requests` with `Retry-After` header when exceeded
- tests for rate limiting behavior

Do not apply rate limiting to read-only album/asset endpoints yet.
