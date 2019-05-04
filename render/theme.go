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
	"encoding/json"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/perrito666/gollery/album"

	"github.com/pkg/errors"
)

// ErrNoPathForTheme should be returned when a Theme, instantiated without a Path is asked to do
// something that requires it.
var ErrNoPathForTheme = errors.New("no path was provided for the theme")

// ErrThemeIsNoDir should be returned when a Theme was instantiated with a path that does not point to
// a folder.
var ErrThemeIsNoDir = errors.New("theme path does not point to a folder")

const (
	templateFolderName     = "templates"
	singleTemplateFileName = "single.html"
	groupTemplateFileName  = "page.html"
	themeConfigFileName    = "theme.json"
)

var themeFolders = []string{
	templateFolderName, // html files will be built with these.
	"js",               // this will be served under /js ideally should contain javasript
	"css",              // this will be serve under /css ideally should contain css styles
	"html",             // this will be served under /html and contains pages other than index.html such as 404.html
	"img",              // this will be suerved under /img and contains images used for theme images
}

// NewTheme constructs a theme with the required parameters.
func NewTheme(name, path string) *Theme {
	return &Theme{
		Path: path,
		Name: name,
	}
}

// Theme is a set of templates to display an album or single file
type Theme struct {
	// Path points to the relative or absolute path where the templates are stored.
	Path               string
	Name               string
	singlePageTemplate *template.Template
	folderTemplate     *template.Template
}

// ensureFolder will make a significant effort to attempt creation of a given theme folder
func ensureFolder(folder string) error {
	info, err := os.Stat(folder)
	// Cant stat on path
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "trying to find out information on %q path", folder)
	}
	// Path exists but is not a folder (or we cannot tell it is)
	if err == nil && !info.IsDir() {
		return ErrThemeIsNoDir
	}
	// path doe not exist
	if os.IsNotExist(err) {
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			return errors.Wrapf(err, "while creating theme folder in %q", folder)
		}
	}
	return nil
}

// writeINonExistent will write the passed contets into a file if said file does not exist.
// most operations in gollery try to preserve data from accidental fingers.
func writeINonExistent(filePath string, contents []byte) error {
	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		return nil
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0655)
	if err != nil {
		return errors.Wrapf(err, "opening template file %q to initialize it", filePath)
	}
	defer file.Close()
	_, err = file.Write(contents)
	if err != nil {
		return errors.Wrap(err, "writing contents to file")
	}
	return nil
}

// Create will create the base files required for a theme to be valid.
func (t *Theme) Create() error {
	if t.Path == "" {
		return ErrNoPathForTheme
	}
	err := ensureFolder(t.Path)
	if err != nil {
		return errors.Wrapf(err, "while making sure theme folder is available %q", t.Path)
	}

	// We will need a few folders to conform to what the server expects as a theme, most likely we will
	// be able to serve it anway but wouldn't it be nicer if all was as expected?
	for _, desiredFolder := range themeFolders {
		desiredPath := filepath.Join(t.Path, desiredFolder)
		err := ensureFolder(desiredPath)
		if err != nil {
			return errors.Wrapf(err, "while making sure theme subfolder %q is available %q", desiredFolder, desiredPath)
		}

	}

	// the fllowing two files are the only ones we actually need for a template, it will be
	// super dull and contain no style whatsoever.
	singleTPLPath := filepath.Join(t.Path, templateFolderName, singleTemplateFileName)
	err = writeINonExistent(singleTPLPath, singleImageDefaultTemplate)
	if err != nil {
		return errors.Wrap(err, "while trying to generate a template for single picture view")
	}

	groupTPLPath := filepath.Join(t.Path, templateFolderName, groupTemplateFileName)
	err = writeINonExistent(groupTPLPath, groupDefaultTemplate)
	if err != nil {
		return errors.Wrap(err, "while trying to generate a template for album folder")
	}

	err = t.WriteConfig()
	if err != nil {
		return errors.Wrap(err, "writing theme config")
	}
	return nil
}

// WriteConfig serializes to json and writes the current theme config, there s no reading of it until
// necessary
func (t *Theme) WriteConfig() error {
	themeConfigFilePath := filepath.Join(t.Path, themeConfigFileName)
	themeConfigFile, err := os.OpenFile(themeConfigFilePath, os.O_CREATE|os.O_WRONLY, 0655)
	if err != nil {
		return errors.Wrap(err, "accessing theme config file")
	}
	defer themeConfigFile.Close()
	themeConfig := map[string]string{"name": t.Name}
	serializedConfig, err := json.MarshalIndent(&themeConfig, "", "    ")
	if err != nil {
		return errors.Wrapf(err, "serializing %q theme config", t.Name)
	}

	_, err = themeConfigFile.Write(serializedConfig)
	if err != nil {
		return errors.Wrap(err, "writing serialized theme configto file")
	}
	return nil
}

func (t *Theme) maybeLoadSingleTemplate() error {
	if t.singlePageTemplate != nil {
		return nil
	}
	filePath := filepath.Join(t.Path, templateFolderName, singleTemplateFileName)
	var err error
	t.singlePageTemplate, err = template.New(singleTemplateFileName).Funcs(template.FuncMap{
		"LastItem": func(i, size int) bool { return i == size-1 },
	}).
		ParseFiles(filePath)
	if err != nil {
		return errors.Wrapf(err, "loading %q template for single images", filePath)
	}
	return nil
}
func (t *Theme) maybeLoadPageTemplate() error {
	if t.folderTemplate != nil {
		return nil
	}
	filePath := filepath.Join(t.Path, templateFolderName, groupTemplateFileName)
	var err error
	t.folderTemplate, err = template.New(groupTemplateFileName).Funcs(template.FuncMap{
		"LastItem": func(i, size int) bool { return i == size-1 },
	}).
		ParseFiles(filePath)
	if err != nil {
		return errors.Wrapf(err, "loading %q template for folders", filePath)
	}
	return nil
}

// RenderPicture render the passed picture page into the passed io.Writer
func (t *Theme) RenderPicture(folder *album.PictureGroup, img *album.SinglePicture, destination io.Writer, meta map[string]string) error {
	err := t.maybeLoadSingleTemplate()
	if err != nil {
		return errors.Wrap(err, "loading template to render single image")
	}
	picture := NewRendereableImage(folder, img, meta)
	err = t.singlePageTemplate.Execute(destination, picture)
	if err != nil {
		return errors.Wrap(err, "rendering image template")
	}
	return nil
}

// RenderFolder renders the passed folder into the passed io.Writer
func (t *Theme) RenderFolder(folder *album.PictureGroup, destination io.Writer, meta map[string]string) error {
	err := t.maybeLoadPageTemplate()
	if err != nil {
		return errors.Wrap(err, "loading template to render folder")
	}
	albumFolder := NewRendereablePage(*folder, true, meta)
	err = t.folderTemplate.Execute(destination, *albumFolder)
	if err != nil {
		return errors.Wrap(err, "rendering folder template")
	}
	return nil
}
