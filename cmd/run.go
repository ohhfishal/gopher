package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ohhfishal/gopher/cache"
	"github.com/ohhfishal/gopher/compile"
	"github.com/ohhfishal/gopher/example"
	"github.com/ohhfishal/gopher/runner"
)

type RunCMD struct {
	Target     string          `arg:"" default:"default" help:"Recipe to run."`
	List       bool            `short:"l" help:"List all targets then exit."`
	GoConfig   runner.GoConfig `embed:"" group:"Golang Flags"`
	GopherDir  string          `default:".gopher" help:"Directory to cache files gopher creates."`
	GopherFile string          `short:"C" default:"gopher.go" help:"File to read from. If gopher.go is not found, defaults to using examples/default.go. (See source code)"`
}

func (config *RunCMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	if err := BuildGopherIfNeeded(config.GopherFile, config.GopherDir, config.GoConfig.GoBin); err != nil {
		return err
	}

	// Following convention of passing in invokation cmd
	args := []string{"./" + compile.BinaryName}
	if config.List {
		args = append(args, "-l")
	}
	args = append(args, config.Target)

	cmd := &exec.Cmd{
		Path:   filepath.Join(config.GopherDir, compile.BinaryName),
		Stdout: stdout,
		Stderr: stdout,
		Args:   args,
	}
	logger.Debug("calling", "path", cmd.Path, "args", cmd.Args)

	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func BuildGopherIfNeeded(file string, directory string, goBin string) error {
	reader, err := GopherFile(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	ok, err := cache.Valid(file, directory, goBin)
	if err != nil {
		return fmt.Errorf("determining if cached: %w", err)
	}

	if ok {
		return nil
	}
	slog.Debug("needs to compile, compiling")
	if err := compile.Compile(content, directory, goBin); err != nil {
		return fmt.Errorf("compiling: %w", err)
	}
	return nil
}

func GopherFile(filepath string) (io.ReadCloser, error) {
	file, err := os.Open(filepath)
	if err != nil {
		// TODO: Probably should print a better error like: No gopher file found run "gopher bootstrap" to get started
		if filepath == DefaultFilePath {
			slog.Info("could not find gopher.go, using default", "err", err)
			return io.NopCloser(strings.NewReader(example.DefaultGopherFile)), nil
		}
		return nil, fmt.Errorf("failed to open file: %s: %w", filepath, err)
	}
	return file, nil
}
