package cache_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/ohhfishal/gopher/cache"
	"github.com/ohhfishal/nibbles/assert"
)

func TestHashDeterministic(t *testing.T) {
	var content = []byte("Testing123")

	hash1 := string(cache.Hash(content))
	hash2 := string(cache.Hash(content))
	if hash1 != hash2 {
		t.Errorf("%s != %s", hash1, hash2)
	}
}

func TestContext(t *testing.T) {
	assert := assert.With(t)
	dir := t.TempDir()

	tempFile, err := os.CreateTemp(dir, "test_context-*.txt")
	assert.Nil(err)
	assert.Nil(tempFile.Close())

	t.Logf("file: %s", tempFile.Name())
	ctx := cache.WithFileCancel(t.Context(), tempFile.Name())
	assert.Nil(ctx.Err())

	// Give the goroutine time to work
	time.Sleep(250 * time.Millisecond)
	assert.Nil(os.WriteFile(tempFile.Name(), []byte("Hello world"), 0777))
	time.Sleep(250 * time.Millisecond)

	select {
	case <-ctx.Done():
	default:
		assert.Unreachable("context is not done")
	}

	cause := context.Cause(ctx)
	assert.True(
		errors.Is(cause, cache.ErrFileChanged),
		"%v == nil", cause,
	)
}
