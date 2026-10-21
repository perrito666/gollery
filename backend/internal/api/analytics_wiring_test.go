package api

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/perrito666/gollery/backend/internal/analytics"
	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/domain"
)

// serverWithRecorder returns a test server whose handlers record view events
// into the returned store.
func serverWithRecorder(t *testing.T) (*Server, *mockAnalyticsStore) {
	t.Helper()

	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)
	srv.SetContentRoot(t.TempDir(), nil)

	store := &mockAnalyticsStore{}
	srv.SetAnalyticsRecorder(NewAnalyticsRecorder(
		store,
		func() map[string]*config.AlbumConfig { return cfgs },
		300,
		"test-salt",
	))
	return srv, store
}

func (m *mockAnalyticsStore) eventsCopy() []analytics.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]analytics.Event, len(m.events))
	copy(out, m.events)
	return out
}

// wantSingleEvent asserts the store holds exactly one event of the given type
// for the given object.
func wantSingleEvent(t *testing.T, store *mockAnalyticsStore, typ analytics.EventType, objectID string) {
	t.Helper()

	events := store.eventsCopy()
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1: %+v", len(events), events)
	}
	if events[0].Type != typ {
		t.Errorf("type = %q, want %q", events[0].Type, typ)
	}
	if events[0].ObjectID != objectID {
		t.Errorf("object_id = %q, want %q", events[0].ObjectID, objectID)
	}
	if events[0].VisitorHash == "" {
		t.Error("visitor hash is empty")
	}
}

func TestAlbumsRootRecordsAlbumView(t *testing.T) {
	srv, store := serverWithRecorder(t)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/root", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	wantSingleEvent(t, store, analytics.EventAlbumView, "alb_root")
}

func TestAlbumByIDRecordsAlbumView(t *testing.T) {
	srv, store := serverWithRecorder(t)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_vac", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	wantSingleEvent(t, store, analytics.EventAlbumView, "alb_vac")
}

func TestAssetByIDRecordsAssetView(t *testing.T) {
	srv, store := serverWithRecorder(t)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/assets/ast_1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	wantSingleEvent(t, store, analytics.EventAssetView, "ast_1")
}

func TestAssetOriginalRecordsOriginalHit(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	// handleAssetOriginal serves the file straight from the content root.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.jpg"), []byte("jpeg"), 0o600); err != nil {
		t.Fatal(err)
	}
	srv.SetContentRoot(dir, nil)

	store := &mockAnalyticsStore{}
	srv.SetAnalyticsRecorder(NewAnalyticsRecorder(
		store,
		func() map[string]*config.AlbumConfig { return cfgs },
		300,
		"test-salt",
	))

	rr := doRequest(srv.Handler(), "GET", "/api/v1/assets/ast_1/original", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	wantSingleEvent(t, store, analytics.EventOriginalHit, "ast_1")
}

func TestUnservedRequestsRecordNothing(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"missing album", "/api/v1/albums/alb_nope"},
		{"missing asset", "/api/v1/assets/ast_nope"},
		// "private" is restricted to alice, and doRequest sends no principal.
		{"restricted album", "/api/v1/albums/alb_priv"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, store := serverWithRecorder(t)

			rr := doRequest(srv.Handler(), "GET", tc.path, nil)
			if rr.Code == http.StatusOK {
				t.Fatalf("status = %d, want a failure", rr.Code)
			}
			if n := store.eventCount(); n != 0 {
				t.Errorf("events = %d, want 0: %+v", n, store.eventsCopy())
			}
		})
	}
}

func TestHandlersWorkWithoutRecorder(t *testing.T) {
	snap, cfgs := testSnapshot()
	srv := NewServer(snap, cfgs)

	// No SetAnalyticsRecorder call: recording must be a no-op, not a panic.
	for _, path := range []string{"/api/v1/albums/root", "/api/v1/assets/ast_1"} {
		rr := doRequest(srv.Handler(), "GET", path, nil)
		if rr.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want %d", path, rr.Code, http.StatusOK)
		}
	}
}

// TestRecordingSurvivesSnapshotSwap guards the reason the handlers record
// outside the lock: a reindex replacing the snapshot must not deadlock
// against an in-flight request that is about to record an event.
func TestRecordingSurvivesSnapshotSwap(t *testing.T) {
	srv, store := serverWithRecorder(t)

	rr := doRequest(srv.Handler(), "GET", "/api/v1/albums/root", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	snap, cfgs := testSnapshot()
	snap.Albums["fresh"] = &domain.Album{ID: "alb_fresh", Path: "fresh", Title: "Fresh"}
	cfgs["fresh"] = &config.AlbumConfig{Access: &config.AccessConfig{View: "public"}}
	srv.SetSnapshot(snap, cfgs)

	rr = doRequest(srv.Handler(), "GET", "/api/v1/albums/alb_fresh", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("status after swap = %d, want %d", rr.Code, http.StatusOK)
	}

	events := store.eventsCopy()
	if len(events) != 2 {
		t.Fatalf("events = %d, want 2: %+v", len(events), events)
	}
	if events[1].ObjectID != "alb_fresh" {
		t.Errorf("second object_id = %q, want %q", events[1].ObjectID, "alb_fresh")
	}
}
