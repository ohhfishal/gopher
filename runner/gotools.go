package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/ohhfishal/gopher/pretty"
)

var _ Runner = &GoBuild{}
var _ Runner = &GoTest{}
var _ Runner = &GoFormat{}
var _ Runner = &GoVet{}

type GoBuild struct {
	Output   string
	Flags    []string
	Packages []string
}

type GoTest struct {
	Path string
}

type GoFormat struct {
	CheckOnly bool
	Path      string
}

type GoVet struct {
	Packages []string
}

func runGoTool(ctx context.Context, printer *pretty.Printer, args RunArgs, cmdArgs []string) (string, error) {
	slog.Debug("running command", "cmd", args.GoConfig.GoBin, "args", cmdArgs)
	cmd := exec.CommandContext(ctx, args.GoConfig.GoBin, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if printer != nil {
		printer.Done(err)
	}
	return string(output), err
}

func (build *GoBuild) Run(ctx context.Context, args RunArgs) error {
	printer := pretty.New(args.Stdout, "Go Build")
	printer.Start()

	cmdArgs := append([]string{"build"}, build.Flags...)
	if build.Output != "" {
		cmdArgs = append(cmdArgs, "-o", build.Output)
	}
	cmdArgs = append(cmdArgs, build.Packages...)

	output, err := runGoTool(ctx, printer, args, cmdArgs)
	fmt.Fprint(args.Stdout, output)
	return err
}

func (test *GoTest) Run(ctx context.Context, args RunArgs) error {
	printer := pretty.New(args.Stdout, "Go Test")
	printer.Start()

	path := test.Path
	if path == "" {
		path = "./..."
	}
	cmdArgs := []string{"test", path}

	output, err := runGoTool(ctx, printer, args, cmdArgs)
	fmt.Fprint(args.Stdout, output)
	return err
}

func (vet *GoVet) Run(ctx context.Context, args RunArgs) error {
	printer := pretty.New(args.Stdout, "Go Vet")
	printer.Start()

	packages := vet.Packages
	if len(packages) == 0 {
		packages = append(packages, "./...")
	}
	cmdArgs := []string{"vet"}
	cmdArgs = append(cmdArgs, packages...)

	output, err := runGoTool(ctx, printer, args, cmdArgs)
	fmt.Fprint(args.Stdout, output)
	return err
}

func (format *GoFormat) Run(ctx context.Context, args RunArgs) error {
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

	path := format.Path
	if path == "" {
		path = "./..."
	}

	cmdArgs := []string{
		"fmt",
		path,
	}

	output, err := runGoTool(ctx, printer, args, cmdArgs)
	fmt.Fprint(args.Stdout, output)
	return err
}
