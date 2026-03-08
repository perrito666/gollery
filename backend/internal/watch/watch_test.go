package watch

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectNewFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "album.json"), `{"title":"root"}`)

	w := New(Config{ContentRoot: root})

	// Initial scan — no dirty paths.
	w.ScanOnce()
	if len(w.DirtyPaths()) != 0 {
		t.Fatalf("expected 0 dirty after initial scan, got %v", w.DirtyPaths())
	}

	// Add a new file.
	writeFile(t, filepath.Join(root, "new.jpg"), "image data")

	dirty := w.DetectChanges()
	if len(dirty) != 1 || dirty[0] != "" {
		t.Errorf("dirty = %v, want [\"\"] (root dir)", dirty)
	}
}

func TestDetectModifiedFile(t *testing.T) {
	root := t.TempDir()
	imgPath := filepath.Join(root, "photo.jpg")
	writeFile(t, imgPath, "v1")

	w := New(Config{ContentRoot: root})
	w.ScanOnce()

	// Modify the file (need different mtime).
	time.Sleep(10 * time.Millisecond)
	writeFile(t, imgPath, "v2 longer content")

	dirty := w.DetectChanges()
	if len(dirty) == 0 {
		t.Error("expected dirty paths after modification")
	}
}

func TestDetectDeletedFile(t *testing.T) {
	root := t.TempDir()
	imgPath := filepath.Join(root, "photo.jpg")
	writeFile(t, imgPath, "data")

	w := New(Config{ContentRoot: root})
	w.ScanOnce()

	os.Remove(imgPath)

	dirty := w.DetectChanges()
	if len(dirty) == 0 {
		t.Error("expected dirty paths after deletion")
	}
}

func TestDetectNewSubdirectory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "album.json"), `{}`)

	w := New(Config{ContentRoot: root})
	w.ScanOnce()

	// Add a new subdirectory with a file.
	sub := filepath.Join(root, "vacation")
	writeFile(t, filepath.Join(sub, "beach.jpg"), "img")

	dirty := w.DetectChanges()
	sort.Strings(dirty)

	found := false
	for _, d := range dirty {
		if d == "vacation" {
			found = true
		}
	}
	if !found {
		t.Errorf("dirty = %v, want vacation in the list", dirty)
	}
}

func TestHiddenDirsSkipped(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "album.json"), `{}`)

	w := New(Config{ContentRoot: root})
	w.ScanOnce()

	// Add a file inside .gallery — should not appear as dirty.
	writeFile(t, filepath.Join(root, ".gallery", "state.json"), "{}")

	dirty := w.DetectChanges()
	for _, d := range dirty {
		if d == ".gallery" {
			t.Error(".gallery should be skipped by watcher")
		}
	}
}

func TestReconcileCallback(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "album.json"), `{}`)

	var mu sync.Mutex
	var reconciled []string

	w := New(Config{
		ContentRoot: root,
		Reconcile: func(_ context.Context, paths []string) error {
			mu.Lock()
			reconciled = append(reconciled, paths...)
			mu.Unlock()
			return nil
		},
	})

	w.ScanOnce()

	// Add a file.
	writeFile(t, filepath.Join(root, "new.jpg"), "data")
	w.ScanOnce()

	// Force reconcile (bypass debounce).
	w.ForceReconcile(context.Background())

	mu.Lock()
	defer mu.Unlock()

	if len(reconciled) == 0 {
		t.Error("reconcile callback should have been called")
	}

	// Dirty paths should be cleared.
	if len(w.DirtyPaths()) != 0 {
		t.Errorf("dirty paths should be empty after reconcile, got %v", w.DirtyPaths())
	}
}

func TestNoReconcileBeforeDebounce(t *testing.T) {
	root := t.TempDir()

	called := false
	w := New(Config{
		ContentRoot:   root,
		DebounceDelay: 1 * time.Hour, // Very long debounce.
		Reconcile: func(_ context.Context, _ []string) error {
			called = true
			return nil
		},
	})

	writeFile(t, filepath.Join(root, "file.txt"), "data")
	w.ScanOnce()

	// Add a dirty path.
	writeFile(t, filepath.Join(root, "new.txt"), "data")
	w.ScanOnce()

	// Try reconcile — should not fire (debounce not elapsed).
	w.maybeReconcile(context.Background())

	if called {
		t.Error("reconcile should not have fired before debounce elapsed")
	}
}

func TestNoDirtyPathsOnCleanDirectory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "stable.jpg"), "data")

	w := New(Config{ContentRoot: root})
	w.ScanOnce() // baseline

	// Scan again — no changes.
	dirty := w.DetectChanges()
	if len(dirty) != 0 {
		t.Errorf("expected 0 dirty paths on stable dir, got %v", dirty)
	}
}
