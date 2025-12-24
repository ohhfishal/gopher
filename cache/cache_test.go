package cache_test

import (
	"testing"

	"github.com/ohhfishal/gopher/cache"
)

func TestHashDeterministic(t *testing.T) {
	var content = []byte("Testing123")

	hash1 := string(cache.Hash(content))
	hash2 := string(cache.Hash(content))
	if hash1 != hash2 {
		t.Errorf("%s != %s", hash1, hash2)
	}
}
