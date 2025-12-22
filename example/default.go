//go:build gopher

// We use a build directive to prevent this file being included in your builds

package main

import (
	"context"
	"fmt"
	. "github.com/ohhfishal/gopher/runtime"
	"os"
	"time"
)

// Devel builds the app as you make changes.
func Devel(ctx context.Context, args RunArgs) error {
	// See runtime.Run
	return Run(ctx, NowAndOn(FileChanged(1*time.Second, ".go")),
		&Printer{}, // Prints an initial status message
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
