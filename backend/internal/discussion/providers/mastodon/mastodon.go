// Package mastodon implements the Mastodon discussion provider.
package mastodon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/perrito666/gollery/backend/internal/discussion"
)

// Config holds the Mastodon provider configuration.
type Config struct {
	// InstanceURL is the base URL of the Mastodon instance (e.g. "https://mastodon.social").
	InstanceURL string `json:"instance_url"`

	// AccessToken is the OAuth access token for posting.
	AccessToken string `json:"access_token"`
}

// Poster abstracts the HTTP call to create a Mastodon status.
// This allows testing without a real Mastodon instance.
type Poster interface {
	PostStatus(ctx context.Context, instanceURL, accessToken, status string) (id, url string, err error)
}

// Provider implements discussion.Provider for Mastodon.
type Provider struct {
	cfg    Config
	poster Poster
}

// New creates a Mastodon discussion provider.
func New(cfg Config) *Provider {
	return &Provider{cfg: cfg, poster: &httpPoster{}}
}

// NewWithPoster creates a Mastodon provider with a custom poster (for testing).
func NewWithPoster(cfg Config, p Poster) *Provider {
	return &Provider{cfg: cfg, poster: p}
}

// Name returns "mastodon".
func (p *Provider) Name() string { return "mastodon" }

// CreateThread posts a new status to Mastodon and returns the thread info.
func (p *Provider) CreateThread(ctx context.Context, title, body string) (*discussion.Thread, error) {
	status := title
	if body != "" {
		status = title + "\n\n" + body
	}

	id, statusURL, err := p.poster.PostStatus(ctx, p.cfg.InstanceURL, p.cfg.AccessToken, status)
	if err != nil {
		return nil, fmt.Errorf("mastodon: %w", err)
	}

	return &discussion.Thread{
		Provider: "mastodon",
		RemoteID: id,
		URL:      statusURL,
		ProviderMeta: map[string]string{
			"instance": p.cfg.InstanceURL,
		},
	}, nil
}

// httpPoster is the real HTTP implementation of Poster.
type httpPoster struct{}

func (h *httpPoster) PostStatus(ctx context.Context, instanceURL, accessToken, status string) (string, string, error) {
	endpoint := instanceURL + "/api/v1/statuses"

	form := url.Values{"status": {status}}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decoding response: %w", err)
	}
	return result.ID, result.URL, nil
}
