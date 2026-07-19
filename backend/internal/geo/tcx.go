package geo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// tcxFile represents the top-level Training Center XML structure.
// The schema places trackpoints under Activities > Activity > Lap > Track,
// with an alternative Courses > Course > Track path used by some exporters.
type tcxFile struct {
	XMLName    xml.Name     `xml:"TrainingCenterDatabase"`
	Activities tcxActivities `xml:"Activities"`
	Courses    tcxCourses    `xml:"Courses"`
}

type tcxActivities struct {
	Activities []tcxActivity `xml:"Activity"`
}

type tcxActivity struct {
	Laps []tcxLap `xml:"Lap"`
}

type tcxLap struct {
	Tracks []tcxTrack `xml:"Track"`
}

type tcxCourses struct {
	Courses []tcxCourse `xml:"Course"`
}

type tcxCourse struct {
	Tracks []tcxTrack `xml:"Track"`
}

type tcxTrack struct {
	Points []tcxTrackpoint `xml:"Trackpoint"`
}

type tcxTrackpoint struct {
	Time     string       `xml:"Time"`
	Position *tcxPosition `xml:"Position"`
}

type tcxPosition struct {
	Lat float64 `xml:"LatitudeDegrees"`
	Lon float64 `xml:"LongitudeDegrees"`
}

// ParseTCXFiles reads one or more TCX (Training Center XML) files and returns
// all trackpoints sorted by time. Points without a parseable timestamp or
// without a <Position> element (e.g. indoor treadmill segments or GPS drop-outs)
// are silently skipped.
func ParseTCXFiles(paths []string) ([]Trackpoint, error) {
	var all []Trackpoint
	for _, path := range paths {
		pts, err := parseOneTCX(path)
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

func parseOneTCX(path string) ([]Trackpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tcx tcxFile
	if err := xml.Unmarshal(data, &tcx); err != nil {
		return nil, err
	}

	var pts []Trackpoint
	appendTracks := func(tracks []tcxTrack) {
		for _, tr := range tracks {
			for _, pt := range tr.Points {
				if pt.Position == nil {
					continue
				}
				t, err := time.Parse(time.RFC3339, pt.Time)
				if err != nil {
					t, err = time.Parse(time.RFC3339Nano, pt.Time)
					if err != nil {
						continue
					}
				}
				pts = append(pts, Trackpoint{
					Lat:  pt.Position.Lat,
					Lon:  pt.Position.Lon,
					Time: t,
				})
			}
		}
	}
	for _, act := range tcx.Activities.Activities {
		for _, lap := range act.Laps {
			appendTracks(lap.Tracks)
		}
	}
	for _, course := range tcx.Courses.Courses {
		appendTracks(course.Tracks)
	}
	return pts, nil
}

// ParseTrackFiles parses a mixed set of GPX and TCX track files, dispatched
// by lowercase file extension, and returns the merged trackpoints sorted by
// time. Unknown extensions return an error so misconfiguration is not silent.
func ParseTrackFiles(paths []string) ([]Trackpoint, error) {
	var gpxPaths, tcxPaths []string
	for _, p := range paths {
		switch strings.ToLower(filepath.Ext(p)) {
		case ".gpx":
			gpxPaths = append(gpxPaths, p)
		case ".tcx":
			tcxPaths = append(tcxPaths, p)
		default:
			return nil, fmt.Errorf("unsupported track file extension: %s", p)
		}
	}

	var all []Trackpoint
	if len(gpxPaths) > 0 {
		pts, err := ParseGPXFiles(gpxPaths)
		if err != nil {
			return nil, err
		}
		all = append(all, pts...)
	}
	if len(tcxPaths) > 0 {
		pts, err := ParseTCXFiles(tcxPaths)
		if err != nil {
			return nil, err
		}
		all = append(all, pts...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Time.Before(all[j].Time)
	})
	return all, nil
}
