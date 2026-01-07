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
	"github.com/ohhfishal/gopher/pretty"
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
	return mainTemplate.Execute(writer, TemplateData{
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

func parseTargets(content []byte) ([]Target, []error, error) {
	tree, err := parser.ParseFile(
		token.NewFileSet(),
		TargetsFile,
		content,
		parser.ParseComments,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse file: %w", err)
	}

	targets := []Target{}
	warnings := []error{}
	for _, decl := range tree.Decls {
		node, ok := decl.(*ast.FuncDecl)
		if !ok || !node.Name.IsExported() {
			continue
		} else if err := isValidFunc(node); err != nil {
			warnings = append(warnings, err)
			continue
		}

		comment := "No target description provided."
		if node.Doc != nil {
			if len(node.Doc.List) == 1 {
				comment = strings.TrimPrefix(node.Doc.List[0].Text, "// ")
			} else {
				warnings = append(warnings, fmt.Errorf("not implemented: multi-line doc comment for targets"))
				continue
			}
		}

		targets = append(targets, Target{
			Name:        node.Name.Name,
			Description: comment,
		})
	}
	return targets, warnings, nil
}

func isValidFunc(fn *ast.FuncDecl) error {
	expected := funcSignature{
		Name:       fn.Name.String(),
		Parameters: []string{"context.Context", "*runtime.Gopher"},
		Returns:    []string{"error"},
	}
	signature := fromFuncDecl(fn)

	errTail := fmt.Errorf("\n\t  have: %s\n\t  want: %s", signature.String(), expected.String())

	if fn.Type.TypeParams != nil && fn.Type.TypeParams.NumFields() != 0 {
		return fmt.Errorf("expected 0 type parameters got: %d", fn.Type.TypeParams.NumFields())
	} else if len(signature.Parameters) != len(expected.Parameters) {
		return fmt.Errorf("expected %d parameters got: %d\n%w",
			len(expected.Parameters),
			len(signature.Parameters),
			errTail,
		)
	} else if len(signature.Returns) != len(expected.Returns) {
		return fmt.Errorf("expected %d return value got: %d\n%w",
			len(expected.Returns),
			len(signature.Returns),
			errTail,
		)
	}

	for i, param := range signature.Parameters {
		expectedParam := expected.Parameters[i]
		if param != expectedParam {
			return fmt.Errorf("param %d: expected %s: got: %s\n%w", i, expectedParam, param, errTail)
		}
	}

	for i, ret := range signature.Returns {
		expectedRet := expected.Returns[i]
		if ret != expectedRet {
			return fmt.Errorf("ret %d: expected %s: got: %s\n%w", i, expectedRet, ret, errTail)
		}
	}
	return nil

}

func getType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + getType(t.Elt)
	case *ast.StarExpr:
		return "*" + getType(t.X)
	case *ast.SelectorExpr:
		return getType(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + getType(t.Key) + "]" + getType(t.Value)
	case *ast.ChanType:
		return "chan " + getType(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return "func"
	default:
		return fmt.Sprintf("%T", expr)
	}
}
