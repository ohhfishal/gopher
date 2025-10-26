package watch

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"io"
	"log/slog"
)

type CMD struct {
	Paths []string `arg:"" default:"." name:"path" help:"Paths to watch (default=${default})" type:"path"`
}

// From https://pkg.go.dev/cmd/go#hdr-Build__json_encoding
type BuildEvent struct {
	ImportPath string
	Action     string
	Output     string

	// The Action field is one of the following:
	// build-output - The toolchain printed output
	// build-fail - The build failed
}

func (config *CMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	for _, path := range config.Paths {
		if err = watcher.Add(path); err != nil {
			return err
		}
		logger.Debug("watching path", "path", path)
	}

	go func() {
		logger.Info("starting")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					logger.Info("modified file", "name", event.Name)
				} else if event.Has(fsnotify.Create) {
					logger.Info("created file", "name", event.Name)
				} else if event.Has(fsnotify.Rename) {
					logger.Info("renamed file", "name", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("error", "error:", err)
			}
		}
	}()

	select {
	case <-ctx.Done(): // Context was canceled or deadline expired
		logger.Info("done", "reason", ctx.Err())
		return nil
	}
}
