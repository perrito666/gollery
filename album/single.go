package album

// SinglePicture contains the information pertaining a one picture of a folder, if nothing
// else the path should be there.
type SinglePicture struct {
	// Path the relative path to the picture represented.
	Path string `json:"path"`
	// Title is a given title for this picture.
	Title string `json:"title,omitempty"`
	// Description is a description of this picture.
	Description string `json:"description,omitempty"`
	// Visible indicates if this picture is displayed.
	Visible bool `json:"visible"`
	// Existing indicates wether the file we represent is present on disk.
	Existing bool `json:"existing"`
	// Accessible tells us if we can actually read the file (or our best guess)
	Accessible bool `json:"accessible"`
}
