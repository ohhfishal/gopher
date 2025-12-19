package runner

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	// "github.com/ohhfishal/gopher/pretty"
	"github.com/fsnotify/fsnotify"
)

var _ Runner = &FileCache{}

type FileCache struct {
	Path    string
	watcher *fsnotify.Watcher
	ok      atomic.Bool
	once    sync.Once
}

func (cache *FileCache) init() {
	if err := cache.initHelper(); err != nil {
		// This means a major upstream dependency (os or fsnotify) is broken
		// If we had asserts this would be an assert instead
		slog.Error("file cache suffered a fatal error", "error", err)
		panic(err)
	}
}

func (cache *FileCache) initHelper() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if cache.Path == "" {
		path, err := os.Getwd()
		if err != nil {
			return err
		}
		cache.Path = path
	}

	if err := watcher.Add(cache.Path); err != nil {
		return fmt.Errorf("adding path to watch list: %w", err)
	}
	if err := filepath.WalkDir(cache.Path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if err := watcher.Add(path); err != nil {
				return fmt.Errorf("adding path to watch list: %w", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walking directories: %w", err)
	}

	cache.watcher = watcher
	// NOTE: Prevents a race condition where Run tries to grab this before cache.work sets it
	cache.ok.Store(true)
	go cache.work()
	return nil

}

func (cache *FileCache) work() {
	// TODO: Respect context?
	for {
		select {
		case event, ok := <-cache.watcher.Events:
			if !ok || filepath.Ext(event.Name) != ".go" || (!event.Has(fsnotify.Write) &&
				!event.Has(fsnotify.Create) &&
				!event.Has(fsnotify.Rename)) {
				continue
			}
			cache.ok.Store(true)
		case err, ok := <-cache.watcher.Errors:
			if !ok {
				continue
			}
			// TODO: I don't think this triggers so letting it panic
			panic(err)
		}
	}
}

func (cache *FileCache) Run(ctx context.Context, args RunArgs) error {
	cache.once.Do(func() { cache.init() })
	if cache.ok.Swap(false) {
		return nil
	}
	return ErrSkip
}
