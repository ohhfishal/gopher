//go:build gopher

package main

import (
	"context"
	"os"
	"time"

	. "github.com/ohhfishal/gopher/runtime"
)

/*
Devel inits git hooks then builds the gopher binary then runs it.
NOTE: This is hardcoded to be used for developers of gopher.
If you are reading this to try and learn, example/default.go may be more useful.
*/
func Devel(ctx context.Context, gopher *Gopher) error {
	if err := InstallGitHook(gopher.Stdout, GitPreCommit, "go run . cicd"); err != nil {
		return err
	}
	var status Status
	return gopher.Run(ctx, NowAnd(OnFileChange(1*time.Second, ".go")),
		status.Start(),
		&GoBuild{},
		&GoFormat{},
		&GoTest{},
		&GoVet{},
		&GoModTidy{},
		status.Done(),
	)
}

/*
CICD runs the entire ci/cd suite.
It should be installed as a pre-commit hook using the Devel target.
This target does not change any code, but instead returns an error if anything
is incorrect.
*/
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

// Calls devel.
func Default(ctx context.Context, gopher *Gopher) error {
	// By defaut gopher tries to call the target "default"
	return Devel(ctx, gopher)
}
