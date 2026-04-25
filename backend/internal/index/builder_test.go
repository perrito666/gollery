package index

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/perrito666/gollery/backend/internal/config"
	"github.com/perrito666/gollery/backend/internal/fswalk"
	"github.com/perrito666/gollery/backend/internal/geo"
	"github.com/perrito666/gollery/backend/internal/state"
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

func float64Ptr(v float64) *float64 { return &v }

func TestResolveCoords_CachedSidecar(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "photo.jpg"))

	as := &state.AssetState{
		ObjectID:    "ast_test",
		Latitude:    float64Ptr(48.8566),
		Longitude:   float64Ptr(2.3522),
		GeoResolved: true,
	}

	lat, lon := resolveCoords(dir, "photo.jpg", as, nil)
	if lat == nil || *lat != 48.8566 {
		t.Errorf("lat = %v, want 48.8566", lat)
	}
	if lon == nil || *lon != 2.3522 {
		t.Errorf("lon = %v, want 2.3522", lon)
	}
}

func TestResolveCoords_GeoResolvedNoCoords(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "photo.jpg"))

	as := &state.AssetState{
		ObjectID:    "ast_test",
		GeoResolved: true,
	}

	lat, lon := resolveCoords(dir, "photo.jpg", as, nil)
	if lat != nil || lon != nil {
		t.Errorf("expected nil coords for resolved-no-coords, got (%v, %v)", lat, lon)
	}
}

func TestResolveCoords_NoExifMarksResolved(t *testing.T) {
	dir := t.TempDir()
	// Write a non-JPEG file (no EXIF possible).
	writeFile(t, filepath.Join(dir, "photo.jpg"))

	as := &state.AssetState{ObjectID: "ast_test"}

	lat, lon := resolveCoords(dir, "photo.jpg", as, nil)
	if lat != nil || lon != nil {
		t.Errorf("expected nil coords, got (%v, %v)", lat, lon)
	}
	if !as.GeoResolved {
		t.Error("GeoResolved should be true after exhausting sources")
	}
}

func TestResolveCoords_GPXMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "photo.jpg"))

	// Simulate: EXIF gave us a DateTaken but no GPS.
	// We can't easily produce a real EXIF file in tests, so test
	// resolveCoords with GPX points directly by pre-populating
	// the asset state as if EXIF extraction yielded nothing.
	// Instead, test the GPX matching function directly.
	pts := []geo.Trackpoint{
		{Lat: 40.0, Lon: -74.0, Time: time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)},
		{Lat: 42.0, Lon: -72.0, Time: time.Date(2024, 6, 15, 10, 2, 0, 0, time.UTC)},
	}

	query := time.Date(2024, 6, 15, 10, 1, 0, 0, time.UTC)
	lat, lon, ok := geo.MatchNearest(pts, query, 30*time.Second)
	if !ok {
		t.Fatal("expected GPX match")
	}
	if lat < 40.9 || lat > 41.1 {
		t.Errorf("lat = %f, want ~41.0", lat)
	}
	if lon < -73.1 || lon > -72.9 {
		t.Errorf("lon = %f, want ~-73.0", lon)
	}
}

func TestBuildSnapshot_AlbumFallbackCoords(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "photo.jpg"))

	lat := 48.8566
	lon := 2.3522
	scan := &fswalk.ScanResult{
		Albums: map[string]*fswalk.ScannedAlbum{
			"": {
				Path: "",
				Config: &config.AlbumConfig{
					Title:     "Paris",
					Latitude:  &lat,
					Longitude: &lon,
				},
				Assets: []fswalk.ScannedAsset{
					{Filename: "photo.jpg", ModTime: time.Now(), SizeBytes: 4},
				},
			},
		},
	}

	snap, err := BuildSnapshot(root, scan)
	if err != nil {
		t.Fatal(err)
	}

	asset := snap.Albums[""].Assets[0]
	if asset.Metadata == nil {
		t.Fatal("expected metadata with album fallback coords")
	}
	if asset.Metadata.Latitude == nil || *asset.Metadata.Latitude != 48.8566 {
		t.Errorf("lat = %v, want 48.8566", asset.Metadata.Latitude)
	}
	if asset.Metadata.Longitude == nil || *asset.Metadata.Longitude != 2.3522 {
		t.Errorf("lon = %v, want 2.3522", asset.Metadata.Longitude)
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
