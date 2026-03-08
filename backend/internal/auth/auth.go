// Package auth provides authentication abstractions and session management.
package auth

import (
	"context"
	"errors"

	"github.com/perrito666/gollery/backend/internal/domain"
)

// ErrInvalidCredentials is returned when login credentials are invalid.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrSessionNotFound is returned when a session token does not map to a user.
var ErrSessionNotFound = errors.New("session not found")

// Authenticator verifies user credentials and returns a Principal.
// Implementations may use a local user store, LDAP, OAuth, etc.
type Authenticator interface {
	// Authenticate checks the given credentials and returns a Principal
	// on success. Returns ErrInvalidCredentials on failure.
	Authenticate(ctx context.Context, username, password string) (*domain.Principal, error)
}

// SessionStore manages session tokens.
type SessionStore interface {
	// Create creates a new session for the given principal and returns a token.
	Create(ctx context.Context, p *domain.Principal) (token string, err error)

	// Lookup returns the principal associated with the given token.
	// Returns ErrSessionNotFound if the token is invalid or expired.
	Lookup(ctx context.Context, token string) (*domain.Principal, error)

	// Delete removes a session by token (logout).
	Delete(ctx context.Context, token string) error
}

// principalKey is the context key for storing the current principal.
type principalKey struct{}

// WithPrincipal returns a new context carrying the given principal.
func WithPrincipal(ctx context.Context, p *domain.Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFromContext retrieves the principal from the context.
// Returns nil if no principal is set (anonymous request).
func PrincipalFromContext(ctx context.Context) *domain.Principal {
	p, _ := ctx.Value(principalKey{}).(*domain.Principal)
	return p
}
