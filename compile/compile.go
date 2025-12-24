package compile

import (
	"context"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ohhfishal/gopher/cache"
	"github.com/ohhfishal/gopher/runtime"
)

var funcMap = template.FuncMap{
	"lower": strings.ToLower,
}

//go:embed template.go.tmpl
var rawMainTemplate string
var mainTemplate = template.Must(template.New("main.go").
	Funcs(funcMap).
	Parse(rawMainTemplate))

const BinaryName = "target"
const TargetsFile = "targets.go"

type Target struct {
	Name        string
	Description string
}

// Compile aa gopher binary using the provided dependencies. Note dir is assumed to exist when called.
func Compile(stdout io.Writer, reader io.Reader, dir string, goBin string) (retErr error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, TargetsFile), content, 0660); err != nil {
		return fmt.Errorf("copying over file: %w", err)
	}

	// Extract info on targets for generating main.go
	targets, err := parseTargets(content)
	if err != nil {
		return fmt.Errorf("parsing targets: %w", err)
	} else if len(targets) == 0 {
		return fmt.Errorf("must include at least one target: %v", targets)
	}
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
	if err := formatter.Run(context.TODO(), &runtime.Gopher{
		GoConfig: runtime.GoConfig{
			GoBin: goBin,
		},
		Stdout: stdout,
	}); err != nil {
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
	return mainTemplate.Execute(writer, TemplateData{
		Targets: targets,
	})
}

func parseTargets(content []byte) ([]Target, error) {
	tree, err := parser.ParseFile(
		token.NewFileSet(),
		TargetsFile,
		content,
		parser.ParseComments,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var targets []Target
	for _, decl := range tree.Decls {
		node, ok := decl.(*ast.FuncDecl)
		if !ok || !node.Name.IsExported() || !isValidFunc(node) {
			continue
		}

		targets = append(targets, Target{
			Name:        node.Name.Name,
			Description: "TODO IMPLEMENT PARSING COMMENT",
		})
	}
	return targets, nil
}

func isValidFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || fn.Type.Params.NumFields() != 2 {
		return false
	}

	if fn.Type.Results == nil || fn.Type.Results.NumFields() != 1 {
		return false
	}
	// TODO: Do more validation to make this more robust
	return true
}
