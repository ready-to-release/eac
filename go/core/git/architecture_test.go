//go:build L1
// +build L1

package git

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoExecImport ensures that the core git package does NOT import os/exec.
// The core git package must use pure go-git only - exec.Command belongs in CLI layers.
// This is an architectural constraint enforced by test.
func TestNoExecImport(t *testing.T) {
	// Get the current package directory
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// If running from a different directory, find the git package
	if !strings.HasSuffix(pkgDir, "git") {
		pkgDir = filepath.Join(pkgDir, "go", "core", "git")
	}

	fset := token.NewFileSet()

	// Parse all Go files in the package (excluding test files)
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		t.Fatalf("Failed to read package directory: %v", err)
	}

	forbiddenImports := []string{
		"os/exec",
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip test files - they may legitimately use exec for setup
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		filePath := filepath.Join(pkgDir, name)
		f, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", name, err)
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			for _, forbidden := range forbiddenImports {
				if importPath == forbidden {
					t.Errorf("ARCHITECTURE VIOLATION: %s imports %q\n"+
						"The core git package must use pure go-git only.\n"+
						"exec.Command operations belong in go/clibase/git/ or go/cli/eac/impl/.\n"+
						"See out/git-abstraction-layer-plan.md for layer separation.",
						name, forbidden)
				}
			}
		}
	}
}

// TestNoExecUsage ensures that exec.Command is never called directly.
// This catches cases where exec might be used without explicit import (unlikely but possible via dot imports).
func TestNoExecUsage(t *testing.T) {
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	if !strings.HasSuffix(pkgDir, "git") {
		pkgDir = filepath.Join(pkgDir, "go", "core", "git")
	}

	fset := token.NewFileSet()

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		t.Fatalf("Failed to read package directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		filePath := filepath.Join(pkgDir, name)
		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", name, err)
		}

		// Walk the AST looking for exec.Command calls
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			// Check for exec.Command or exec.CommandContext
			if ident.Name == "exec" && (sel.Sel.Name == "Command" || sel.Sel.Name == "CommandContext") {
				pos := fset.Position(call.Pos())
				t.Errorf("ARCHITECTURE VIOLATION: %s:%d uses exec.%s\n"+
					"The core git package must use pure go-git only.\n"+
					"exec.Command operations belong in go/clibase/git/ or go/cli/eac/impl/.",
					name, pos.Line, sel.Sel.Name)
			}

			return true
		})
	}
}
