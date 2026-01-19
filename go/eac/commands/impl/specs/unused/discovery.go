// Package unused provides detection of unused godog step definitions.
package unused

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// ImplSpecsPair represents a mapping between step implementations and feature specs.
type ImplSpecsPair struct {
	ImplDir      string   // absolute path to impl directory
	SpecsPath    string   // absolute path to specs directory
	StepFiles    []string // step definition files (absolute paths)
	FeatureFiles []string // feature files (absolute paths)
	UsesInternal bool     // whether this pair uses shared internal steps
}

// DiscoverPairs finds all impl↔specs pairs by scanning modules with test-impl components.
// Each module's test-impl component defines where to find godog_test.go.
func DiscoverPairs(repoRoot string) ([]ImplSpecsPair, error) {
	// Load EAC config (properly merged with defaults) for path resolution
	eacCfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to load EAC config: %w", err)
	}

	var pairs []ImplSpecsPair

	// Iterate modules that have test-impl components
	for _, module := range eacCfg.Repository.Modules {
		comp, hasTestImpl := module.Components["test-impl"]
		if !hasTestImpl || comp == nil || comp.Root == "" {
			continue
		}

		implDir := filepath.Join(repoRoot, comp.Root)
		godogFile := filepath.Join(implDir, "godog_test.go")

		// Check if godog_test.go exists in this module's test-impl root
		if _, err := os.Stat(godogFile); os.IsNotExist(err) {
			continue
		}

		pair, err := parsePairFromGodogFile(godogFile, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", godogFile, err)
		}
		pairs = append(pairs, pair)
	}

	return pairs, nil
}

// parsePairFromGodogFile extracts pair information from a godog_test.go file.
func parsePairFromGodogFile(godogFile, repoRoot string) (ImplSpecsPair, error) {
	content, err := os.ReadFile(godogFile)
	if err != nil {
		return ImplSpecsPair{}, fmt.Errorf("failed to read file: %w", err)
	}

	implDir := filepath.Dir(godogFile)

	// Extract SpecsPath from the file
	specsPath, err := extractSpecsPath(string(content), repoRoot)
	if err != nil {
		return ImplSpecsPair{}, err
	}

	// Find step files in impl directory
	stepFiles, err := findStepFiles(implDir)
	if err != nil {
		return ImplSpecsPair{}, fmt.Errorf("failed to find step files: %w", err)
	}

	// Find feature files in specs directory
	featureFiles, err := findFeatureFiles(specsPath)
	if err != nil {
		return ImplSpecsPair{}, fmt.Errorf("failed to find feature files: %w", err)
	}

	// Check if this pair uses internal steps
	usesInternal := checkUsesInternalFromFiles(stepFiles)

	return ImplSpecsPair{
		ImplDir:      implDir,
		SpecsPath:    specsPath,
		StepFiles:    stepFiles,
		FeatureFiles: featureFiles,
		UsesInternal: usesInternal,
	}, nil
}

// extractSpecsPath parses the SpecsPath value from godog_test.go content.
// It extracts the specs-relative path and resolves it from repoRoot.
func extractSpecsPath(content, repoRoot string) (string, error) {
	// Match: SpecsPath: "../../../../specs/eac-commands",
	pattern := regexp.MustCompile(`SpecsPath:\s*"([^"]+)"`)
	matches := pattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		return "", fmt.Errorf("SpecsPath not found in file")
	}

	relativePath := matches[1]

	// Extract the specs/<module>/<feature> part from the relative path
	// The path always contains "specs/" - extract from that point
	specsIdx := strings.Index(relativePath, paths.SpecsDir+"/")
	if specsIdx == -1 {
		return "", fmt.Errorf("SpecsPath does not contain '%s/': %s", paths.SpecsDir, relativePath)
	}

	// Get the specs-relative path (e.g., "specs/eac-commands")
	specsRelPath := relativePath[specsIdx:]

	// Resolve from repo root
	absPath := filepath.Join(repoRoot, specsRelPath)

	// Verify path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("specs path does not exist: %s", absPath)
	}

	return absPath, nil
}

// findStepFiles finds all step*.go files in the given directory (recursively).
func findStepFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		// Match steps.go, steps_*.go but not *_test.go
		if strings.HasPrefix(name, "steps") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// findFeatureFiles finds all .feature files recursively in the given directory.
func findFeatureFiles(dir string) ([]string, error) {
	var files []string

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil // Return empty, not an error
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".feature") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// checkUsesInternalFromFiles checks if any step file imports the internal package.
func checkUsesInternalFromFiles(stepFiles []string) bool {
	for _, file := range stepFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if strings.Contains(string(content), `"github.com/ready-to-release/eac/go/eac/specs/internal"`) {
			return true
		}
	}
	return false
}

// GetInternalStepsFile returns the path to the shared internal steps file.
func GetInternalStepsFile(repoRoot string) string {
	return filepath.Join(repoRoot, "go", "eac", "specs", "internal", "steps.go")
}
