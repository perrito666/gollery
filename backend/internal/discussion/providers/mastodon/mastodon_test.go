package mastodon

import (
	"context"
	"errors"
	"testing"
)

type fakePoster struct {
	lastStatus string
	returnID   string
	returnURL  string
	returnErr  error
}

func (f *fakePoster) PostStatus(_ context.Context, _, _, status string) (string, string, error) {
	f.lastStatus = status
	return f.returnID, f.returnURL, f.returnErr
}

func TestProvider_Name(t *testing.T) {
	p := New(Config{})
	if p.Name() != "mastodon" {
		t.Errorf("name = %q", p.Name())
	}
}

func TestProvider_CreateThread(t *testing.T) {
	poster := &fakePoster{
		returnID:  "12345",
		returnURL: "https://mastodon.social/@user/12345",
	}
	p := NewWithPoster(Config{InstanceURL: "https://mastodon.social"}, poster)

	thread, err := p.CreateThread(context.Background(), "My Album", "Check it out")
	if err != nil {
		t.Fatal(err)
	}

	if thread.RemoteID != "12345" {
		t.Errorf("remote_id = %q", thread.RemoteID)
	}
	if thread.URL != "https://mastodon.social/@user/12345" {
		t.Errorf("url = %q", thread.URL)
	}
	if thread.Provider != "mastodon" {
		t.Errorf("provider = %q", thread.Provider)
	}
	if thread.ProviderMeta["instance"] != "https://mastodon.social" {
		t.Error("missing instance in provider_meta")
	}
	if poster.lastStatus != "My Album\n\nCheck it out" {
		t.Errorf("status = %q", poster.lastStatus)
	}
}

func TestProvider_CreateThread_TitleOnly(t *testing.T) {
	poster := &fakePoster{returnID: "1", returnURL: "https://example.com/1"}
	p := NewWithPoster(Config{InstanceURL: "https://example.com"}, poster)

	_, err := p.CreateThread(context.Background(), "Title Only", "")
	if err != nil {
		t.Fatal(err)
	}
	if poster.lastStatus != "Title Only" {
		t.Errorf("status = %q, want just title", poster.lastStatus)
	}
}

func TestProvider_CreateThread_Error(t *testing.T) {
	poster := &fakePoster{returnErr: errors.New("network error")}
	p := NewWithPoster(Config{}, poster)

	_, err := p.CreateThread(context.Background(), "T", "B")
	if err == nil {
		t.Error("expected error")
	}
}
