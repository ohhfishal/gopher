//go:build gopher

// This file is the default Gopher File loaded if none are found.

package main

import (
	"context"
	. "github.com/ohhfishal/gopher/runner"
	"os"
	"time"
)

// TODO: Maybe create a Cached Runner, so they can make changes.
//       That way go fmt doesn't trigger another build?
// func CacheExampel(ctx context.Context, args RunArgs) error {
// 	return Run(ctx, NowAnd(Every(3*time.Second)),
// 		Cached(
// 			&GoBuild{
// 				Output: "target/dev",
// 			},
// 			&GoFormat{},
// 		),
// 	)
// }

// Devel builds the gopher binary then runs it
func Devel(ctx context.Context, args RunArgs) error {
	return Run(ctx, NowAnd(Every(3*time.Second)),
		&FileCache{},
		&GoBuild{
			Output: "target/dev",
		},
		&GoFormat{},
		// TODO: Implement cmd runner so gopher can bootstrap itself during development
		// NOTE: Also requires a "closer" interface to kill the process before rerunning
	)
}

// Removes all local build artifacts.
func Clean(ctx context.Context, args RunArgs) error {
	return os.RemoveAll("target")
}
