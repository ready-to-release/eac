// tscucumber.go - Test runner for TypeScript cucumber-js tests
package runners

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

func init() {
	Register(&TsCucumberRunner{})
}

// TsCucumberRunner handles TypeScript cucumber-js test execution.
type TsCucumberRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *TsCucumberRunner) TestTypes() []string {
	return []string{"tscucumber"}
}

// GetTestInfo extracts structured test metadata from a TypeScript cucumber test reference.
func (r *TsCucumberRunner) GetTestInfo(test testing.TestReference, workspaceRoot string, cfg *config.EACConfig) *TestInfo {
	// Calculate relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)

	info := &TestInfo{Language: "ts"}

	// Extract module moniker from specs path
	specsPrefix := cfg.Repository.Paths.SpecsRoot + "/"
	specRelPath := strings.TrimPrefix(relPath, specsPrefix)
	specRelPath = filepath.ToSlash(specRelPath)

	// Get module moniker from first path component
	parts := strings.Split(specRelPath, "/")
	if len(parts) == 0 {
		return nil
	}
	info.ModuleMoniker = parts[0]

	// Verify module exists
	if cfg.Repository.GetByMoniker(info.ModuleMoniker) == nil {
		return nil
	}

	// Find test root
	info.TestRoot = r.FindTestRoot(relPath, cfg)
	if info.TestRoot == "" {
		return nil
	}

	// Build package key and display name
	featureFolderName := extractTsFeatureFolderName(relPath)
	info.PackageKey = featureFolderName + ":" + info.TestRoot + ":" + relPath
	info.DisplayName = featureFolderName + ":" + info.TestRoot

	return info
}

// FindTestRoot finds the module root for a TypeScript cucumber feature file.
// The test runner (cucumber-js) is located in the module's root directory.
func (r *TsCucumberRunner) FindTestRoot(featurePath string, cfg *config.EACConfig) string {
	// Extract module moniker from specs path
	specsPrefix := cfg.Repository.Paths.SpecsRoot + "/"
	relPath := strings.TrimPrefix(filepath.ToSlash(featurePath), specsPrefix)
	relPath = strings.TrimPrefix(relPath, strings.ReplaceAll(specsPrefix, "/", "\\"))
	relPath = filepath.ToSlash(relPath)

	// Get module moniker (first path component)
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}
	moniker := parts[0]

	// Look up the module by moniker
	module, ok := cfg.Repository.GetModule(moniker)
	if !ok {
		return ""
	}

	// Return the module's root directory where cucumber-js should be
	return filepath.ToSlash(module.Files.Root)
}

// BuildPackagePath constructs the package path for test grouping.
// Returns "featureFolderName:moduleRoot:featurePath" format for cleaner display.
func (r *TsCucumberRunner) BuildPackagePath(testRoot string, featurePath string) string {
	if testRoot == "" {
		return ""
	}
	if featurePath != "" {
		// Extract feature folder name from path like:
		// "specs/some-module/feature-name/specification.feature"
		// -> "feature-name"
		featureFolderName := extractTsFeatureFolderName(featurePath)
		// Store full feature path after second colon for Execute() to use
		return featureFolderName + ":" + testRoot + ":" + featurePath
	}
	return testRoot
}

// extractTsFeatureFolderName extracts the feature folder name from a feature path.
func extractTsFeatureFolderName(featurePath string) string {
	featurePath = filepath.ToSlash(featurePath)
	// Remove the filename (specification.feature)
	dir := filepath.Dir(featurePath)
	// Get the last directory component (feature folder name)
	return filepath.Base(dir)
}

