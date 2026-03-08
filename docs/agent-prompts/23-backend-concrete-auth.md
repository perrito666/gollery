# Prompt 23 — Backend concrete auth implementation

Implement a concrete authentication backend.

Implement:
- a file-based user store (`internal/auth/userstore.go`) that reads a `users.json` file with username, bcrypt-hashed password, groups, and is_admin
- `Authenticate(username, password)` that verifies credentials against the store
- a cookie-based `SessionStore` implementation (`internal/auth/sessions.go`) using secure, HttpOnly, SameSite=Lax cookies with HMAC-signed session tokens
- session creation, lookup, and deletion
- tests for authentication and session lifecycle

Do not implement HTTP handlers or middleware yet.
Do not implement CSRF protection yet.
