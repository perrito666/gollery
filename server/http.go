package server

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
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/perrito666/gollery/logs"
	"github.com/perrito666/gollery/render"

	"github.com/perrito666/gollery/album"
	"github.com/pkg/errors"
)

var thumbRE = regexp.MustCompile(`.*_\d+_x_\d+\z`)

func isThumbRequest(fileNameAndWidth string) (string, int, int, bool) {
	if !thumbRE.Match([]byte(fileNameAndWidth)) {
		return "", 0, 0, false
	}
	var height, width int
	parts := strings.Split(fileNameAndWidth, "_")
	if len(parts) > 3 {
		h := parts[len(parts)-1]
		w := parts[len(parts)-3]
		var err error
		height, err = strconv.Atoi(h)
		if err != nil {
			return "", 0, 0, false
		}
		width, err = strconv.Atoi(w)
		if err != nil {
			return "", 0, 0, false
		}
	}
	fileName := strings.TrimRight(fileNameAndWidth, fmt.Sprintf("_%d_x_%d", width, height))
	return fileName, width, height, true
}

// AlbumServer holds the necessary data to serve a given album.
type AlbumServer struct {
	RootFolder *album.PictureFolder
	Port       int64
	Host       string
	ThemePath  string
	Theme      *render.Theme
	Logger     *logs.Logger
	Metadata   map[string]string
}

func serveFolder(w http.ResponseWriter, r *http.Request, folder *album.PictureFolder, theme *render.Theme, meta map[string]string) {
	w.Header().Set("Content-Type", "text/html")
	var b []byte
	buf := bytes.NewBuffer(b)
	err := theme.RenderFolder(folder, buf, meta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("500 boom %v", err)))
		return
	}
	w.Write(buf.Bytes())
}

func (a *AlbumServer) handler(w http.ResponseWriter, r *http.Request) {
	folder := a.RootFolder
	if r.URL.Path == "/" {
		serveFolder(w, r, folder, a.Theme, a.Metadata)
		return
	}
	// TODO traverse path for folders so we know what to do when ends with /
	URLPath := strings.Trim(r.URL.Path, "/")
	pathComponents := strings.Split(URLPath, "/")
	for i, component := range pathComponents {
		if folder.HasSubFolder(component) {
			folder = folder.SubGroups[component]
			if i == len(pathComponents)-1 {
				serveFolder(w, r, folder, a.Theme, a.Metadata)
				return
			}
			continue
		}
		// Serve an image
		if folder.HasImage(component) {
			if r.URL.Query().Get("raw") == "true" {
				// Serve image data
				w.Header().Set("Content-Type", mime.TypeByExtension(r.URL.Path))
				albumImage := folder.GetImage(component)
				http.ServeFile(w, r, albumImage.Path)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			var b []byte
			buf := bytes.NewBuffer(b)
			err := a.Theme.RenderPicture(folder, folder.Pictures[component], buf, a.Metadata)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				a.Logger.Errorf("ERROR: %v", err)
				return
			}
			w.Write(buf.Bytes())
			return
		}
		// Serve a thumbnail
		if actualFile, width, height, isThumb := isThumbRequest(component); isThumb {
			allowedSize := false
			for _, s := range folder.AllowedThumbSizes {
				if s.Height == uint(height) && s.Width == uint(width) {
					allowedSize = true
					break
				}
			}
			if allowedSize {
				w.Header().Set("Content-Type", mime.TypeByExtension(actualFile))
				albumImage := folder.GetImage(actualFile)
				http.ServeFile(w, r, albumImage.FileSystemThumb(uint(width), uint(height)))
				return
			}
			w.WriteHeader(http.StatusForbidden)
			return
		}

	}
	w.WriteHeader(http.StatusNotFound)
}

type themedNotFound struct {
	http.ResponseWriter
	html404Path string
	html500Path string
	status      int
}

func (t *themedNotFound) WriteHeader(status int) {
	t.status = status
	if status == http.StatusNotFound || status == http.StatusInternalServerError {
		t.Header().Set("Content-type", "text/html")
	}
	t.ResponseWriter.WriteHeader(status)
}

func (t *themedNotFound) Write(p []byte) (int, error) {
	var statusFile []byte
	var err error
	if t.status == http.StatusNotFound && t.html404Path != "" {
		statusFile, err = ioutil.ReadFile(t.html404Path)
	}
	if t.status == http.StatusInternalServerError && t.html500Path != "" {
		statusFile, err = ioutil.ReadFile(t.html500Path)
	}
	if err != nil {
		return 0, errors.Wrap(err, "wrapping the error in custom theme")
	}
	if len(statusFile) == 0 {
		return t.ResponseWriter.Write(p)
	}

	b := bytes.NewBuffer(statusFile)

	_, err = b.WriteTo(t.ResponseWriter)
	if err != nil {
		return 0, errors.Wrap(err, "writing wrapped error")
	}
	return len(p), nil
}

// Start starts serving this AlbumServer's root album
func (a *AlbumServer) Start() {
	html404Path := filepath.Join(a.ThemePath, "html", "404.html")
	_, err := os.Stat(html404Path)
	if err != nil {
		html404Path = ""
	}
	html500Path := filepath.Join(a.ThemePath, "html", "500.html")
	_, err = os.Stat(html500Path)
	if err != nil {
		html500Path = ""
	}

	themedFileServer := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			tfs := &themedNotFound{ResponseWriter: w,
				html404Path: html404Path,
				html500Path: html500Path}
			h.ServeHTTP(tfs, r)
		}
	}
	cssServer := themedFileServer(http.FileServer(http.Dir(path.Join(a.ThemePath, "css"))))
	http.Handle("/css/", http.StripPrefix("/css", cssServer))
	jsServer := themedFileServer(http.FileServer(http.Dir(path.Join(a.ThemePath, "js"))))
	http.Handle("/js/", http.StripPrefix("/js", jsServer))
	htmlServer := themedFileServer(http.FileServer(http.Dir(path.Join(a.ThemePath, "html"))))
	http.Handle("/html/", http.StripPrefix("/html", htmlServer))
	imgServer := themedFileServer(http.FileServer(http.Dir(path.Join(a.ThemePath, "img"))))
	http.Handle("/img/", http.StripPrefix("/img", imgServer))
	http.HandleFunc("/", a.handler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", a.Host, a.Port), nil))
}
