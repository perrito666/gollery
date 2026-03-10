package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/domain"
)

func TestShareAssetPublic(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/share/assets/ast_1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}

	// Should contain og:title with filename (no custom title set).
	if !strings.Contains(body, "hello.jpg") {
		t.Error("expected og:title to contain filename hello.jpg")
	}

	// Should contain og:image with preview URL.
	if !strings.Contains(body, "/api/v1/assets/ast_1/preview?size=1200") {
		t.Error("expected og:image with preview URL")
	}

	// Should contain meta refresh redirect.
	if !strings.Contains(body, `http-equiv="refresh"`) {
		t.Error("expected meta refresh tag")
	}

	// Should contain twitter:card summary_large_image (has image).
	if !strings.Contains(body, "summary_large_image") {
		t.Error("expected twitter:card summary_large_image")
	}
}

func TestShareAssetNotFound(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/share/assets/ast_nonexistent", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Page not found") {
		t.Error("expected 'Page not found' in body")
	}
}

func TestShareAssetRestricted(t *testing.T) {
	snap, cfgs := testSnapshot()
	// Add an asset to the private album.
	snap.Albums["private"].Assets = []domain.Asset{
		{ID: "ast_priv", Filename: "secret.jpg", AlbumPath: "private"},
	}
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/share/assets/ast_priv", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Private photo") {
		t.Error("expected 'Private photo' placeholder for restricted asset")
	}
	// Should NOT contain the real filename.
	if strings.Contains(body, "secret.jpg") {
		t.Error("should not leak filename for restricted asset")
	}
}

func TestShareAlbumPublic(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/share/albums/alb_vac", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// Should contain og:title with album title.
	if !strings.Contains(body, "Vacation") {
		t.Error("expected og:title to contain album title 'Vacation'")
	}

	// Should contain og:type website.
	if !strings.Contains(body, `og:type" content="website"`) {
		t.Error("expected og:type website")
	}
}

func TestShareAlbumNotFound(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/share/albums/alb_nonexistent", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestShareAlbumRestricted(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/share/albums/alb_priv", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Private album") {
		t.Error("expected 'Private album' placeholder for restricted album")
	}
}

func TestResolvePublicBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		xProto   string
		expected string
	}{
		{
			name:     "no X-Forwarded-Proto",
			host:     "example.com",
			xProto:   "",
			expected: "http://example.com",
		},
		{
			name:     "with X-Forwarded-Proto https",
			host:     "example.com",
			xProto:   "https",
			expected: "https://example.com",
		},
		{
			name:     "with port",
			host:     "example.com:8080",
			xProto:   "http",
			expected: "http://example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Host = tt.host
			if tt.xProto != "" {
				req.Header.Set("X-Forwarded-Proto", tt.xProto)
			}
			got := resolvePublicBaseURL(req)
			if got != tt.expected {
				t.Errorf("resolvePublicBaseURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestShareAlbumPublicWithImage(t *testing.T) {
	snap, cfgs := testSnapshot()
	// The vacation album already has a public asset (beach.jpg / ast_2).
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/share/albums/alb_vac", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	// Should contain og:image with first asset's thumbnail.
	if !strings.Contains(body, "/api/v1/assets/ast_2/thumbnail?size=1200") {
		t.Error("expected og:image with first public asset thumbnail URL")
	}
}

func TestShareAlbumNoPublicAssets(t *testing.T) {
	snap, cfgs := testSnapshot()
	// Create an album that is public but all its assets are restricted.
	snap.Albums["mixed"] = &domain.Album{
		ID:    "alb_mixed",
		Path:  "mixed",
		Title: "Mixed",
		Assets: []domain.Asset{
			{
				ID:        "ast_restricted",
				Filename:  "restricted.jpg",
				AlbumPath: "mixed",
				Access:    &domain.AccessOverride{View: "restricted"},
			},
		},
	}
	cfgs["mixed"] = &config.AlbumConfig{
		Title:  "Mixed",
		Access: &config.AccessConfig{View: "public"},
	}
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/share/albums/alb_mixed", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	// Should use summary (no image) twitter card.
	if strings.Contains(body, "summary_large_image") {
		t.Error("expected twitter:card summary (no image), got summary_large_image")
	}
	if !strings.Contains(body, `twitter:card" content="summary"`) {
		t.Error("expected twitter:card summary")
	}
}
