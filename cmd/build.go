package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/ohhfishal/gopher/cache"
)

type GoBuild struct {
	Output   string
	Flags    []string
	Packages []string
	DisableCache bool
	cache *cache.Cache

}

func (build *GoBuild) Run(ctx context.Context, args RunArgs) error {
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

	var output = build.Output
	if output == "" {
		output = "testBin"
	}
	cmdArgs := []string{
		"build", "-o", output,
	}
	cmdArgs = append(cmdArgs, build.Flags...)
	cmdArgs = append(cmdArgs, build.Packages...)

	cmd := exec.CommandContext(ctx, args.GoBin, cmdArgs...)
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
		fmt.Fprintf(args.Stdout, "%s\n", slurp)
	} else {
		fmt.Fprintln(args.Stdout, "OK")
	}
	return nil
}

func NowAnd(when RunEvent) RunEvent {
	return func(yield func(any) bool) {
		for range Now() {
			if !yield(nil) {
				break
			}
		}
		for range when {
			if !yield(nil) {
				return
			}
		}
	}
}

func Now() RunEvent {
	return func(yield func(_ any) bool) {
		_ = yield(nil)
	}
}

func Every(duration time.Duration) RunEvent {
	ticker := time.NewTicker(duration)
	return func(yield func(_ any) bool) {
		defer ticker.Stop()
		for range ticker.C {
			if !yield(nil) {
				return
			}
		}
	}
}
