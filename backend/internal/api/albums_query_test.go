package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAlbumsByPath_Root(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums?path=", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.ID != "alb_root" {
		t.Errorf("id = %q", resp.ID)
	}
}

func TestAlbumsByPath_SubAlbum(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums?path=vacation", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	var resp AlbumResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Title != "Vacation" {
		t.Errorf("title = %q", resp.Title)
	}
}

func TestAlbumsByPath_NormalizesSlashes(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums?path=/vacation/", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestAlbumsByPath_NotFound(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums?path=nonexistent", nil)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestAlbumsByPath_ACLDeniesAnonymous(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums?path=private", nil)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}
