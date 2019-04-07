package render

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
	"github.com/perrito666/gollery/album"
)

// RendereableImage wraps an album.SinglePicture into a template friendly structure
// to ease converting it into a page.
type RendereableImage struct {
	// I am pretty sure I am buying a headache in the near future here.
	*album.SinglePicture
	*FSChild

	// The only reason that the following fields are rendereables instead of the base types
	// for the images and groups is because they will likely gain methods to ease template
	// rendering and easier done now.

	// Siblings holds a list of the other images in the same folders as this.
	Siblings []*RendereableImage
	// Children contains the folders that live with this image.
	Children []*RendereablePage

	previous int
	next     int

	// Current is set to true if this image is rendering itself.
	Current bool

	inflated bool
}

// First returns the first image in the current folder.
func (r *RendereableImage) First() *RendereableImage {
	if len(r.Siblings) == 0 {
		return nil
	}
	return r.Siblings[0]
}

// Previous returns the image before this one in the current folder
func (r *RendereableImage) Previous() *RendereableImage {
	if len(r.Siblings) == 0 {
		return nil
	}
	if r.previous < 0 {
		return nil
	}
	return r.Siblings[r.previous]
}

// Next returns the image before this one in the current folder.
func (r *RendereableImage) Next() *RendereableImage {
	if len(r.Siblings) == 0 {
		return nil
	}
	if r.next < 0 {
		return nil
	}
	return r.Siblings[r.next]
}

// Last returns the last image in this folder.
func (r *RendereableImage) Last() *RendereableImage {
	if len(r.Siblings) == 0 {
		return nil
	}
	return r.Siblings[len(r.Siblings)-1]
}

// NewRendereableImage returns a struct that can be use to render an image template.
func NewRendereableImage(imageFolder *album.PictureGroup, image *album.SinglePicture) *RendereableImage {
	img := &RendereableImage{
		SinglePicture: image,
		FSChild:       &FSChild{},
		Siblings:      []*RendereableImage{},
		Children:      []*RendereablePage{}}

	// I am pretty sure this is costing me in garbage collection.
	for i, k := range imageFolder.Order {
		img.Siblings = append(img.Siblings, &RendereableImage{SinglePicture: imageFolder.Pictures[k]})
		if img.Siblings[i].FileName == img.FileName {
			img.Siblings[i].Current = true
			// -1 means no previous
			img.previous = i - 1
			// we deal with the possible out of range of this later
			img.next = i + 1
		}
	}
	// just in case there is no next
	if img.next > len(img.Siblings)-1 {
		img.next = -1
	}
	for _, k := range imageFolder.SubGroupOrder {
		img.Children = append(img.Children, &RendereablePage{PictureGroup: imageFolder.SubGroups[k]})
	}
	img.buildParentTree(imageFolder)

	return img
}
