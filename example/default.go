//go:build gopher

// We use a build directive to prevent this file being included in your builds

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/ohhfishal/gopher/runtime"
)

// Devel builds the app as you make changes.
func Devel(ctx context.Context, gopher *Gopher) error {
	// See runtime.Gopher.Run
	return gopher.Run(ctx, NowAndOn(FileChanged(1*time.Second, ".go")),
		&Printer{}, // Prints an initial status message
		&GoBuild{ // Runner that wraps Go build
			Output: "target/dev",
		},
	)
}

// Prints hello world.
func Hello(ctx context.Context, _ *Gopher) error {
	// Can write normal go code.
	_, err := fmt.Println("Hello world")
	return err
}

// Removes all local build artifacts.
func Clean(ctx context.Context, _ *Gopher) error {
	return os.RemoveAll("target")
}
