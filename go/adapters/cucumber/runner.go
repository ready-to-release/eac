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
		TestType:            "tscucumber",
		IsBDD:               true,
		ComponentType:       "gherkin",
		MonikerStyle:        "feature",
		Language:            "ts",
		ParentComponentType: "typescript",
		ModuleManifestFile:  "package.json",
		DefaultLevel:        "@L2",
		DefaultDepTag:       "@deps:npm",
		OutputArtifacts: []testrunners.ArtifactPattern{
			{ID: "cucumber-report", Pattern: "cucumber.json", Type: "cucumber-report"},
		},
		FeatureTestTypeResolver: func(info testrunners.FeatureModuleInfo) bool {
			// TypeScript Cucumber owns features for TypeScript modules
			return info.HasTypeScript
		},
		DefaultInferences: []testrunners.Inference{
			{
				TestTypes:   []string{"tscucumber"},
				ThenAddTags: []string{"@deps:npm"},
				Description: "TypeScript Cucumber tests require Node.js/npm",
			},
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
func (r *TsCucumberRunner) GetTestInfo(ref testing.TestReference, workspaceRoot string, cfg any) *testrunners.TestInfo {
	eacCfg := cfg.(*config.EACConfig)

	// Calculate relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, ref.FilePath)
	if err != nil {
		return nil
	}
	relPath = filepath.ToSlash(relPath)

	// Language from descriptor — no hardcoded string
	lang := "ts"
	if desc := testrunners.GetDescriptor("tscucumber"); desc != nil && desc.Language != "" {
		lang = desc.Language
	}
	info := &testrunners.TestInfo{Language: lang}

	// Extract module moniker from specs path
	specsPrefix := eacCfg.Repository.Paths.SpecsRoot + "/"
	specRelPath := strings.TrimPrefix(relPath, specsPrefix)
	specRelPath = filepath.ToSlash(specRelPath)

	// Get module moniker from first path component
	parts := strings.Split(specRelPath, "/")
	if len(parts) == 0 {
		return nil
	}
	info.ModuleMoniker = parts[0]

	// Verify module exists
	if eacCfg.Repository.GetByMoniker(info.ModuleMoniker) == nil {
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
func (r *TsCucumberRunner) FindTestRoot(featurePath string, cfg any) string {
	eacCfg := cfg.(*config.EACConfig)

	// Extract module moniker from specs path
	specsPrefix := eacCfg.Repository.Paths.SpecsRoot + "/"
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
	module, ok := eacCfg.Repository.GetModule(moniker)
	if !ok {
		return ""
	}

	// Return the module's parent component root where cucumber-js should be.
	// Use GetFirstByType (not GetComponentRoot) because the component's map key
	// is derived from the root path, not the type (e.g., key="vscode-commit"
	// for root="typescript/vscode-commit").
	// ParentComponentType comes from the descriptor (e.g., "typescript").
	parentType := "typescript"
	if desc := testrunners.GetDescriptor("tscucumber"); desc != nil && desc.ParentComponentType != "" {
		parentType = desc.ParentComponentType
	}
	if _, comp := module.Components.GetFirstByType(parentType); comp != nil && comp.Root != "" {
		return filepath.ToSlash(comp.Root)
	}
	return ""
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

// parsedCucumberPkg holds the parsed components of a cucumber package path.
type parsedCucumberPkg struct {
	displayName    string
	relPkgPath     string
	relFeatureFile string
}

// parseCucumberPackagePath parses a package path using the BDDPackagePath protocol.
func parseCucumberPackagePath(pkgPath string) parsedCucumberPkg {
	p := testrunners.ParseBDDPackagePath(pkgPath)
	return parsedCucumberPkg{
		displayName:    p.DisplayName(),
		relPkgPath:     p.TestRoot,
		relFeatureFile: p.FeaturePath,
	}
}

// prepareCucumberNpmEnv validates the module root, creates an isolated npm environment,
// and runs npm ci/install. Returns the isolated environment on success, or nil with
// result.PackageFailed set on failure.
func prepareCucumberNpmEnv(moduleRoot string, cfg testrunners.RunConfig, logWriter io.Writer, result *testrunners.RunResult) *npm.IsolatedEnv {
	// Check if module manifest file exists (from descriptor: "package.json")
	manifestFile := "package.json"
	if desc := testrunners.GetDescriptor("tscucumber"); desc != nil && desc.ModuleManifestFile != "" {
		manifestFile = desc.ModuleManifestFile
	}
	packageJSON := filepath.Join(moduleRoot, manifestFile)
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		fmt.Fprintf(logWriter, "No package.json found at %s\n", packageJSON)
		result.PackageFailed = true
		return nil
	}

	// Prepare isolated npm environment keyed by UoW output dir
	isolation := npm.NewNpmIsolation(cfg.WorkspaceRoot)
	env, err := isolation.PrepareIsolatedEnv(moduleRoot, cfg.OutputDir)
	if err != nil {
		fmt.Fprintf(logWriter, "Failed to prepare isolated environment: %v\n", err)
		result.PackageFailed = true
		return nil
	}

	fmt.Fprintf(logWriter, "Using isolated environment: %s\n", env.WorkDir)

	// Skip npm ci/install when dependencies haven't changed
	if env.DepsUpToDate {
		fmt.Fprintf(logWriter, "Dependencies up-to-date (cached node_modules), skipping npm install\n")
		return env
	}

	// Run npm ci/install
	packageLock := filepath.Join(moduleRoot, "package-lock.json")
	var npmCmd string
	if _, err := os.Stat(packageLock); err == nil {
		npmCmd = "ci"
	} else {
		npmCmd = "install"
	}

	fmt.Fprintf(logWriter, "Installing npm dependencies (npm %s)...\n", npmCmd)
	var installResult *tool.ExecutionResult
	var installErr error
	lockErr := npm.WithInstallLock(func() error {
		installToolDef := tool.GlobalRegistry().GetOrAdhoc("npm")
		installExecCtx := &tool.ExecutionContext{
			ModuleRoot:    env.WorkDir,
			FullEnv:       env.Env,
			ArgsOverrides: []string{npmCmd},
		}
		installResult, installErr = tool.GlobalExecutor().Execute(context.Background(), installToolDef, installExecCtx)
		return installErr
	})
	if installResult != nil {
		fmt.Fprintf(logWriter, "%s%s\n", installResult.Stdout, installResult.Stderr)
	}
	if lockErr != nil || installErr != nil || (installResult != nil && installResult.ExitCode != 0) {
		fmt.Fprintf(logWriter, "npm %s failed: %v\n", npmCmd, installErr)
		result.PackageFailed = true
		return nil
	}
	npm.MarkNpmDepsInstalled(env.WorkDir, moduleRoot)

	return env
}

// buildCucumberArgs constructs the npx cucumber-js command arguments including
// output format, tag filters, and the specific feature file path.
func buildCucumberArgs(cfg testrunners.RunConfig, env *npm.IsolatedEnv, relFeatureFile string) []string {
	// Use --no-install to prevent npx from fetching packages from the registry.
	// The @cucumber/cucumber package installs a "cucumber-js" binary locally;
	// without --no-install, npx falls through to the npm registry and resolves
	// a dependency-confusion placeholder package that exits 0 (false pass).
	args := []string{"--no-install", "cucumber-js"}

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

	return args
}

// executeCucumberTests runs npx cucumber-js and checks for dependency-confusion attacks.
// It populates result with pass/fail counts based on the execution outcome.
func executeCucumberTests(env *npm.IsolatedEnv, args []string, tests []testing.TestReference, logWriter io.Writer, result *testrunners.RunResult) {
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

	failed := runErr != nil || (execResult != nil && execResult.ExitCode != 0)
	// Guard against dependency confusion: if stderr mentions a placeholder or
	// npx could not find the local binary, treat as failure even if exit code is 0.
	if !failed && execResult != nil && (strings.Contains(string(execResult.Stderr), "placeholder") ||
		strings.Contains(string(execResult.Stderr), "dependency confusion") ||
		strings.Contains(string(execResult.Stderr), "not found and will be installed")) {
		failed = true
		fmt.Fprintf(logWriter, "cucumber-js: detected unexpected package resolution (possible dependency confusion)\n")
	}

	// Report 1 passed or 1 failed as a conservative fallback.
	// len(tests) is the feature-file count, not the scenario count, so using it
	// would inflate metrics. Real counts come from cucumber JSON when available.
	if failed {
		result.PackageFailed = true
		result.TestsFailed = 1
		fmt.Fprintf(logWriter, "cucumber-js failed\n")
	} else {
		result.TestsPassed = 1
		fmt.Fprintf(logWriter, "cucumber-js passed\n")
	}

	result.TestsTotal = 1
}

// Execute runs TypeScript cucumber-js tests for a package.
func (r *TsCucumberRunner) Execute(pkgPath string, tests []testing.TestReference, logWriter io.Writer, cfg testrunners.RunConfig) testrunners.RunResult {
	start := time.Now()

	pkg := parseCucumberPackagePath(pkgPath)
	result := testrunners.RunResult{
		PackageName:   pkg.displayName,
		ModuleMoniker: cfg.ModuleMoniker,
	}

	moduleRoot := filepath.Join(cfg.WorkspaceRoot, pkg.relPkgPath)

	env := prepareCucumberNpmEnv(moduleRoot, cfg, logWriter, &result)
	if env == nil {
		return result
	}

	args := buildCucumberArgs(cfg, env, pkg.relFeatureFile)

	fmt.Fprintf(logWriter, "=== Testing TypeScript cucumber specs ===\n")
	fmt.Fprintf(logWriter, "Module root: %s (isolated: %s)\n", moduleRoot, env.WorkDir)
	fmt.Fprintf(logWriter, "Command: npx %s\n\n", strings.Join(args, " "))

	executeCucumberTests(env, args, tests, logWriter, &result)

	result.Duration = time.Since(start)
	return result
}
