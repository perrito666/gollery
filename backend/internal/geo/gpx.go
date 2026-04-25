// Package geo provides GPS coordinate resolution from GPX tracklogs.
package geo

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

// Trackpoint is a single GPS position with a timestamp.
type Trackpoint struct {
	Lat  float64
	Lon  float64
	Time time.Time
}

// gpxFile represents the top-level GPX XML structure.
type gpxFile struct {
	XMLName xml.Name `xml:"gpx"`
	Tracks  []gpxTrk `xml:"trk"`
}

type gpxTrk struct {
	Segments []gpxTrkSeg `xml:"trkseg"`
}

type gpxTrkSeg struct {
	Points []gpxTrkPt `xml:"trkpt"`
}

type gpxTrkPt struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Time string  `xml:"time"`
}

// ParseGPXFiles reads one or more GPX files and returns all trackpoints
// sorted by time. Points without a parseable timestamp are silently skipped.
func ParseGPXFiles(paths []string) ([]Trackpoint, error) {
	var all []Trackpoint
	for _, path := range paths {
		pts, err := parseOneGPX(path)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		all = append(all, pts...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Time.Before(all[j].Time)
	})
	return all, nil
}

func parseOneGPX(path string) ([]Trackpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var gpx gpxFile
	if err := xml.Unmarshal(data, &gpx); err != nil {
		return nil, err
	}

	var pts []Trackpoint
	for _, trk := range gpx.Tracks {
		for _, seg := range trk.Segments {
			for _, pt := range seg.Points {
				t, err := time.Parse(time.RFC3339, pt.Time)
				if err != nil {
					// Try RFC3339Nano as fallback.
					t, err = time.Parse(time.RFC3339Nano, pt.Time)
					if err != nil {
						continue
					}
				}
				pts = append(pts, Trackpoint{
					Lat:  pt.Lat,
					Lon:  pt.Lon,
					Time: t,
				})
			}
		}
	}
	return pts, nil
}

// MatchNearest finds the best GPS coordinate for a given timestamp from a
// sorted trackpoint slice. It first looks for the single closest point
// within tolerance. If no single point qualifies but the timestamp falls
// between two consecutive points, it linearly interpolates between them.
// Returns ok=false if no match is possible.
func MatchNearest(points []Trackpoint, t time.Time, tolerance time.Duration) (lat, lon float64, ok bool) {
	if len(points) == 0 {
		return 0, 0, false
	}

	// Binary search for the insertion point.
	idx := sort.Search(len(points), func(i int) bool {
		return !points[i].Time.Before(t)
	})

	// Check the nearest candidates (idx-1 and idx).
	bestDist := time.Duration(math.MaxInt64)
	bestIdx := -1

	if idx > 0 {
		d := t.Sub(points[idx-1].Time)
		if d < bestDist {
			bestDist = d
			bestIdx = idx - 1
		}
	}
	if idx < len(points) {
		d := points[idx].Time.Sub(t)
		if d < bestDist {
			bestDist = d
			bestIdx = idx
		}
	}

	// Exact or near match within tolerance.
	if bestIdx >= 0 && bestDist <= tolerance {
		return points[bestIdx].Lat, points[bestIdx].Lon, true
	}

	// Interpolation: t falls between two consecutive points.
	if idx > 0 && idx < len(points) {
		before := points[idx-1]
		after := points[idx]
		span := after.Time.Sub(before.Time)
		if span > 0 {
			frac := float64(t.Sub(before.Time)) / float64(span)
			lat = before.Lat + frac*(after.Lat-before.Lat)
			lon = before.Lon + frac*(after.Lon-before.Lon)
			return lat, lon, true
		}
	}

	return 0, 0, false
}
