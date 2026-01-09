package cache

import (
	"context"
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/ohhfishal/nibbles/assert"
	"log/slog"
)

var ErrFileChanged = errors.New("file changed")

// Returns a context that gets cancels using [ErrFileChanged] when the file is chcanged.
func WithFileCancel(ctx context.Context, file string) context.Context {
	newCtx, cancel := context.WithCancelCause(ctx)

	// TODO: This goroutine might be leaking
	go func() {
		err := watch(newCtx, file)
		cancel(err)

		select {
		case <-newCtx.Done():
			return
		default:
			assert.Unreachable("context not canceled despite call")
		}
	}()
	return newCtx
}

func watch(ctx context.Context, file string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(file); err != nil {
		return err
	}

	for {
		select {
		case event := <-watcher.Events:
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) && !event.Has(fsnotify.Rename) {
				continue
			}
			slog.Info("file changed returning error", "event", event)
			return ErrFileChanged
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
