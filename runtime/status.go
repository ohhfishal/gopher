package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TODO: Think this only works on Linux
const clearCharacter = "\033[H\033[2J"

// [Runner] that prints well formatted before and after messages using [Status.Start] and [Status.Done]
type Status struct {
	lastStart time.Time
}

func (status *Status) Done() Runner {
	return RunnerFunc(func(ctx context.Context, gopher *Gopher) error {
		if status.lastStart.IsZero() {
			return errors.New("Done().Run called before a successful Start().Run")
		}
		_, err := fmt.Fprintf(gopher.Stdout,
			"---\nDone: %s\n",
			time.Now().Format(time.DateTime),
		)
		return err
	})
}

func (status *Status) Start() Runner {
	return RunnerFunc(func(ctx context.Context, gopher *Gopher) error {
		now := time.Now()
		// // TODO: Make this output better
		_, err := fmt.Fprintf(gopher.Stdout,
			"%sStarting: %s\n---\n",
			clearCharacter,
			now.Format(time.DateTime),
		)
		if err != nil {
			status.lastStart = time.Time{}
			return err
		}
		status.lastStart = now
		return nil
	})
}
