package index

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/fswalk"
)

func writeAlbumJSON(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "album.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildSnapshot_SingleAlbum(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "My Gallery"}`)
	writeFile(t, filepath.Join(root, "photo.jpg"))

	scan, err := fswalk.Scan(root)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := BuildSnapshot(root, scan)
	if err != nil {
		t.Fatal(err)
	}

	if len(snap.Albums) != 1 {
		t.Fatalf("expected 1 album, got %d", len(snap.Albums))
	}

	album := snap.Albums[""]
	if album == nil {
		t.Fatal("root album missing")
	}
	if !strings.HasPrefix(album.ID, "alb_") {
		t.Errorf("album ID = %q, want alb_ prefix", album.ID)
	}
	if album.Title != "My Gallery" {
		t.Errorf("title = %q", album.Title)
	}
	if len(album.Assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(album.Assets))
	}
	if !strings.HasPrefix(album.Assets[0].ID, "ast_") {
		t.Errorf("asset ID = %q, want ast_ prefix", album.Assets[0].ID)
	}
}

func TestBuildSnapshot_IDsAreStable(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root"}`)
	writeFile(t, filepath.Join(root, "img.png"))

	scan1, _ := fswalk.Scan(root)
	snap1, err := BuildSnapshot(root, scan1)
	if err != nil {
		t.Fatal(err)
	}

	// Second scan+build should produce the same IDs.
	scan2, _ := fswalk.Scan(root)
	snap2, err := BuildSnapshot(root, scan2)
	if err != nil {
		t.Fatal(err)
	}

	if snap1.Albums[""].ID != snap2.Albums[""].ID {
		t.Error("album ID should be stable across builds")
	}
	if snap1.Albums[""].Assets[0].ID != snap2.Albums[""].Assets[0].ID {
		t.Error("asset ID should be stable across builds")
	}
}

func TestBuildSnapshot_Hierarchy(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root"}`)

	sub := filepath.Join(root, "vacation")
	writeAlbumJSON(t, sub, `{"title": "Vacation"}`)
	writeFile(t, filepath.Join(sub, "beach.jpg"))

	scan, _ := fswalk.Scan(root)
	snap, err := BuildSnapshot(root, scan)
	if err != nil {
		t.Fatal(err)
	}

	if len(snap.Albums) != 2 {
		t.Fatalf("expected 2 albums, got %d", len(snap.Albums))
	}

	rootAlbum := snap.Albums[""]
	if rootAlbum == nil {
		t.Fatal("root album missing")
	}
	if len(rootAlbum.Children) != 1 || rootAlbum.Children[0] != "vacation" {
		t.Errorf("root children = %v", rootAlbum.Children)
	}

	vacAlbum := snap.Albums["vacation"]
	if vacAlbum == nil {
		t.Fatal("vacation album missing")
	}
	if vacAlbum.ParentPath != "" {
		t.Errorf("vacation parent = %q, want empty (root)", vacAlbum.ParentPath)
	}
	if vacAlbum.Title != "Vacation" {
		t.Errorf("title = %q", vacAlbum.Title)
	}
}

func TestBuildSnapshot_EmptyScan(t *testing.T) {
	root := t.TempDir()
	scan := &fswalk.ScanResult{Albums: map[string]*fswalk.ScannedAlbum{}}

	snap, err := BuildSnapshot(root, scan)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Albums) != 0 {
		t.Errorf("expected 0 albums, got %d", len(snap.Albums))
	}
	if snap.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should be set")
	}
}

func TestBuildSnapshot_FromManualScanResult(t *testing.T) {
	root := t.TempDir()

	// Build a ScanResult manually to test without filesystem scanning.
	scan := &fswalk.ScanResult{
		Albums: map[string]*fswalk.ScannedAlbum{
			"": {
				Path:   "",
				Config: &config.AlbumConfig{Title: "Manual"},
			},
		},
	}

	snap, err := BuildSnapshot(root, scan)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Albums[""].Title != "Manual" {
		t.Errorf("title = %q", snap.Albums[""].Title)
	}
}
