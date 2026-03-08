package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/perrito666/gollery/backend/internal/app"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "gollery.json", "path to config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("galleryd %s\n", version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("galleryd starting", "version", version)

	if err := app.Run(ctx, *configPath); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}
