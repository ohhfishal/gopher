package cmd

import (
	"context"
	"errors"
	"io"
	"log/slog"
)

type BootstrapCMD struct {
}

func (config *BootstrapCMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	/*
		TODO:
			- Create gopher file at root of git dir
			- add .gopher to gitignore
			- Go mod init? (Do in compile?) and get the runner installed
	*/
	return errors.New("not implemented: bootstrap command")
}
