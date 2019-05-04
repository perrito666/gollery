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
	"html/template"

	"github.com/perrito666/gollery/album"
)

// FSChild represents an item in a file system.
type FSChild struct {
	// ParentTree contains a list of folders that are parents to this one, the order
	// goes from shallow to deep.
	ParentTree []RendereablePage
	// Metadata contains all the data we might want to pass to the site
	// it comes from command line.
	Metadata map[string]string
}

// buildParentTree crawls a PictureGroup until it reaches the top.
func (f *FSChild) buildParentTree(imageFolder *album.PictureGroup) {
	f.ParentTree = []RendereablePage{}
	parent := imageFolder
	for {
		if parent == nil {
			break
		}
		// not sure if I need this, but I have not slept in many hours
		var newParent = parent
		f.ParentTree = append(f.ParentTree, RendereablePage{PictureGroup: *newParent})
		parent = parent.Parent
	}
	rParentTree := make([]RendereablePage, len(f.ParentTree), len(f.ParentTree))
	for i := range f.ParentTree {
		rParentTree[len(f.ParentTree)-1-i] = f.ParentTree[i]
	}
	f.ParentTree = rParentTree
}

// RendereablePage wraps an album.PictureGroup into a strcuture that has enough convenience
// methods to make rendering templates nicer.
type RendereablePage struct {
	album.PictureGroup
	FSChild
	// Siblings holds a list of the other folders in the same folder as this.
	Siblings []*RendereablePage
	// Children are folders below this one
	Children []*RendereablePage

	// Images contains all the rendereable images in this folder
	Images []*RendereableImage
	// inflated indicates if this has been properly constructed or is just a bare minimum
	inflated bool
	// useful marker for navtrees
	Current bool
}

// DescriptionRich returns an un-escaped description,use at your own Risk
func (r RendereablePage) DescriptionRich() template.HTML {
	return template.HTML(r.Description)
}

// populateSiblings will load the folders that live next to this one if this is not the top level one
func (r *RendereablePage) populateSiblings() {
	if r.Parent == nil {
		return
	}
	// I will not populate them that much, if we really find a case for it
	// we will inflate them, but for now as shallow as a b movie villain
	for i := range r.Parent.SubGroupOrder {
		current := r.Parent.SubGroupOrder[i] == r.FolderName
		r.Siblings = append(r.Siblings, &RendereablePage{
			Current:      current,
			PictureGroup: *r.Parent.SubGroups[r.Parent.SubGroupOrder[i]],
			FSChild:      FSChild{}})
	}
}

// TraversePath just passes through the underlying picture group traverse path for template
func (r RendereablePage) TraversePath() string {
	return r.PictureGroup.TraversePath()
}

// populateChildren adds Children to this page
func (r *RendereablePage) populateChildren(inflate bool) {
	// I will not populate them that much, if we really find a case for it
	// we will inflate them, but for now as shallow as a b movie villain
	for i := range r.SubGroupOrder {
		child := r.SubGroups[r.SubGroupOrder[i]]
		if child == nil {
			continue
		}
		page := NewRendereablePage(*child, inflate, r.Metadata)
		r.Children = append(r.Children, page)
	}
}

// populateImages will load images or this folder, if any.
func (r *RendereablePage) populateImages() {
	if len(r.Order) == 0 {
		return
	}
	for _, k := range r.Order {
		if _, ok := r.Pictures[k]; !ok {
			continue
			// Missing image
		}
		if r.Pictures[k].Visible && r.Pictures[k].Existing {
			r.Images = append(r.Images, &RendereableImage{
				SinglePicture: r.Pictures[k],
				FSChild:       &FSChild{}})
		}
	}
}

// NewRendereablePage constructs a new RendereablePage with the passed folder
func NewRendereablePage(folder album.PictureGroup, inflate bool, meta map[string]string) *RendereablePage {
	page := &RendereablePage{
		PictureGroup: folder,
		FSChild: FSChild{
			Metadata: meta,
		},
		Siblings: []*RendereablePage{},
		Children: []*RendereablePage{},
		Images:   []*RendereableImage{}}
	page.buildParentTree(&folder)
	if inflate {
		page.populateChildren(false)
		page.populateImages()
		page.populateSiblings()
		page.inflated = true
	}
	return page
}
