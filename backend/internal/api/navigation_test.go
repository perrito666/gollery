package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/domain"
)

func navSnapshot() (*domain.Snapshot, map[string]*config.AlbumConfig) {
	snap := &domain.Snapshot{
		GeneratedAt: time.Now(),
		Albums: map[string]*domain.Album{
			"photos": {
				ID:    "alb_photos",
				Path:  "photos",
				Title: "Photos",
				Assets: []domain.Asset{
					{ID: "ast_b", Filename: "bravo.jpg", AlbumPath: "photos"},
					{ID: "ast_a", Filename: "alpha.jpg", AlbumPath: "photos"},
					{ID: "ast_c", Filename: "charlie.jpg", AlbumPath: "photos"},
				},
			},
			"single": {
				ID:    "alb_single",
				Path:  "single",
				Title: "Single",
				Assets: []domain.Asset{
					{ID: "ast_only", Filename: "only.jpg", AlbumPath: "single"},
				},
			},
		},
	}
	configs := map[string]*config.AlbumConfig{
		"photos": {Title: "Photos", Access: &config.AccessConfig{View: "public"}},
		"single": {Title: "Single", Access: &config.AccessConfig{View: "public"}},
	}
	return snap, configs
}

func navSnapshotDateSorted() (*domain.Snapshot, map[string]*config.AlbumConfig) {
	t1 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) // newest
	t2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // oldest
	t3 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) // middle

	snap := &domain.Snapshot{
		GeneratedAt: time.Now(),
		Albums: map[string]*domain.Album{
			"photos": {
				ID:    "alb_photos",
				Path:  "photos",
				Title: "Photos",
				Assets: []domain.Asset{
					{ID: "ast_b", Filename: "bravo.jpg", AlbumPath: "photos", ModTime: t1},
					{ID: "ast_a", Filename: "alpha.jpg", AlbumPath: "photos", ModTime: t2},
					{ID: "ast_c", Filename: "charlie.jpg", AlbumPath: "photos", ModTime: t3},
				},
			},
		},
	}
	configs := map[string]*config.AlbumConfig{
		"photos": {Title: "Photos", Access: &config.AccessConfig{View: "public"}, SortOrder: "date"},
	}
	return snap, configs
}

func TestAsset_PrevNext_Middle(t *testing.T) {
	snap, configs := navSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	// bravo.jpg is in the middle when sorted: alpha, bravo, charlie
	rr := doRequest(handler, "GET", "/api/v1/assets/ast_b", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp AssetResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.PrevAssetID == nil || *resp.PrevAssetID != "ast_a" {
		t.Errorf("prev = %v, want ast_a", resp.PrevAssetID)
	}
	if resp.NextAssetID == nil || *resp.NextAssetID != "ast_c" {
		t.Errorf("next = %v, want ast_c", resp.NextAssetID)
	}
}

func TestAsset_PrevNext_First(t *testing.T) {
	snap, configs := navSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	// alpha.jpg is first when sorted
	rr := doRequest(handler, "GET", "/api/v1/assets/ast_a", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp AssetResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.PrevAssetID != nil {
		t.Errorf("prev = %v, want nil", *resp.PrevAssetID)
	}
	if resp.NextAssetID == nil || *resp.NextAssetID != "ast_b" {
		t.Errorf("next = %v, want ast_b", resp.NextAssetID)
	}
}

func TestAsset_PrevNext_Last(t *testing.T) {
	snap, configs := navSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	// charlie.jpg is last when sorted
	rr := doRequest(handler, "GET", "/api/v1/assets/ast_c", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp AssetResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.PrevAssetID == nil || *resp.PrevAssetID != "ast_b" {
		t.Errorf("prev = %v, want ast_b", resp.PrevAssetID)
	}
	if resp.NextAssetID != nil {
		t.Errorf("next = %v, want nil", *resp.NextAssetID)
	}
}

func TestAsset_PrevNext_DateSorted(t *testing.T) {
	snap, configs := navSnapshotDateSorted()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	// Date order: alpha(Jan1) -> charlie(Jan2) -> bravo(Jan3)
	// charlie is in the middle when sorted by date
	rr := doRequest(handler, "GET", "/api/v1/assets/ast_c", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp AssetResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.PrevAssetID == nil || *resp.PrevAssetID != "ast_a" {
		t.Errorf("prev = %v, want ast_a (alpha, oldest)", resp.PrevAssetID)
	}
	if resp.NextAssetID == nil || *resp.NextAssetID != "ast_b" {
		t.Errorf("next = %v, want ast_b (bravo, newest)", resp.NextAssetID)
	}
}

func TestAsset_PrevNext_DateSorted_First(t *testing.T) {
	snap, configs := navSnapshotDateSorted()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	// alpha is oldest (first when sorted by date)
	rr := doRequest(handler, "GET", "/api/v1/assets/ast_a", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp AssetResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.PrevAssetID != nil {
		t.Errorf("prev = %v, want nil (first in date order)", *resp.PrevAssetID)
	}
	if resp.NextAssetID == nil || *resp.NextAssetID != "ast_c" {
		t.Errorf("next = %v, want ast_c", resp.NextAssetID)
	}
}

func TestAsset_PrevNext_SingleAsset(t *testing.T) {
	snap, configs := navSnapshot()
	srv := NewServer(snap, configs)
	handler := srv.Handler()

	rr := doRequest(handler, "GET", "/api/v1/assets/ast_only", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp AssetResponse
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.PrevAssetID != nil {
		t.Errorf("prev = %v, want nil", *resp.PrevAssetID)
	}
	if resp.NextAssetID != nil {
		t.Errorf("next = %v, want nil", *resp.NextAssetID)
	}
}
