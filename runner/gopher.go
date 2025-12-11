package runner

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/ohhfishal/gopher/compile"
	"github.com/ohhfishal/gopher/example"
	"io"
	"log/slog"
	"os"
	"strings"
)

var ErrNeedsCompile = errors.New("needs to compile gopherfile")

const DefaultFilePath = "gopher.go"

type Executable func() error

type Gopher struct {
	// TODO: I don't think this goes here? Gopher is the compiler?
	GoBin  string       `default:"go" help:"Go binary to use for commands."`
	Logger *slog.Logger `kong:"-"`
}

func (gopher *Gopher) Load(file string, directory string) (Executable, error) {
	reader, err := gopher.GopherFile(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if err := gopher.Cached(content); !errors.Is(ErrNeedsCompile, err) {
		return nil, err
	} else if err != nil {
		gopher.Logger.Debug("needs to compile, compiling")
		if err := compile.Compile(content, directory); err != nil {
			return nil, fmt.Errorf("compiling: %w", err)
		}
	}
	return gopher.Executable()
}

func (gopher *Gopher) Executable() (Executable, error) {
	return nil, errors.New("not implemented: exec")
}

func (gopher *Gopher) Cached(content []byte) error {
	gopher.Logger.Warn("caching not implemented")
	return ErrNeedsCompile
}

func (gopher *Gopher) GopherFile(filepath string) (io.ReadCloser, error) {
	file, err := os.Open(filepath)
	if err != nil {
		if filepath == DefaultFilePath {
			gopher.Logger.Debug("could not open gopher.go, using default", "err", err)
			return io.NopCloser(strings.NewReader(example.DefaultGopherFile)), nil
		}
		return nil, fmt.Errorf("failed to open file: %s: %w", filepath, err)
	}
	return file, nil
}

// TODO: Move?
func (gopher *Gopher) Run(ctx context.Context, event RunEvent, runners ...Runner) error {
	for range event {
		for _, runner := range runners {
			if ctx.Err() != nil {
				return nil
			}

			err := runner.Run(ctx, RunArgs{
				GoBin:  gopher.GoBin,
				Stdout: os.Stdout,
			})

			if errors.Is(ErrOK, err) {
				// TODO: ????
				// Eventually print Go Build: OK
				fmt.Println("OK")
			} else if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
