package api

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/perrito666/gollery/backend/internal/analytics"
	"github.com/perrito666/gollery/backend/internal/config"
)

type mockAnalyticsStore struct {
	mu     sync.Mutex
	events []analytics.Event
}

func (m *mockAnalyticsStore) RecordEvent(_ context.Context, e analytics.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
	return nil
}

func (m *mockAnalyticsStore) QueryPopularity(_ context.Context, objectID string) (*analytics.PopularitySummary, error) {
	return &analytics.PopularitySummary{}, nil
}

func (m *mockAnalyticsStore) AggregateDailyPopularity(_ context.Context, _ time.Time) error { return nil }
func (m *mockAnalyticsStore) PurgeOldEvents(_ context.Context, _ time.Time) error           { return nil }
func (m *mockAnalyticsStore) Close() error                                                   { return nil }

func (m *mockAnalyticsStore) eventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func TestAnalyticsRecorder_RecordsEvents(t *testing.T) {
	store := &mockAnalyticsStore{}
	configs := map[string]*config.AlbumConfig{
		"": {Analytics: &config.AlbumAnalyticsConfig{Enabled: boolPtr(true)}},
	}
	configsFn := func() map[string]*config.AlbumConfig { return configs }

	ar := NewAnalyticsRecorder(store, configsFn, 1, "salt")

	req := httptest.NewRequest("GET", "/api/v1/albums/root", nil)
	req.RemoteAddr = "192.0.2.1:1234"

	ar.RecordAlbumView(req, "", "alb_root")

	if store.eventCount() != 1 {
		t.Errorf("events = %d, want 1", store.eventCount())
	}
}

func TestAnalyticsRecorder_NilStoreIsNoop(t *testing.T) {
	configsFn := func() map[string]*config.AlbumConfig { return nil }
	ar := NewAnalyticsRecorder(nil, configsFn, 300, "salt")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"

	// Should not panic.
	ar.RecordAlbumView(req, "", "alb_root")
}

func TestAnalyticsRecorder_DedupWindow(t *testing.T) {
	store := &mockAnalyticsStore{}
	configsFn := func() map[string]*config.AlbumConfig { return nil }

	ar := NewAnalyticsRecorder(store, configsFn, 300, "salt")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"

	ar.RecordAlbumView(req, "", "alb_root")
	ar.RecordAlbumView(req, "", "alb_root") // duplicate

	if store.eventCount() != 1 {
		t.Errorf("events = %d, want 1 (dedup)", store.eventCount())
	}

	// Different object should not be deduped.
	ar.RecordAssetView(req, "", "ast_1")
	if store.eventCount() != 2 {
		t.Errorf("events = %d, want 2", store.eventCount())
	}
}

func TestAnalyticsRecorder_DisabledForAlbum(t *testing.T) {
	store := &mockAnalyticsStore{}
	configs := map[string]*config.AlbumConfig{
		"private": {Analytics: &config.AlbumAnalyticsConfig{Enabled: boolPtr(false)}},
	}
	configsFn := func() map[string]*config.AlbumConfig { return configs }

	ar := NewAnalyticsRecorder(store, configsFn, 1, "salt")

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"

	ar.RecordAlbumView(req, "private", "alb_priv")

	if store.eventCount() != 0 {
		t.Errorf("events = %d, want 0 (disabled)", store.eventCount())
	}
}

func boolPtr(v bool) *bool { return &v }
