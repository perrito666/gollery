// Package analytics defines the optional popularity analytics service.
package analytics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// EventType identifies the kind of analytics event.
type EventType string

const (
	EventAlbumView       EventType = "album_view"
	EventAssetView       EventType = "asset_view"
	EventOriginalHit     EventType = "original_hit"
	EventDiscussionClick EventType = "discussion_click"
)

// Event represents a single analytics event to be recorded.
type Event struct {
	Type        EventType
	ObjectID    string
	VisitorHash string
	CreatedAt   time.Time
}

// PopularitySummary holds aggregate popularity data for an object.
type PopularitySummary struct {
	ObjectID         string `json:"object_id"`
	TotalViews       int64  `json:"total_views"`
	Views7d          int64  `json:"views_7d"`
	Views30d         int64  `json:"views_30d"`
	OriginalHits     int64  `json:"original_hits"`
	DiscussionClicks int64  `json:"discussion_clicks"`
}

// Store defines the interface for analytics persistence.
// Implementations must not be required for gallery correctness.
type Store interface {
	// RecordEvent persists a single analytics event.
	RecordEvent(ctx context.Context, e Event) error

	// QueryPopularity returns the popularity summary for an object.
	QueryPopularity(ctx context.Context, objectID string) (*PopularitySummary, error)

	// AggregateDailyPopularity rolls up raw events into the daily
	// popularity table. Intended to be called periodically.
	AggregateDailyPopularity(ctx context.Context, day time.Time) error

	// PurgeOldEvents deletes events older than the given retention period.
	PurgeOldEvents(ctx context.Context, olderThan time.Time) error

	// Close releases any resources held by the store.
	Close() error
}

// HashVisitorID produces a privacy-safe hash from a raw identifier
// (e.g. IP address) and a server-side salt.
func HashVisitorID(raw, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
