package geo

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleTCX = `<?xml version="1.0" encoding="UTF-8"?>
<TrainingCenterDatabase xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2">
  <Activities>
    <Activity Sport="Biking">
      <Lap StartTime="2024-06-15T10:00:00Z">
        <Track>
          <Trackpoint>
            <Time>2024-06-15T10:00:00Z</Time>
            <Position>
              <LatitudeDegrees>48.8566</LatitudeDegrees>
              <LongitudeDegrees>2.3522</LongitudeDegrees>
            </Position>
          </Trackpoint>
          <Trackpoint>
            <Time>2024-06-15T10:01:00Z</Time>
            <Position>
              <LatitudeDegrees>48.8570</LatitudeDegrees>
              <LongitudeDegrees>2.3530</LongitudeDegrees>
            </Position>
          </Trackpoint>
          <Trackpoint>
            <Time>2024-06-15T10:02:00Z</Time>
          </Trackpoint>
          <Trackpoint>
            <Time>2024-06-15T10:03:00Z</Time>
            <Position>
              <LatitudeDegrees>48.8580</LatitudeDegrees>
              <LongitudeDegrees>2.3540</LongitudeDegrees>
            </Position>
          </Trackpoint>
        </Track>
      </Lap>
    </Activity>
  </Activities>
</TrainingCenterDatabase>`

const courseTCX = `<?xml version="1.0" encoding="UTF-8"?>
<TrainingCenterDatabase xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2">
  <Courses>
    <Course>
      <Track>
        <Trackpoint>
          <Time>2024-06-15T09:00:00Z</Time>
          <Position>
            <LatitudeDegrees>40.7128</LatitudeDegrees>
            <LongitudeDegrees>-74.0060</LongitudeDegrees>
          </Position>
        </Trackpoint>
      </Track>
    </Course>
  </Courses>
</TrainingCenterDatabase>`

func writeFileGeo(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseTCXFiles_Basic(t *testing.T) {
	dir := t.TempDir()
	path := writeFileGeo(t, dir, "workout.tcx", sampleTCX)

	pts, err := ParseTCXFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	// One trackpoint has no Position — should be skipped.
	if len(pts) != 3 {
		t.Fatalf("got %d points, want 3", len(pts))
	}
	if pts[0].Lat != 48.8566 || pts[0].Lon != 2.3522 {
		t.Errorf("first point = (%f, %f), want (48.8566, 2.3522)", pts[0].Lat, pts[0].Lon)
	}
}

func TestParseTCXFiles_CoursesSection(t *testing.T) {
	dir := t.TempDir()
	path := writeFileGeo(t, dir, "course.tcx", courseTCX)

	pts, err := ParseTCXFiles([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 1 {
		t.Fatalf("got %d points, want 1", len(pts))
	}
	if pts[0].Lat != 40.7128 {
		t.Errorf("lat = %f, want 40.7128", pts[0].Lat)
	}
}

func TestParseTCXFiles_Multiple_SortedByTime(t *testing.T) {
	dir := t.TempDir()
	p1 := writeFileGeo(t, dir, "a.tcx", sampleTCX)
	p2 := writeFileGeo(t, dir, "b.tcx", courseTCX)

	pts, err := ParseTCXFiles([]string{p1, p2})
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 4 {
		t.Fatalf("got %d points, want 4", len(pts))
	}
	// courseTCX point at 09:00 should sort ahead of the 10:00 activity.
	if pts[0].Lat != 40.7128 {
		t.Errorf("first point lat = %f, want 40.7128 (earliest)", pts[0].Lat)
	}
}

func TestParseTCXFiles_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFileGeo(t, dir, "bad.tcx", "not xml at all {{{")

	_, err := ParseTCXFiles([]string{path})
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestParseTCXFiles_MissingFile(t *testing.T) {
	_, err := ParseTCXFiles([]string{"/nonexistent/workout.tcx"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseTrackFiles_MixedGPXAndTCX(t *testing.T) {
	dir := t.TempDir()
	gpxPath := writeFileGeo(t, dir, "track.gpx", sampleGPX)
	tcxPath := writeFileGeo(t, dir, "course.tcx", courseTCX)

	pts, err := ParseTrackFiles([]string{gpxPath, tcxPath})
	if err != nil {
		t.Fatal(err)
	}
	// 3 from GPX + 1 from TCX course.
	if len(pts) != 4 {
		t.Fatalf("got %d points, want 4", len(pts))
	}
	// Merged output must be sorted by time; TCX course point at 09:00
	// should precede all GPX points (10:00+).
	if pts[0].Lat != 40.7128 {
		t.Errorf("first point lat = %f, want 40.7128 (earliest across formats)", pts[0].Lat)
	}
}

func TestParseTrackFiles_UnknownExtension(t *testing.T) {
	dir := t.TempDir()
	path := writeFileGeo(t, dir, "log.kml", "<kml/>")

	_, err := ParseTrackFiles([]string{path})
	if err == nil {
		t.Error("expected error for unsupported extension")
	}
}

func TestParseTrackFiles_EmptyInput(t *testing.T) {
	pts, err := ParseTrackFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 0 {
		t.Errorf("expected 0 points, got %d", len(pts))
	}
}
