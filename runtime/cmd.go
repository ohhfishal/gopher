package runtime

import (
	"context"
	"fmt"
	"os/exec"
)

/*
Runner that uses [exec.CommandContext] to create and run commands.
This needs to be done to ensure the command can be canceled and invoked
several times.
You may use [ExecCommand] to initialize the struct with a similar API to [exec.Command].
*/
type ExecCmdRunner struct {
	// TODO: Provide more options akin to exec.Shell
	Name       string   // Same as [exec.CommandContext]
	Args       []string // Same as [exec.CommandContext]
	Dir        string   // Same as [exec.CommandContext]
	HideOutput bool     // When true, does not print command output to [Gopher].Stdout
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

func (runner *ExecCmdRunner) Run(ctx context.Context, args *Gopher) error {
	cmd := exec.CommandContext(ctx, runner.Name, runner.Args...)
	cmd.Dir = runner.Dir
	output, err := cmd.CombinedOutput()
	if !runner.HideOutput {
		fmt.Fprint(args.Stdout, string(output))
	}
	if err != nil {
		return err
	}
	return nil
}
