package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/domain"
)

func paginationSnapshot() (*domain.Snapshot, map[string]*config.AlbumConfig) {
	assets := make([]domain.Asset, 10)
	for i := range assets {
		assets[i] = domain.Asset{
			ID:        fmt.Sprintf("ast_%d", i),
			Filename:  fmt.Sprintf("img_%02d.jpg", i),
			AlbumPath: "photos",
		}
	}

	snap := &domain.Snapshot{
		GeneratedAt: time.Now(),
		Albums: map[string]*domain.Album{
			"": {
				ID: "alb_root", Path: "", Title: "Root",
				Children: []string{"photos"},
			},
			"photos": {
				ID: "alb_photos", Path: "photos", Title: "Photos",
				Assets: assets,
			},
		},
	}
	configs := map[string]*config.AlbumConfig{
		"":       {Access: &config.AccessConfig{View: "public"}},
		"photos": {Access: &config.AccessConfig{View: "public"}},
	}
	return snap, configs
}

func TestPagination_Default(t *testing.T) {
	snap, configs := paginationSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_photos", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.TotalAssets != 10 {
		t.Errorf("total_assets = %d, want 10", resp.TotalAssets)
	}
	if len(resp.Assets) != 10 {
		t.Errorf("assets len = %d, want 10 (all fit in default limit)", len(resp.Assets))
	}
}

func TestPagination_CustomLimit(t *testing.T) {
	snap, configs := paginationSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_photos?limit=3", nil)
	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.TotalAssets != 10 {
		t.Errorf("total_assets = %d, want 10", resp.TotalAssets)
	}
	if len(resp.Assets) != 3 {
		t.Errorf("assets len = %d, want 3", len(resp.Assets))
	}
	if resp.Assets[0].ID != "ast_0" {
		t.Errorf("first asset = %s, want ast_0", resp.Assets[0].ID)
	}
}

func TestPagination_Offset(t *testing.T) {
	snap, configs := paginationSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_photos?offset=7&limit=5", nil)
	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.TotalAssets != 10 {
		t.Errorf("total_assets = %d, want 10", resp.TotalAssets)
	}
	// Only 3 assets remain after offset 7.
	if len(resp.Assets) != 3 {
		t.Errorf("assets len = %d, want 3", len(resp.Assets))
	}
	if resp.Assets[0].ID != "ast_7" {
		t.Errorf("first asset = %s, want ast_7", resp.Assets[0].ID)
	}
}

func TestPagination_OffsetBeyondTotal(t *testing.T) {
	snap, configs := paginationSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_photos?offset=100", nil)
	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.TotalAssets != 10 {
		t.Errorf("total_assets = %d, want 10", resp.TotalAssets)
	}
	if len(resp.Assets) != 0 {
		t.Errorf("assets len = %d, want 0", len(resp.Assets))
	}
}

func TestPagination_LimitCapped(t *testing.T) {
	snap, configs := paginationSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_photos?limit=9999", nil)
	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	// Should be capped at maxPageLimit but since we only have 10 assets,
	// we get all 10.
	if len(resp.Assets) != 10 {
		t.Errorf("assets len = %d, want 10", len(resp.Assets))
	}
}

func TestPagination_DateSortOrder(t *testing.T) {
	t1 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	snap := &domain.Snapshot{
		GeneratedAt: time.Now(),
		Albums: map[string]*domain.Album{
			"photos": {
				ID: "alb_photos", Path: "photos", Title: "Photos",
				Assets: []domain.Asset{
					{ID: "ast_c", Filename: "charlie.jpg", AlbumPath: "photos", ModTime: t1},
					{ID: "ast_a", Filename: "alpha.jpg", AlbumPath: "photos", ModTime: t2},
					{ID: "ast_b", Filename: "bravo.jpg", AlbumPath: "photos", ModTime: t3},
				},
			},
		},
	}
	configs := map[string]*config.AlbumConfig{
		"photos": {Access: &config.AccessConfig{View: "public"}, SortOrder: "date"},
	}

	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_photos", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	// Date order: alpha(Jan1), bravo(Jan2), charlie(Jan3)
	if len(resp.Assets) != 3 {
		t.Fatalf("assets len = %d, want 3", len(resp.Assets))
	}
	if resp.Assets[0].ID != "ast_a" {
		t.Errorf("first asset = %s, want ast_a (oldest)", resp.Assets[0].ID)
	}
	if resp.Assets[1].ID != "ast_b" {
		t.Errorf("second asset = %s, want ast_b (middle)", resp.Assets[1].ID)
	}
	if resp.Assets[2].ID != "ast_c" {
		t.Errorf("third asset = %s, want ast_c (newest)", resp.Assets[2].ID)
	}
}

func TestPagination_DefaultSortIsFilename(t *testing.T) {
	t1 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	snap := &domain.Snapshot{
		GeneratedAt: time.Now(),
		Albums: map[string]*domain.Album{
			"photos": {
				ID: "alb_photos", Path: "photos", Title: "Photos",
				Assets: []domain.Asset{
					{ID: "ast_c", Filename: "charlie.jpg", AlbumPath: "photos", ModTime: t1},
					{ID: "ast_a", Filename: "alpha.jpg", AlbumPath: "photos", ModTime: t2},
					{ID: "ast_b", Filename: "bravo.jpg", AlbumPath: "photos", ModTime: t3},
				},
			},
		},
	}
	// No sort_order set — should default to filename.
	configs := map[string]*config.AlbumConfig{
		"photos": {Access: &config.AccessConfig{View: "public"}},
	}

	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/albums/alb_photos", nil)
	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	// Filename order: alpha, bravo, charlie
	if resp.Assets[0].ID != "ast_a" {
		t.Errorf("first asset = %s, want ast_a (alpha)", resp.Assets[0].ID)
	}
	if resp.Assets[1].ID != "ast_b" {
		t.Errorf("second asset = %s, want ast_b (bravo)", resp.Assets[1].ID)
	}
	if resp.Assets[2].ID != "ast_c" {
		t.Errorf("third asset = %s, want ast_c (charlie)", resp.Assets[2].ID)
	}
}

func TestPagination_RootAlbum(t *testing.T) {
	snap, configs := paginationSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/albums/root", nil)
	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.TotalAssets != 0 {
		t.Errorf("total_assets = %d, want 0", resp.TotalAssets)
	}
}
