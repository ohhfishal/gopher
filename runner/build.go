package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/ohhfishal/gopher/cache"
	"github.com/ohhfishal/gopher/pretty"
)

var _ Runner = &GoBuild{}

type GoBuild struct {
	Output       string
	Flags        []string
	Packages     []string
	DisableCache bool
	cache        *cache.Cache
}

func (build *GoBuild) Run(ctx context.Context, args RunArgs) (retErr error) {
	if !build.DisableCache && build.cache == nil {
		pwd, _ := os.Getwd()
		cache, err := cache.NewFileCache(pwd)
		if err != nil {
			return fmt.Errorf("creating file cache:", err)
		}
		build.cache = cache
		return nil
	}
	if build.cache != nil && !build.cache.Ready() {
		return nil
	}
	printer := pretty.New(args.Stdout, "Go Build")
	printer.Start()
	defer func() { printer.Done(retErr) }()

	var output = build.Output
	if output == "" {
		output = "testBin"
	}
	cmdArgs := []string{
		"build",
	}
	cmdArgs = append(cmdArgs, build.Flags...)
	cmdArgs = append(cmdArgs, "-o", output)
	cmdArgs = append(cmdArgs, build.Packages...)

	slog.Debug("running command", "cmd", args.GoConfig.GoBin, "args", cmdArgs)
	cmd := exec.CommandContext(ctx, args.GoConfig.GoBin, cmdArgs...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("getting stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	slurp, err := io.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("reading stderr")
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s", slurp)
	}
	return nil
}
