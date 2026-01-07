package compile

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ohhfishal/gopher/cache"
	"github.com/ohhfishal/gopher/pretty"
	"github.com/ohhfishal/gopher/runtime"
)

//go:embed template.go
var mainSource string

var targetsTemplate = template.Must(template.New("main.go").
	Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).
	Parse(rawTargetsTemplate))

const rawTargetsTemplate = `
var targets = map[string]Target{
	   {{range .Targets}}

	   	{{printf "%q" .Name | lower }}: Target{
	   	  Name: {{printf "%q" .Name}},
	   	  Description: {{printf "%q" .Description}},
	   	  Func: {{ printf "%s" .Name}},
	   	},

	   {{end}}
}
`

const BinaryName = "target"
const TargetsFile = "targets.go"

type Target struct {
	Name        string
	Description string
}

// Compile a gopher binary using the provided dependencies. Note dir is assumed to exist when called.
func Compile(stdout io.Writer, reader io.Reader, dir string, goBin string) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	stdout = pretty.NewIndentedWriter(stdout, "  ")
	gopher := runtime.Gopher{
		GoConfig: runtime.GoConfig{
			GoBin: goBin,
		},
		Stdout: stdout,
	}

	if err := initGoModule(context.TODO(), gopher, dir); err != nil {
		return fmt.Errorf("init module: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, TargetsFile), content, 0660); err != nil {
		return fmt.Errorf("copying over file: %w", err)
	}

	// Extract info on targets for generating main.go
	printer := pretty.New(stdout, "Parsing Targets", pretty.Indent)
	printer.Start()
	targets, warnings, err := parseTargets(content)
	printer.Warn(warnings...)
	if err != nil {
		printer.Done(err)
		return fmt.Errorf("parsing targets: %w", err)
	} else if len(targets) == 0 {
		err := fmt.Errorf("must include at least one target: %v", targets)
		printer.Done(err)
		return err
	}
	printer.Done(nil)
	slog.Debug("parsed targets", "count", len(targets), "targets", targets)

	// Write main.go
	mainPath := filepath.Join(dir, "main.go")
	mainFile, err := os.OpenFile(mainPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("opening main.go: %w", err)
	}
	defer mainFile.Close()

	err = writeMain(mainFile, targets)
	if err != nil {
		return fmt.Errorf("writing main.go: %w", err)
	}

	formatter := runtime.GoFormat{Packages: []string{mainPath}}
	if err := formatter.Run(context.TODO(), &gopher); err != nil {
		return fmt.Errorf("formatting main.go: %w", err)
	}

	// Build gopher targets binary
	if err := buildBinary(stdout, dir, goBin); err != nil {
		return fmt.Errorf("building binary: %w", err)
	}

	// Write cache file
	if err := cache.WriteCacheMetadata(content, dir, goBin); err != nil {
		return fmt.Errorf("caching build metadata: %w", err)
	}
	return nil
}

func buildBinary(stdout io.Writer, dir string, goBin string) error {
	builder := runtime.GoBuild{
		Output:   BinaryName,
		Flags:    []string{"-C", dir},
		Packages: []string{"main.go", TargetsFile},
	}
	return builder.Run(context.TODO(), &runtime.Gopher{
		GoConfig: runtime.GoConfig{
			GoBin: goBin,
		},
		Stdout: stdout,
	})
}

type TemplateData struct {
	Targets []Target
}

func writeMain(writer io.Writer, targets []Target) error {
	fmt.Fprintln(writer, mainSource)
	return targetsTemplate.Execute(writer, TemplateData{
		Targets: targets,
	})
}

func initGoModule(ctx context.Context, gopher runtime.Gopher, dir string) (retErr error) {
	printer := pretty.New(gopher.Stdout, fmt.Sprintf("Initializing Go Module (%s)", dir))
	printer.Start()
	defer func() { printer.Done(retErr) }()
	stdout := pretty.NewIndentedWriter(printer, "  ")

	var output strings.Builder
	gopher.Stdout = &output
	// TODO: Validate that if there is an error its since go.mod already exists
	runner := &runtime.ExecCmdRunner{
		Name: gopher.GoConfig.GoBin,
		Args: []string{"mod", "init", "gopher-scripts"},
		Dir:  dir,
	}

	err := runner.Run(ctx, &gopher)
	fmt.Fprint(stdout, output.String())
	if err != nil && !strings.Contains(output.String(), "already exists") {
		return fmt.Errorf("unhandled error: %w: %s", err, output.String())
	} else if err != nil {
		// TODO: Update dependencies here?
		return nil
	}

	gopher.Stdout = stdout
	// TODO: HACK: Can be a lot smarter with this
	runner = &runtime.ExecCmdRunner{
		Name: gopher.GoConfig.GoBin,
		Args: []string{"get", "github.com/ohhfishal/gopher/runtime"},
		Dir:  dir,
	}
	if err := runner.Run(ctx, &gopher); err != nil {
		return fmt.Errorf("installing runtime dependencies: %w", err)
	}
	return nil
}

func NormalizeComment(comment string) string {
	switch {
	case strings.HasPrefix(comment, "//"):
		return strings.TrimPrefix(comment, "// ")
	case strings.HasPrefix(comment, "/*"):
		return strings.TrimSpace(
			strings.TrimSuffix(
				strings.TrimPrefix(comment, "/*"),
				"*/",
			),
		)
	default:
		return comment
	}
}
