package domain

import (
	"testing"
	"time"
)

func TestAlbumZeroValue(t *testing.T) {
	var a Album
	if a.ID != "" {
		t.Error("zero Album should have empty ID")
	}
	if a.Path != "" {
		t.Error("zero Album should have empty Path")
	}
	if len(a.Children) != 0 {
		t.Error("zero Album should have no children")
	}
	if len(a.Assets) != 0 {
		t.Error("zero Album should have no assets")
	}
}

func TestAssetFields(t *testing.T) {
	now := time.Now()
	a := Asset{
		ID:        "ast_01EXAMPLE",
		Filename:  "photo.jpg",
		AlbumPath: "vacation/beach",
		ModTime:   now,
		SizeBytes: 4096,
	}
	if a.ID != "ast_01EXAMPLE" {
		t.Errorf("ID = %q", a.ID)
	}
	if a.SizeBytes != 4096 {
		t.Errorf("SizeBytes = %d", a.SizeBytes)
	}
}

func TestSnapshotAlbumsMap(t *testing.T) {
	s := Snapshot{
		GeneratedAt: time.Now(),
		Albums:      make(map[string]*Album),
	}
	s.Albums["root"] = &Album{ID: "alb_ROOT", Path: ""}
	s.Albums["vacation"] = &Album{ID: "alb_VAC", Path: "vacation"}

	if len(s.Albums) != 2 {
		t.Errorf("albums count = %d, want 2", len(s.Albums))
	}
	if s.Albums["root"].ID != "alb_ROOT" {
		t.Error("root album ID mismatch")
	}
}

func TestPrincipalFields(t *testing.T) {
	p := Principal{
		Username: "horacio",
		Groups:   []string{"family", "admins"},
		IsAdmin:  true,
	}
	if !p.IsAdmin {
		t.Error("expected admin")
	}
	if len(p.Groups) != 2 {
		t.Errorf("groups = %v, want 2 entries", p.Groups)
	}
}

func TestDiscussionBindingFields(t *testing.T) {
	b := DiscussionBinding{
		Provider:  "mastodon",
		RemoteID:  "12345",
		URL:       "https://mastodon.social/@user/12345",
		CreatedBy: "horacio",
		ProviderMeta: map[string]string{
			"instance": "mastodon.social",
		},
	}
	if b.Provider != "mastodon" {
		t.Errorf("provider = %q", b.Provider)
	}
	if b.ProviderMeta["instance"] != "mastodon.social" {
		t.Error("provider_meta missing instance")
	}
}
