package runner

import (
	"context"
	"fmt"
	"os/exec"
)

type ExecCmdRunner struct {
	Cmd *exec.Cmd
}

func ExecCommand(name string, arg ...string) Runner {
	return &ExecCmdRunner{
		Cmd: exec.Command(name, arg...),
	}
}

func (cmd *ExecCmdRunner) Run(ctx context.Context, args RunArgs) (retErr error) {
	// TODO: Use command context and copy over values
	output, err := cmd.Cmd.Output()
	fmt.Fprint(args.Stdout, string(output))
	if err != nil {
		return err
	}
	return nil
}
