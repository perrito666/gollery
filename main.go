package main

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
	"log"
	"os"
	"path/filepath"

	"github.com/juju/gnuflag"
	"github.com/perrito666/gollery/album"
	"github.com/perrito666/gollery/render"
	"github.com/perrito666/gollery/server"
	"github.com/pkg/errors"
)

type configFields struct {
	buildMetadata bool
	buildTheme    string
	themeFolder   string
	recursive     bool
	albumPath     string
}

var config = configFields{}

var (
	// ErrNoAlbumPath should be returned when the user did not specify an album path to work on.
	ErrNoAlbumPath = errors.New("album path not passed")
)

func init() {
	gnuflag.CommandLine.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s [flags] /path/to/root/album:\n\n", os.Args[0])
		gnuflag.PrintDefaults()
	}
	buildMetadata := gnuflag.CommandLine.Bool("createmeta", false, "create the metadata files for album and optionally subfolders")
	buildTheme := gnuflag.CommandLine.String("createtheme", "", "create a theme for the album in the passed path with the passed name")
	themePath := gnuflag.CommandLine.String("theme", "", "use the passed in folder as a theme (included for creation)")
	recursive := gnuflag.CommandLine.Bool("recursive", true, "apply any action to the root album and subfolders")

	var err error
	defer func() {
		if err != nil {
			gnuflag.CommandLine.Usage()
			os.Exit(1)
		}
	}()
	err = gnuflag.CommandLine.Parse(true, os.Args[1:])
	if err != nil {
		return
	}

	args := gnuflag.CommandLine.Args()
	if len(args) != 1 {
		err = ErrNoAlbumPath
	}
	config.albumPath = args[0]
	config.buildMetadata = *buildMetadata
	config.recursive = *recursive

	config.buildTheme = *buildTheme
	config.themeFolder = *themePath
	config.themeFolder, err = filepath.Abs(config.themeFolder)
	if err != nil {
		err = errors.Wrap(err, "something is fishy with theme path")
	}
}

func main() {
	// TODO: make some reload for server when files change
	logger := log.New(os.Stdout, "GOLLERY: ", log.Ldate|log.Ltime|log.Lshortfile)
	var err error
	defer func() {
		if err != nil {
			logger.Printf("found error while running: %v", err)
			os.Exit(1)
		}
	}()

	var currentTheme *render.Theme
	if config.buildTheme != "" {
		if config.themeFolder == "" {
			logger.Fatal("cannot create a template without a template path")
			return
		}
		currentTheme = render.NewTheme(config.buildTheme, config.themeFolder)
		err = currentTheme.Create()
		if err != nil {
			logger.Fatalf("initializing theme: %v", err)
			return
		}
	}

	//fmt.Printf("%#v\n", config) // ye ole print debug, sue me
	var pictureGroup *album.PictureGroup
	pictureGroup, err = album.NewPictureGroup(
		config.albumPath,
		[]*album.ThumbSize{&album.DefaultThumbSize},
		false, config.recursive, nil)
	if err != nil {
		logger.Fatalf("initializing album: %v", err)
		return
	}
	if config.buildMetadata {
		successMsg := fmt.Sprintf("succesfully generated metadata for %q", config.albumPath)
		if config.recursive {
			successMsg += " and subfolders"
		}
		logger.Printf(successMsg)
		return
	}

	// TODO Serve
	// TODO add traverse where each folder receives a path, removes it's part and passes the rest

	logger.Print("STARTED")
	srv := &server.AlbumServer{
		RootFolder: pictureGroup,
		Port:       8080,
		ThemePath:  config.themeFolder,
		Theme:      render.NewTheme("something", config.themeFolder),
	}
	srv.Start()
}
