//go:build gopher

package main

import (
	"context"
	. "github.com/ohhfishal/gopher/runtime"
	"os"
	"time"
)

// Devel builds the gopher binary then runs it
func Devel(ctx context.Context, args RunArgs) error {
	return Run(ctx, NowAnd(OnFileChange(1*time.Second, ".go")),
		&Printer{},
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
			Output: "target/cicd",
		},
		&GoFormat{
			CheckOnly: true,
		},
		&GoTest{},
		&GoVet{},
		ExecCommand("echo", "CICD OK"),
	)
}

// Removes all local build artifacts.
func Clean(ctx context.Context, args RunArgs) error {
	return os.RemoveAll("target")
}
