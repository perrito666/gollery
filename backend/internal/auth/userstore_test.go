package auth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(h)
}

func writeUsersFile(t *testing.T, entries []UserEntry) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFileUserStore_Authenticate(t *testing.T) {
	path := writeUsersFile(t, []UserEntry{
		{Username: "alice", Password: hashPassword(t, "pass123"), Groups: []string{"editors"}, IsAdmin: false},
		{Username: "admin", Password: hashPassword(t, "admin"), Groups: nil, IsAdmin: true},
	})

	store, err := LoadFileUserStore(path)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Correct credentials.
	p, err := store.Authenticate(ctx, "alice", "pass123")
	if err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if p.Username != "alice" {
		t.Errorf("username = %q", p.Username)
	}
	if len(p.Groups) != 1 || p.Groups[0] != "editors" {
		t.Errorf("groups = %v", p.Groups)
	}
	if p.IsAdmin {
		t.Error("alice should not be admin")
	}

	// Admin user.
	p, err = store.Authenticate(ctx, "admin", "admin")
	if err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if !p.IsAdmin {
		t.Error("admin should have IsAdmin=true")
	}
}

func TestFileUserStore_WrongPassword(t *testing.T) {
	path := writeUsersFile(t, []UserEntry{
		{Username: "alice", Password: hashPassword(t, "correct")},
	})

	store, err := LoadFileUserStore(path)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Authenticate(context.Background(), "alice", "wrong")
	if err != ErrInvalidCredentials {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestFileUserStore_UnknownUser(t *testing.T) {
	path := writeUsersFile(t, []UserEntry{
		{Username: "alice", Password: hashPassword(t, "pass")},
	})

	store, err := LoadFileUserStore(path)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Authenticate(context.Background(), "bob", "pass")
	if err != ErrInvalidCredentials {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoadFileUserStore_MissingFile(t *testing.T) {
	_, err := LoadFileUserStore("/nonexistent/users.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFileUserStore_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")
	if err := os.WriteFile(path, []byte("{bad"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFileUserStore(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
