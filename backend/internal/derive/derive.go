// Package derive generates image derivatives (thumbnails and previews) from
// source images, using the [cache] package for storage.
//
// # How it works
//
// Each derivative endpoint (thumbnail, preview) follows the same pattern:
//
//  1. Compute the expected cache path via [cache.Layout.ThumbPath] or
//     [cache.Layout.PreviewPath].
//  2. If the file already exists on disk ([cache.Exists]), return the path
//     immediately — cache hit.
//  3. Otherwise, decode the source image, scale it with CatmullRom
//     interpolation (high quality, moderate cost), encode as JPEG at
//     quality 85, and write to the cache path.
//
// # Scaling algorithm
//
// Images are scaled so the longest edge equals the requested size, preserving
// the original aspect ratio. If both dimensions are already within the
// requested size, the image is returned at its original dimensions (no
// upscaling). The [golang.org/x/image/draw] CatmullRom scaler is used for
// visual quality; it is slower than NearestNeighbor or ApproxBiLinear but
// produces noticeably sharper results for photographic content.
//
// # Supported input formats
//
// JPEG and PNG sources are supported (PNG decoder is registered via a blank
// import). The output is always JPEG regardless of the source format.
//
// # Error handling
//
// If encoding fails after the output file has been created, the partial file
// is removed before the error is returned to avoid leaving corrupt cache
// entries.
//
// # Concurrency
//
// [GenerateThumbnail] and [GeneratePreview] are safe to call concurrently
// for different assets. Two concurrent calls for the same asset+size may
// both decode and encode, but the result is identical so the race is benign.
// No locking is performed beyond what the filesystem provides.
//
// # Integration with the API
//
// The API handlers in [api.handleAssetThumbnail] and [api.handleAssetPreview]
// call these functions, then serve the resulting file path via
// [http.ServeFile]. The source path is constructed from scanner-populated
// data (album path + filename), never from raw user input.
package derive

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // register PNG decoder
	"os"

	"golang.org/x/image/draw"

	"github.com/perrito666/gollery/backend/internal/cache"
)

// GenerateThumbnail creates a thumbnail for the given source image.
// If the cached thumbnail already exists, it does nothing.
func GenerateThumbnail(layout *cache.Layout, assetID, sourcePath string, size int) (string, error) {
	outPath := layout.ThumbPath(assetID, size)
	if cache.Exists(outPath) {
		return outPath, nil
	}
	if err := layout.EnsureDirs(); err != nil {
		return "", err
	}
	return outPath, resizeAndSave(sourcePath, outPath, size)
}

// GeneratePreview creates a preview for the given source image.
// If the cached preview already exists, it does nothing.
func GeneratePreview(layout *cache.Layout, assetID, sourcePath string, size int) (string, error) {
	outPath := layout.PreviewPath(assetID, size)
	if cache.Exists(outPath) {
		return outPath, nil
	}
	if err := layout.EnsureDirs(); err != nil {
		return "", err
	}
	return outPath, resizeAndSave(sourcePath, outPath, size)
}

// resizeAndSave decodes an image, scales it so the longest edge equals
// maxSize (preserving aspect ratio), and saves as JPEG.
func resizeAndSave(srcPath, dstPath string, maxSize int) error {
	src, err := decodeImage(srcPath)
	if err != nil {
		return fmt.Errorf("decoding %s: %w", srcPath, err)
	}

	bounds := src.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	newW, newH := fitDimensions(origW, origH, maxSize)
	if newW <= 0 || newH <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d -> %dx%d", origW, origH, newW, newH)
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("creating output: %w", err)
	}
	defer out.Close()

	if err := jpeg.Encode(out, dst, &jpeg.Options{Quality: 85}); err != nil {
		os.Remove(dstPath)
		return fmt.Errorf("encoding JPEG: %w", err)
	}
	return nil
}

func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// fitDimensions scales width and height so the longest edge equals maxSize.
func fitDimensions(w, h, maxSize int) (int, int) {
	if w <= 0 || h <= 0 || maxSize <= 0 {
		return 0, 0
	}
	if w <= maxSize && h <= maxSize {
		return w, h
	}
	if w >= h {
		return maxSize, h * maxSize / w
	}
	return w * maxSize / h, maxSize
}
