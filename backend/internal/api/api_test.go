package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/perrito666/gollery/backend/internal/auth"
	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/domain"
)

func testSnapshot() (*domain.Snapshot, map[string]*config.AlbumConfig) {
	snap := &domain.Snapshot{
		GeneratedAt: time.Now(),
		Albums: map[string]*domain.Album{
			"": {
				ID:       "alb_root",
				Path:     "",
				Title:    "Root",
				Children: []string{"vacation"},
				Assets: []domain.Asset{
					{ID: "ast_1", Filename: "hello.jpg", AlbumPath: "", SizeBytes: 1024},
				},
			},
			"vacation": {
				ID:         "alb_vac",
				Path:       "vacation",
				Title:      "Vacation",
				ParentPath: "",
				Assets: []domain.Asset{
					{ID: "ast_2", Filename: "beach.jpg", AlbumPath: "vacation", SizeBytes: 2048},
				},
			},
			"private": {
				ID:    "alb_priv",
				Path:  "private",
				Title: "Private",
			},
		},
	}
	configs := map[string]*config.AlbumConfig{
		"":         {Title: "Root", Access: &config.AccessConfig{View: "public"}},
		"vacation": {Title: "Vacation", Access: &config.AccessConfig{View: "public"}},
		"private": {
			Title: "Private",
			Access: &config.AccessConfig{
				View:         "restricted",
				AllowedUsers: []string{"alice"},
			},
		},
	}
	return snap, configs
}

func doRequest(handler http.Handler, method, path string, principal *domain.Principal) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if principal != nil {
		req = req.WithContext(auth.WithPrincipal(req.Context(), principal))
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestGetAlbumsRoot(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/root", nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp AlbumResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.ID != "alb_root" {
		t.Errorf("id = %q", resp.ID)
	}
	if resp.Title != "Root" {
		t.Errorf("title = %q", resp.Title)
	}
	if len(resp.Children) != 1 {
		t.Errorf("children = %v", resp.Children)
	}
	if len(resp.Assets) != 1 {
		t.Errorf("assets = %v", resp.Assets)
	}
}

func TestGetAlbumByID(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_vac", nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Title != "Vacation" {
		t.Errorf("title = %q", resp.Title)
	}
}

func TestGetAlbumByID_NotFound(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_nonexistent", nil)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetAssetByID(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	rr := doRequest(srv.Handler(), "GET", "/api/v1/assets/ast_2", nil)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp AssetResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Filename != "beach.jpg" {
		t.Errorf("filename = %q", resp.Filename)
	}
	if resp.AlbumID != "alb_vac" {
		t.Errorf("album_id = %q", resp.AlbumID)
	}
}

