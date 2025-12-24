package runtime

import (
	"context"
	"fmt"
	"time"
)

// [Runner] that prints a nice status message when called.
type Printer struct {
}

// TODO: Think this only works on Linux
const clearCharacter = "\033[H\033[2J"

func (printer *Printer) Run(ctx context.Context, args *Gopher) error {
	// TODO: Make this output better
	_, err := fmt.Fprintf(args.Stdout,
		"%sStarting: %s\n---\n",
		clearCharacter,
		time.Now().Format(time.DateTime),
	)
	return err
}
