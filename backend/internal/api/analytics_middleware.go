package api

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/perrito666/gollery/backend/internal/analytics"
	"github.com/perrito666/gollery/backend/internal/config"
)

// AnalyticsRecorder records analytics events for API requests.
type AnalyticsRecorder struct {
	store           analytics.Store
	configs         func() map[string]*config.AlbumConfig
	dedupWindow     time.Duration
	ipSalt          string

	mu    sync.Mutex
	dedup map[string]time.Time // key: "visitorHash:objectID:eventType"
}

// NewAnalyticsRecorder creates a new recorder. store may be nil (no-op).
func NewAnalyticsRecorder(store analytics.Store, configsFn func() map[string]*config.AlbumConfig, dedupWindowSecs int, ipSalt string) *AnalyticsRecorder {
	if dedupWindowSecs <= 0 {
		dedupWindowSecs = 300
	}
	return &AnalyticsRecorder{
		store:       store,
		configs:     configsFn,
		dedupWindow: time.Duration(dedupWindowSecs) * time.Second,
		ipSalt:      ipSalt,
		dedup:       make(map[string]time.Time),
	}
}

// RecordAlbumView records an album view event.
func (ar *AnalyticsRecorder) RecordAlbumView(r *http.Request, albumPath, albumID string) {
	ar.record(r, albumPath, albumID, analytics.EventAlbumView)
}

// RecordAssetView records an asset view event.
func (ar *AnalyticsRecorder) RecordAssetView(r *http.Request, albumPath, assetID string) {
	ar.record(r, albumPath, assetID, analytics.EventAssetView)
}

// RecordOriginalHit records an original download event.
func (ar *AnalyticsRecorder) RecordOriginalHit(r *http.Request, albumPath, assetID string) {
	ar.record(r, albumPath, assetID, analytics.EventOriginalHit)
}

func (ar *AnalyticsRecorder) record(r *http.Request, albumPath, objectID string, eventType analytics.EventType) {
	if ar.store == nil {
		return
	}

	// Check if analytics is enabled for this album.
	if !ar.isEnabledForAlbum(albumPath) {
		return
	}

	visitorHash := analytics.HashVisitorID(extractIP(r), ar.ipSalt)

	// Dedup check.
	dedupKey := visitorHash + ":" + objectID + ":" + string(eventType)
	if ar.isDuplicate(dedupKey) {
		return
	}

	event := analytics.Event{
		Type:        eventType,
		ObjectID:    objectID,
		VisitorHash: visitorHash,
		CreatedAt:   time.Now(),
	}

	if err := ar.store.RecordEvent(r.Context(), event); err != nil {
		slog.Error("recording analytics event", "event_type", eventType, "object_id", objectID, "error", err)
	}
}

func (ar *AnalyticsRecorder) isEnabledForAlbum(albumPath string) bool {
	configs := ar.configs()
	if configs == nil {
		return true // default: enabled
	}

	// Walk up the path to find the most specific config with analytics setting.
	path := albumPath
	for {
		if cfg, ok := configs[path]; ok && cfg.Analytics != nil && cfg.Analytics.Enabled != nil {
			return *cfg.Analytics.Enabled
		}
		if path == "" {
			break
		}
		idx := strings.LastIndex(path, "/")
		if idx < 0 {
			path = ""
		} else {
			path = path[:idx]
		}
	}
	return true // default: enabled
}

func (ar *AnalyticsRecorder) isDuplicate(key string) bool {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	now := time.Now()

	// Clean old entries periodically.
	if len(ar.dedup) > 10000 {
		for k, t := range ar.dedup {
			if now.Sub(t) > ar.dedupWindow {
				delete(ar.dedup, k)
			}
		}
	}

	if t, ok := ar.dedup[key]; ok && now.Sub(t) < ar.dedupWindow {
		return true
	}
	ar.dedup[key] = now
	return false
}

func extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
