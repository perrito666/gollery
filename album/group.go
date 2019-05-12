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
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// MetaDataFileName is te name of the file that will contain the folder/album metadata.
const MetaDataFileName = "metadata.json"

var imageExtensions = []string{".jpg", ".jpeg", ".png", ".tiff", ".gif"}

func isImage(fileName string) bool {
	lowerFile := strings.ToLower(fileName)
	// very likely to be there so let's fail early
	if strings.HasSuffix(lowerFile, ".json") {
		return false
	}
	for _, ext := range imageExtensions {
		if strings.HasSuffix(lowerFile, ext) {
			return true
		}
	}
	return false
}

func nonRecursiveWalk(path string, walk filepath.WalkFunc) error {
	matches, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return errors.Wrapf(err, "listing contents of %q", err)
	}
	for _, fileName := range matches {
		fInfo, err := os.Stat(fileName)
		err = walk(fileName, fInfo, err)
		if err != nil {
			return err
		}
	}
	return nil
}

// ThumbSize represents the size of a thumbnail image.
type ThumbSize struct {
	Width  uint `json:"width"`
	Height uint `json:"height"`
}

// DefaultThumbSize is the size we swill use for thumbnails if none is defined
// the number is super retro and it was picked arbitrarily beacause it's the only
// number I can think out of the top of my head on a plane.
var DefaultThumbSize = ThumbSize{
	Width:  320,
	Height: 213,
}

// PictureGroup holds information about a group of pictures from an album
type PictureGroup struct {
	// Parent is the blatant irony spitting in my face reminding us that, even if you don't think
	// you will ever use a linked list in real life scenarios, life is long.
	// it also contains a reference to the containing folder of this one, if any (notice that we
	// will only go as deep as the path passed to initialize the outmost album)
	Parent *PictureGroup `json:"-"`
	// Path is the absolute disk path of this folder/album
	Path string `json:"-"`
	// FolderName holds the name for this group's folder.
	FolderName string `json:"-"`
	// Title is a title representing this picture group/folder/album.
	Title string `json:"title"`
	// Description is a description of this folder.
	Description string `json:"description"`
	// Pictures holds a map of references to all the pictures in this album.
	Pictures map[string]*SinglePicture `json:"pictures"`
	// Order contains a list of strings ordered in the way the items in this album
	// are expected to be accessed.
	Order []string `json:"order"`

	// Recursion
	// SubGroups are folders/albums found inside this one.
	SubGroups map[string]*PictureGroup `json:"-"`
	// SubGroupOrder is the ordering of display for the subgroups, we store this.
	SubGroupOrder []string `json:"sub-group-order"`

	// AllowedThumbSizes holds all the allowed thumbnail sizes for this and children groups
	// it will override parent sizes and be inherited by children that do not specify them, if
	// none is defined it will default to one sane size.
	AllowedThumbSizes []*ThumbSize `json:"allowed-thumb-sizes"`
	// Logger holds a logger
	Logger *log.Logger `json:"-"`
}

// HasSubAlbum returns true if the passed in name matches a subFolder.
func (pg *PictureGroup) HasSubAlbum(name string) bool {
	_, exists := pg.SubGroups[name]
	return exists
}

// HasImage returns true if the passed name matches an image in this folder.
func (pg *PictureGroup) HasImage(name string) bool {
	_, exists := pg.Pictures[name]
	return exists
}

// GetImage returns the SinglePicture for the <name> file if it exists.
func (pg *PictureGroup) GetImage(name string) *SinglePicture {
	key, exists := pg.Pictures[name]
	if !exists {
		return nil
	}
	return key
}

// TraversePath returns the path of the current folder relative to the root album.
func (pg *PictureGroup) TraversePath() string {
	if pg.Parent == nil {
		return "/"
	}
	// the use of path instead of filepath is intentional, this is an internet path
	return path.Join(pg.Parent.TraversePath(), pg.FolderName)
}

// TraverseFileSystemPath returns the path of the current folder relative to the root of file system.
func (pg *PictureGroup) TraverseFileSystemPath() string {
	if pg.Parent == nil {
		absFilePath, err := filepath.Abs(pg.Path)
		if err == nil {
			return absFilePath
		}
		return "/"
	}
	// the use of path instead of filepath is intentional, this is an internet path
	return path.Join(pg.Parent.TraverseFileSystemPath(), pg.FolderName)
}

