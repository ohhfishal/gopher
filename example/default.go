//go:build gopher

// This file is the default Gopher File loaded if none are found.

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
	return Run(ctx, NowAnd(Every(3*time.Second)),
		&GoBuild{
			Output: "target/dev",
		},
	)
}

// Prints hello world.
func Hello(ctx context.Context, args RunArgs) error {
	_, err := fmt.Println("Hello world")
	return err
}

// Removes all local build artifacts.
func Clean(ctx context.Context, args RunArgs) error {
	return os.RemoveAll("target")
}
