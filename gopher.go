//go:build gopher

// This file is the default Gopher File loaded if none are found.

package main

import (
	"context"
	. "github.com/ohhfishal/gopher/runner"
	"os"
	"time"
)

// Devel builds the gopher binary then runs it
func Devel(ctx context.Context, args RunArgs) error {
	return Run(ctx, NowAnd(Every(10*time.Second)),
		&FileCache{},
		&Printer{},
		&GoBuild{
			Output: "target/dev",
		},
		&GoFormat{},
		&GoTest{},
		// TODO: Find a way to hot-swap the binary so we can bootstrap outself
		// NOTE: Also maybe a "closer" interface to kill the process before rerunning
		// ExecCommand("target/dev", "devel"),
		ExecCommand("echo", "---"),
		ExecCommand("echo", "DEVEL OK"),
	)
}

// cicd runs the entire ci/cd suite
func CICD(ctx context.Context, args RunArgs) error {
	return Run(ctx, Now(),
		&Printer{},
		&GoBuild{
			Output: "target/dev",
		},
		&GoFormat{
			CheckOnly: true,
		},
		&GoTest{},
		// TODO: Find a way to hot-swap the binary so we can bootstrap outself
		// NOTE: Also maybe a "closer" interface to kill the process before rerunning
		// ExecCommand("target/dev", "devel"),
		ExecCommand("echo", "CICD OK"),
	)
}

// Removes all local build artifacts.
func Clean(ctx context.Context, args RunArgs) error {
	return os.RemoveAll("target")
}