// Execute runs TypeScript cucumber-js tests for a package.
func (r *TsCucumberRunner) Execute(pkgPath string, tests []testing.TestReference, tuiWriter io.Writer, cfg RunConfig) RunResult {
	start := time.Now()

	// Parse package path - new format: "featureName:moduleRoot:featurePath" or "moduleRoot"
	var displayName, relPkgPath, relFeatureFile string
	parts := strings.Split(pkgPath, ":")
	if len(parts) == 3 {
		// BDD format: featureName:moduleRoot:featurePath
		// Display as "featureName:moduleRoot" (without full feature path)
		displayName = parts[0] + ":" + parts[1]
		relPkgPath = parts[1]     // moduleRoot
		relFeatureFile = parts[2] // full feature path
	} else if len(parts) == 1 {
		// Just the module path
		displayName = parts[0]
		relPkgPath = parts[0]
	} else {
		// Fallback for unexpected format
		displayName = pkgPath
		relPkgPath = pkgPath
	}

	result := RunResult{
		PackageName:   displayName,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	// moduleRoot is the TypeScript module directory
	moduleRoot := filepath.Join(cfg.WorkspaceRoot, relPkgPath)

	// Create log directory using module-based output path if available
	outputPath := cfg.ModuleOutputPath
	if outputPath == "" {
		outputPath = sanitizePathForLog(pkgPath)
	}
	logDir := filepath.Join(cfg.TestRunDir, outputPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(tuiWriter, "Failed to create log directory: %v\n", err)
		result.PackageFailed = true
		return result
	}

	// Create log file
	logFilePath := filepath.Join(logDir, "test.log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Fprintf(tuiWriter, "Failed to create log file: %v\n", err)
		result.PackageFailed = true
		return result
	}
	defer logFile.Close()
	result.LogFilePath = logFilePath

	// Check if package.json exists
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(tuiWriter, "No package.json found\n")
		fmt.Fprintf(logFile, "No package.json found at %s\n", packageJSON)
		result.PackageFailed = true
		return result
	}

	// Build cucumber-js command
	args := []string{"cucumber-js"}

	// Add cucumber.json output format
	cucumberJSONPath := filepath.Join(logDir, "cucumber.json")
	args = append(args, "--format", fmt.Sprintf("json:%s", cucumberJSONPath))

	// Add tag filter if provided
	if cfg.SuiteTagFilter != "" {
		tagExpr := convertToCucumberTagExpr(cfg.SuiteTagFilter)
		if tagExpr != "" {
			args = append(args, "--tags", tagExpr)
		}
	}

	// Add the specific feature file if provided
	if relFeatureFile != "" {
		featurePath := filepath.Join(cfg.WorkspaceRoot, relFeatureFile)
		relPath, err := filepath.Rel(moduleRoot, featurePath)
		if err == nil {
			args = append(args, relPath)
		}
	}

	// Log command
	fmt.Fprintf(logFile, "=== Testing TypeScript cucumber specs ===\n")
	fmt.Fprintf(logFile, "Module root: %s\n", moduleRoot)
	fmt.Fprintf(logFile, "Command: npx %s\n\n", strings.Join(args, " "))

	// Execute npx cucumber-js
	wrappedName, wrappedArgs := platform.WrapCommand("npx", args...)
	cmd := exec.Command(wrappedName, wrappedArgs...)
	cmd.Dir = moduleRoot
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "R2R_TEST_LOGGING_ACTIVE=true")

	// Capture output
	output, runErr := cmd.CombinedOutput()
	fmt.Fprintf(logFile, "%s\n", output)

	// Parse results
	if runErr != nil {
		result.PackageFailed = true
		result.TestsFailed = len(tests)
		fmt.Fprintf(tuiWriter, "cucumber-js failed\n")
	} else {
		result.TestsPassed = len(tests)
		fmt.Fprintf(tuiWriter, "cucumber-js passed\n")
	}

	result.TestsTotal = len(tests)
	result.Duration = time.Since(start)

	return result
}

// convertToCucumberTagExpr converts godog tag expression to cucumber-js tag expression.
// Godog format: "@L0,@L1 && ~@skip:broken"
// Cucumber format: "(@L0 or @L1) and not @skip:broken"
func convertToCucumberTagExpr(godogTags string) string {
	if godogTags == "" {
		return ""
	}

	var parts []string
	// Split by && for AND conditions
	andParts := strings.Split(godogTags, " && ")
	for _, part := range andParts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "~") {
			// Negation: ~@tag -> not @tag
			parts = append(parts, "not "+strings.TrimPrefix(part, "~"))
		} else if strings.Contains(part, ",") {
			// OR: @L0,@L1 -> (@L0 or @L1)
			orTags := strings.Split(part, ",")
			parts = append(parts, "("+strings.Join(orTags, " or ")+")")
		} else {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " and ")
}
