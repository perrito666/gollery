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

func accessServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	snap, cfgs := testSnapshot()
	snap.Albums["vacation"].Assets = []domain.Asset{
		{ID: "ast_beach", Filename: "beach.jpg", AlbumPath: "vacation", SizeBytes: 2048},
	}

	srv := NewServer(snap, cfgs)
	srv.SetContentRoot(t.TempDir(), nil)

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

func TestAlbumAccess_AdminOnly(t *testing.T) {
	_, handler := accessServer(t)

	// Anonymous.
	rr := doRequest(handler, "GET", "/api/v1/albums/alb_root/access", nil)
	if rr.Code != http.StatusForbidden {
		t.Errorf("anonymous status = %d, want 403", rr.Code)
	}

	// Non-admin.
	cookie, _ := loginAs(t, handler, "alice", "pass")
	req := httptest.NewRequest("GET", "/api/v1/albums/alb_root/access", nil)
	req.AddCookie(cookie)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusForbidden {
		t.Errorf("non-admin status = %d, want 403", rr2.Code)
	}

	// Admin.
	cookie, _ = loginAs(t, handler, "admin", "admin")
	req2 := httptest.NewRequest("GET", "/api/v1/albums/alb_root/access", nil)
	req2.AddCookie(cookie)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req2)
	if rr3.Code != http.StatusOK {
		t.Errorf("admin status = %d, want 200", rr3.Code)
	}

	var resp AccessResponse
	json.NewDecoder(rr3.Body).Decode(&resp)
	if resp.View != "public" {
		t.Errorf("view = %q, want public", resp.View)
	}
}

func TestAssetAccess_AdminOnly(t *testing.T) {
	_, handler := accessServer(t)

	// Admin can view.
	cookie, _ := loginAs(t, handler, "admin", "admin")
	req := httptest.NewRequest("GET", "/api/v1/assets/ast_beach/access", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("admin status = %d, want 200", rr.Code)
	}
}

func TestAssetAccess_NotFound(t *testing.T) {
	_, handler := accessServer(t)

	cookie, _ := loginAs(t, handler, "admin", "admin")
	req := httptest.NewRequest("GET", "/api/v1/assets/nonexistent/access", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestAssetAccessPatch_Validation(t *testing.T) {
	_, handler := accessServer(t)

	cookie, csrf := loginAs(t, handler, "admin", "admin")

	// Invalid access mode.
	body := `{"view":"invalid_mode"}`
	req := httptest.NewRequest("PATCH", "/api/v1/assets/ast_beach/access", strings.NewReader(body))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("invalid mode status = %d, want 400", rr.Code)
	}

	// Invalid JSON.
	req2 := httptest.NewRequest("PATCH", "/api/v1/assets/ast_beach/access", strings.NewReader("{bad"))
	req2.AddCookie(cookie)
	req2.Header.Set("X-CSRF-Token", csrf)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("bad json status = %d, want 400", rr2.Code)
	}
}

func TestAssetAccessPatch_AdminRequired(t *testing.T) {
	_, handler := accessServer(t)

	cookie, csrf := loginAs(t, handler, "alice", "pass")
	body := `{"view":"restricted"}`
	req := httptest.NewRequest("PATCH", "/api/v1/assets/ast_beach/access", strings.NewReader(body))
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("non-admin patch status = %d, want 403", rr.Code)
	}
}
