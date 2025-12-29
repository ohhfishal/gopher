package runtime

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/ohhfishal/gopher/pretty"
)

var _ Runner = &GoBuild{}
var _ Runner = &GoTest{}
var _ Runner = &GoFormat{}
var _ Runner = &GoVet{}
var _ Runner = &GoModTidy{}

/*
[GoBuild] implements the [Runner] interface and exec's `go build`.
*/
type GoBuild struct {
	Output   string   // Binary file produced. Effectively go build -o.
	Flags    []string // Any additional flags to be passed to go build
	Packages []string // Positional args. If empty, defaults to ["./..."].
}

/*
[GoTest] implements the [Runner] interface and exec's `go test`.
*/
type GoTest struct {
	Packages []string // Positional args. If empty, defaults to ["./..."].
}

/*
[GoFormat] implements the [Runner] interface and exec's `go fmt`.
*/
type GoFormat struct {
	CheckOnly bool     // If true, `gofmt -l` is used to exit non-zero if formatting is incorrect.
	Packages  []string // Positional args. If empty, defaults to ["./..."].
}

/*
[GoVet] implements the [Runner] interface and exec's `go vet`.
*/
type GoVet struct {
	Packages []string // Positional args. If empty, defaults to ["./..."].
}

/*
[GoModTidy] implements the [Runner] interface and exec's `go mod tidy`.
*/
type GoModTidy struct {
}

func runGoTool(ctx context.Context, stdout io.Writer, gopher Gopher, cmdArgs []string) error {
	gopher.Stdout = pretty.NewIndentedWriter(stdout, "  ")
	runner := &ExecCmdRunner{
		Name: gopher.GoConfig.GoBin,
		Args: cmdArgs,
	}
	return runner.Run(ctx, &gopher)
}

func (build *GoBuild) Run(ctx context.Context, args *Gopher) error {
	printer := pretty.New(args.Stdout, "Go Build")
	printer.Start()

	cmdArgs := append([]string{"build"}, build.Flags...)
	if build.Output != "" {
		cmdArgs = append(cmdArgs, "-o", build.Output)
	}
	cmdArgs = append(cmdArgs, build.Packages...)

	err := runGoTool(ctx, printer, *args, cmdArgs)
	printer.Done(err)
	return err
}

func (test *GoTest) Run(ctx context.Context, args *Gopher) error {
	printer := pretty.New(args.Stdout, "Go Test")
	printer.Start()

	packages := test.Packages
	if len(packages) == 0 {
		packages = append(packages, "./...")
	}
	cmdArgs := []string{"test"}
	cmdArgs = append(cmdArgs, packages...)

	err := runGoTool(ctx, printer, *args, cmdArgs)
	printer.Done(err)
	return err
}

func (vet *GoVet) Run(ctx context.Context, args *Gopher) error {
	printer := pretty.New(args.Stdout, "Go Vet")
	printer.Start()

	packages := vet.Packages
	if len(packages) == 0 {
		packages = append(packages, "./...")
	}
	cmdArgs := []string{"vet"}
	cmdArgs = append(cmdArgs, packages...)

	err := runGoTool(ctx, printer, *args, cmdArgs)
	printer.Done(err)
	return err
}

func (tidy *GoModTidy) Run(ctx context.Context, args *Gopher) error {
	printer := pretty.New(args.Stdout, "Go Mod Tidy")
	printer.Start()

	cmdArgs := []string{"mod", "tidy"}

	err := runGoTool(ctx, printer, *args, cmdArgs)
	printer.Done(err)
	return err
}

func (format *GoFormat) Run(ctx context.Context, args *Gopher) error {
	printer := pretty.New(args.Stdout, "Go Format")
	printer.Start()

	if format.CheckOnly {
		// TODO: This is a hack
		// slog.Debug("running command", "cmd", args.GoConfig.GoBin, "args", cmdArgs)
		cmd := exec.CommandContext(ctx, "gofmt", "-l", ".")
		outputBytes, err := cmd.CombinedOutput()
		output := string(outputBytes)
		if len(strings.TrimSpace(output)) != 0 {
			err := fmt.Errorf("%s", output)
			printer.Done(err)
			return err
		}
		if printer != nil {
			printer.Done(err)
		}
		return err
	}

	packages := format.Packages
	if len(packages) == 0 {
		packages = append(packages, "./...")
	}

	cmdArgs := []string{"fmt"}
	cmdArgs = append(cmdArgs, packages...)

	err := runGoTool(ctx, printer, *args, cmdArgs)
	printer.Done(err)
	return err
}
