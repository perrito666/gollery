package album

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
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

// PictureGroup holds information about a group of pictures from an album
type PictureGroup struct {
	// Path is the absolute disk path of this folder/album
	Path string `json:"-"`
	// Title is a title representing this picture group/folder/album.
	Title string `json:"title"`
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
}

// WriteMetadata will write metadata for this picture group to the passed io.Writer.
func (pg *PictureGroup) WriteMetadata(metaDestination io.Writer) error {
	serializedPG, err := json.Marshal(pg)
	if err != nil {
		return errors.Wrapf(err, "serializing %q picture group", pg.Path)
	}

	_, err = metaDestination.Write(serializedPG)
	if err != nil {
		return errors.Wrap(err, "writing serialized picture group to file")
	}
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
	return nil
}

// AddImage constructs the SingleImage for this path and then adds it to the corresponding
// references of this group.
func (pg *PictureGroup) AddImage(path string) error {
	var accessible = true

	_, err := os.Stat(path)
	if err != nil {
		accessible = false
	}

	image := &SinglePicture{
		Path:        path,
		Title:       "",
		Description: "",
		Visible:     true,
		Existing:    true,
		Accessible:  accessible,
	}

	_, exists := pg.Pictures[path]
	pg.Pictures[path] = image
	if !exists {
		pg.Order = append(pg.Order, path)
	}
	return nil
}

// AddSubGroup construct a SubGroup that lives under this one
func (pg *PictureGroup) AddSubGroup(fullPath, folderName string, recursive bool) error {
	newPg, err := NewPictureGroup(fullPath, false, recursive)
	if err != nil {
		return errors.Wrap(err, "constructing sub picture group")
	}

	err = newPg.ConstructMetadata(recursive)
	if err != nil {
		return errors.Wrapf(err, "constructing subfoler %q", folderName)
	}

	pg.SubGroups[folderName] = newPg
	pg.SubGroupOrder = append(pg.SubGroupOrder, folderName)
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
			err = pg.AddSubGroup(path, dName, recursive)
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
			err = pg.AddSubGroup(path, dName, recursive)
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
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "crawling picture group")
	}

	for _, picture := range pg.Pictures {
		picture.Existing = false
	}
	for _, picturePath := range files {
		pic, ok := pg.Pictures[picturePath]
		if !ok {
			return errors.Errorf("consistency error, %q detail dissapeared from our records", picturePath)
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
func NewPictureGroup(path string, update, recursive bool) (*PictureGroup, error) {
	pg := &PictureGroup{
		Path:          path,
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

	if created && update {
		// We will write to this file again so we rewind it.
		metaDataFile.Seek(0, io.SeekStart)
		err = pg.WriteMetadata(metaDataFile)
		if err != nil {
			return nil, errors.Wrap(err, "writing up-to-date metadata to file")
		}
	}

	return pg, nil
}
