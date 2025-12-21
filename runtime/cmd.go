package runtime

import (
	"context"
	"fmt"
	"os/exec"
)

/*
Uses [exec.CommandContext] to create and run commands.
*/
type ExecCmdRunner struct {
	// TODO: Provide more options akin to exec.Shell
	Name       string   // Same as [exec.CommandContext]
	Args       []string // Same as [exec.CommandContext]
	HideOutput bool     // When true, does not print command output to [RunArgs].Stdout
}

/*
Shorthand syntax for creating a [ExecCmdRunner] that exposes only the same parameters as [exec.Command].
*/
func ExecCommand(name string, args ...string) Runner {
	return &ExecCmdRunner{
		Name: name,
		Args: args,
	}
}

func (runner *ExecCmdRunner) Run(ctx context.Context, args RunArgs) error {
	cmd := exec.CommandContext(ctx, runner.Name, runner.Args...)
	output, err := cmd.CombinedOutput()
	// if !runner.HideOutput {
	// 	fmt.Fprint(args.Stdout, string(output))
	// }
	fmt.Fprint(args.Stdout, string(output))
	if err != nil {
		return err
	}
	return nil
}
