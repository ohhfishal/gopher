package cache_test

import (
	"github.com/ohhfishal/gopher/cache"
	"testing"
)

func TestHashDeterministic(t *testing.T) {
	var content = "Testing123"

	hash1 := cache.Hash(content)
	hash2 := cache.Hash(content)
	if hash1 != hash2 {
		t.Errorf("%s != %s", hash1, hash2)
	}
}
