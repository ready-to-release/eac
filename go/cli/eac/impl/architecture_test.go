//go:build L1

package impl_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoDirectExecForTooledBinaries ensures that git, gh, and go commands
// are routed through the tool system (tool.GlobalToolSystem().RunTool)
// rather than using direct exec.Command calls.
//
// This architectural constraint ensures consistent command routing via the
// tool registry's executor-mode configuration.
func TestNoDirectExecForTooledBinaries(t *testing.T) {
	implDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to get impl directory: %v", err)
	}

	// Binaries that must use the tool system
	tooledBinaries := map[string]string{
		"git": "tool.GlobalToolSystem().RunTool",
		"gh":  "tool.GlobalToolSystem().RunTool",
		"go":  "tool.GlobalToolSystem().RunTool",
	}

	// Files/patterns exempt from this rule
	exemptions := map[string]string{
		// Test files can use direct exec for setup
		"_test.go": "test setup",
		// This architecture test itself
		"architecture_test.go": "architecture test",
		// Docker login needs stdin pipe (Category C from plan)
		"build/builders/buildx.go": "docker login needs stdin pipe",
	}

	var violations []string

	err = filepath.Walk(implDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Check exemptions
		relPath, _ := filepath.Rel(implDir, path)
		for pattern, _ := range exemptions {
			if strings.HasSuffix(relPath, pattern) || strings.Contains(relPath, pattern) {
				return nil
			}
		}

		// Parse the file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		// Check if file imports os/exec
		importsExec := false
		for _, imp := range node.Imports {
			if imp.Path.Value == `"os/exec"` {
				importsExec = true
				break
			}
		}

		if !importsExec {
			return nil
		}

		// Walk AST looking for exec.Command or exec.CommandContext calls
		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check for exec.Command or exec.CommandContext
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != "exec" {
				return true
			}

			if sel.Sel.Name != "Command" && sel.Sel.Name != "CommandContext" {
				return true
			}

			// Found exec.Command or exec.CommandContext - check first argument
			argIndex := 0
			if sel.Sel.Name == "CommandContext" {
				argIndex = 1 // Skip context argument
			}

			if len(call.Args) <= argIndex {
				return true
			}

			// Get the binary name from the first string argument
			lit, ok := call.Args[argIndex].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			// Remove quotes from string literal
			binary := strings.Trim(lit.Value, `"`)

			// Check if it's a tooled binary
			if helper, isTooled := tooledBinaries[binary]; isTooled {
				pos := fset.Position(call.Pos())
				violations = append(violations,
					fmt.Sprintf("%s:%d: use %s.Run() instead of exec.Command(\"%s\", ...)",
						relPath, pos.Line, helper, binary))
			}

			return true
		})

		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk impl directory: %v", err)
	}

	if len(violations) > 0 {
		t.Errorf("Found %d direct exec.Command calls for tooled binaries:\n\n%s\n\n"+
			"Use the tool system instead:\n"+
			"  tool.GlobalToolSystem().RunTool(ctx, \"git\", workDir, args...)\n"+
			"  tool.GlobalToolSystem().RunTool(ctx, \"gh\", workDir, args...)\n"+
			"  tool.GlobalToolSystem().RunTool(ctx, \"go\", workDir, args...)",
			len(violations), strings.Join(violations, "\n"))
	}
}
