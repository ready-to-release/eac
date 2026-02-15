package gotest

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

// tagLevelRe matches positive (non-negated) @L0..@L4 tags.
var tagLevelRe = regexp.MustCompile(`(?:^|[^~])@(L[0-4])`)

// tagDepsRe matches @deps:<name> tags (letters, digits, hyphens).
var tagDepsRe = regexp.MustCompile(`@deps:([\w-]+)`)

// findModuleForPath finds the module moniker for a given relative path.
func findModuleForPath(relPath string, cfg *config.EACConfig) string {
	// Iterate through modules to find the one that owns this path
	for i := range cfg.Repository.Modules {
		module := &cfg.Repository.Modules[i]
		// Check all package roots
		for _, entry := range module.Components {
			if entry == nil || entry.Root == "" {
				continue
			}
			pkgRoot := filepath.ToSlash(entry.Root)
			if strings.HasPrefix(relPath, pkgRoot+"/") || relPath == pkgRoot {
				return module.Moniker
			}
		}
	}
	return ""
}

// extractFeatureFolderName extracts the feature folder name from a feature path.
// Input: "specs/repository/no-build-tags-in-steps/specification.feature"
// Output: "no-build-tags-in-steps".
func extractFeatureFolderName(featurePath string) string {
	featurePath = filepath.ToSlash(featurePath)
	// Remove the filename (specification.feature)
	dir := filepath.Dir(featurePath)
	// Get the last directory component (feature folder name)
	return filepath.Base(dir)
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// findModuleRoot walks up from dir to find the directory containing go.mod.
func findModuleRoot(dir string) string {
	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// hasGenerateDirectives checks whether any .go file under root contains a
// //go:generate directive. It walks the tree and scans each file line-by-line,
// returning true as soon as the first directive is found.
func hasGenerateDirectives(root string) (bool, error) {
	found := false
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if found {
			return filepath.SkipAll
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "//go:generate ") || line == "//go:generate" {
				found = true
				return filepath.SkipAll
			}
		}
		return scanner.Err()
	})
	return found, err
}

// runGoGenerate runs go generate for a package directory.
// It first checks whether any Go files contain //go:generate directives
// and skips execution when none are found.
func runGoGenerate(ctx context.Context, pkgDir string, logWriter io.Writer) error {
	moduleRoot := findModuleRoot(pkgDir)
	if moduleRoot == "" {
		return nil
	}

	// Skip go generate when no directives exist (OO-005).
	hasDirectives, err := hasGenerateDirectives(moduleRoot)
	if err != nil {
		return fmt.Errorf("scanning for go:generate directives: %w", err)
	}
	if !hasDirectives {
		return nil
	}

	toolDef := tool.GlobalRegistry().GetOrAdhoc("go")
	fullEnv := append(os.Environ(), "CLIE_TEST_LOGGING_ACTIVE=true")
	execCtx := &tool.ExecutionContext{
		ModuleRoot:    moduleRoot,
		FullEnv:       fullEnv,
		ArgsOverrides: []string{"generate", "./..."},
	}
	result, err := tool.GlobalExecutor().Execute(ctx, toolDef, execCtx)
	if result != nil && (len(result.Stdout) > 0 || len(result.Stderr) > 0) {
		fmt.Fprintf(logWriter, "go generate output:\n%s%s\n", result.Stdout, result.Stderr)
	}
	if err != nil {
		return err
	}
	if result != nil && result.ExitCode != 0 {
		return fmt.Errorf("go generate exited with code %d", result.ExitCode)
	}
	return nil
}

// insertBeforeLast inserts an element before the last element of a slice.
// Used to add flags before the package path in go test args.
func insertBeforeLast(args []string, elem string) []string {
	if len(args) == 0 {
		return []string{elem}
	}
	result := make([]string, 0, len(args)+1)
	result = append(result, args[:len(args)-1]...)
	result = append(result, elem)
	result = append(result, args[len(args)-1])
	return result
}

// extractGoBuildTags extracts Go build tags from a godog-format suite tag filter string.
// Input: "@L0,@L1 && ~@skip:wip" or "@L0,@L1,@L2" or "@deps:gh-token"
// Output: "L0,L1" or "L0,L1,L2" or "L0,L1,deps_gh_token" (comma-separated Go build tags)
//
// Only positive (non-negated) @L0..@L4 tags are included; ~@L2 is excluded.
// @deps:<name> tags are translated to deps_<name> with hyphens replaced by underscores.
func extractGoBuildTags(suiteTagFilter string) string {
	if suiteTagFilter == "" {
		return ""
	}

	var tags []string
	seen := map[string]bool{}

	// Match positive @L0..@L4 tags (not preceded by ~)
	for _, m := range tagLevelRe.FindAllStringSubmatch(suiteTagFilter, -1) {
		level := m[1]
		if !seen[level] {
			seen[level] = true
			tags = append(tags, level)
		}
	}

	// Match @deps:<name> tags and translate to deps_<name>
	for _, m := range tagDepsRe.FindAllStringSubmatch(suiteTagFilter, -1) {
		depName := m[1]
		goBuildTag := "deps_" + strings.ReplaceAll(depName, "-", "_")
		if !seen[goBuildTag] {
			seen[goBuildTag] = true
			tags = append(tags, goBuildTag)
		}
	}

	return strings.Join(tags, ",")
}
