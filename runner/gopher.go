package runner

import (
	"context"
	// "errors"
	// "fmt"
	"os"
)

type Gopher struct {
}

func (gopher *Gopher) Run(ctx context.Context, event RunEvent, runners ...Runner) error {
	for range event {
		for _, runner := range runners {
			if ctx.Err() != nil {
				return nil
			}

			err := runner.Run(ctx, RunArgs{
				// GoBin: gopher.GoBin,
				// TODO: FIX HACK
				GoConfig: GoConfig{
					GoBin: "go",
				},
				Stdout: os.Stdout,
			})
			if err != nil {
				break
			}

			// if errors.Is(ErrOK, err) {
			// 	// TODO: ????
			// 	// Eventually print Go Build: OK
			// 	fmt.Println("OK")
			// } else if err != nil {
			// 	fmt.Println(err)
			// }
		}
	}
	return nil
}
