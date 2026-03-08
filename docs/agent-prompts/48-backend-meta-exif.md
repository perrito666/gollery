# Prompt 48 — Backend image metadata extraction

Implement the `meta` package for EXIF data extraction.

Implement:
- `meta.Extract(filePath string) (*Metadata, error)` that reads EXIF from JPEG files
- `Metadata` struct with: camera make/model, date taken, dimensions, GPS coordinates (optional), orientation
- use `github.com/rwcarlsen/goexif/exif` or a similar lightweight library
- return a zero `Metadata` (not an error) for files without EXIF
- add `Metadata` field to `domain.Asset` (optional, populated during snapshot build)
- tests with a sample JPEG fixture

Do not expose metadata through the API yet.
