package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/perrito666/gollery/backend/internal/auth"
	"github.com/perrito666/gollery/backend/internal/discussion"
	"github.com/perrito666/gollery/backend/internal/domain"
)

// fakeProvider implements discussion.Provider for tests.
type fakeProvider struct{}

func (f *fakeProvider) Name() string { return "test" }
func (f *fakeProvider) CreateThread(_ context.Context, title, body string) (*discussion.Thread, error) {
	return &discussion.Thread{
		Provider: "test",
		RemoteID: "remote_123",
		URL:      "https://example.com/thread/123",
	}, nil
}

func discussionServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	snap, cfgs := testSnapshot()
	// Add an asset to the vacation album so we can test asset discussions.
	snap.Albums["vacation"].Assets = []domain.Asset{
		{ID: "ast_beach", Filename: "beach.jpg", AlbumPath: "vacation", SizeBytes: 2048},
	}

	srv := NewServer(snap, cfgs)
	srv.SetContentRoot(t.TempDir(), nil)

	discSvc := discussion.NewService(&fakeProvider{})
	srv.SetDiscussions(discSvc)

	fa := &fakeAuthenticator{
		users: map[string]*domain.Principal{
			"admin:admin": {Username: "admin", IsAdmin: true},
			"alice:pass":  {Username: "alice"},
		},
	}
	sessions := auth.NewCookieSessionStore("test-secret")
	srv.SetAuth(fa, sessions, "csrf-test-secret", nil)

	return srv, srv.Handler()
}

