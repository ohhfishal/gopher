package runtime

import (
	"context"
	"os"
)

func ExampleGopher_Run() {
	var status Status
	gopher := Gopher{
		GoConfig: GoConfig{
			GoBin: "go",
		},
		Stdout: os.Stdout,
	}
	err := gopher.Run(context.Background(), Now(),
		status.Start(),
		&GoBuild{
			Output: "target/dev",
		},
		&GoFormat{},
		status.Done(),
	)
	if err != nil {
		panic(err)
	}
}

func ExampleGopher_RunNow() {
	var status Status
	gopher := Gopher{
		GoConfig: GoConfig{
			GoBin: "go",
		},
		Stdout: os.Stdout,
	}
	err := gopher.RunNow(context.Background(),
		status.Start(),
		&GoBuild{
			Output: "target/dev",
		},
		&GoFormat{},
		status.Done(),
	)
	if err != nil {
		panic(err)
	}
}
