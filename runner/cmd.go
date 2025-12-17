package runner

import (
	"context"
	"fmt"
	"os/exec"
)

/*
TODO: Provide more options akin to exec.Shell
*/
type ExecCmdRunner struct {
	Name string
	Args []string
}

func ExecCommand(name string, args ...string) Runner {
	return &ExecCmdRunner{
		Name: name,
		Args: args,
	}
}

func (runner *ExecCmdRunner) Run(ctx context.Context, args RunArgs) error {
	cmd := exec.CommandContext(ctx, runner.Name, runner.Args...)
	output, err := cmd.CombinedOutput()
	fmt.Fprint(args.Stdout, string(output))
	if err != nil {
		return err
	}
	return nil
}
