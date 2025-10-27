package watch

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"time"
)

type CMD struct {
	Path     string        `arg:"" default:"." name:"path" help:"Paths to watch (default=${default})" type:"path"`
	Interval time.Duration `default:"3s" help:""`
}

func (config *CMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close() //nolint: errcheck

	if err := watcher.Add(config.Path); err != nil {
		return fmt.Errorf("adding path to watch list: %w", err)
	}
	if err := filepath.WalkDir(config.Path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if err := watcher.Add(path); err != nil {
				return fmt.Errorf("adding path to watch list: %w", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walking directories: %w", err)
	}

	go config.Work(ctx, watcher, logger)

	<-ctx.Done()
	logger.Info("done", "reason", ctx.Err())
	return nil
}

func (config CMD) Work(ctx context.Context, watcher *fsnotify.Watcher, logger *slog.Logger) {
	refresh := time.NewTicker(config.Interval / 10)
	defer refresh.Stop()

	// TODO: Control this dependency injection somewhere
	config.Build(ctx, logger, goBuild)
	lastBuild := time.Now()
	build := false

	for {
		select {
		case <-ctx.Done():
			logger.Info("closing")
			return
		case <-refresh.C:
			if build && time.Since(lastBuild) > config.Interval {
				config.Build(ctx, logger, goBuild)
				lastBuild = time.Now()
				build = false
			}
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Filter out events we don't want to handle
			if filepath.Ext(event.Name) != ".go" {
				continue
			}
			if !event.Has(fsnotify.Write) &&
				!event.Has(fsnotify.Create) &&
				!event.Has(fsnotify.Rename) {
				continue
			}
			build = true
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Error("error", "error:", err)
		}
	}
}
