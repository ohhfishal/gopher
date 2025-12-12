package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/ohhfishal/gopher/compile"
	"github.com/ohhfishal/gopher/example"
	"github.com/ohhfishal/gopher/runner"
	konghelp "github.com/ohhfishal/kong-help"
)

var ErrDone = errors.New("program ready to exit")
var ErrNeedsCompile = errors.New("needs to compile gopherfile")

const DefaultFilePath = "gopher.go"

type Executable func() ([]byte, error)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	if err := Run(ctx, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func Run(ctx context.Context, stdout io.Writer, args []string) error {
	var exit bool
	var cmd CMD
	parser, err := kong.New(
		&cmd,
		kong.Exit(func(_ int) { exit = true }),
		konghelp.Help(),
		kong.BindTo(ctx, new(context.Context)),
		kong.BindTo(stdout, new(io.Writer)),
	)
	if err != nil {
		return err
	}

	parser.Stdout = stdout
	parser.Stderr = stdout

	context, err := parser.Parse(
		os.Args[1:],
	)
	if errors.Is(err, ErrDone) {
		return nil
	} else if err != nil || exit {
		return err
	}

	if cmd.Debug {
		cmd.LogConfig.Level = slog.LevelDebug
	}

	logger := cmd.LogConfig.NewLogger(stdout)
	if err := context.Run(logger); err != nil {
		logger.Error("failed to run", "error", err)
		return nil
	}
	return nil
}

type CMD struct {
	LogConfig  LogConfig       `embed:"" group:"Logging Flags:"`
	Debug      bool            `help:"Turn on debugging features."`
	Target     string          `arg:"" default:"default" help:"Recipe to run."`
	List       bool            `short:"l" help:"List all targets then exit."`
	GoConfig   runner.GoConfig `embed:"" group:"Golang Flags"`
	GopherDir  string          `default:".gopher" help:"Directory to cache files gopher creates."`
	GopherFile string          `short:"C" default:"gopher.go" help:"File to read from. If gopher.go is not found, defaults to using examples/default.go. (See source code)"`
}

func (config *CMD) Run(ctx context.Context, stdout io.Writer, logger *slog.Logger) error {
	if err := Load(config.GopherFile, config.GopherDir, config.GoConfig.GoBin); err != nil {
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

func Load(file string, directory string, goBin string) error {
	reader, err := GopherFile(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	if err := Cached(content); !errors.Is(ErrNeedsCompile, err) {
		return err
	} else if err != nil {
		slog.Debug("needs to compile, compiling")
		if err := compile.Compile(content, directory, goBin); err != nil {
			return fmt.Errorf("compiling: %w", err)
		}
	}
	return nil
}

func Cached(content []byte) error {
	slog.Warn("caching not implemented")
	return ErrNeedsCompile
}

func GopherFile(filepath string) (io.ReadCloser, error) {
	file, err := os.Open(filepath)
	if err != nil {
		if filepath == DefaultFilePath {
			slog.Debug("could not open gopher.go, using default", "err", err)
			return io.NopCloser(strings.NewReader(example.DefaultGopherFile)), nil
		}
		return nil, fmt.Errorf("failed to open file: %s: %w", filepath, err)
	}
	return file, nil
}

// TODO: Move?
// func (config *Config) Run(ctx context.Context, event RunEvent, runners ...Runner) error {
// 	for range event {
// 		for _, runner := range runners {
// 			if ctx.Err() != nil {
// 				return nil
// 			}
//
// 			err := runner.Run(ctx, RunArgs{
// 				GoBin:  config.GoBin,
// 				Stdout: os.Stdout,
// 			})
//
// 			if errors.Is(ErrOK, err) {
// 				// TODO: ????
// 				// Eventually print Go Build: OK
// 				fmt.Println("OK")
// 			} else if err != nil {
// 				fmt.Println(err)
// 			}
// 		}
// 	}
// 	return nil
// }
