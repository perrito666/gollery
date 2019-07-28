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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/juju/gnuflag"
	"github.com/perrito666/gollery/album"
	"github.com/perrito666/gollery/logs"
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
	bindToHost    string
	bindToPort    int64
	logLevel      string
	metadata      map[string]string
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
	bindToPort := gnuflag.CommandLine.Int64("port", 8080, "use this port to serve")
	bindToHost := gnuflag.CommandLine.String("host", "127.0.0.1", "bind to this host address")
	metaData := gnuflag.CommandLine.String("extrametafile", "", "a json file containing key/values for extra metadata that one wants available in templates")
	logLevel := gnuflag.CommandLine.String("loglevel", "info", "the level of logging")

	var err error
	defer func() {
		if err != nil {
			fmt.Println(fmt.Sprintf("Failures encountered: %v\n\n", err))
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
	config.bindToHost = *bindToHost
	config.bindToPort = *bindToPort

	config.metadata = map[string]string{}
	if *metaData != "" {
		f, err := os.OpenFile(*metaData, os.O_RDONLY, 0655)
		if err != nil {
			err = errors.Wrap(err, "opening extra meta file")
			return
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			err = errors.Wrap(err, "loading extra metadata from file")
			return
		}
		if len(data) == 0 {
			return
		}
		if err = json.Unmarshal(data, &config.metadata); err != nil {
			err = errors.Wrap(err, "parsing extra metadata from file")
		}
	}
	config.logLevel = *logLevel
}

func main() {
	// TODO: make some reload for server when files change
	logger := logs.New("GOLLERY")
	var err error
	defer func() {
		if err != nil {
			logger.Errorf("found error while running: %v", err)
			os.Exit(1)
		}
	}()
	switch config.logLevel {
	case "error":
		logger.SetLevel(LvlError)
	case "warn", "warning":
		logger.SetLevel(LvlWarning)
	case "info":
		logger.SetLevel(LvlInfo)
	case "debug":
		logger.SetLevel(LvlDebug)
	case "trace":
		logger.SetLevel(LvlTrace)
	}

	var currentTheme *render.Theme
	if config.buildTheme != "" {
		if config.themeFolder == "" {
			logger.Errorf("cannot create a template without a template path")
			os.Exit(1)
		}
		currentTheme = render.NewTheme(config.buildTheme, config.themeFolder)
		err = currentTheme.Create()
		if err != nil {
			logger.Errorf("initializing theme: %v", err)
			os.Exit(1)
		}
	}

	var pictureFolder *album.PictureFolder
	pictureFolder, err = album.NewPictureGroup(logger,
		config.albumPath,
		[]*album.ThumbSize{&album.DefaultThumbSize},
		config.buildMetadata, config.recursive, nil)
	if err != nil {
		logger.Errorf("initializing album: %v", err)
		os.Exit(1)
	}
	if config.buildMetadata {
		successMsg := fmt.Sprintf("succesfully generated metadata for %q", config.albumPath)
		if config.recursive {
			successMsg += " and subfolders"
		}
		logger.Infof(successMsg)
		return
	}

	logger.Info("STARTED")
	srv := &server.AlbumServer{
		Logger:     logger,
		RootFolder: pictureFolder,
		Port:       config.bindToPort,
		Host:       config.bindToHost,
		ThemePath:  config.themeFolder,
		Theme:      render.NewTheme("something", config.themeFolder),
		Metadata:   config.metadata,
	}
	srv.Start()
}
