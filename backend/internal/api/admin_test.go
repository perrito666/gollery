package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/perrito666/gollery/backend/internal/auth"
	"github.com/perrito666/gollery/backend/internal/domain"
)

func adminServer(t *testing.T) http.Handler {
	t.Helper()
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	srv.SetContentRoot(t.TempDir(), nil)
	srv.SetAdmin(func() error { return nil })
	srv.SetScanErrors([]string{"warn: orphaned file"})

	fa := &fakeAuthenticator{
		users: map[string]*domain.Principal{
			"admin:admin": {Username: "admin", IsAdmin: true},
			"alice:pass":  {Username: "alice"},
		},
	}
	sessions := auth.NewCookieSessionStore("test-secret")
	srv.SetAuth(fa, sessions, "csrf-test-secret", nil)
	return srv.Handler()
}

func TestAdminStatus_AdminOnly(t *testing.T) {
	handler := adminServer(t)

	// Anonymous.
	rr := doRequest(handler, "GET", "/api/v1/admin/status", nil)
	if rr.Code != http.StatusForbidden {
		t.Errorf("anonymous status = %d, want 403", rr.Code)
	}

	// Non-admin.
	cookie, _ := loginAs(t, handler, "alice", "pass")
	req := httptest.NewRequest("GET", "/api/v1/admin/status", nil)
	req.AddCookie(cookie)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusForbidden {
		t.Errorf("non-admin status = %d, want 403", rr2.Code)
	}

	// Admin.
	cookie, _ = loginAs(t, handler, "admin", "admin")
	req2 := httptest.NewRequest("GET", "/api/v1/admin/status", nil)
	req2.AddCookie(cookie)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req2)
	if rr3.Code != http.StatusOK {
		t.Fatalf("admin status = %d, want 200", rr3.Code)
	}

	var resp StatusResponse
	json.NewDecoder(rr3.Body).Decode(&resp)
	if resp.AlbumCount != 3 {
		t.Errorf("album_count = %d, want 3", resp.AlbumCount)
	}
}

func TestAdminReindex(t *testing.T) {
	handler := adminServer(t)
	cookie, csrf := loginAs(t, handler, "admin", "admin")

	req := httptest.NewRequest("POST", "/api/v1/admin/reindex", nil)
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrf)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestAdminDiagnostics(t *testing.T) {
	handler := adminServer(t)
	cookie, _ := loginAs(t, handler, "admin", "admin")

	req := httptest.NewRequest("GET", "/api/v1/admin/diagnostics", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp DiagnosticsResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.ScanErrors) != 1 {
		t.Errorf("scan_errors count = %d, want 1", len(resp.ScanErrors))
	}
}
