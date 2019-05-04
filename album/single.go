package album

/*
MIT License

Copyright (c) 2019 Horacio Duran <horacio.duran@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/nfnt/resize"
	"github.com/pkg/errors"
)

// SinglePicture contains the information pertaining a one picture of a folder, if nothing
// else the path should be there.
type SinglePicture struct {
	Parent *PictureGroup `json:"-"`
	// Path the relative path to the picture represented.
	Path string `json:"path"`
	// FileName holds the name for this picture's file.
	FileName string `json:"file-name"`
	// Title is a given title for this picture.
	Title string `json:"title"`
	// Description is a description of this picture.
	Description string `json:"description"`
	// Visible indicates if this picture is displayed.
	Visible bool `json:"visible"`
	// Existing indicates wether the file we represent is present on disk.
	Existing bool `json:"existing"`
	// Accessible tells us if we can actually read the file (or our best guess)
	Accessible bool `json:"accessible"`
}

// ThumbName returns the file name of the thumb for this picture with the given size
func (s *SinglePicture) ThumbName(width, height uint) string {
	return fmt.Sprintf("%s_%d_x_%d", s.RelativePath(), width, height)
}

// FileSystemThumb returns the file path of the thumb for this picture with the given size
func (s *SinglePicture) FileSystemThumb(width, height uint) string {
	return fmt.Sprintf("%s_%d_x_%d", filepath.Join(s.Parent.TraverseFileSystemPath(), s.FileName), width, height)
}

// ensureThumbnail makes a very best effort to create a thumbnail in the given sizes
// for this picture
// resize.Resize(width, height uint, img image.Image, interp resize.InterpolationFunction) image.Image
// resize.Thumbnail(maxWidth, maxHeight uint, img image.Image, interp resize.InterpolationFunction) image.Image
func (s *SinglePicture) ensureThumbnail(width, height uint) error {
	path := s.FileSystemThumb(width, height)
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "checking metadata for thumbnail file %q existence", path)
	}
	if err == nil {
		return nil
	}
	var thumbFile *os.File
	var imageFile *os.File

	imageFile, err = os.OpenFile(s.Path, os.O_RDONLY, 0655)
	if err != nil {
		return errors.Wrapf(err, "opening image file %q to generate thumbnail", s.Path)
	}
	defer imageFile.Close()
	imgData, imageDecoder, err := image.Decode(imageFile)
	if err != nil {
		return errors.Wrapf(err, "decoding image in %q cannot read this type of image", s.Path)
	}

	thumbFile, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0655)
	if err != nil {
		return errors.Wrapf(err, "opening thumbnail file %q for writing", path)
	}
	defer thumbFile.Close()

	thumbImgData := resize.Thumbnail(width, height, imgData, resize.Bilinear)

	switch imageDecoder {
	case "png":
		err = png.Encode(thumbFile, thumbImgData)
	case "jpeg", "jpg":
		err = jpeg.Encode(thumbFile, thumbImgData, nil)
	case "gif":
		err = gif.Encode(thumbFile, thumbImgData, nil)
	default:
		err = jpeg.Encode(thumbFile, thumbImgData, nil)
	}
	if err != nil {
		return errors.Wrap(err, "encoding and storing image")
	}
	s.Parent.Logger.Printf("Created thumb %s", path)
	return nil
}

// Thumbnail returns a ReadCloser pointing to the desired size's thumbnail, bear in mind
// that the thumbnail might be created if it does not exist.
// Caller must close thumbnail after use.
func (s *SinglePicture) Thumbnail(width, height uint) (io.ReadCloser, error) {
	err := s.ensureThumbnail(width, height)
	if err != nil {
		return nil, errors.Wrap(err, "cannot ensure thumbnails existence")
	}

	path := s.ThumbName(width, height)
	thumbFile, err := os.OpenFile(path, os.O_RDONLY, 0655)
	if err != nil {
		return nil, errors.Wrapf(err, "opening thumbnail file %q for reading", path)
	}

	return thumbFile, nil
}

// RelativePath returns this file's path relative to the root album.
func (s *SinglePicture) RelativePath() string {
	if s.Parent == nil {
		// FIXME parent should never be nil
		return path.Join("/", s.FileName)
	}
	return path.Join(s.Parent.TraversePath(), s.FileName)
}
