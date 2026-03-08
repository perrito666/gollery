package meta

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestExtract_NoEXIF(t *testing.T) {
	// Create a plain JPEG with no EXIF data.
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.jpg")

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, nil); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	m, err := Extract(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("metadata should not be nil")
	}
	// No EXIF, so fields should be zero.
	if m.CameraMake != "" {
		t.Errorf("camera make = %q, want empty", m.CameraMake)
	}
	if m.DateTaken != nil {
		t.Errorf("date taken should be nil")
	}
}

func TestExtract_NonExistentFile(t *testing.T) {
	_, err := Extract("/nonexistent/file.jpg")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestExtract_NonJPEG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "text.txt")
	if err := os.WriteFile(path, []byte("not an image"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := Extract(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return empty metadata, not error.
	if m == nil {
		t.Fatal("metadata should not be nil")
	}
}
