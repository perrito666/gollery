package render

import (
	"github.com/perrito666/gollery/album"
)

// FSChild represents an item in a file system.
type FSChild struct {
	// ParentTree contains a list of folders that are parents to this one, the order
	// goes from shallow to deep.
	ParentTree []*RendereablePage
}

// buildParentTree crawls a PictureGroup until it reaches the top.
func (f *FSChild) buildParentTree(imageFolder *album.PictureGroup) {
	parent := imageFolder.Parent
	for {
		if parent.Parent == nil {
			break
		}
		// not sure if I need this, but I have not slept in many hours
		var newParent = parent.Parent
		f.ParentTree = append(f.ParentTree, &RendereablePage{PictureGroup: newParent})
		parent = parent.Parent
	}
}

// RendereablePage wraps an album.PictureGroup into a strcuture that has enough convenience
// methods to make rendering templates nicer.
type RendereablePage struct {
	*album.PictureGroup
	*FSChild
	// Siblings holds a list of the other folders in the same folder as this.
	Siblings []*RendereablePage

	// Images contains all the rendereable images in this folder
	Images []*RendereableImage
	// inflated indicates if this has been properly constructed or is just a bare minimum
	inflated bool
}

// populateSiblings will load the folders that live next to this one if this is not the top level one
func (r *RendereablePage) populateSiblings() {
	if r.Parent == nil {
		return
	}
	// I will not populate them that much, if we really find a case for it
	// we will inflate them, but for now as shallow as a b movie villain
	for i := range r.Parent.SubGroupOrder {
		r.Siblings = append(r.Siblings, &RendereablePage{
			PictureGroup: r.Parent.SubGroups[r.Parent.SubGroupOrder[i]],
			FSChild:      &FSChild{}})
	}
}

// populateImages will load images or this folder, if any.
func (r *RendereablePage) populateImages() {
	if len(r.Order) == 0 {
		return
	}
	for _, k := range r.Order {
		r.Images = append(r.Images, &RendereableImage{
			SinglePicture: r.Pictures[k],
			FSChild:       &FSChild{}})
	}
}

// NewRendereablePage constructs a new RendereablePage with the passed folder
func NewRendereablePage(folder *album.PictureGroup) *RendereablePage {
	page := &RendereablePage{PictureGroup: folder, FSChild: &FSChild{}}
	page.buildParentTree(folder)

	return page
}
