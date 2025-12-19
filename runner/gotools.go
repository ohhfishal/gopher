package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/ohhfishal/gopher/pretty"
)

var _ Runner = &GoBuild{}
var _ Runner = &GoTest{}
var _ Runner = &GoFormat{}

type GoBuild struct {
	Output   string
	Flags    []string
	Packages []string
}

type GoTest struct {
	Path string
}

type GoFormat struct {
}

func runGoTool(ctx context.Context, printer *pretty.Printer, args RunArgs, cmdArgs []string) error {
	slog.Debug("running command", "cmd", args.GoConfig.GoBin, "args", cmdArgs)
	cmd := exec.CommandContext(ctx, args.GoConfig.GoBin, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if printer != nil {
		printer.Done(err)
	}
	fmt.Fprint(args.Stdout, string(output))
	return err
}

func (build *GoBuild) Run(ctx context.Context, args RunArgs) error {
	printer := pretty.New(args.Stdout, "Go Build")
	printer.Start()

	cmdArgs := append([]string{"build"}, build.Flags...)
	if build.Output != "" {
		cmdArgs = append(cmdArgs, "-o", build.Output)
	}
	cmdArgs = append(cmdArgs, build.Packages...)

	return runGoTool(ctx, printer, args, cmdArgs)
}

func (test *GoTest) Run(ctx context.Context, args RunArgs) error {
	printer := pretty.New(args.Stdout, "Go Test")
	printer.Start()

	path := test.Path
	if path == "" {
		path = "./..."
	}
	cmdArgs := []string{"test", path}

	return runGoTool(ctx, printer, args, cmdArgs)
}

func (format *GoFormat) Run(ctx context.Context, args RunArgs) error {
	printer := pretty.New(args.Stdout, "Go Format")
	printer.Start()

	cmdArgs := []string{
		"fmt",
		"./...", // TODO: Extract this to be from the struct
	}

	return runGoTool(ctx, printer, args, cmdArgs)
}
