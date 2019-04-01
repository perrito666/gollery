package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/juju/gnuflag"
	"github.com/perrito666/gollery/album"
)

type configFields struct {
	buildMetadata bool
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
}

func main() {
	// TODO: Make path relative to abs
	// Store both name and path in individual elements
	// make some reload for server when files change
	logger := log.New(os.Stdout, "GOLLERY: ", log.Ldate|log.Ltime|log.Lshortfile)
	var err error
	defer func() {
		if err != nil {
			logger.Printf("found error while running: %v", err)
			os.Exit(1)
		}
	}()
	fmt.Printf("%#v\n", config)
	if config.buildMetadata {
		_, err = album.NewPictureGroup(config.albumPath, false, config.recursive)
		if err != nil {
			return
		}
		successMsg := fmt.Sprintf("succesfully generated metadata for %q", config.albumPath)
		if config.recursive {
			successMsg += " and subfolders"
		}
		logger.Printf(successMsg)
		return
	}

	logger.Print("STARTED")
}
