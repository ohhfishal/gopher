//go:build gopher

// This file is the default Gopher File loaded if none are found.

package example

import (
	"context"
	_ "github.com/ohhfishal/gopher/runner"
	"time"
)

func init() {
	// Learn how to do this automatically?
	Register("devel", Devel, "Run a developer shell that rebuilds as you edit source.")
	Register("prod", Prod, "Run a developer shell that rebuilds as you edit source.")
	// RegisterDefault("devel")
}

// Builds the app as you make changes.
func Devel(ctx context.Context, gopher Gopher) error {
	return gopher.Run(ctx, NowAnd(Every(3*time.Second)),
		&GoBuild{
			Output: "target/dev",
		},
	)
}

// Performs static analysis, tests,and builds the final binary.
func Prod(ctx context.Context, gopher Gopher) error {
	return gopher.Run(ctx, Now(),
		&GoBuild{
			Output: "target/prod",
		},
	)
}

// func init() {
// 	// Maybe do automatically?
// 	gopher.Register("devel", Devel, "Run a developer shell that rebuilds as you edit source.")
// 	gopher.Register("prod", Prod, "Build the production binary")
// 	gopher.Register("deploy", Deploy, "Push the production binary to live server.")
// 	/* Now allows in the shell
// 	gopher devel -> Rebuild and test on changes
// 	gopher devel -h -> Print what the stage does. (--dry-run?)
// 	gopher prod -> One time creates binary ./target/prod/myApp
//
// 	gopher -l -> Prints all the registered targets
// 	*/
// }
//
//
// // Have build be any exported function with 0/1 arg and a err or no return
//
// func Devel(ctx context.Context) error {
// 	target := "./target/devel"
// 	if err := gopher.UseGo(GoConfig{
// 		Bin: "/some/path/go"
// 		AssertVersion: ">=1.21",
// 	}); err != nil {
// 		return err
// 	}
// 	return gopher.RunStagesWhen(ctx, SourceChanges(AtMost(1, 10 * time.Second)),
// 		GoBuild{
// 			Flags: "-json",
// 			LDFlags: fmt.Sprintf("-X main.version={version}", GitTag()),
// 			BuildDir: "./src",,
// 			Target: target,
// 		},
// 		GoTest{
// 			Dest: "./...",
// 		},
// 		GoFmt{
// 			Change: false,
// 		},
// 		func() error {
// 			gopher.Log("You can even write you own custom runners!")
// 			return nil
// 		},
//
// 	).Then(
// 		gopher.Shell(target, ...),
// 	)
// }
//
// func Prod(ctx context.Context) error {
// 	defer gopher.Assert(FileExists(target))
// 	return GoBuild{
// 		// Alternative call
// 	}.Run(ctx)
// }
//
// func Deploy(ctx context.Context) error {
// 	gopher = gopher.With(Verbosity, CommandPrompts)
// 	return gopher.RunStages(ctx, ...)
// }
