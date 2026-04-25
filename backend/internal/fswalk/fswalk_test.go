package fswalk

import (
	"os"
	"path/filepath"
	"testing"
)

// writeAlbumJSON writes an album.json file in the given directory.
func writeAlbumJSON(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "album.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// writeFile creates a file with minimal content.
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

func TestScan_EmptyRoot(t *testing.T) {
	root := t.TempDir()
	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Albums) != 0 {
		t.Errorf("expected 0 albums, got %d", len(result.Albums))
	}
}

func TestScan_SingleAlbum(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root"}`)
	writeFile(t, filepath.Join(root, "photo1.jpg"))
	writeFile(t, filepath.Join(root, "photo2.png"))
	writeFile(t, filepath.Join(root, "notes.txt"))

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Albums) != 1 {
		t.Fatalf("expected 1 album, got %d", len(result.Albums))
	}

	album := result.Albums[""]
	if album == nil {
		t.Fatal("root album not found")
	}
	if album.Config.Title != "Root" {
		t.Errorf("title = %q, want %q", album.Config.Title, "Root")
	}
	if len(album.Assets) != 2 {
		t.Errorf("expected 2 assets, got %d", len(album.Assets))
	}
}

func TestScan_NestedAlbums(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root", "access": {"view": "public"}}`)

	sub := filepath.Join(root, "vacation")
	writeAlbumJSON(t, sub, `{"title": "Vacation"}`)
	writeFile(t, filepath.Join(sub, "beach.jpg"))

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Albums) != 2 {
		t.Fatalf("expected 2 albums, got %d", len(result.Albums))
	}

	vacAlbum := result.Albums["vacation"]
	if vacAlbum == nil {
		t.Fatal("vacation album not found")
	}
	if vacAlbum.Config.Title != "Vacation" {
		t.Errorf("title = %q, want %q", vacAlbum.Config.Title, "Vacation")
	}
	// Access should be inherited from root.
	if vacAlbum.Config.Access == nil || vacAlbum.Config.Access.View != "public" {
		t.Error("access.view should be inherited as public")
	}
}

func TestScan_InheritedConfig(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{
		"title": "Root",
		"access": {"view": "authenticated"},
		"analytics": {"enabled": true}
	}`)

	// Subfolder without album.json inherits parent config.
	sub := filepath.Join(root, "inherited")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "img.webp"))

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}

	album := result.Albums["inherited"]
	if album == nil {
		t.Fatal("inherited album not found")
	}
	if album.Config.Access == nil || album.Config.Access.View != "authenticated" {
		t.Error("should inherit access.view from parent")
	}
}

func TestScan_UnpublishedSubtreeIgnored(t *testing.T) {
	root := t.TempDir()

	// No album.json at root. Create subdirs with images but no album.json.
	unpub := filepath.Join(root, "private")
	if err := os.MkdirAll(unpub, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(unpub, "secret.jpg"))

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Albums) != 0 {
		t.Errorf("expected 0 albums (no published subtree), got %d", len(result.Albums))
	}
}

func TestScan_HiddenDirsSkipped(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root"}`)

	// .gallery should be skipped.
	hidden := filepath.Join(root, ".gallery")
	if err := os.MkdirAll(hidden, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(hidden, "state.json"))

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result.Albums[".gallery"]; ok {
		t.Error(".gallery should be skipped")
	}
}

func TestScan_InvalidAlbumJSON(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root"}`)

	sub := filepath.Join(root, "broken")
	writeAlbumJSON(t, sub, `{invalid json}`)

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}

	// Broken config should be recorded as error.
	if len(result.Errors) == 0 {
		t.Error("expected scan errors for invalid album.json")
	}

	// Folder should still be published via inherited config.
	album := result.Albums["broken"]
	if album == nil {
		t.Fatal("broken album should still appear with inherited config")
	}
	if album.Config.Title != "Root" {
		t.Errorf("should inherit title from parent, got %q", album.Config.Title)
	}
}

func TestScan_ChildPaths(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root"}`)

	for _, name := range []string{"a", "b"} {
		sub := filepath.Join(root, name)
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatal(err)
		}
	}

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}

	rootAlbum := result.Albums[""]
	if rootAlbum == nil {
		t.Fatal("root album not found")
	}
	if len(rootAlbum.ChildPaths) != 2 {
		t.Errorf("expected 2 child paths, got %d: %v", len(rootAlbum.ChildPaths), rootAlbum.ChildPaths)
	}
}

func TestScan_GPXFilesDiscovered(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root"}`)
	writeFile(t, filepath.Join(root, "photo.jpg"))
	writeFile(t, filepath.Join(root, "track.gpx"))
	writeFile(t, filepath.Join(root, "route.gpx"))

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}

	album := result.Albums[""]
	if len(album.Assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(album.Assets))
	}
	if len(album.GPXFiles) != 2 {
		t.Errorf("expected 2 GPX files, got %d: %v", len(album.GPXFiles), album.GPXFiles)
	}
	// GPX files should be absolute paths.
	for _, f := range album.GPXFiles {
		if !filepath.IsAbs(f) {
			t.Errorf("GPX path should be absolute: %s", f)
		}
	}
}

func TestScan_ImageExtensions(t *testing.T) {
	root := t.TempDir()
	writeAlbumJSON(t, root, `{"title": "Root"}`)

	exts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp"}
	for _, ext := range exts {
		writeFile(t, filepath.Join(root, "photo"+ext))
	}
	writeFile(t, filepath.Join(root, "readme.txt"))
	writeFile(t, filepath.Join(root, "data.csv"))

	result, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}

	album := result.Albums[""]
	if len(album.Assets) != len(exts) {
		t.Errorf("expected %d assets, got %d", len(exts), len(album.Assets))
	}
}