// WriteMetadata will write metadata for this picture group to the passed io.Writer.
func (pg *PictureGroup) WriteMetadata(metaDestination io.Writer) error {
	serializedPG, err := json.MarshalIndent(pg, "", "    ")
	if err != nil {
		return errors.Wrapf(err, "serializing %q picture group", pg.Path)
	}

	_, err = metaDestination.Write(serializedPG)
	if err != nil {
		return errors.Wrap(err, "writing serialized picture group to file")
	}
	pg.Logger.Printf("will write metadata:\n%s", string(serializedPG))
	return nil
}

// ReadMetadata will load metadata for a picture group from the passed io.Reader.
func (pg *PictureGroup) ReadMetadata(metaOrigin io.Reader) error {
	data, err := ioutil.ReadAll(metaOrigin)
	if err != nil {
		return errors.Wrap(err, "loading metadata from file")
	}
	if len(data) == 0 {
		return nil
	}
	err = json.Unmarshal(data, pg)
	if err != nil {
		return errors.Wrap(err, "deserializing picture group from file data")
	}
	// reparent de-serialized images
	for _, v := range pg.Pictures {
		v.Parent = pg
	}
	var newOrder = []string{}
	for i, v := range pg.SubGroupOrder {
		fullPath := filepath.Join(pg.TraverseFileSystemPath(), v)
		_, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			continue
		}
		newOrder = append(newOrder, pg.SubGroupOrder[i])
		pg.AddSubGroup(fullPath, v, false, false)
	}
	pg.SubGroupOrder = newOrder
	return nil
}

// AddImage constructs the SingleImage for this path and then adds it to the corresponding
// references of this group.
func (pg *PictureGroup) AddImage(path string) error {
	pg.Logger.Printf("considering file: %s\n", path)
	var accessible = true

	_, err := os.Stat(path)
	if err != nil {
		accessible = false
	}

	_, fileName := filepath.Split(path)

	sp, exists := pg.Pictures[fileName]
	if !exists {
		image := &SinglePicture{
			Parent:      pg,
			Path:        path,
			FileName:    fileName,
			Title:       "",
			Description: "",
			Visible:     true,
			Existing:    true,
			Accessible:  accessible,
		}
		pg.Pictures[fileName] = image
		pg.Order = append(pg.Order, fileName)
		sp = image
	} else {
		sp.Parent = pg
	}
	for _, ts := range pg.AllowedThumbSizes {
		err = sp.ensureThumbnail(ts.Width, ts.Height)
		if err != nil {
			return errors.Wrapf(err, "creathing thumb %d x %d", ts.Width, ts.Height)
		}
	}
	return nil
}

// AddSubGroup construct a SubGroup that lives under this one
func (pg *PictureGroup) AddSubGroup(fullPath, folderName string, recursive, update bool) error {
	newPg, err := NewPictureGroup(pg.Logger, fullPath, pg.AllowedThumbSizes, update, recursive, pg)
	if err != nil {
		return errors.Wrap(err, "constructing sub picture group")
	}

	err = newPg.ConstructMetadata(recursive)
	if err != nil {
		return errors.Wrapf(err, "constructing subfoler %q", folderName)
	}

	pg.SubGroups[newPg.FolderName] = newPg
	for _, v := range pg.SubGroupOrder {
		if newPg.FolderName == v {
			return nil
		}
	}
	pg.SubGroupOrder = append(pg.SubGroupOrder, newPg.FolderName)
	return nil
}

// ConstructMetadata will fill the in-memory metadata of this folder/album from the filesystem
func (pg *PictureGroup) ConstructMetadata(recursive bool) error {
	//if not recursive just put folder names in the subgrouporder

	// use abs Path please
	err := nonRecursiveWalk(pg.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walking into path %q", path)
		}
		if info.IsDir() {
			if !recursive {
				return nil
			}
			_, dName := filepath.Split(path)
			err = pg.AddSubGroup(path, dName, recursive, false)
			if err != nil {
				return errors.Wrapf(err, "processing folder %q", path)
			}
			return nil
		}

		filename := info.Name()
		if isImage(filename) {
			err = pg.AddImage(path)
			if err != nil {
				return errors.Wrapf(err, "processing image %q", path)
			}
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "crawling picture group")
	}

	return nil
}

