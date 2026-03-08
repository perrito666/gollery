package analytics

import (
	"testing"
)

func TestHashVisitorID(t *testing.T) {
	hash1 := HashVisitorID("192.168.1.1", "secret-salt")
	hash2 := HashVisitorID("192.168.1.1", "secret-salt")

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if len(hash1) != 16 {
		t.Errorf("hash length = %d, want 16", len(hash1))
	}

	// Different IP should produce different hash.
	hash3 := HashVisitorID("10.0.0.1", "secret-salt")
	if hash1 == hash3 {
		t.Error("different IPs should produce different hashes")
	}

	// Different salt should produce different hash.
	hash4 := HashVisitorID("192.168.1.1", "other-salt")
	if hash1 == hash4 {
		t.Error("different salts should produce different hashes")
	}
}

func TestHashVisitorID_NoRawIP(t *testing.T) {
	hash := HashVisitorID("192.168.1.1", "salt")
	if hash == "192.168.1.1" {
		t.Error("hash should not equal raw IP")
	}
}

func TestEventTypes(t *testing.T) {
	types := []EventType{EventAlbumView, EventAssetView, EventOriginalHit, EventDiscussionClick}
	seen := make(map[EventType]bool)
	for _, et := range types {
		if seen[et] {
			t.Errorf("duplicate event type: %s", et)
		}
		seen[et] = true
		if string(et) == "" {
			t.Error("event type should not be empty")
		}
	}
}
