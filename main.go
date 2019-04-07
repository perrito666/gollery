package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/juju/gnuflag"
	"github.com/perrito666/gollery/album"
	"github.com/perrito666/gollery/render"
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

	fmt.Printf("%#v\n", config)
	if config.buildMetadata {
		_, err = album.NewPictureGroup(
			config.albumPath,
			[]*album.ThumbSize{&album.DefaultThumbSize},
			false, config.recursive)
		if err != nil {
			logger.Fatalf("initializing album: %v", err)
			return
		}
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
}
