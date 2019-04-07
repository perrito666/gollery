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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

// tempDir creates a temp folder with  "image" files inside and folders
// and returns it's path along with a cleanup function or error.
func tempDir(baseDir string, recurse bool) (string, func(), error) {
	noop := func() {}
	dirPath, err := ioutil.TempDir(baseDir, "pictureGroup")
	if err != nil {
		return "", noop, errors.Wrap(err, "creating a temp file without directories for recursion")
	}

	files := make([]string, len(imageExtensions), len(imageExtensions))
	for i, ext := range imageExtensions {
		f, err := ioutil.TempFile(dirPath, fmt.Sprintf("test*image%s", ext))
		if err != nil {
			panic("cannot create temporary files for tests")
		}
		files[i] = f.Name()
		f.Close()
	}

	var innerDirCleaner = func() {}
	if recurse {
		_, innerDirCleaner, err = tempDir(dirPath, false)
	}

	cleanup := func() {
		innerDirCleaner()
		for _, fname := range files {
			os.Remove(fname)
		}
		os.RemoveAll(dirPath)
	} // clean up
	return dirPath, cleanup, nil

}

func TestPictureGroup_ConstructMetadata(t *testing.T) {
	tests := []struct {
		name      string
		pg        *PictureGroup
		recursive bool
		wantErr   bool
	}{
		{
			name: "picture group construct non recursive",
			pg: &PictureGroup{
				Pictures:      make(map[string]*SinglePicture),
				Order:         []string{},
				SubGroups:     map[string]*PictureGroup{},
				SubGroupOrder: []string{},
			},
			recursive: false,
			wantErr:   false,
		},
		{
			name: "picture group construct recursive",
			pg: &PictureGroup{
				Pictures:      make(map[string]*SinglePicture),
				Order:         []string{},
				SubGroups:     map[string]*PictureGroup{},
				SubGroupOrder: []string{},
			},
			recursive: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		// Non Recursive
		title := tt.name
		if tt.recursive {
			title += " recursive"
		}
		t.Run(title, func(t *testing.T) {
			dirPath, closer, err := tempDir("", true)
			if err != nil {
				t.Errorf("could not create necessary folders %v", err)
				return
			}
			defer closer()
			tt.pg.Path = dirPath
			if err := tt.pg.ConstructMetadata(tt.recursive); (err != nil) != tt.wantErr {
				t.Errorf("PictureGroup.ConstructMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tt.pg.Order) != len(imageExtensions) {
				t.Logf("obtained %d images expected %d", len(tt.pg.Order), len(imageExtensions))
				t.Logf("Images")
				t.Logf("%#v", tt.pg.Order)
				t.Logf("Folders")
				t.Logf("%#v", tt.pg.SubGroupOrder)
				t.FailNow()
			}
			if !tt.recursive {
				return
			}

			if len(tt.pg.SubGroupOrder) != 1 {
				t.Logf("obtained %d subfolders, expected 1", len(tt.pg.SubGroupOrder))
				t.FailNow()
			}

			subOrder := tt.pg.SubGroups[tt.pg.SubGroupOrder[0]]
			if len(subOrder.Order) != len(imageExtensions) {
				t.Logf("obtained %d sub images expected %d", len(subOrder.Order), len(imageExtensions))
				t.Logf("Sub Images")
				t.Logf("%#v", subOrder.Order)
				t.Logf("Sub Folders")
				t.Logf("%#v", subOrder.SubGroupOrder)
				t.FailNow()
			}
		})

	}
}

func TestNewPictureGroup(t *testing.T) {
	type args struct {
		path      string
		update    bool
		recursive bool
	}
	tests := []struct {
		name    string
		args    args
		want    *PictureGroup
		wantErr bool
	}{
		{name: "New PG on fresh folder",
			args: args{
				path:      "",
				update:    false,
				recursive: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirPath, cleaner, err := tempDir("", true)
			if err != nil {
				t.Errorf("could not create necessary folders %v", err)
				t.FailNow()
			}
			defer cleaner()
			tt.args.path = dirPath
			got, err := NewPictureGroup(tt.args.path, []*ThumbSize{&DefaultThumbSize}, tt.args.update, tt.args.recursive, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPictureGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			metaData, err := ioutil.ReadFile(filepath.Join(dirPath, MetaDataFileName))
			if err != nil {
				t.Errorf("could not read metadata file to check it: %v", err)
				return
			}
			metaDataText := string(metaData)
			for _, image := range got.Order {
				appearances := strings.Count(metaDataText, fmt.Sprintf("%q", image))
				// in orders, in the map key and in the single image path
				if appearances != 3 {
					t.Logf("%s is present %d times in json but should be 1", image, appearances)
					t.Logf("-------------------------------------------------\n")
					t.Logf(metaDataText)
					t.Logf("-------------------------------------------------\n")
					t.FailNow()
				}
			}
			if strings.Count(metaDataText, fmt.Sprintf("%q", got.SubGroupOrder[0])) != 1 {
				t.Log(fmt.Sprintf("subgroup [%q] not found", got.SubGroupOrder[0]))
				t.Logf("-------------------------------------------------\n")
				t.Logf(metaDataText)
				t.Logf("-------------------------------------------------\n")
				t.FailNow()
			}

		})
	}
}

func TestPictureGroup_NonDestructiveUpdateMetadata(t *testing.T) {

	t.Run("Updates Metadata when added and deleted files", func(t *testing.T) {
		dirPath, cleaner, err := tempDir("", true)
		if err != nil {
			t.Errorf("could not create necessary folders %v", err)
			t.FailNow()
		}
		defer cleaner()
		pg, err := NewPictureGroup(dirPath, []*ThumbSize{&DefaultThumbSize}, false, true, nil)
		if err != nil {
			t.Errorf("NewPictureGroup() error = %v", err)
			return
		}
		// Not the fanciest rename, but it should work
		oldName := pg.Order[0]
		oldPath := pg.Pictures[oldName].Path
		newPath := oldPath + ".jpg"
		newName := oldName + ".jpg"
		os.Rename(oldPath, newPath)
		if err := pg.NonDestructiveUpdateMetadata(true); err != nil {
			t.Errorf("PictureGroup.NonDestructiveUpdateMetadata() error = %v", err)
			return
		}

		if pg.Pictures[oldName].Existing {
			t.Logf("%q should not be flagged as existing", oldName)
			t.FailNow()
		}

		if _, ok := pg.Pictures[newName]; !ok {
			t.Logf("%q should exist", newName)
			t.FailNow()
		}

		if len(pg.Order) != len(imageExtensions)+1 {
			t.Logf("we expect %d entries but obtained %d", len(pg.Order), len(imageExtensions)+1)
			t.FailNow()
		}

	})

}
