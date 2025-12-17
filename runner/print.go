package runner

import (
	"context"
	"fmt"
	"time"
)

type Printer struct {
}

// TODO: Think this only works on Linux
const ClearCharacter = "\033[H\033[2J"

func (printer *Printer) Run(ctx context.Context, args RunArgs) error {
	_, err := fmt.Fprintf(args.Stdout,
		"%sStarting: %s\n---\n",
		ClearCharacter,
		time.Now().Format(time.DateTime),
	)
	return err
}
