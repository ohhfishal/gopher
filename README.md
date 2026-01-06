# gopher

Declarative, Golang Makefile that sits on your directory. Run `go build` and other tools while you work.

I wanted something like [magefile](https://github.com/magefile/mage) that is *declarative* and closer to Nix.

(My primary use case is to keep running go builds and tools so code is always ready for testing.)

## Getting Started

```
# Alternatively use go install
go get -tool github.com/ohhfishal/gopher

# Confirm installation
go tool gopher version

# Get a sample gopher.go file (Akin to makefile)
wget https://raw.githubusercontent.com/ohhfishal/gopher/refs/heads/main/example/default.go -O gopher.go

# Run the hello target to confirm everything is configured
go tool gopher hello
```

After that point, open `gopher.go` and add/edit targets as desired. All target functions must have exactly 2 parameters `context.Context` and `*gopher/runtime.Gopher`. (See example.)

## Example
See [example/default.go](example/default.go).
```go
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
```


## TODO
- [ ] More examples. (GitHub CICD?)
- [ ] Ctrl + R reset??
- [ ] Support more gotools
    - [ ] Add more options to those supported
- [ ] Better validate functions in gopher files (better errors)
- [ ] Comments in `gopher.go` visible when using `-l`
