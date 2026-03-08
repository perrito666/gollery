package bluesky

import (
	"context"
	"errors"
	"testing"
)

type fakePoster struct {
	lastText  string
	returnURI string
	returnCID string
	returnErr error
}

func (f *fakePoster) CreatePost(_ context.Context, _, _, _, text string) (string, string, error) {
	f.lastText = text
	return f.returnURI, f.returnCID, f.returnErr
}

func TestProvider_Name(t *testing.T) {
	p := New(Config{})
	if p.Name() != "bluesky" {
		t.Errorf("name = %q", p.Name())
	}
}

func TestProvider_CreateThread(t *testing.T) {
	poster := &fakePoster{
		returnURI: "at://did:plc:abc/app.bsky.feed.post/xyz123",
		returnCID: "bafyreiabc",
	}
	p := NewWithPoster(Config{
		ServiceURL: "https://bsky.social",
		Handle:     "user.bsky.social",
	}, poster)

	thread, err := p.CreateThread(context.Background(), "Gallery Update", "New photos")
	if err != nil {
		t.Fatal(err)
	}

	if thread.Provider != "bluesky" {
		t.Errorf("provider = %q", thread.Provider)
	}
	if thread.RemoteID != "at://did:plc:abc/app.bsky.feed.post/xyz123" {
		t.Errorf("remote_id = %q", thread.RemoteID)
	}
	if thread.URL != "https://bsky.app/profile/user.bsky.social/post/xyz123" {
		t.Errorf("url = %q", thread.URL)
	}
	if thread.ProviderMeta["cid"] != "bafyreiabc" {
		t.Error("missing cid in provider_meta")
	}
	if thread.ProviderMeta["handle"] != "user.bsky.social" {
		t.Error("missing handle in provider_meta")
	}
	if poster.lastText != "Gallery Update\n\nNew photos" {
		t.Errorf("text = %q", poster.lastText)
	}
}

func TestProvider_CreateThread_TitleOnly(t *testing.T) {
	poster := &fakePoster{returnURI: "at://did/col/key", returnCID: "cid"}
	p := NewWithPoster(Config{Handle: "u.bsky.social"}, poster)

	_, err := p.CreateThread(context.Background(), "Just Title", "")
	if err != nil {
		t.Fatal(err)
	}
	if poster.lastText != "Just Title" {
		t.Errorf("text = %q", poster.lastText)
	}
}

func TestProvider_CreateThread_Error(t *testing.T) {
	poster := &fakePoster{returnErr: errors.New("api error")}
	p := NewWithPoster(Config{}, poster)

	_, err := p.CreateThread(context.Background(), "T", "B")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRKeyFromURI(t *testing.T) {
	tests := []struct {
		uri  string
		want string
	}{
		{"at://did:plc:abc/app.bsky.feed.post/xyz123", "xyz123"},
		{"at://did/col/key", "key"},
		{"noslash", "noslash"},
	}
	for _, tt := range tests {
		got := rKeyFromURI(tt.uri)
		if got != tt.want {
			t.Errorf("rKeyFromURI(%q) = %q, want %q", tt.uri, got, tt.want)
		}
	}
}
