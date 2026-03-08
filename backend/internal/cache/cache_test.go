package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestThumbPath(t *testing.T) {
	l := NewLayout("/data/cache")
	got := l.ThumbPath("ast_abc", 400)
	want := filepath.Join("/data/cache", "thumbs", "ast_abc_400.jpg")
	if got != want {
		t.Errorf("ThumbPath = %q, want %q", got, want)
	}
}

func TestPreviewPath(t *testing.T) {
	l := NewLayout("/data/cache")
	got := l.PreviewPath("ast_abc", 1600)
	want := filepath.Join("/data/cache", "previews", "ast_abc_1600.jpg")
	if got != want {
		t.Errorf("PreviewPath = %q, want %q", got, want)
	}
}

func TestEnsureDirs(t *testing.T) {
	root := t.TempDir()
	l := NewLayout(root)
	if err := l.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	for _, sub := range []string{"thumbs", "previews"} {
		p := filepath.Join(root, sub)
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("%s not created: %v", sub, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", sub)
		}
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")

	if Exists(path) {
		t.Error("should not exist yet")
	}

	if err := os.WriteFile(path, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	if !Exists(path) {
		t.Error("should exist after write")
	}
}

func TestPurgeOrphans(t *testing.T) {
	root := t.TempDir()
	l := NewLayout(root)
	if err := l.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	// Create cache files: 2 for known asset, 2 for orphan.
	for _, f := range []string{
		"ast_known_400.jpg",
		"ast_known_1600.jpg",
		"ast_orphan_400.jpg",
		"ast_orphan_1600.jpg",
	} {
		if err := os.WriteFile(filepath.Join(l.ThumbDir(), f), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(l.PreviewDir(), f), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	known := map[string]bool{"ast_known": true}
	removed, err := PurgeOrphans(l, known)
	if err != nil {
		t.Fatal(err)
	}

	// 2 orphan thumbs + 2 orphan previews = 4 removed.
	if removed != 4 {
		t.Errorf("removed = %d, want 4", removed)
	}

	// Verify known files remain.
	if !Exists(filepath.Join(l.ThumbDir(), "ast_known_400.jpg")) {
		t.Error("known thumb should still exist")
	}
	if !Exists(filepath.Join(l.PreviewDir(), "ast_known_1600.jpg")) {
		t.Error("known preview should still exist")
	}

	// Verify orphan files are gone.
	if Exists(filepath.Join(l.ThumbDir(), "ast_orphan_400.jpg")) {
		t.Error("orphan thumb should be removed")
	}
}

func TestPurgeOrphans_EmptyCache(t *testing.T) {
	root := t.TempDir()
	l := NewLayout(root)
	// Don't create dirs — should handle missing dirs gracefully.
	removed, err := PurgeOrphans(l, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if removed != 0 {
		t.Errorf("removed = %d, want 0", removed)
	}
}

func TestExtractAssetID(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"ast_abc123_400.jpg", "ast_abc123"},
		{"ast_xyz_1600.jpg", "ast_xyz"},
		{"nounderscores.jpg", ""},
		{"", ""},
	}
	for _, tc := range tests {
		got := extractAssetID(tc.name)
		if got != tc.want {
			t.Errorf("extractAssetID(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestLayout_PathsContainAssetID(t *testing.T) {
	l := NewLayout("/cache")
	if !strings.Contains(l.ThumbPath("ast_xyz", 200), "ast_xyz") {
		t.Error("thumb path should contain asset ID")
	}
	if !strings.Contains(l.PreviewPath("ast_xyz", 800), "ast_xyz") {
		t.Error("preview path should contain asset ID")
	}
}
