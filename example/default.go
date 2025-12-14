//go:build gopher

/*
	This file is the default Gopher File loaded if none are specified.
	We use a go:build gopher tag to ignore this file when doing normal go builds.
*/

package main

import (
	"context"
	"fmt"
	. "github.com/ohhfishal/gopher/runner"
	"os"
	"time"
)

// Devel builds the app as you make changes.
func Devel(ctx context.Context, args RunArgs) error {
	// See runner.Run
	return Run(ctx, NowAnd(Every(3*time.Second)),
		&FileCache{}, // Only allows the next runner to run if go files have changed
		&GoBuild{ // Runner that wraps Go build
			Output: "target/dev",
		},
	)
}

// Prints hello world.
func Hello(ctx context.Context, args RunArgs) error {
	// Can write normal go code.
	_, err := fmt.Println("Hello world")
	return err
}

// Removes all local build artifacts.
func Clean(ctx context.Context, args RunArgs) error {
	return os.RemoveAll("target")
}
