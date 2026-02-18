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

	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

// godogParentComponentType returns the parent component type from the godog descriptor.
func godogParentComponentType() string {
	if desc := testrunners.GetDescriptor("godog"); desc != nil && desc.ParentComponentType != "" {
		return desc.ParentComponentType
	}
	return "go"
}

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

// findGodogRootAcrossModules searches all modules' Go component roots for a
// directory matching the qualifier that contains (or has subdirectories with)
// the godog test file. This handles the case where spec directories are named
// after the CLI (e.g., "eac_work") but the godog runners live in a different
// module (e.g., "commands" module at go/commands/repository/work/).
//
// Returns the candidate path (e.g., "go/commands/repository/work") so that
// FindTestRoot can check for godog_test.go directly or in subdirectories.
func findGodogRootAcrossModules(cfg *config.EACConfig, qualifier, godogTestFile, workspaceRoot string) string {
	parentType := godogParentComponentType()
	// First pass: find exact matches (godog_test.go directly in the candidate dir)
	for i := range cfg.Repository.Modules {
		module := &cfg.Repository.Modules[i]
		for _, comp := range module.Components {
			if comp == nil || comp.Root == "" || comp.Type != parentType {
				continue
			}
			// Check <componentRoot>/<qualifier>/godogTestFile
			candidate := filepath.ToSlash(filepath.Join(comp.Root, qualifier))
			if fileExists(filepath.Join(workspaceRoot, candidate, godogTestFile)) {
				goRunnerLog.Debugf("findGodogRootAcrossModules: found %s at %s (module %s)", godogTestFile, candidate, module.Moniker)
				return candidate
			}
			// Check if the component root itself matches the qualifier
			// (e.g., root="go/commands/test" and qualifier="test")
			if filepath.Base(comp.Root) == qualifier {
				if fileExists(filepath.Join(workspaceRoot, comp.Root, godogTestFile)) {
					goRunnerLog.Debugf("findGodogRootAcrossModules: root matches qualifier %s at %s (module %s)", qualifier, comp.Root, module.Moniker)
					return comp.Root
				}
			}
		}
	}
	// Second pass: find directory matches (qualifier dir exists but godog_test.go
	// is in subdirectories — e.g., eac_create where runners are in create/commit-message/)
	for i := range cfg.Repository.Modules {
		module := &cfg.Repository.Modules[i]
		for _, comp := range module.Components {
			if comp == nil || comp.Root == "" || comp.Type != parentType {
				continue
			}
			candidate := filepath.ToSlash(filepath.Join(comp.Root, qualifier))
			if dirExists(filepath.Join(workspaceRoot, candidate)) {
				goRunnerLog.Debugf("findGodogRootAcrossModules: dir exists at %s (module %s), deferring to subdirectory search", candidate, module.Moniker)
				return candidate
			}
		}
	}
	return ""
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// getRunnerSearchDirs returns the runner_search_dirs from the godog component kind.
func getRunnerSearchDirs(cfg *config.EACConfig) []string {
	if cfg.ComponentKinds == nil {
		return nil
	}
	godogKind := cfg.ComponentKinds.Get("godog")
	if godogKind == nil {
		return nil
	}
	return godogKind.RunnerSearchDirs
}

// findGodogRootForMoniker searches all modules' Go component roots and their
// runner_search_dirs for a directory matching the moniker that contains
// godog_test.go. Handles modules with no Go components whose godog tests live
// in another module (e.g., implicit-cli specs tested at go/cli/eac/specs/implicit-cli/).
func findGodogRootForMoniker(cfg *config.EACConfig, moniker, godogTestFile, workspaceRoot string) string {
	parentType := godogParentComponentType()
	searchDirs := getRunnerSearchDirs(cfg)
	for i := range cfg.Repository.Modules {
		module := &cfg.Repository.Modules[i]
		for _, comp := range module.Components {
			if comp == nil || comp.Root == "" || comp.Type != parentType {
				continue
			}
			// Check <componentRoot>/<moniker>/godogTestFile
			candidate := filepath.ToSlash(filepath.Join(comp.Root, moniker))
			if fileExists(filepath.Join(workspaceRoot, candidate, godogTestFile)) {
				goRunnerLog.Debugf("findGodogRootForMoniker: found %s at %s (module %s)", godogTestFile, candidate, module.Moniker)
				return candidate
			}
			// Check runner_search_dirs: <componentRoot>/<searchDir>/<moniker>/godogTestFile
			for _, dir := range searchDirs {
				if dir == "." {
					continue
				}
				candidate = filepath.ToSlash(filepath.Join(comp.Root, dir, moniker))
				if fileExists(filepath.Join(workspaceRoot, candidate, godogTestFile)) {
					goRunnerLog.Debugf("findGodogRootForMoniker: found %s at %s via search dir %s (module %s)", godogTestFile, candidate, dir, module.Moniker)
					return candidate
				}
			}
		}
	}
	return ""
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
