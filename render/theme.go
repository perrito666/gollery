package render

import (
	"encoding/json"
	"os"
	"path/filepath"

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
	"html",             // this will be served under / and contains pages other than index.html
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
	Path string
	Name string
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
