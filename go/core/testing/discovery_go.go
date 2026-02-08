package testing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// DiscoverGoTestTags discovers Go test functions and their build tags.
// Public API - creates config internally.
func DiscoverGoTestTags(pkgPath string) ([]TestReference, error) {
	dc, err := NewDiscoveryConfig()
	if err != nil {
		return nil, err
	}
	return discoverGoTestTagsInPath(pkgPath, dc)
}

// discoverGoTestTagsInPath discovers Go test functions using provided config.
func discoverGoTestTagsInPath(pkgPath string, dc *DiscoveryConfig) ([]TestReference, error) {
	refs := []TestReference{}

	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == "testdata" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		if dc.IsRunnerFile(info.Name()) {
			return nil
		}

		fileRefs, err := parseGoTestFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		refs = append(refs, fileRefs...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return refs, nil
}

// parseGoTestFile parses a single Go test file.
func parseGoTestFile(filePath string) ([]TestReference, error) {
	fset := token.NewFileSet()

	// Parse with comments to get build tags
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Extract build tags
	tags := extractBuildTags(file)

	// Find all Test* functions
	refs := []TestReference{}
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if function name starts with Test
		if !strings.HasPrefix(funcDecl.Name.Name, "Test") {
			continue
		}

		// Check if it has testing.T parameter
		if !hasTestingParam(funcDecl) {
			continue
		}

		refs = append(refs, TestReference{
			FilePath: filePath,
			Type:     "gotest",
			TestName: funcDecl.Name.Name,
			Tags:     copyTags(tags),
		})
	}

	return refs, nil
}

// extractBuildTags extracts build constraint tags from file comments.
func extractBuildTags(file *ast.File) []string {
	tags := []string{}

	// Check all comment groups
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text

			// Check for //go:build directive
			if strings.HasPrefix(text, "//go:build ") {
				buildExpr := strings.TrimPrefix(text, "//go:build ")
				buildExpr = strings.TrimSpace(buildExpr)

				// Simple parsing: look for L0, L1, L2 tags
				// TODO: Handle complex expressions if needed
				if strings.Contains(buildExpr, "L0") {
					tags = append(tags, "@L0")
				} else if strings.Contains(buildExpr, "L1") {
					tags = append(tags, "@L1")
				} else if strings.Contains(buildExpr, "L2") {
					tags = append(tags, "@L2")
				}
			}

			// Also check old-style // +build
			if strings.HasPrefix(text, "// +build ") {
				buildExpr := strings.TrimPrefix(text, "// +build ")
				buildExpr = strings.TrimSpace(buildExpr)

				if strings.Contains(buildExpr, "L0") && !slices.Contains(tags, "@L0") {
					tags = append(tags, "@L0")
				} else if strings.Contains(buildExpr, "L1") && !slices.Contains(tags, "@L1") {
					tags = append(tags, "@L1")
				} else if strings.Contains(buildExpr, "L2") && !slices.Contains(tags, "@L2") {
					tags = append(tags, "@L2")
				}
			}
		}
	}

	return tags
}

// hasTestingParam checks if function has *testing.T parameter.
func hasTestingParam(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0 {
		return false
	}

	// Check first parameter
	param := funcDecl.Type.Params.List[0]

	// Check if it's *testing.T or *testing.B
	starExpr, ok := param.Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	selExpr, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	// Check if it's testing.T or testing.B
	return ident.Name == "testing" && (selExpr.Sel.Name == "T" || selExpr.Sel.Name == "B")
}

// copyTags creates a copy of tags slice.
func copyTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}
	copied := make([]string, len(tags))
	copy(copied, tags)
	return copied
}
