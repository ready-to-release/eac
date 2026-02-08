package gotest

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

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

// runGoGenerate runs go generate for a package directory.
func runGoGenerate(ctx context.Context, pkgDir string, logWriter io.Writer) error {
	moduleRoot := findModuleRoot(pkgDir)
	if moduleRoot == "" {
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

// extractGoBuildTags extracts Go build tags from a godog-format suite tag filter string.
// Input: "@L0,@L1 && ~@skip:wip" or "@L0,@L1,@L2" or "@deps:gh-token"
// Output: "L0,L1" or "L0,L1,L2" or "L0,L1,deps_gh_token" (comma-separated Go build tags)
func extractGoBuildTags(suiteTagFilter string) string {
	if suiteTagFilter == "" {
		return ""
	}

	var tags []string

	// Look for L-level tags (@L0, @L1, @L2, @L3, @L4)
	for _, level := range []string{"L0", "L1", "L2", "L3", "L4"} {
		tag := "@" + level
		idx := 0
		for {
			pos := strings.Index(suiteTagFilter[idx:], tag)
			if pos == -1 {
				break
			}
			absPos := idx + pos
			if absPos == 0 || suiteTagFilter[absPos-1] != '~' {
				tags = append(tags, level)
				break
			}
			idx = absPos + len(tag)
		}
	}

	// Look for @deps:<name> tags and translate to deps_<name>
	depsPrefix := "@deps:"
	idx := 0
	for {
		pos := strings.Index(suiteTagFilter[idx:], depsPrefix)
		if pos == -1 {
			break
		}
		start := idx + pos + len(depsPrefix)
		end := start
		for end < len(suiteTagFilter) {
			c := suiteTagFilter[end]
			if c == ' ' || c == ',' || c == '&' || c == ')' {
				break
			}
			end++
		}
		if end > start {
			depName := suiteTagFilter[start:end]
			goBuildTag := "deps_" + strings.ReplaceAll(depName, "-", "_")
			tags = append(tags, goBuildTag)
		}
		idx = end
	}

	if len(tags) == 0 {
		return ""
	}

	return strings.Join(tags, ",")
}
