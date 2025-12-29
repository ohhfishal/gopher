package runtime

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ohhfishal/nibbles/assert"
	"golang.org/x/time/rate"
)

/*
Cache that stops additional runners from running until a file has been chaged.
*/
type FileCache struct {
	Extensions []string      // List of extensions to watch for changes.
	Path       string        // Directory to watch for file changes. If empty, defaults to [os.Getwd].
	Interval   time.Duration // Minimum duration between updates.
}

func (cache *FileCache) newWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if cache.Path == "" {
		path, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		cache.Path = path
	}

	if err := watcher.Add(cache.Path); err != nil {
		return nil, fmt.Errorf("adding path to watch list: %w", err)
	}
	if err := filepath.WalkDir(cache.Path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if err := watcher.Add(path); err != nil {
				return fmt.Errorf("adding path to watch list: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking directories: %w", err)
	}
	return watcher, nil
}

// Event returns an event that yields when a file is changed.
func (cache *FileCache) Event() (Event, error) {
	watcher, err := cache.newWatcher()
	if err != nil {
		return nil, fmt.Errorf("adding files to watch: %w", err)
	}
	limiter := rate.NewLimiter(rate.Every(cache.Interval), 1)
	return func(yield func(_ any) bool) {
		for {
			select {
			case event, ok := <-watcher.Events:
				// filter out events we don't care about
				if !ok || !slices.Contains(cache.Extensions, filepath.Ext(event.Name)) {
					continue
				}
				if !limiter.Allow() {
					continue
				}
				// NOTE: This delay is to allow editors to fully write their changes
				time.Sleep(120 * time.Millisecond)
				if !yield(nil) {
					return
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				// TODO: I don't think this triggers so letting it panic
				assert.Unreachable(err.Error())
			}
		}
	}, nil
}
