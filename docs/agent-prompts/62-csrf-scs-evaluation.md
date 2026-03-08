# Prompt 62 — Evaluate CSRF with scs (read-only)

After switching to scs sessions (prompt 61), evaluate whether the hand-rolled CSRF implementation should also be replaced.

## Investigate

1. Read the current CSRF implementation in `api.go` (`csrfMiddleware`, `handleCSRFToken`).
2. Check if scs provides CSRF protection or recommends a companion library (e.g., `justinas/nosurf`, `gorilla/csrf`).
3. Evaluate:
   - Does scs's cookie-based approach give us SameSite protection for free?
   - Is the current HMAC-based CSRF token scheme still appropriate with scs?
   - Would `nosurf` or `gorilla/csrf` reduce maintenance burden?
   - What changes would the frontend need (how it obtains and sends the CSRF token)?

## Report

Recommend one of:
- **Replace** CSRF with a library (specify which, with migration steps)
- **Keep** current CSRF (with justification for why scs doesn't change the calculus)
- **Simplify** to SameSite-only (if the threat model allows it)

## Do not

- Modify any source files
