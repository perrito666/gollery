// Package discussion defines the provider-pluggable discussion abstraction.
package discussion

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/perrito666/gollery/backend/internal/state"
)

// ErrProviderNotFound is returned when a requested provider is not registered.
var ErrProviderNotFound = errors.New("discussion provider not found")

// Thread represents a discussion thread on an external platform.
type Thread struct {
	Provider     string
	RemoteID     string
	URL          string
	ProviderMeta map[string]string
}

// Provider is the interface that discussion backends must implement.
type Provider interface {
	// Name returns the provider identifier (e.g. "mastodon", "bluesky").
	Name() string

	// CreateThread creates a new discussion thread for the given content.
	// Returns the created thread with remote ID and URL.
	CreateThread(ctx context.Context, title, body string) (*Thread, error)
}

// Service manages discussion bindings for gallery objects.
type Service struct {
	providers map[string]Provider
}

// NewService creates a new discussion service with the given providers.
func NewService(providers ...Provider) *Service {
	m := make(map[string]Provider, len(providers))
	for _, p := range providers {
		m[p.Name()] = p
	}
	return &Service{providers: m}
}

// CreateBinding creates a new discussion thread and persists the binding
// in the sidecar state. albumAbsPath is the filesystem path to the album.
// objectType is "album" or "asset". filename is required for assets.
func (s *Service) CreateBinding(ctx context.Context, providerName, albumAbsPath, objectType, filename, title, body, createdBy string) (*state.DiscussionBinding, error) {
	p, ok := s.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, providerName)
	}

	thread, err := p.CreateThread(ctx, title, body)
	if err != nil {
		return nil, fmt.Errorf("creating thread: %w", err)
	}

	binding := state.DiscussionBinding{
		Provider:     providerName,
		RemoteID:     thread.RemoteID,
		URL:          thread.URL,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		CreatedBy:    createdBy,
		ProviderMeta: thread.ProviderMeta,
	}

	switch objectType {
	case "album":
		return &binding, s.addAlbumBinding(albumAbsPath, binding)
	case "asset":
		return &binding, s.addAssetBinding(albumAbsPath, filename, binding)
	default:
		return nil, fmt.Errorf("unknown object type: %s", objectType)
	}
}

// ListBindings returns the discussion bindings for an object.
func (s *Service) ListBindings(albumAbsPath, objectType, filename string) ([]state.DiscussionBinding, error) {
	switch objectType {
	case "album":
		st, err := state.LoadAlbumState(albumAbsPath)
		if err != nil {
			return nil, err
		}
		if st == nil {
			return nil, nil
		}
		return st.Discussions, nil
	case "asset":
		st, err := state.LoadAssetState(albumAbsPath, filename)
		if err != nil {
			return nil, err
		}
		if st == nil {
			return nil, nil
		}
		return st.Discussions, nil
	default:
		return nil, fmt.Errorf("unknown object type: %s", objectType)
	}
}

func (s *Service) addAlbumBinding(albumAbsPath string, binding state.DiscussionBinding) error {
	st, err := state.LoadAlbumState(albumAbsPath)
	if err != nil {
		return err
	}
	if st == nil {
		st = &state.AlbumState{}
	}
	st.Discussions = append(st.Discussions, binding)
	return state.SaveAlbumState(albumAbsPath, st)
}

func (s *Service) addAssetBinding(albumAbsPath, filename string, binding state.DiscussionBinding) error {
	st, err := state.LoadAssetState(albumAbsPath, filename)
	if err != nil {
		return err
	}
	if st == nil {
		st = &state.AssetState{}
	}
	st.Discussions = append(st.Discussions, binding)
	return state.SaveAssetState(albumAbsPath, filename, st)
}
