package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/perrito666/gollery/backend/internal/domain"
)

func TestCookieSessionStore_Lifecycle(t *testing.T) {
	store := NewCookieSessionStore("test-secret")
	ctx := context.Background()
	p := &domain.Principal{Username: "alice", Groups: []string{"editors"}}

	// Create session.
	token, err := store.Create(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Lookup session.
	got, err := store.Lookup(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if got.Username != "alice" {
		t.Errorf("username = %q", got.Username)
	}

	// Delete session.
	if err := store.Delete(ctx, token); err != nil {
		t.Fatal(err)
	}

	// Lookup should fail.
	_, err = store.Lookup(ctx, token)
	if err != ErrSessionNotFound {
		t.Errorf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestCookieSessionStore_LookupInvalidToken(t *testing.T) {
	store := NewCookieSessionStore("test-secret")
	_, err := store.Lookup(context.Background(), "bogus-token")
	if err != ErrSessionNotFound {
		t.Errorf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestCookieSessionStore_UniqueTokens(t *testing.T) {
	store := NewCookieSessionStore("test-secret")
	ctx := context.Background()
	p := &domain.Principal{Username: "alice"}

	t1, _ := store.Create(ctx, p)
	t2, _ := store.Create(ctx, p)
	if t1 == t2 {
		t.Error("expected different tokens for different sessions")
	}
}

func TestCookieSessionStore_SetCookie(t *testing.T) {
	store := NewCookieSessionStore("test-secret")
	w := httptest.NewRecorder()
	store.SetCookie(w, "test-token")

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	c := cookies[0]
	if c.Name != SessionCookieName {
		t.Errorf("name = %q", c.Name)
	}
	if c.Value != "test-token" {
		t.Errorf("value = %q", c.Value)
	}
	if !c.HttpOnly {
		t.Error("expected HttpOnly")
	}
	if !c.Secure {
		t.Error("expected Secure")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v", c.SameSite)
	}
}

func TestCookieSessionStore_ClearCookie(t *testing.T) {
	store := NewCookieSessionStore("test-secret")
	w := httptest.NewRecorder()
	store.ClearCookie(w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", cookies[0].MaxAge)
	}
}

func TestTokenFromRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "my-token"})

	got := TokenFromRequest(req)
	if got != "my-token" {
		t.Errorf("token = %q, want %q", got, "my-token")
	}
}

func TestTokenFromRequest_NoCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	got := TokenFromRequest(req)
	if got != "" {
		t.Errorf("token = %q, want empty", got)
	}
}

func TestCookieSessionStore_DeleteNonexistent(t *testing.T) {
	store := NewCookieSessionStore("test-secret")
	// Delete of nonexistent should not error.
	if err := store.Delete(context.Background(), "nonexistent"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
