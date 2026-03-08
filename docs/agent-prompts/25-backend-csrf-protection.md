# Prompt 25 — Backend CSRF protection

Add CSRF protection for state-changing endpoints.

Implement:
- a CSRF middleware that validates a token on POST/PATCH/DELETE requests
- token generation and validation using HMAC with the session secret
- `GET /api/v1/auth/csrf-token` endpoint that returns a fresh token
- skip CSRF check for the login endpoint (no session yet)
- tests for token validation, missing token, and invalid token

Do not change existing endpoint behavior for GET requests.
