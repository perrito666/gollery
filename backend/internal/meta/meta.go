// Package meta extracts image metadata such as EXIF data.
package meta

import (
	"fmt"
	"os"

	"github.com/rwcarlsen/goexif/exif"

	"github.com/perrito666/gollery/backend/internal/domain"
)

// Extract reads EXIF metadata from a JPEG file.
// Returns a zero ImageMetadata (not an error) for files without EXIF data.
func Extract(filePath string) (*domain.ImageMetadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		// No EXIF data — return empty metadata, not an error.
		return &domain.ImageMetadata{}, nil
	}

	m := &domain.ImageMetadata{}

	if tag, err := x.Get(exif.Make); err == nil {
		m.CameraMake, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Model); err == nil {
		m.CameraModel, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Orientation); err == nil {
		if v, err := tag.Int(0); err == nil {
			m.Orientation = v
		}
	}
	if tag, err := x.Get(exif.PixelXDimension); err == nil {
		if v, err := tag.Int(0); err == nil {
			m.Width = v
		}
	}
	if tag, err := x.Get(exif.PixelYDimension); err == nil {
		if v, err := tag.Int(0); err == nil {
			m.Height = v
		}
	}
	if t, err := x.DateTime(); err == nil {
		m.DateTaken = &t
	}
	if lat, lon, err := x.LatLong(); err == nil {
		m.Latitude = &lat
		m.Longitude = &lon
	}

	return m, nil
}