// NonDestructiveUpdateMetadata will see to concile all the metadata in the file, if existing and then update
// in-memory metadata.
func (pg *PictureGroup) NonDestructiveUpdateMetadata(recursive bool) error {
	//if not recursive just put folder names in the subgrouporder
	files := []string{}
	directories := []string{}

	// use abs Path please
	err := nonRecursiveWalk(pg.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walking into path %q", path)
		}
		if info.IsDir() {
			directories = append(directories, path)
			if !recursive {
				return nil
			}
			_, dName := filepath.Split(path)
			err = pg.AddSubGroup(path, dName, recursive, true)
			if err != nil {
				return errors.Wrapf(err, "processing folder %q", path)
			}

			return nil
		}

		filename := info.Name()
		if isImage(filename) {
			err = pg.AddImage(path)
			if err != nil {
				return errors.Wrapf(err, "processing image %q", path)
			}
			files = append(files, filename)
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "crawling picture group")
	}

	for _, picture := range pg.Pictures {
		picture.Existing = false
	}
	for _, pictureName := range files {
		pic, ok := pg.Pictures[pictureName]
		if !ok {
			return errors.Errorf("consistency error, %q detail dissapeared from our records", pictureName)
		}
		pic.Existing = true
	}

	return nil
}

// metaFileAccessor is a mux of all the file characteristics we need.
type metaFileAccessor interface {
	io.Writer
	io.Reader
	io.Seeker
	io.Closer
	Sync() error
	Truncate(int64) error
}

// ensureMetaFile opens and optionally creates a metadata file and returns a reference to it along
// with information about creation or error.
func ensureMetaFile(path string) (metaFileAccessor, bool, error) {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, false, errors.Wrapf(err, "checking metadata file %q existence", path)
	}
	created := false
	fileOpenFlags := os.O_RDWR
	if os.IsNotExist(err) {
		fileOpenFlags = fileOpenFlags | os.O_CREATE
		created = true
	}
	f, err := os.OpenFile(path, fileOpenFlags, 0655)
	if err != nil {
		return nil, false, errors.Wrapf(err, "opening meta file with flags %v", fileOpenFlags)
	}
	return f, created, nil
}

// NewPictureGroup returns a PictureGroup reference with loaded metadata which might be optionally
// up to date. (this implies conciliation of filesystem with metadata files.)
func NewPictureGroup(logger *log.Logger, path string, allowedThumbsSizes []*ThumbSize, update, recursive bool, parent *PictureGroup) (*PictureGroup, error) {
	_, fileName := filepath.Split(path)
	pg := &PictureGroup{
		Logger:        logger,
		Parent:        parent,
		Path:          path,
		FolderName:    fileName,
		Pictures:      make(map[string]*SinglePicture),
		Order:         []string{},
		SubGroups:     map[string]*PictureGroup{},
		SubGroupOrder: []string{},
	}
	metaFilePath := filepath.Join(path, MetaDataFileName)
	metaDataFile, created, err := ensureMetaFile(metaFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "accessing metadata file")
	}
	defer metaDataFile.Close()

	err = pg.ReadMetadata(metaDataFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading metadata from meta file")
	}

	if pg.AllowedThumbSizes == nil {
		pg.AllowedThumbSizes = allowedThumbsSizes
	}

	// if update was mandated but we had to create the file there is more to be done here
	// most likely this is a path into a recursive lookup that was not there before.
	if update && !created {
		err = pg.NonDestructiveUpdateMetadata(recursive)
		if err != nil {
			return nil, errors.Wrap(err, "conciling metadata from disk and file")
		}
	}

	if created {
		err = pg.ConstructMetadata(recursive)
		if err != nil {
			return nil, errors.Wrap(err, "constructing metadata from disk and file")
		}
	}

	if created || update {
		logger.Printf("writing metadata to %s\n", metaFilePath)
		// We will write to this file again so we rewind it.
		metaDataFile.Seek(0, io.SeekStart)
		err = metaDataFile.Truncate(0)
		if err != nil {
			return nil, errors.Wrap(err, "truncating to rewrite conf")
		}
		err = pg.WriteMetadata(metaDataFile)
		if err != nil {
			return nil, errors.Wrap(err, "writing up-to-date metadata to file")
		}
		err = metaDataFile.Sync()
		if err != nil {
			return nil, errors.Wrap(err, "syncing after writing conf")
		}
	}

	return pg, nil
}
