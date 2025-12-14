package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"

	"github.com/ohhfishal/gopher/pretty"
)

var _ Runner = &GoBuild{}
var _ Runner = &GoTest{}

type GoBuild struct {
	Output   string
	Flags    []string
	Packages []string
}

func (build *GoBuild) Run(ctx context.Context, args RunArgs) (retErr error) {
	printer := pretty.New(args.Stdout, "Go Build")
	printer.Start()
	defer func() { printer.Done(retErr) }()

	var output = build.Output
	if output == "" {
		output = "testBin"
	}
	cmdArgs := []string{
		"build",
	}
	cmdArgs = append(cmdArgs, build.Flags...)
	cmdArgs = append(cmdArgs, "-o", output)
	cmdArgs = append(cmdArgs, build.Packages...)

	slog.Debug("running command", "cmd", args.GoConfig.GoBin, "args", cmdArgs)
	cmd := exec.CommandContext(ctx, args.GoConfig.GoBin, cmdArgs...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("getting stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	slurp, err := io.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("reading stderr")
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s", slurp)
	}
	return nil
}

type GoTest struct {
	Path string
}

func (test *GoTest) Run(ctx context.Context, args RunArgs) (retErr error) {
	fmt.Fprintln(args.Stdout, "Running Go Test:")

	path := test.Path
	if path == "" {
		path = "./..."
	}
	cmdArgs := []string{
		"test",
		path,
	}

	slog.Debug("running command", "cmd", args.GoConfig.GoBin, "args", cmdArgs)
	cmd := exec.CommandContext(ctx, args.GoConfig.GoBin, cmdArgs...)

	output, err := cmd.CombinedOutput()
	fmt.Fprint(args.Stdout, string(output))
	return err
}
