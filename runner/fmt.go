package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"

	"github.com/ohhfishal/gopher/pretty"
)

var _ Runner = &GoFormat{}

type GoFormat struct {
}

func (format *GoFormat) Run(ctx context.Context, args RunArgs) (retErr error) {
	printer := pretty.New(args.Stdout, "Go Format")
	printer.Start()
	defer func() { printer.Done(retErr) }()

	cmdArgs := []string{
		"fmt",
		"./...", // TODO: Extract this to be from the struct
	}

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
