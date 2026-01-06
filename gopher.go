//go:build gopher

package main

import (
	"context"
	"os"
	"time"

	. "github.com/ohhfishal/gopher/runtime"
)

// Devel inits git hooks then builds the gopher binary then runs it
func Devel(ctx context.Context, gopher *Gopher) error {
	if err := InstallGitHook(gopher.Stdout, GitPreCommit, "go run . cicd"); err != nil {
		return err
	}
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
		// TODO: Find a way to hot-swap the binary so we can bootstrap outself
		// NOTE: Also maybe a "closer" interface to kill the process before rerunning
		// NOTE 25/12/20: Should be done via closing their context?
		status.Done(),
	)
}

// CICD runs the entire ci/cd suite
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

// Removes all local build artifacts.
func Clean(ctx context.Context, gopher *Gopher) error {
	return os.RemoveAll("target")
}
