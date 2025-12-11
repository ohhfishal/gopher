package cache

import (
	"io/fs"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"github.com/fsnotify/fsnotify"
)

type Cache struct {
	watcher *fsnotify.Watcher
	ok atomic.Bool
}

func (cache *Cache) Close() error {
	return cache.watcher.Close()
}

func NewFileCache(path string) (*Cache, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(path); err != nil {
		return nil, fmt.Errorf("adding path to watch list: %w", err)
	}
	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if err := watcher.Add(path); err != nil {
				return fmt.Errorf("adding path to watch list: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking directories: %w", err)
	}
	cache := &Cache{
		watcher: watcher,
	}
	go cache.work()
	return cache, nil
}

func (cache *Cache) work()  {
	// TODO: Respect context?
	cache.ok.Store(true)
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

func (cache *Cache) Ready() bool {
	return cache.ok.Swap(false)
}
