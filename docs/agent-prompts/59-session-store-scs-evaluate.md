# Prompt 59 — Evaluate session library replacement (read-only)

Evaluate replacing the hand-rolled HMAC session store with `alexedwards/scs` v2. This is a **read-only research prompt** — do not modify the repository.

## Investigate

1. Read the current auth implementation:
   - `backend/internal/auth/auth.go` — interfaces and concrete `CookieSessionStore`
   - `backend/internal/api/api.go` — how `authMiddleware` and `csrfMiddleware` use sessions

2. Read the scs v2 documentation (or source) and evaluate:
   - Does scs support HMAC-signed cookie sessions (no server-side storage)?
   - How does scs handle CSRF? Does it provide its own CSRF protection?
   - What is the migration path for the `Authenticator` interface? Does it need to change?
   - What cookie attributes does scs set by default (SameSite, Secure, HttpOnly)?
   - Does scs work with `net/http` middleware pattern?

3. Identify risks:
   - Will existing session cookies be invalidated?
   - Does scs require changes to the frontend login/logout flow?
   - Are there any behavioral differences in token rotation or expiry?

## Report

Write findings as a brief document. Recommend either:
- **Proceed** with scs adoption (with specific migration steps), or
- **Keep** the current implementation (with justification)

## Do not

- Modify any source files
- Add dependencies
- Change tests
