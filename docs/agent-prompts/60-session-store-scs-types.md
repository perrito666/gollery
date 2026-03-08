# Prompt 60 — Add scs session store adapter

Add `alexedwards/scs/v2` and create an adapter that implements the existing `auth.SessionStore` interface.

**Only execute this prompt if prompt 59 recommended proceeding with scs.**

## Implement

1. Add dependency: `go get github.com/alexedwards/scs/v2`

2. Create `backend/internal/auth/scs_store.go`:
   - Define `SCSSessionStore` struct wrapping `*scs.SessionManager`.
   - Implement the `SessionStore` interface (`Create`, `Lookup`, `Delete`).
   - `Create`: put principal data into scs session, return the token.
   - `Lookup`: load session by token, extract principal.
   - `Delete`: destroy session.

3. Create `backend/internal/auth/scs_store_test.go`:
   - Test Create/Lookup/Delete round-trip.
   - Test Lookup with invalid token returns `ErrSessionNotFound`.
   - Test Delete of non-existent session is not an error.

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Remove the existing `CookieSessionStore` yet
- Change `app.go` wiring yet
- Change the CSRF implementation yet
- Modify any API handlers
