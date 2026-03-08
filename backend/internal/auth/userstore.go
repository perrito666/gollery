package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/perrito666/gollery/backend/internal/domain"
)

// UserEntry represents a user in the users.json file.
type UserEntry struct {
	Username string   `json:"username"`
	Password string   `json:"password"` // bcrypt hash
	Groups   []string `json:"groups"`
	IsAdmin  bool     `json:"is_admin"`
}

// FileUserStore implements Authenticator backed by a JSON file.
type FileUserStore struct {
	users map[string]UserEntry
}

// LoadFileUserStore reads a users.json file and returns a FileUserStore.
func LoadFileUserStore(path string) (*FileUserStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading users file: %w", err)
	}

	var entries []UserEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing users file: %w", err)
	}

	store := &FileUserStore{
		users: make(map[string]UserEntry, len(entries)),
	}
	for _, e := range entries {
		if e.Username == "" {
			continue
		}
		store.users[e.Username] = e
	}
	return store, nil
}

// Authenticate verifies credentials against the file store.
func (s *FileUserStore) Authenticate(_ context.Context, username, password string) (*domain.Principal, error) {
	entry, ok := s.users[username]
	if !ok {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(entry.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return &domain.Principal{
		Username: entry.Username,
		Groups:   entry.Groups,
		IsAdmin:  entry.IsAdmin,
	}, nil
}
