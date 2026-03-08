package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/perrito666/gollery/backend/internal/domain"
)

const (
	// SessionCookieName is the name of the session cookie.
	SessionCookieName = "gollery_session"

	// sessionTokenBytes is the number of random bytes in a session token.
	sessionTokenBytes = 32
)

// CookieSessionStore implements SessionStore using secure cookies
// with HMAC-signed tokens stored in memory.
type CookieSessionStore struct {
	secret []byte

	mu       sync.RWMutex
	sessions map[string]*sessionData
}

type sessionData struct {
	principal *domain.Principal
	createdAt time.Time
}

// NewCookieSessionStore creates a new session store with the given HMAC secret.
func NewCookieSessionStore(secret string) *CookieSessionStore {
	return &CookieSessionStore{
		secret:   []byte(secret),
		sessions: make(map[string]*sessionData),
	}
}

// Create creates a new session and returns the token.
func (s *CookieSessionStore) Create(_ context.Context, p *domain.Principal) (string, error) {
	raw := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	token := s.sign(raw)

	s.mu.Lock()
	s.sessions[token] = &sessionData{
		principal: p,
		createdAt: time.Now(),
	}
	s.mu.Unlock()

	return token, nil
}

// Lookup returns the principal for a session token.
func (s *CookieSessionStore) Lookup(_ context.Context, token string) (*domain.Principal, error) {
	s.mu.RLock()
	sess, ok := s.sessions[token]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrSessionNotFound
	}
	return sess.principal, nil
}

// Delete removes a session.
func (s *CookieSessionStore) Delete(_ context.Context, token string) error {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
	return nil
}

// SetCookie writes the session cookie to the response.
func (s *CookieSessionStore) SetCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearCookie removes the session cookie.
func (s *CookieSessionStore) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// TokenFromRequest extracts the session token from the request cookie.
func TokenFromRequest(r *http.Request) string {
	c, err := r.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// sign produces an HMAC-signed hex token from raw bytes.
func (s *CookieSessionStore) sign(raw []byte) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write(raw)
	sig := mac.Sum(nil)
	// token = hex(raw) + hex(sig)
	return hex.EncodeToString(raw) + hex.EncodeToString(sig)
}
