# Prompt 24 — Backend auth middleware and routes

Wire authentication into the HTTP layer.

Implement:
- auth middleware that extracts the session cookie, looks up the session, and sets the principal in context
- `POST /api/v1/auth/login` handler
- `GET /api/v1/auth/me` handler
- `POST /api/v1/auth/logout` handler
- tests for each endpoint (valid login, bad credentials, session restore, logout)

Do not implement CSRF protection yet.
Do not implement rate limiting yet.
