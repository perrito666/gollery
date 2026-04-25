package geo

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const sampleGPX = `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1">
  <trk>
    <trkseg>
      <trkpt lat="48.8566" lon="2.3522">
        <time>2024-06-15T10:00:00Z</time>
      </trkpt>
      <trkpt lat="48.8570" lon="2.3530">
        <time>2024-06-15T10:01:00Z</time>
      </trkpt>
      <trkpt lat="48.8580" lon="2.3540">
        <time>2024-06-15T10:02:00Z</time>
      </trkpt>
    </trkseg>
  </trk>
</gpx>`

func writeGPX(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseGPXFiles_Basic(t *testing.T) {
	dir := t.TempDir()
	path := writeGPX(t, dir, "track.gpx", sampleGPX)

	pts, err := ParseGPXFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 3 {
		t.Fatalf("got %d points, want 3", len(pts))
	}
	if pts[0].Lat != 48.8566 || pts[0].Lon != 2.3522 {
		t.Errorf("first point = (%f, %f), want (48.8566, 2.3522)", pts[0].Lat, pts[0].Lon)
	}
}

func TestParseGPXFiles_Multiple(t *testing.T) {
	dir := t.TempDir()
	gpx2 := `<?xml version="1.0"?>
<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1">
  <trk><trkseg>
    <trkpt lat="40.7128" lon="-74.0060">
      <time>2024-06-15T09:00:00Z</time>
    </trkpt>
  </trkseg></trk>
</gpx>`
	p1 := writeGPX(t, dir, "a.gpx", sampleGPX)
	p2 := writeGPX(t, dir, "b.gpx", gpx2)

	pts, err := ParseGPXFiles([]string{p1, p2})
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 4 {
		t.Fatalf("got %d points, want 4", len(pts))
	}
	// Should be sorted by time — NYC point (09:00) comes first.
	if pts[0].Lat != 40.7128 {
		t.Errorf("first point lat = %f, want 40.7128 (earliest)", pts[0].Lat)
	}
}

func TestParseGPXFiles_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := writeGPX(t, dir, "bad.gpx", "not xml at all {{{")

	_, err := ParseGPXFiles([]string{path})
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestParseGPXFiles_MissingFile(t *testing.T) {
	_, err := ParseGPXFiles([]string{"/nonexistent/track.gpx"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestMatchNearest_ExactMatch(t *testing.T) {
	pts := []Trackpoint{
		{Lat: 48.8566, Lon: 2.3522, Time: time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)},
		{Lat: 48.8570, Lon: 2.3530, Time: time.Date(2024, 6, 15, 10, 1, 0, 0, time.UTC)},
	}

	lat, lon, ok := MatchNearest(pts, pts[0].Time, 30*time.Second)
	if !ok {
		t.Fatal("expected match")
	}
	if lat != 48.8566 || lon != 2.3522 {
		t.Errorf("got (%f, %f), want (48.8566, 2.3522)", lat, lon)
	}
}

func TestMatchNearest_WithinTolerance(t *testing.T) {
	pts := []Trackpoint{
		{Lat: 48.8566, Lon: 2.3522, Time: time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)},
	}

	// 20 seconds after — within 30s tolerance.
	query := pts[0].Time.Add(20 * time.Second)
	lat, lon, ok := MatchNearest(pts, query, 30*time.Second)
	if !ok {
		t.Fatal("expected match within tolerance")
	}
	if lat != 48.8566 {
		t.Errorf("lat = %f, want 48.8566", lat)
	}
	_ = lon
}

func TestMatchNearest_OutsideTolerance_Interpolates(t *testing.T) {
	t1 := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 6, 15, 10, 2, 0, 0, time.UTC) // 2 minutes apart

	pts := []Trackpoint{
		{Lat: 40.0, Lon: -74.0, Time: t1},
		{Lat: 42.0, Lon: -72.0, Time: t2},
	}

	// Midpoint: 1 minute in.
	query := t1.Add(1 * time.Minute)
	lat, lon, ok := MatchNearest(pts, query, 30*time.Second)
	if !ok {
		t.Fatal("expected interpolated match")
	}
	if math.Abs(lat-41.0) > 0.001 {
		t.Errorf("lat = %f, want ~41.0 (midpoint)", lat)
	}
	if math.Abs(lon-(-73.0)) > 0.001 {
		t.Errorf("lon = %f, want ~-73.0 (midpoint)", lon)
	}
}

func TestMatchNearest_OutsideTolerance_QuarterPoint(t *testing.T) {
	t1 := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 6, 15, 10, 4, 0, 0, time.UTC) // 4 minutes apart

	pts := []Trackpoint{
		{Lat: 40.0, Lon: -74.0, Time: t1},
		{Lat: 44.0, Lon: -70.0, Time: t2},
	}

	// 1 minute in = 25% of the way.
	query := t1.Add(1 * time.Minute)
	lat, lon, ok := MatchNearest(pts, query, 30*time.Second)
	if !ok {
		t.Fatal("expected interpolated match")
	}
	if math.Abs(lat-41.0) > 0.001 {
		t.Errorf("lat = %f, want ~41.0 (25%%)", lat)
	}
	if math.Abs(lon-(-73.0)) > 0.001 {
		t.Errorf("lon = %f, want ~-73.0 (25%%)", lon)
	}
}

func TestMatchNearest_BeforeAllPoints(t *testing.T) {
	pts := []Trackpoint{
		{Lat: 48.8566, Lon: 2.3522, Time: time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)},
	}

	// Way before the only point — no interpolation possible.
	query := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)
	_, _, ok := MatchNearest(pts, query, 30*time.Second)
	if ok {
		t.Error("expected no match for time far before all points")
	}
}

func TestMatchNearest_AfterAllPoints(t *testing.T) {
	pts := []Trackpoint{
		{Lat: 48.8566, Lon: 2.3522, Time: time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)},
	}

	// Way after the only point.
	query := time.Date(2024, 6, 15, 11, 0, 0, 0, time.UTC)
	_, _, ok := MatchNearest(pts, query, 30*time.Second)
	if ok {
		t.Error("expected no match for time far after all points")
	}
}

func TestMatchNearest_EmptyPoints(t *testing.T) {
	_, _, ok := MatchNearest(nil, time.Now(), 30*time.Second)
	if ok {
		t.Error("expected no match for empty points")
	}
}
