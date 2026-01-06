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

// gopher -l

// gopher hello

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

// Devel inits git hooks then builds the gopher binary then runs it
func Devel(ctx context.Context, gopher *Gopher) error {
	// Install a pre-commit hook if it does not exist
	if err := InstallGitHook(gopher.Stdout, GitPreCommit, "gopher cicd"); err != nil {
		return err
	}
	// Variable that let's us print a start and end message
	var status Status
	return gopher.Run(ctx, NowAnd(OnFileChange(1*time.Second, ".go")),
		status.Start(),
		&GoBuild{
			Output: "target/dev",
		},
		&GoFormat{},
		&GoTest{},
		&GoVet{},
		&GoModTidy{},
		status.Done(),
	)
}

// cicd runs the entire ci/cd suite
func CICD(ctx context.Context, gopher *Gopher) error {
	var status Status
	return gopher.Run(ctx, Now(),
		status.Start(),
		&GoBuild{
			Output: "target/cicd",
		},
		&GoFormat{
			CheckOnly: true,
		},
		&GoTest{},
		&GoVet{},
		status.Done(),
	)
}
