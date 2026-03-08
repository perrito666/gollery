package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/perrito666/gollery/backend/internal/auth"
	"github.com/perrito666/gollery/backend/internal/domain"
)

func rateLimitedServer() http.Handler {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	fa := &fakeAuthenticator{
		users: map[string]*domain.Principal{
			"alice:pass123": {Username: "alice"},
		},
	}
	sessions := auth.NewCookieSessionStore("test-secret")
	rlCfg := &RateLimitConfig{Rate: 2, Burst: 2}
	srv.SetAuth(fa, sessions, "csrf-test-secret", rlCfg)
	return srv.Handler()
}

func TestRateLimit_LoginThrottled(t *testing.T) {
	handler := rateLimitedServer()

	// First 2 requests should succeed (burst=2).
	for i := 0; i < 2; i++ {
		body := `{"username":"alice","password":"pass123"}`
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
		req.RemoteAddr = "192.0.2.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, rr.Code)
		}
	}

	// Third request should be throttled.
	body := `{"username":"alice","password":"pass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.RemoteAddr = "192.0.2.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusTooManyRequests)
	}

	if rr.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header")
	}
}

func TestRateLimit_DifferentIPsNotThrottled(t *testing.T) {
	handler := rateLimitedServer()

	// Different IPs should each get their own bucket.
	for i := 0; i < 3; i++ {
		body := `{"username":"alice","password":"pass123"}`
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
		req.RemoteAddr = "192.0.2." + string(rune('1'+i)) + ":12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("ip %d: status = %d, want 200", i, rr.Code)
		}
	}
}

func TestRateLimit_GetNotThrottled(t *testing.T) {
	handler := rateLimitedServer()

	// GET requests to non-rate-limited paths should not be throttled.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/v1/albums/root", nil)
		req.RemoteAddr = "192.0.2.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d: unexpected rate limit on GET", i)
		}
	}
}

func TestRateLimit_ResponseFormat(t *testing.T) {
	handler := rateLimitedServer()

	// Exhaust burst.
	for i := 0; i < 2; i++ {
		body := `{"username":"alice","password":"pass123"}`
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
		req.RemoteAddr = "192.0.2.99:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	body := `{"username":"alice","password":"pass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.RemoteAddr = "192.0.2.99:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var apiErr APIError
	if err := json.NewDecoder(rr.Body).Decode(&apiErr); err != nil {
		t.Fatal(err)
	}
	if apiErr.Status != http.StatusTooManyRequests {
		t.Errorf("status = %d", apiErr.Status)
	}
}