func loginAs(t *testing.T, handler http.Handler, username, password string) (*http.Cookie, string) {
	t.Helper()
	body := `{"username":"` + username + `","password":"` + password + `"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == auth.SessionCookieName {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("no session cookie")
	}

	csrfReq := httptest.NewRequest("GET", "/api/v1/auth/csrf-token", nil)
	csrfReq.AddCookie(sessionCookie)
	csrfRR := httptest.NewRecorder()
	handler.ServeHTTP(csrfRR, csrfReq)
	var csrfResp map[string]string
	json.NewDecoder(csrfRR.Body).Decode(&csrfResp)

	return sessionCookie, csrfResp["token"]
}

func TestAlbumDiscussionsList_Empty(t *testing.T) {
	_, handler := discussionServer(t)

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_vac/discussion-threads", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp []DiscussionBindingResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty, got %d", len(resp))
	}
}

func TestAlbumDiscussionsList_NotFound(t *testing.T) {
	_, handler := discussionServer(t)
	rr := doRequest(handler, "GET", "/api/v1/albums/nonexistent/discussion-threads", nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestAlbumDiscussionsCreate_AdminOnly(t *testing.T) {
	_, handler := discussionServer(t)

	// Non-admin should be denied.
	cookie, csrf := loginAs(t, handler, "alice", "pass")
	body := `{"provider":"test","title":"Hello","body":"World"}`
	req := httptest.NewRequest("POST", "/api/v1/albums/alb_vac/discussion-threads", strings.NewReader(body))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("non-admin status = %d, want 403", rr.Code)
	}

	// Admin should succeed.
	cookie, csrf = loginAs(t, handler, "admin", "admin")
	req2 := httptest.NewRequest("POST", "/api/v1/albums/alb_vac/discussion-threads", strings.NewReader(body))
	req2.AddCookie(cookie)
	req2.Header.Set("X-CSRF-Token", csrf)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusCreated {
		t.Errorf("admin status = %d, want 201", rr2.Code)
	}

	var resp DiscussionBindingResponse
	json.NewDecoder(rr2.Body).Decode(&resp)
	if resp.Provider != "test" {
		t.Errorf("provider = %q", resp.Provider)
	}
	if resp.URL != "https://example.com/thread/123" {
		t.Errorf("url = %q", resp.URL)
	}
}

func TestAssetDiscussionsList_Empty(t *testing.T) {
	_, handler := discussionServer(t)

	rr := doRequest(handler, "GET", "/api/v1/assets/ast_beach/discussion-threads", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp []DiscussionBindingResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty, got %d", len(resp))
	}
}

func TestAssetDiscussionsCreate_AdminOnly(t *testing.T) {
	_, handler := discussionServer(t)

	// Non-admin should be denied.
	cookie, csrf := loginAs(t, handler, "alice", "pass")
	body := `{"provider":"test","title":"Beach","body":"Discussion"}`
	req := httptest.NewRequest("POST", "/api/v1/assets/ast_beach/discussion-threads", strings.NewReader(body))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("non-admin status = %d, want 403", rr.Code)
	}

	// Admin should succeed.
	cookie, csrf = loginAs(t, handler, "admin", "admin")
	req2 := httptest.NewRequest("POST", "/api/v1/assets/ast_beach/discussion-threads", strings.NewReader(body))
	req2.AddCookie(cookie)
	req2.Header.Set("X-CSRF-Token", csrf)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusCreated {
		t.Errorf("admin status = %d, want 201", rr2.Code)
	}
}

func TestDiscussions_ACLCheck(t *testing.T) {
	_, handler := discussionServer(t)

	// Private album should deny anonymous.
	rr := doRequest(handler, "GET", "/api/v1/albums/alb_priv/discussion-threads", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAssetDiscussionsCreate_WithURL(t *testing.T) {
	_, handler := discussionServer(t)

	cookie, csrf := loginAs(t, handler, "admin", "admin")

	// Valid URL should create a linked binding (201) and skip provider.
	body := `{"url":"https://mastodon.social/@user/123456"}`
	req := httptest.NewRequest("POST", "/api/v1/assets/ast_beach/discussion-threads", strings.NewReader(body))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rr.Code, rr.Body.String())
	}

	var resp DiscussionBindingResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.URL != "https://mastodon.social/@user/123456" {
		t.Errorf("url = %q, want the linked URL", resp.URL)
	}
	if resp.Provider != "mastodon" {
		t.Errorf("provider = %q, want mastodon (default)", resp.Provider)
	}
	if resp.CreatedBy != "admin" {
		t.Errorf("created_by = %q", resp.CreatedBy)
	}
}

func TestAssetDiscussionsCreate_WithURL_ExplicitProvider(t *testing.T) {
	_, handler := discussionServer(t)
	cookie, csrf := loginAs(t, handler, "admin", "admin")

	body := `{"url":"https://bsky.app/profile/user/post/abc","provider":"bluesky"}`
	req := httptest.NewRequest("POST", "/api/v1/assets/ast_beach/discussion-threads", strings.NewReader(body))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rr.Code)
	}

	var resp DiscussionBindingResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Provider != "bluesky" {
		t.Errorf("provider = %q, want bluesky", resp.Provider)
	}
}

func TestAssetDiscussionsCreate_WithInvalidURL(t *testing.T) {
	_, handler := discussionServer(t)
	cookie, csrf := loginAs(t, handler, "admin", "admin")

	// http:// should be rejected
	body := `{"url":"http://mastodon.social/@user/123456"}`
	req := httptest.NewRequest("POST", "/api/v1/assets/ast_beach/discussion-threads", strings.NewReader(body))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("http:// status = %d, want 400", rr.Code)
	}
}

func TestIsValidDiscussionURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://mastodon.social/@user/123", true},
		{"https://example.com/thread", true},
		{"http://mastodon.social/@user/123", false},
		{"ftp://example.com/file", false},
		{"not-a-url", false},
		{"https://", false},
		{"", false},
		{"javascript:alert(1)", false},
		{"data:text/html,<script>alert(1)</script>", false},
		{"https://example.com/" + strings.Repeat("a", 2048), false},
	}
	for _, tt := range tests {
		got := isValidDiscussionURL(tt.url)
		if got != tt.want {
			t.Errorf("isValidDiscussionURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}
