// Package bluesky implements the Bluesky discussion provider.
package bluesky

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/perrito666/gollery/backend/internal/discussion"
)

// Config holds the Bluesky provider configuration.
type Config struct {
	// ServiceURL is the Bluesky PDS URL (e.g. "https://bsky.social").
	ServiceURL string `json:"service_url"`

	// Handle is the Bluesky handle (e.g. "user.bsky.social").
	Handle string `json:"handle"`

	// AppPassword is the app-specific password for API access.
	AppPassword string `json:"app_password"`
}

// Poster abstracts the HTTP calls to create a Bluesky post.
type Poster interface {
	CreatePost(ctx context.Context, serviceURL, handle, appPassword, text string) (uri, cid string, err error)
}

// Provider implements discussion.Provider for Bluesky.
type Provider struct {
	cfg    Config
	poster Poster
}

// New creates a Bluesky discussion provider.
func New(cfg Config) *Provider {
	return &Provider{cfg: cfg, poster: &httpPoster{}}
}

// NewWithPoster creates a Bluesky provider with a custom poster (for testing).
func NewWithPoster(cfg Config, p Poster) *Provider {
	return &Provider{cfg: cfg, poster: p}
}

// Name returns "bluesky".
func (p *Provider) Name() string { return "bluesky" }

// CreateThread creates a new post on Bluesky and returns the thread info.
func (p *Provider) CreateThread(ctx context.Context, title, body string) (*discussion.Thread, error) {
	text := title
	if body != "" {
		text = title + "\n\n" + body
	}

	uri, cid, err := p.poster.CreatePost(ctx, p.cfg.ServiceURL, p.cfg.Handle, p.cfg.AppPassword, text)
	if err != nil {
		return nil, fmt.Errorf("bluesky: %w", err)
	}

	// Build a web URL from the handle and the rkey portion of the AT URI.
	webURL := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", p.cfg.Handle, rKeyFromURI(uri))

	return &discussion.Thread{
		Provider: "bluesky",
		RemoteID: uri,
		URL:      webURL,
		ProviderMeta: map[string]string{
			"cid":    cid,
			"handle": p.cfg.Handle,
		},
	}, nil
}

// rKeyFromURI extracts the record key from an AT URI like "at://did:plc:xxx/app.bsky.feed.post/abc123".
func rKeyFromURI(uri string) string {
	// Find the last '/' in the URI.
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' {
			return uri[i+1:]
		}
	}
	return uri
}

// httpPoster is the real HTTP implementation of Poster.
type httpPoster struct{}

func (h *httpPoster) CreatePost(ctx context.Context, serviceURL, handle, appPassword, text string) (string, string, error) {
	// Step 1: Create session.
	sessionURL := serviceURL + "/xrpc/com.atproto.server.createSession"
	sessionBody, _ := json.Marshal(map[string]string{
		"identifier": handle,
		"password":   appPassword,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", sessionURL, bytes.NewReader(sessionBody))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("session creation failed: %d: %s", resp.StatusCode, string(body))
	}

	var session struct {
		DID         string `json:"did"`
		AccessJwt   string `json:"accessJwt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return "", "", fmt.Errorf("decoding session: %w", err)
	}

	// Step 2: Create post.
	postURL := serviceURL + "/xrpc/com.atproto.repo.createRecord"
	record := map[string]any{
		"repo":       session.DID,
		"collection": "app.bsky.feed.post",
		"record": map[string]any{
			"$type":     "app.bsky.feed.post",
			"text":      text,
			"createdAt": "now",
		},
	}
	postBody, _ := json.Marshal(record)

	req2, err := http.NewRequestWithContext(ctx, "POST", postURL, bytes.NewReader(postBody))
	if err != nil {
		return "", "", err
	}
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+session.AccessJwt)

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return "", "", err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		return "", "", fmt.Errorf("post creation failed: %d: %s", resp2.StatusCode, string(body))
	}

	var result struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decoding post: %w", err)
	}
	return result.URI, result.CID, nil
}
