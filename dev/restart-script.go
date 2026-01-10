package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	if os.Getenv("GOPHER_DEV") == "" {
		fmt.Println("set $GOPHER_DEV")
		os.Exit(1)
	}
	for {
		cmd := exec.CommandContext(ctx, "./gopher", os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()
		if exitCode := cmd.ProcessState.ExitCode(); exitCode != 42 {
			os.Exit(exitCode)
		}
	}
}
