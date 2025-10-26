package watch

import (
	"context"
	"io"
	"log/slog"
)

type CMD struct {
}

func (cmd *CMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	return nil
}
