package derive

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/perrito666/gollery/backend/internal/cache"
)

// createTestPNG creates a solid-color PNG file.
func createTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestFitDimensions(t *testing.T) {
	tests := []struct {
		name           string
		w, h, max      int
		wantW, wantH   int
	}{
		{"landscape", 1000, 500, 400, 400, 200},
		{"portrait", 500, 1000, 400, 200, 400},
		{"square", 800, 800, 400, 400, 400},
		{"already fits", 200, 100, 400, 200, 100},
		{"zero max", 100, 100, 0, 0, 0},
		{"zero input", 0, 100, 400, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := fitDimensions(tt.w, tt.h, tt.max)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Errorf("fitDimensions(%d,%d,%d) = %d,%d; want %d,%d",
					tt.w, tt.h, tt.max, gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestGenerateThumbnail(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.png")
	createTestPNG(t, srcPath, 800, 600)

	cacheDir := filepath.Join(dir, "cache")
	layout := cache.NewLayout(cacheDir)

	outPath, err := GenerateThumbnail(layout, "ast_test", srcPath, 200)
	if err != nil {
		t.Fatal(err)
	}

	if !cache.Exists(outPath) {
		t.Error("thumbnail file should exist")
	}

	// Second call should be a no-op (cached).
	outPath2, err := GenerateThumbnail(layout, "ast_test", srcPath, 200)
	if err != nil {
		t.Fatal(err)
	}
	if outPath2 != outPath {
		t.Error("second call should return same path")
	}
}

func TestGeneratePreview(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.png")
	createTestPNG(t, srcPath, 2000, 1500)

	cacheDir := filepath.Join(dir, "cache")
	layout := cache.NewLayout(cacheDir)

	outPath, err := GeneratePreview(layout, "ast_test", srcPath, 1600)
	if err != nil {
		t.Fatal(err)
	}

	if !cache.Exists(outPath) {
		t.Error("preview file should exist")
	}
}

func TestGenerateThumbnail_MissingSource(t *testing.T) {
	dir := t.TempDir()
	layout := cache.NewLayout(filepath.Join(dir, "cache"))

	_, err := GenerateThumbnail(layout, "ast_x", "/nonexistent.png", 200)
	if err == nil {
		t.Error("expected error for missing source")
	}
}

func TestGeneratedThumbnail_Dimensions(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "wide.png")
	createTestPNG(t, srcPath, 1000, 500)

	layout := cache.NewLayout(filepath.Join(dir, "cache"))
	outPath, err := GenerateThumbnail(layout, "ast_wide", srcPath, 400)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 {
		t.Errorf("width = %d, want 400", bounds.Dx())
	}
	if bounds.Dy() != 200 {
		t.Errorf("height = %d, want 200", bounds.Dy())
	}
}
