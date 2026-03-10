package discussion

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/perrito666/gollery/backend/internal/state"
)

// fakeProvider is an in-memory discussion provider for testing.
type fakeProvider struct {
	name     string
	nextID   int
	failNext bool
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) CreateThread(_ context.Context, title, body string) (*Thread, error) {
	if f.failNext {
		return nil, errors.New("provider error")
	}
	f.nextID++
	return &Thread{
		Provider:     f.name,
		RemoteID:     fmt.Sprintf("%d", f.nextID),
		URL:          fmt.Sprintf("https://%s.example.com/thread/%d", f.name, f.nextID),
		ProviderMeta: map[string]string{"title": title},
	}, nil
}

func TestCreateBinding_Album(t *testing.T) {
	dir := t.TempDir()
	// Pre-create album state with an ID.
	state.SaveAlbumState(dir, &state.AlbumState{ObjectID: "alb_test"})

	fake := &fakeProvider{name: "fake"}
	svc := NewService(fake)

	binding, err := svc.CreateBinding(context.Background(), "fake", dir, "album", "", "Test Album", "Check this out", "horacio")
	if err != nil {
		t.Fatal(err)
	}

	if binding.Provider != "fake" {
		t.Errorf("provider = %q", binding.Provider)
	}
	if binding.RemoteID != "1" {
		t.Errorf("remote_id = %q", binding.RemoteID)
	}
	if binding.CreatedBy != "horacio" {
		t.Errorf("created_by = %q", binding.CreatedBy)
	}
	if binding.URL == "" {
		t.Error("URL should not be empty")
	}

	// Verify it was persisted.
	bindings, err := svc.ListBindings(dir, "album", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
	if bindings[0].RemoteID != "1" {
		t.Errorf("persisted remote_id = %q", bindings[0].RemoteID)
	}
}

func TestCreateBinding_Asset(t *testing.T) {
	dir := t.TempDir()
	state.SaveAssetState(dir, "photo.jpg", &state.AssetState{ObjectID: "ast_test"})

	fake := &fakeProvider{name: "fake"}
	svc := NewService(fake)

	binding, err := svc.CreateBinding(context.Background(), "fake", dir, "asset", "photo.jpg", "Photo", "Nice photo", "alice")
	if err != nil {
		t.Fatal(err)
	}
	if binding.Provider != "fake" {
		t.Errorf("provider = %q", binding.Provider)
	}

	bindings, err := svc.ListBindings(dir, "asset", "photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
}

func TestCreateBinding_MultipleBindings(t *testing.T) {
	dir := t.TempDir()
	state.SaveAlbumState(dir, &state.AlbumState{ObjectID: "alb_multi"})

	fake := &fakeProvider{name: "fake"}
	svc := NewService(fake)

	for i := 0; i < 3; i++ {
		_, err := svc.CreateBinding(context.Background(), "fake", dir, "album", "", "Title", "Body", "user")
		if err != nil {
			t.Fatal(err)
		}
	}

	bindings, err := svc.ListBindings(dir, "album", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 3 {
		t.Errorf("expected 3 bindings, got %d", len(bindings))
	}
}

func TestCreateBinding_ProviderNotFound(t *testing.T) {
	dir := t.TempDir()
	svc := NewService() // no providers

	_, err := svc.CreateBinding(context.Background(), "nonexistent", dir, "album", "", "T", "B", "u")
	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestCreateBinding_ProviderError(t *testing.T) {
	dir := t.TempDir()
	state.SaveAlbumState(dir, &state.AlbumState{ObjectID: "alb_fail"})

	fake := &fakeProvider{name: "fake", failNext: true}
	svc := NewService(fake)

	_, err := svc.CreateBinding(context.Background(), "fake", dir, "album", "", "T", "B", "u")
	if err == nil {
		t.Error("expected error from provider")
	}
}

func TestLinkBinding_Asset(t *testing.T) {
	dir := t.TempDir()
	state.SaveAssetState(dir, "photo.jpg", &state.AssetState{ObjectID: "ast_link"})

	svc := NewService() // no providers needed for linking

	binding, err := svc.LinkBinding("mastodon", "https://mastodon.social/@user/123", dir, "asset", "photo.jpg", "admin")
	if err != nil {
		t.Fatal(err)
	}

	if binding.Provider != "mastodon" {
		t.Errorf("provider = %q, want mastodon", binding.Provider)
	}
	if binding.URL != "https://mastodon.social/@user/123" {
		t.Errorf("url = %q", binding.URL)
	}
	if binding.RemoteID != "" {
		t.Errorf("remote_id = %q, want empty", binding.RemoteID)
	}
	if binding.CreatedBy != "admin" {
		t.Errorf("created_by = %q", binding.CreatedBy)
	}
	if binding.CreatedAt == "" {
		t.Error("created_at should be set")
	}

	// Verify persistence.
	bindings, err := svc.ListBindings(dir, "asset", "photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
	if bindings[0].URL != "https://mastodon.social/@user/123" {
		t.Errorf("persisted url = %q", bindings[0].URL)
	}
}

func TestLinkBinding_Album(t *testing.T) {
	dir := t.TempDir()
	state.SaveAlbumState(dir, &state.AlbumState{ObjectID: "alb_link"})

	svc := NewService()

	binding, err := svc.LinkBinding("mastodon", "https://mastodon.social/@user/456", dir, "album", "", "editor")
	if err != nil {
		t.Fatal(err)
	}
	if binding.Provider != "mastodon" {
		t.Errorf("provider = %q", binding.Provider)
	}

	bindings, err := svc.ListBindings(dir, "album", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 1 {
		t.Fatalf("expected 1, got %d", len(bindings))
	}
}

func TestListBindings_Empty(t *testing.T) {
	dir := t.TempDir()
	svc := NewService()

	bindings, err := svc.ListBindings(dir, "album", "")
	if err != nil {
		t.Fatal(err)
	}
	if bindings != nil {
		t.Errorf("expected nil for no state, got %v", bindings)
	}
}

func TestListBindings_UnknownType(t *testing.T) {
	dir := t.TempDir()
	svc := NewService()

	_, err := svc.ListBindings(dir, "unknown", "")
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestMultipleProviders(t *testing.T) {
	dir := t.TempDir()
	state.SaveAlbumState(dir, &state.AlbumState{ObjectID: "alb_mp"})

	p1 := &fakeProvider{name: "mastodon"}
	p2 := &fakeProvider{name: "bluesky"}
	svc := NewService(p1, p2)

	_, err := svc.CreateBinding(context.Background(), "mastodon", dir, "album", "", "T", "B", "u")
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateBinding(context.Background(), "bluesky", dir, "album", "", "T", "B", "u")
	if err != nil {
		t.Fatal(err)
	}

	bindings, err := svc.ListBindings(dir, "album", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
	if bindings[0].Provider != "mastodon" {
		t.Errorf("first provider = %q", bindings[0].Provider)
	}
	if bindings[1].Provider != "bluesky" {
		t.Errorf("second provider = %q", bindings[1].Provider)
	}
}