func TestGetAssetByID_NotFound(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	rr := doRequest(srv.Handler(), "GET", "/api/v1/assets/ast_missing", nil)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestACL_RestrictedDeniesAnonymous(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_priv", nil)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestACL_RestrictedDeniesUnauthorizedUser(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	p := &domain.Principal{Username: "eve"}
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_priv", p)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestACL_RestrictedAllowsAuthorizedUser(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	p := &domain.Principal{Username: "alice"}
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_priv", p)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestACL_GlobalAdminOverrides(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	p := &domain.Principal{Username: "superadmin", IsAdmin: true}
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_priv", p)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestACL_AssetInheritsAlbumACL(t *testing.T) {
	snap, cfgs := testSnapshot()
	// Add an asset to the private album.
	snap.Albums["private"].Assets = []domain.Asset{
		{ID: "ast_priv", Filename: "secret.jpg", AlbumPath: "private"},
	}
	srv := NewServer(snap, cfgs)

	// Anonymous should be denied.
	rr := doRequest(srv.Handler(), "GET", "/api/v1/assets/ast_priv", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	// Allowed user should succeed.
	p := &domain.Principal{Username: "alice"}
	rr = doRequest(srv.Handler(), "GET", "/api/v1/assets/ast_priv", p)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestErrorResponseFormat(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/nope", nil)

	var apiErr APIError
	if err := json.NewDecoder(rr.Body).Decode(&apiErr); err != nil {
		t.Fatal(err)
	}
	if apiErr.Status != http.StatusNotFound {
		t.Errorf("status = %d", apiErr.Status)
	}
	if apiErr.Error != "Not Found" {
		t.Errorf("error = %q", apiErr.Error)
	}
}

// fakeAuthenticator implements auth.Authenticator for tests.
type fakeAuthenticator struct {
	users map[string]*domain.Principal // key: "user:pass"
}

func (f *fakeAuthenticator) Authenticate(_ context.Context, username, password string) (*domain.Principal, error) {
	key := username + ":" + password
	if p, ok := f.users[key]; ok {
		return p, nil
	}
	return nil, auth.ErrInvalidCredentials
}

func authServer() (*Server, http.Handler) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	fa := &fakeAuthenticator{
		users: map[string]*domain.Principal{
			"alice:pass123": {Username: "alice", Groups: []string{"editors"}},
			"admin:admin":   {Username: "admin", IsAdmin: true},
		},
	}
	sessions := auth.NewCookieSessionStore("test-secret")
	srv.SetAuth(fa, sessions, "csrf-test-secret", nil)
	return srv, srv.Handler()
}

func TestLogin_Success(t *testing.T) {
	_, handler := authServer()

	body := `{"username":"alice","password":"pass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp MeResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Username != "alice" {
		t.Errorf("username = %q", resp.Username)
	}

	// Should have set a cookie.
	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == auth.SessionCookieName && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected session cookie")
	}
}

func TestLogin_BadCredentials(t *testing.T) {
	_, handler := authServer()

	body := `{"username":"alice","password":"wrong"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestLogin_EmptyFields(t *testing.T) {
	_, handler := authServer()

	body := `{"username":"","password":""}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestMe_Authenticated(t *testing.T) {
	_, handler := authServer()

	// Login first.
	body := `{"username":"alice","password":"pass123"}`
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	loginRR := httptest.NewRecorder()
	handler.ServeHTTP(loginRR, loginReq)

	// Extract session cookie.
	var sessionCookie *http.Cookie
	for _, c := range loginRR.Result().Cookies() {
		if c.Name == auth.SessionCookieName {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("no session cookie from login")
	}

	// Call /me with cookie.
	meReq := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	meRR := httptest.NewRecorder()
	handler.ServeHTTP(meRR, meReq)

	if meRR.Code != http.StatusOK {
		t.Fatalf("status = %d", meRR.Code)
	}

	var resp MeResponse
	json.NewDecoder(meRR.Body).Decode(&resp)
	if resp.Username != "alice" {
		t.Errorf("username = %q", resp.Username)
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	_, handler := authServer()

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// loginAndGetSession is a helper that logs in and returns the session cookie and CSRF token.
func loginAndGetSession(t *testing.T, handler http.Handler) (*http.Cookie, string) {
	t.Helper()
	body := `{"username":"alice","password":"pass123"}`
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	loginRR := httptest.NewRecorder()
	handler.ServeHTTP(loginRR, loginReq)

	var sessionCookie *http.Cookie
	for _, c := range loginRR.Result().Cookies() {
		if c.Name == auth.SessionCookieName {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("no session cookie from login")
	}

	// Get CSRF token.
	csrfReq := httptest.NewRequest("GET", "/api/v1/auth/csrf-token", nil)
	csrfReq.AddCookie(sessionCookie)
	csrfRR := httptest.NewRecorder()
	handler.ServeHTTP(csrfRR, csrfReq)

	var csrfResp map[string]string
	json.NewDecoder(csrfRR.Body).Decode(&csrfResp)
	return sessionCookie, csrfResp["token"]
}

func TestLogout(t *testing.T) {
	_, handler := authServer()
	sessionCookie, csrfToken := loginAndGetSession(t, handler)

	// Logout with CSRF token.
	logoutReq := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutReq.Header.Set("X-CSRF-Token", csrfToken)
	logoutRR := httptest.NewRecorder()
	handler.ServeHTTP(logoutRR, logoutReq)

	if logoutRR.Code != http.StatusNoContent {
		t.Errorf("logout status = %d, want %d", logoutRR.Code, http.StatusNoContent)
	}

	// /me should now fail.
	meReq := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	meRR := httptest.NewRecorder()
	handler.ServeHTTP(meRR, meReq)

	if meRR.Code != http.StatusUnauthorized {
		t.Errorf("me after logout status = %d, want %d", meRR.Code, http.StatusUnauthorized)
	}
}

func TestCSRF_PostWithoutToken(t *testing.T) {
	_, handler := authServer()
	sessionCookie, _ := loginAndGetSession(t, handler)

	// POST logout without CSRF token should be rejected.
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestCSRF_PostWithInvalidToken(t *testing.T) {
	_, handler := authServer()
	sessionCookie, _ := loginAndGetSession(t, handler)

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.AddCookie(sessionCookie)
	req.Header.Set("X-CSRF-Token", "invalid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestCSRF_LoginExempt(t *testing.T) {
	_, handler := authServer()

	// Login should succeed without CSRF token.
	body := `{"username":"alice","password":"pass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("login status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestCSRF_GetCSRFToken(t *testing.T) {
	_, handler := authServer()
	sessionCookie, csrfToken := loginAndGetSession(t, handler)

	if csrfToken == "" {
		t.Fatal("expected non-empty CSRF token")
	}

	// Unauthenticated should fail.
	req := httptest.NewRequest("GET", "/api/v1/auth/csrf-token", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated csrf-token status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	// Authenticated should succeed.
	req2 := httptest.NewRequest("GET", "/api/v1/auth/csrf-token", nil)
	req2.AddCookie(sessionCookie)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("authenticated csrf-token status = %d, want %d", rr2.Code, http.StatusOK)
	}
}

func TestMutationAuth_AnonymousRejected(t *testing.T) {
	_, handler := authServer()

	// Anonymous POST (not login) should get 401, not CSRF error.
	req := httptest.NewRequest("POST", "/api/v1/admin/reindex", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("anonymous POST status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	// Anonymous PATCH should get 401.
	req2 := httptest.NewRequest("PATCH", "/api/v1/assets/ast_sunset/metadata", strings.NewReader(`{"title":"x"}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusUnauthorized {
		t.Errorf("anonymous PATCH status = %d, want %d", rr2.Code, http.StatusUnauthorized)
	}

	// Anonymous GET should still work (not blocked by mutation middleware).
	rr3 := doRequest(handler, "GET", "/api/v1/albums/root", nil)
	if rr3.Code != http.StatusOK {
		t.Errorf("anonymous GET status = %d, want %d", rr3.Code, http.StatusOK)
	}

	// Login (POST) should be exempt from mutation auth.
	body := `{"username":"alice","password":"pass123"}`
	req4 := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req4)

	if rr4.Code != http.StatusOK {
		t.Errorf("login status = %d, want %d", rr4.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_SetsContext(t *testing.T) {
	_, handler := authServer()

	// Login as alice.
	body := `{"username":"alice","password":"pass123"}`
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	loginRR := httptest.NewRecorder()
	handler.ServeHTTP(loginRR, loginReq)

	var sessionCookie *http.Cookie
	for _, c := range loginRR.Result().Cookies() {
		if c.Name == auth.SessionCookieName {
			sessionCookie = c
		}
	}

	// Access private album (requires alice).
	req := httptest.NewRequest("GET", "/api/v1/albums/alb_priv", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (alice should access private)", rr.Code, http.StatusOK)
	}
}

func TestSetSnapshotUpdatesIndex(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	// Initially alb_vac exists.
	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_vac", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d before update", rr.Code)
	}

	// Replace with a snapshot missing the vacation album.
	snap2 := &domain.Snapshot{
		GeneratedAt: time.Now(),
		Albums: map[string]*domain.Album{
			"": {ID: "alb_root", Path: "", Title: "Root Only"},
		},
	}
	cfgs2 := map[string]*config.AlbumConfig{
		"": {Title: "Root Only", Access: &config.AccessConfig{View: "public"}},
	}
	srv.SetSnapshot(snap2, cfgs2)

	rr = doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_vac", nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d after update, want 404", rr.Code)
	}
}
