package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAlbumID(t *testing.T) {
	id, err := GenerateAlbumID()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(id, "alb_") {
		t.Errorf("album ID %q should start with alb_", id)
	}
	if len(id) != len("alb_")+idRandomBytes*2 {
		t.Errorf("album ID %q has unexpected length %d", id, len(id))
	}
}

func TestGenerateAssetID(t *testing.T) {
	id, err := GenerateAssetID()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(id, "ast_") {
		t.Errorf("asset ID %q should start with ast_", id)
	}
}

func TestGenerateIDs_Unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := GenerateAlbumID()
		if err != nil {
			t.Fatal(err)
		}
		if ids[id] {
			t.Fatalf("duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestAlbumState_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := &AlbumState{ObjectID: "alb_test123"}

	if err := SaveAlbumState(dir, s); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadAlbumState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil {
		t.Fatal("loaded state is nil")
	}
	if loaded.ObjectID != "alb_test123" {
		t.Errorf("ObjectID = %q, want %q", loaded.ObjectID, "alb_test123")
	}
}

func TestAlbumState_LoadMissing(t *testing.T) {
	dir := t.TempDir()
	s, err := LoadAlbumState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s != nil {
		t.Error("expected nil for missing state")
	}
}

func TestAssetState_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := &AssetState{ObjectID: "ast_test456"}

	if err := SaveAssetState(dir, "photo.jpg", s); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadAssetState(dir, "photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil {
		t.Fatal("loaded state is nil")
	}
	if loaded.ObjectID != "ast_test456" {
		t.Errorf("ObjectID = %q, want %q", loaded.ObjectID, "ast_test456")
	}

	// Verify file location.
	expected := filepath.Join(dir, ".gallery", "assets", "photo.jpg.json")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected file at %s: %v", expected, err)
	}
}

func TestAssetState_LoadMissing(t *testing.T) {
	dir := t.TempDir()
	s, err := LoadAssetState(dir, "nonexistent.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if s != nil {
		t.Error("expected nil for missing state")
	}
}

func TestEnsureAlbumID_CreatesNew(t *testing.T) {
	dir := t.TempDir()
	s, created, err := EnsureAlbumID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Error("expected created=true for new album")
	}
	if !strings.HasPrefix(s.ObjectID, "alb_") {
		t.Errorf("ObjectID = %q, expected alb_ prefix", s.ObjectID)
	}

	// Second call should reuse the ID.
	s2, created2, err := EnsureAlbumID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Error("expected created=false for existing album")
	}
	if s2.ObjectID != s.ObjectID {
		t.Errorf("ID changed: %q -> %q", s.ObjectID, s2.ObjectID)
	}
}

func TestEnsureAssetID_CreatesNew(t *testing.T) {
	dir := t.TempDir()
	s, created, err := EnsureAssetID(dir, "beach.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Error("expected created=true")
	}
	if !strings.HasPrefix(s.ObjectID, "ast_") {
		t.Errorf("ObjectID = %q, expected ast_ prefix", s.ObjectID)
	}

	// Second call reuses.
	s2, created2, err := EnsureAssetID(dir, "beach.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Error("expected created=false")
	}
	if s2.ObjectID != s.ObjectID {
		t.Errorf("ID changed: %q -> %q", s.ObjectID, s2.ObjectID)
	}
}

func TestAtomicWrite_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	if err := atomicWriteJSON(path, map[string]string{"key": "val"}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"key"`) {
		t.Errorf("file content = %q, expected JSON with key", string(data))
	}
}

func TestAtomicWrite_NoTempFileLeftOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	if err := atomicWriteJSON(path, "hello"); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestGalleryDir(t *testing.T) {
	got := GalleryDir("/photos/vacation")
	want := filepath.Join("/photos/vacation", ".gallery")
	if got != want {
		t.Errorf("GalleryDir = %q, want %q", got, want)
	}
}
