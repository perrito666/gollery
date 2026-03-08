# Prompt 61 — Wire scs session store into app

Replace the hand-rolled `CookieSessionStore` with the scs adapter from prompt 60.

## Implement

1. In `app.go`, update `setupAuth()`:
   - Create an `scs.SessionManager` with appropriate settings (cookie name, lifetime, SameSite, Secure, HttpOnly).
   - Use `auth.NewSCSSessionStore(manager)` instead of `auth.NewCookieSessionStore(secret)`.
   - Pass the scs `LoadAndSave` middleware into the handler chain.

2. In `api.go`, update `Handler()`:
   - If using scs, wrap the mux with `sessionManager.LoadAndSave()` before auth middleware.
   - Update `authMiddleware` to use the `SessionStore` interface (should already work if the interface is unchanged).

3. Remove the old `CookieSessionStore` type and its tests from `auth.go` / `auth_test.go`.

4. Update the `Deps` struct (from prompt 53) if the session field type changed.

5. Update `api_test.go` and `app_test.go` to use the new session store in test setup.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Change CSRF handling (evaluate separately)
- Change the `Authenticator` interface or `FileUserStore`
- Change login/logout API behavior
