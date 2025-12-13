//go:build gopher

// This file is the default Gopher File loaded if none are found.

package main

import (
	"context"
	. "github.com/ohhfishal/gopher/runner"
	"os"
	"time"
)

// Devel builds the app as you make changes.
func Devel(ctx context.Context, args RunArgs) error {
	/// TODO: Write a runner that does this....
	// if !build.DisableCache && build.cache == nil {
	// 	pwd, _ := os.Getwd()
	// 	cache, err := cache.NewFileCache(pwd)
	// 	if err != nil {
	// 		return fmt.Errorf("creating file cache:", err)
	// 	}
	// 	build.cache = cache
	// 	return nil
	// }
	return Run(ctx, NowAnd(Every(3*time.Second)),
		&GoBuild{
			Output: "target/dev",
		},
		&GoFormat{},
	)
}

// Removes all local build artifacts.
func Clean(ctx context.Context, args RunArgs) error {
	return os.RemoveAll("target")
}
