// Package cucumber provides a test runner adapter for TypeScript cucumber-js tests.
package cucumber

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/adapters/npm"
	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	testrunners.Register(&TsCucumberRunner{})
	testrunners.RegisterDescriptor(&testrunners.TestTypeDescriptor{
		TestType:      "tscucumber",
		IsBDD:         true,
		ComponentType: "gherkin",
		MonikerStyle:  "feature",
		FeatureTestTypeResolver: func(info testrunners.FeatureModuleInfo) bool {
			// TypeScript Cucumber owns features for TypeScript modules
			return info.HasTypeScript
		},
	})
}

// TsCucumberRunner handles TypeScript cucumber-js test execution.
type TsCucumberRunner struct{}

// TestTypes returns the test types this runner handles.
func (r *TsCucumberRunner) TestTypes() []string {
	return []string{"tscucumber"}
}

// IsBDD returns true because cucumber-js is a BDD test framework.
func (r *TsCucumberRunner) IsBDD() bool {
	return true
}

// GetTestInfo extracts structured test metadata from a TypeScript cucumber test reference.
func (r *TsCucumberRunner) GetTestInfo(test testing.TestReference, workspaceRoot string, cfg *config.EACConfig) *testrunners.TestInfo {
	// Calculate relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)

	info := &testrunners.TestInfo{Language: "ts"}

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

	// Return the module's typescript package root where cucumber-js should be
	tsRoot := module.Components.GetComponentRoot("typescript")
	if tsRoot == "" {
		return ""
	}
	return filepath.ToSlash(tsRoot)
}

// BuildPackagePath constructs the package path for test grouping.
// Returns "featureFolderName:moduleRoot:featurePath" format for cleaner display.
func (r *TsCucumberRunner) BuildPackagePath(testRoot, featurePath string) string {
	if testRoot == "" {
		return ""
	}
	if featurePath != "" {
		featureFolderName := extractTsFeatureFolderName(featurePath)
		return featureFolderName + ":" + testRoot + ":" + featurePath
	}
	return testRoot
}

// extractTsFeatureFolderName extracts the feature folder name from a feature path.
func extractTsFeatureFolderName(featurePath string) string {
	featurePath = filepath.ToSlash(featurePath)
	dir := filepath.Dir(featurePath)
	return filepath.Base(dir)
}

// Execute runs TypeScript cucumber-js tests for a package.
func (r *TsCucumberRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()

	// Parse package path - format: "featureName:moduleRoot:featurePath" or "moduleRoot"
	var displayName, relPkgPath, relFeatureFile string
	parts := strings.Split(pkgPath, ":")
	if len(parts) == 3 {
		displayName = parts[0] + ":" + parts[1]
		relPkgPath = parts[1]
		relFeatureFile = parts[2]
	} else if len(parts) == 1 {
		displayName = parts[0]
		relPkgPath = parts[0]
	} else {
		displayName = pkgPath
		relPkgPath = pkgPath
	}

	result := testrunners.RunResult{
		PackageName:   displayName,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	moduleRoot := filepath.Join(cfg.WorkspaceRoot, relPkgPath)

	// Check if package.json exists
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "No package.json found at %s\n", packageJSON)
		result.PackageFailed = true
		return result
	}

	// Prepare isolated npm environment
	isolation := npm.NewNpmIsolation(cfg.WorkspaceRoot)
	env, err := isolation.PrepareIsolatedEnv(moduleRoot, cfg.ModuleMoniker)
	if err != nil {
		fmt.Fprintf(logWriter, "Failed to prepare isolated environment: %v\n", err)
		result.PackageFailed = true
		return result
	}

	fmt.Fprintf(logWriter, "Using isolated environment: %s\n", env.WorkDir)

	// Run npm ci/install
	packageLock := filepath.Join(moduleRoot, "package-lock.json")
	var npmCmd string
	if _, err := os.Stat(packageLock); err == nil {
		npmCmd = "ci"
	} else {
		npmCmd = "install"
	}

	fmt.Fprintf(logWriter, "Installing npm dependencies (npm %s)...\n", npmCmd)
	npm.NpmInstallMu.Lock()
	installToolDef := tool.GlobalRegistry().GetOrAdhoc("npm")
	installExecCtx := &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       env.Env,
		ArgsOverrides: []string{npmCmd},
	}
	installResult, installErr := tool.GlobalExecutor().Execute(context.Background(), installToolDef, installExecCtx)
	npm.NpmInstallMu.Unlock()
	if installResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", installResult.Stdout, installResult.Stderr)
	}
	if installErr != nil || (installResult != nil && installResult.ExitCode != 0) {
		fmt.Fprintf(logWriter, "npm %s failed: %v\n", npmCmd, installErr)
		result.PackageFailed = true
		return result
	}

	// Build cucumber-js command
	args := []string{"cucumber-js"}

	// Add cucumber.json output format
	if cfg.OutputDir != "" {
		cucumberJSONPath := filepath.Join(cfg.OutputDir, "cucumber.json")
		args = append(args, "--format", fmt.Sprintf("json:%s", cucumberJSONPath))
	}

	// Add tag filter if provided (translate from TagFilter to cucumber-js syntax)
	if len(cfg.SuiteTagFilter.Selectors) > 0 {
		translator := &CucumberTagTranslator{}
		tagExpr := translator.TranslateTagFilter(cfg.SuiteTagFilter)
		if tagExpr != "" {
			args = append(args, "--tags", tagExpr)
		}
	}

	// Add the specific feature file if provided
	if relFeatureFile != "" {
		featurePath := filepath.Join(cfg.WorkspaceRoot, relFeatureFile)
		relPath, err := filepath.Rel(env.WorkDir, featurePath)
		if err == nil {
			args = append(args, relPath)
		}
	}

	fmt.Fprintf(logWriter, "=== Testing TypeScript cucumber specs ===\n")
	fmt.Fprintf(logWriter, "Module root: %s (isolated: %s)\n", moduleRoot, env.WorkDir)
	fmt.Fprintf(logWriter, "Command: npx %s\n\n", strings.Join(args, " "))

	// Execute npx cucumber-js
	npxToolDef := tool.GlobalRegistry().GetOrAdhoc("npx")
	npxExecCtx := &tool.ExecutionContext{
		ModuleRoot:    env.WorkDir,
		FullEnv:       append(env.Env, "CLIE_TEST_LOGGING_ACTIVE=true"),
		ArgsOverrides: args,
	}
	execResult, runErr := tool.GlobalExecutor().Execute(context.Background(), npxToolDef, npxExecCtx)
	if execResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", execResult.Stdout, execResult.Stderr)
	}

	if runErr != nil || (execResult != nil && execResult.ExitCode != 0) {
		result.PackageFailed = true
		result.TestsFailed = len(tests)
		fmt.Fprintf(logWriter, "cucumber-js failed\n")
	} else {
		result.TestsPassed = len(tests)
		fmt.Fprintf(logWriter, "cucumber-js passed\n")
	}

	result.TestsTotal = len(tests)
	result.Duration = time.Since(start)

	return result
}
