package cache_test

import (
	"errors"
	"github.com/ohhfishal/gopher/cache"
	"strings"
	"testing"
)

func TestHashDeterministic(t *testing.T) {
	var content = "Testing123"

	hash1, err1 := cache.HashFrom(strings.NewReader(content))
	hash2, err2 := cache.HashFrom(strings.NewReader(content))
	if err := errors.Join(err1, err2); err != nil {
		t.Errorf("Error hashing files: %s", err)
	}
	if hash1 != hash2 {
		t.Errorf("%s != %s", hash1, hash2)
	}
}
