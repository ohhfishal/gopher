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
	return errors.New("not implemented: bootstrap command")
}
