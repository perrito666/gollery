package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/perrito666/gollery/backend/internal/auth"
	"github.com/perrito666/gollery/backend/internal/domain"
)

type fakeAnalyticsStore struct{}

func (f *fakeAnalyticsStore) QueryPopularity(_ context.Context, objectID string) (int64, int64, int64, error) {
	return 100, 10, 40, nil
}

func (f *fakeAnalyticsStore) QueryPopularAssets(_ context.Context, albumID string, limit int) ([]PopularAsset, error) {
	return []PopularAsset{
		{AssetID: "ast_1", TotalViews: 50},
	}, nil
}

func (f *fakeAnalyticsStore) QueryOverview(_ context.Context) (*AnalyticsOverview, error) {
	return &AnalyticsOverview{
		TotalEvents:      1000,
		UniqueVisitors7d: 100,
		TotalViews7d:     500,
		TotalViews30d:    2000,
	}, nil
}

func analyticsServer(t *testing.T) http.Handler {
	t.Helper()
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	srv.SetContentRoot(t.TempDir(), nil)
	srv.SetAnalytics(&fakeAnalyticsStore{})

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

func TestAlbumStats(t *testing.T) {
	handler := analyticsServer(t)

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_root/stats", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp StatsResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.TotalViews != 100 {
		t.Errorf("total_views = %d", resp.TotalViews)
	}
	if resp.Views7d != 10 {
		t.Errorf("views_7d = %d", resp.Views7d)
	}
}

func TestAlbumStats_NotFound(t *testing.T) {
	handler := analyticsServer(t)
	rr := doRequest(handler, "GET", "/api/v1/albums/nonexistent/stats", nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestAlbumStats_ACL(t *testing.T) {
	handler := analyticsServer(t)
	// Private album should deny anonymous.
	rr := doRequest(handler, "GET", "/api/v1/albums/alb_priv/stats", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestPopularAssets(t *testing.T) {
	handler := analyticsServer(t)

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_root/popular-assets", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp []PopularAssetResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Fatalf("count = %d", len(resp))
	}
	if resp[0].ID != "ast_1" {
		t.Errorf("id = %q", resp[0].ID)
	}
}

func TestAnalyticsOverview_AdminOnly(t *testing.T) {
	handler := analyticsServer(t)

	// Anonymous.
	rr := doRequest(handler, "GET", "/api/v1/admin/analytics/overview", nil)
	if rr.Code != http.StatusForbidden {
		t.Errorf("anonymous status = %d, want 403", rr.Code)
	}

	// Admin.
	cookie, _ := loginAs(t, handler, "admin", "admin")
	req := httptest.NewRequest("GET", "/api/v1/admin/analytics/overview", nil)
	req.AddCookie(cookie)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Fatalf("admin status = %d", rr2.Code)
	}

	var resp AnalyticsOverview
	json.NewDecoder(rr2.Body).Decode(&resp)
	if resp.TotalEvents != 1000 {
		t.Errorf("total_events = %d", resp.TotalEvents)
	}
}

func TestPopularityStatus_Enabled(t *testing.T) {
	handler := analyticsServer(t)
	rr := doRequest(handler, "GET", "/api/v1/popularity/status", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var resp map[string]bool
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp["enabled"] {
		t.Error("expected enabled=true")
	}
}

func TestPopularityStatus_Disabled(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	// No analytics store set.
	rr := doRequest(srv.Handler(), "GET", "/api/v1/popularity/status", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var resp map[string]bool
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["enabled"] {
		t.Error("expected enabled=false")
	}
}
