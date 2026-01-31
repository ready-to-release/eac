// Package cacheinvalidation contains godog step implementations for specs/eac-core/cache-invalidation.
//
// This file contains step definitions for the cache invalidation system tests.
// These tests verify that the build cache correctly detects when modules need rebuilding.
package cacheinvalidation

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"

	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/go/eac/core/hash"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// cacheContext holds test state for cache invalidation scenarios.
// This is ONLY for cache-specific state; command execution uses TestContext.
type cacheContext struct {
	// parsedYAML stores the parsed YAML output from last command
	parsedYAML map[string]interface{}
	// mockedCIStatus stores mocked CI responses for CI scenarios
	mockedCIStatus map[string]mockedModuleCI
	// changedFiles stores files marked as changed for CI mocking
	changedFiles []string
	// currentHeadSHA is the simulated HEAD SHA for CI mocking
	currentHeadSHA string
}

// mockedModuleCI represents mocked CI status for a module.
type mockedModuleCI struct {
	LastSuccessSHA   string
	HasFilesChanged  bool
	HasValidCIAtHead bool
	NoHistory        bool
}

var cacheCtx cacheContext

// resetCacheContext resets the cache context for a new scenario.
// Called at the start of RegisterSteps to ensure clean state.
func resetCacheContext() {
	cacheCtx = cacheContext{
		mockedCIStatus: make(map[string]mockedModuleCI),
	}
}

// RegisterSteps registers step definitions for cache invalidation feature specs.
// NOTE: Common steps like "I run", "the exit code is" are registered by RegisterCommonSteps.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Reset cache context for clean state
	resetCacheContext()

	// Background / Setup
	sc.Step(`^I am in an isolated test repository$`, func() error {
		return setupIsolatedRepository(ctx)
	})
	sc.Step(`^the repository has a multi-module structure with:$`, func(table *godog.Table) error {
		return setupMultiModuleStructure(ctx, table)
	})
	sc.Step(`^there is no build state file$`, func() error {
		return ensureNoBuildState(ctx)
	})
	sc.Step(`^I have built all modules successfully$`, func() error {
		return buildAllModules(ctx)
	})
	sc.Step(`^no files have been modified$`, func() error {
		return ensureNoModifications(ctx)
	})

	// File manipulation
	sc.Step(`^I append "([^"]*)" to file "([^"]*)"$`, func(content, filePath string) error {
		return appendToFile(ctx, content, filePath)
	})
	sc.Step(`^I delete the file "([^"]*)"$`, func(filePath string) error {
		return deleteFile(ctx, filePath)
	})
	sc.Step(`^I delete the build state directory$`, func() error {
		return ensureNoBuildState(ctx)
	})

	// Exit code assertion (not 0) - common steps don't have "not"
	sc.Step(`^the exit code is not (\d+)$`, func(notExpected int) error {
		if ctx.ExitCode == notExpected {
			return fmt.Errorf("expected exit code NOT to be %d, but it was\nOutput: %s", notExpected, ctx.CommandOutput)
		}
		return nil
	})

	// YAML output assertions (cache-specific)
	sc.Step(`^the YAML output field "([^"]*)" is "([^"]*)"$`, func(fieldPath, expected string) error {
		return yamlFieldEquals(ctx, fieldPath, expected)
	})
	sc.Step(`^the YAML output field "([^"]*)" contains "([^"]*)"$`, func(fieldPath, expected string) error {
		return yamlFieldContains(ctx, fieldPath, expected)
	})
	sc.Step(`^the YAML output field "([^"]*)" does not contain "([^"]*)"$`, func(fieldPath, expected string) error {
		return yamlFieldNotContains(ctx, fieldPath, expected)
	})
	sc.Step(`^the YAML output field "([^"]*)" is empty$`, func(fieldPath string) error {
		return yamlFieldEmpty(ctx, fieldPath)
	})
	sc.Step(`^the YAML output field "([^"]*)" contains exactly (\d+) occurrence of "([^"]*)"$`, func(fieldPath string, count int, expected string) error {
		return yamlFieldContainsExactly(ctx, fieldPath, count, expected)
	})
	sc.Step(`^the YAML output field "([^"]*)" indicates under (\d+) milliseconds$`, func(fieldPath string, maxMs int) error {
		return yamlFieldUnderMs(ctx, fieldPath, maxMs)
	})

	// Output text assertions (cache-specific patterns)
	sc.Step(`^the output indicates "([^"]*)" would be built$`, func(module string) error {
		return outputIndicatesWouldBuild(ctx, module)
	})
	sc.Step(`^the output indicates "([^"]*)" is up-to-date or would be skipped$`, func(module string) error {
		return outputIndicatesSkipped(ctx, module)
	})
	sc.Step(`^the output indicates "([^"]*)" would be linted$`, func(module string) error {
		return outputIndicatesWouldLint(ctx, module)
	})
	sc.Step(`^the output indicates "([^"]*)" would be skipped$`, func(module string) error {
		return outputIndicatesSkipped(ctx, module)
	})
	sc.Step(`^the output contains "([^"]*)"$`, func(expected string) error {
		return internal.OutputContains(ctx, expected)
	})

	// CI mocking (run with mocked CI - different from regular "I run")
	sc.Step(`^the mocked CI status shows:$`, func(table *godog.Table) error {
		return setupMockedCIStatus(ctx, table)
	})
	sc.Step(`^the mocked CI shows "([^"]*)" has valid CI at current HEAD$`, func(module string) error {
		return mockValidCIAtHead(ctx, module)
	})
	sc.Step(`^the mocked CI shows "([^"]*)" has valid CI at HEAD$`, func(module string) error {
		return mockValidCIAtHead(ctx, module)
	})
	sc.Step(`^the mocked CI shows "([^"]*)" has no successful runs$`, func(module string) error {
		return mockNoCIHistory(ctx, module)
	})
	sc.Step(`^the mocked CI shows "([^"]*)" CI at different SHA$`, func(module string) error {
		return mockCIAtDifferentSHA(ctx, module)
	})
	sc.Step(`^I run "([^"]*)" with mocked CI$`, func(cmdLine string) error {
		return runCommandWithMockedCI(ctx, cmdLine)
	})
	sc.Step(`^the only changed file is "([^"]*)"$`, func(filePath string) error {
		cacheCtx.changedFiles = []string{filePath}
		return nil
	})
	sc.Step(`^the changed files are "([^"]*)" and "([^"]*)"$`, func(file1, file2 string) error {
		cacheCtx.changedFiles = []string{file1, file2}
		return nil
	})

	// Lint state
	sc.Step(`^I have linted "([^"]*)" successfully$`, func(module string) error {
		return lintModuleSuccessfully(ctx, module)
	})
	sc.Step(`^I have a lint state showing "([^"]*)" failed$`, func(module string) error {
		return setLintStateFailed(ctx, module)
	})
	sc.Step(`^no files in "([^"]*)" have been modified$`, func(dir string) error {
		return ensureNoModificationsInDir(ctx, dir)
	})

	// Edge cases
	sc.Step(`^I have built modules "([^"]*)" and "([^"]*)"$`, func(mod1, mod2 string) error {
		return buildSpecificModules(ctx, mod1, mod2)
	})
	sc.Step(`^a new module "([^"]*)" is configured with go_root "([^"]*)"$`, func(moniker, goRoot string) error {
		return addNewModule(ctx, moniker, goRoot)
	})
	sc.Step(`^modules are configured with circular dependency:$`, func(table *godog.Table) error {
		return setupCircularDependency(ctx, table)
	})
	sc.Step(`^the build state file contains invalid JSON "([^"]*)"$`, func(content string) error {
		return corruptBuildState(ctx, content)
	})
}

// === Implementation Functions ===

func setupIsolatedRepository(ctx *internal.TestContext) error {
	// The @env:isolated-test-project tag should trigger SetupIsolation
	// If we're here without isolation, set it up manually
	if ctx.IsolatedDir == "" {
		if err := ctx.SetupIsolation(); err != nil {
			return fmt.Errorf("failed to setup isolation: %w", err)
		}
	}

	// Get current HEAD SHA for CI mocking
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.Output()
	if err == nil {
		cacheCtx.currentHeadSHA = strings.TrimSpace(string(output))
	}

	return nil
}

func setupMultiModuleStructure(ctx *internal.TestContext, table *godog.Table) error {
	ctx.MustBeIsolated()

	// Create .r2r directory structure
	r2rDir := filepath.Join(ctx.IsolatedDir, ".r2r", "eac")
	if err := os.MkdirAll(r2rDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .r2r directory: %w", err)
	}

	// Parse table and build repository.yml
	var modules []moduleConfig
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header
		}
		if len(row.Cells) < 3 {
			continue
		}

		modules = append(modules, moduleConfig{
			Moniker:   row.Cells[0].Value,
			DependsOn: row.Cells[1].Value,
			GoRoot:    row.Cells[2].Value,
		})
	}

	// Generate repository.yml
	repoYAML := generateRepositoryYAML(modules)
	repoPath := filepath.Join(r2rDir, "repository.yml")
	if err := os.WriteFile(repoPath, []byte(repoYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	// Create module source files
	for _, mod := range modules {
		goDir := filepath.Join(ctx.IsolatedDir, mod.GoRoot)
		if err := os.MkdirAll(goDir, 0o755); err != nil {
			return fmt.Errorf("failed to create go directory: %w", err)
		}

		// Create main.go
		mainPath := filepath.Join(goDir, "main.go")
		mainContent := fmt.Sprintf("package main\n\nfunc main() {\n\t// %s entry point\n}\n", mod.Moniker)
		if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
			return fmt.Errorf("failed to write main.go: %w", err)
		}

		// Create helper.go for modules that need it
		helperPath := filepath.Join(goDir, "helper.go")
		helperContent := fmt.Sprintf("package main\n\n// Helper functions for %s\n", mod.Moniker)
		if err := os.WriteFile(helperPath, []byte(helperContent), 0o644); err != nil {
			return fmt.Errorf("failed to write helper.go: %w", err)
		}

		// Create internal directory with handler.go for nested file testing
		internalDir := filepath.Join(goDir, "internal")
		if err := os.MkdirAll(internalDir, 0o755); err != nil {
			return fmt.Errorf("failed to create internal directory: %w", err)
		}

		handlerPath := filepath.Join(internalDir, "handler.go")
		handlerContent := fmt.Sprintf("package internal\n\n// Handler for %s\n", mod.Moniker)
		if err := os.WriteFile(handlerPath, []byte(handlerContent), 0o644); err != nil {
			return fmt.Errorf("failed to write handler.go: %w", err)
		}

		// Create CHANGELOG.md for CI exclusion testing
		changelogPath := filepath.Join(goDir, "CHANGELOG.md")
		changelogContent := fmt.Sprintf("# Changelog for %s\n\n## v1.0.0\n- Initial release\n", mod.Moniker)
		if err := os.WriteFile(changelogPath, []byte(changelogContent), 0o644); err != nil {
			return fmt.Errorf("failed to write CHANGELOG.md: %w", err)
		}
	}

	// Create CI workflow files for get changed-modules-ci filtering
	workflowsDir := filepath.Join(ctx.IsolatedDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}
	for _, mod := range modules {
		workflowPath := filepath.Join(workflowsDir, fmt.Sprintf("ci-%s.yaml", mod.Moniker))
		workflowContent := fmt.Sprintf("name: CI %s\non: [push]\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n", mod.Moniker)
		if err := os.WriteFile(workflowPath, []byte(workflowContent), 0o644); err != nil {
			return fmt.Errorf("failed to write CI workflow for %s: %w", mod.Moniker, err)
		}
	}

	// Create .gitignore to exclude build output directory
	// This is important because the build state file (out/build/.build-state.json)
	// should not cause uncommitted changes when computing git state hashes
	gitignorePath := filepath.Join(ctx.IsolatedDir, ".gitignore")
	gitignoreContent := "out/\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	// Git commit the structure
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Add multi-module structure for cache invalidation tests")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	// Update HEAD SHA after commit
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.Output()
	if err == nil {
		cacheCtx.currentHeadSHA = strings.TrimSpace(string(output))
	}

	return nil
}

type moduleConfig struct {
	Moniker   string
	DependsOn string
	GoRoot    string
}

func generateRepositoryYAML(modules []moduleConfig) string {
	var sb strings.Builder
	sb.WriteString("# Auto-generated repository.yml for cache invalidation tests\n\n")

	// Repository section (required)
	sb.WriteString("repository:\n")
	sb.WriteString("  type: mono\n")
	sb.WriteString("  remote:\n")
	sb.WriteString("    owner: test-owner\n")
	sb.WriteString("    repo: test-repo\n\n")

	// Modules section (array format)
	sb.WriteString("modules:\n")

	for _, mod := range modules {
		sb.WriteString(fmt.Sprintf("  - moniker: %s\n", mod.Moniker))
		sb.WriteString(fmt.Sprintf("    name: %s Module\n", mod.Moniker))
		sb.WriteString(fmt.Sprintf("    description: Test module for %s\n", mod.Moniker))
		if mod.DependsOn != "" {
			sb.WriteString("    depends_on:\n")
			for _, dep := range strings.Split(mod.DependsOn, ",") {
				dep = strings.TrimSpace(dep)
				if dep != "" {
					sb.WriteString(fmt.Sprintf("      - %s\n", dep))
				}
			}
		}
		sb.WriteString("    components:\n")
		sb.WriteString("      go:\n")
		sb.WriteString(fmt.Sprintf("        root: %s\n", mod.GoRoot))
		sb.WriteString("        patterns:\n")
		sb.WriteString("          source: [\"**/*.go\"]\n")
	}

	return sb.String()
}

func ensureNoBuildState(ctx *internal.TestContext) error {
	ctx.MustBeIsolated()
	// Remove the entire out/build directory to clear all per-module state files
	// The new workunit system stores state at out/build/<module>/_module/_/state.json
	buildDir := filepath.Join(ctx.IsolatedDir, "out", "build")
	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("failed to remove build state directory: %w", err)
	}
	return nil
}

func buildAllModules(ctx *internal.TestContext) error {
	ctx.MustBeIsolated()

	// Set R2R_CONTAINER_ROOT so that component-type defaults are loaded from the original repo
	// This matches what the subprocess command sees (see context.go buildMockingEnvironment)
	// Without this, patterns would differ between buildAllModules and get changed-modules-local
	origContainerRoot := os.Getenv("R2R_CONTAINER_ROOT")
	os.Setenv("R2R_CONTAINER_ROOT", ctx.OriginalRepoRoot)
	defer func() {
		if origContainerRoot == "" {
			os.Unsetenv("R2R_CONTAINER_ROOT")
		} else {
			os.Setenv("R2R_CONTAINER_ROOT", origContainerRoot)
		}
	}()

	// Load the module registry from the test repository.yml
	// This ensures we use the EXACT same file discovery as get changed-modules-local
	reg, err := modules.LoadFromWorkspaceNoValidation(ctx.IsolatedDir)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	// Save build state for each module using workunit StateManager
	stateMgr := workunit.NewStateManager(ctx.IsolatedDir)
	for _, contract := range reg.All() {
		// Expand glob patterns to get source files
		files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
		if err != nil {
			return fmt.Errorf("failed to expand patterns for %s: %w", contract.Moniker, err)
		}
		// Compute source hash
		sourceHash, err := hash.Files(ctx.IsolatedDir, files)
		if err != nil {
			return fmt.Errorf("failed to hash files for %s: %w", contract.Moniker, err)
		}
		// Save module state as passed build
		if err := stateMgr.SaveModuleResult(workunit.ContextBuild, contract.Moniker, true, sourceHash); err != nil {
			return fmt.Errorf("failed to update build state for %s: %w", contract.Moniker, err)
		}
	}

	return nil
}

func ensureNoModifications(ctx *internal.TestContext) error {
	ctx.MustBeIsolated()
	// Git checkout to reset any changes
	cmd := exec.Command("git", "checkout", "--", ".")
	cmd.Dir = ctx.IsolatedDir
	_ = cmd.Run() // Ignore errors if nothing to reset
	return nil
}

func appendToFile(ctx *internal.TestContext, content, filePath string) error {
	ctx.MustBeIsolated()
	fullPath := filepath.Join(ctx.IsolatedDir, filePath)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file for appending (create if not exists)
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + content + "\n"); err != nil {
		return fmt.Errorf("failed to append to file: %w", err)
	}

	return nil
}

func deleteFile(ctx *internal.TestContext, filePath string) error {
	ctx.MustBeIsolated()
	fullPath := filepath.Join(ctx.IsolatedDir, filePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// parseYAMLOutput parses the command output as YAML if possible.
func parseYAMLOutput(ctx *internal.TestContext) {
	cacheCtx.parsedYAML = nil
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(ctx.CommandOutput), &parsed); err == nil {
		cacheCtx.parsedYAML = parsed
	}
}

func runCommandWithMockedCI(ctx *internal.TestContext, cmdLine string) error {
	ctx.MustBeIsolated()

	// Set up mock environment variables for CI mocking
	mockJSON, err := json.Marshal(cacheCtx.mockedCIStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal mock CI status: %w", err)
	}

	// Write mock file
	mockDir := filepath.Join(ctx.IsolatedDir, "out", "test-mocks")
	if err := os.MkdirAll(mockDir, 0o755); err != nil {
		return fmt.Errorf("failed to create mock dir: %w", err)
	}

	mockPath := filepath.Join(mockDir, "ci-status.json")
	if err := os.WriteFile(mockPath, mockJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write mock file: %w", err)
	}

	// Set mock overrides on the test context
	ctx.SetMockOverride("R2R_MOCK_CI_STATUS", mockPath)
	ctx.SetMockOverride("R2R_MOCK_HEAD_SHA", cacheCtx.currentHeadSHA)
	ctx.SetMockOverride("R2R_MOCK_CHANGED_FILES", strings.Join(cacheCtx.changedFiles, ","))

	// Run command using TestContext.RunCommand (which applies mock overrides)
	if err := ctx.RunCommand(cmdLine); err != nil {
		return err
	}

	// Parse YAML output
	parseYAMLOutput(ctx)

	return nil
}

func yamlFieldEquals(ctx *internal.TestContext, fieldPath, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	actual := fmt.Sprintf("%v", value)
	if actual != expected {
		return fmt.Errorf("expected %s to be %q, got %q", fieldPath, expected, actual)
	}
	return nil
}

func yamlFieldContains(ctx *internal.TestContext, fieldPath, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		// Debug: print full output when field extraction fails
		if os.Getenv("DEBUG_CACHE_TEST") != "" {
			fmt.Printf("[DEBUG] Failed to extract %s, full output:\n%s\n", fieldPath, ctx.CommandOutput)
		}
		return err
	}

	switch v := value.(type) {
	case string:
		if !strings.Contains(v, expected) {
			return fmt.Errorf("expected %s to contain %q, got %q\nFull output:\n%s", fieldPath, expected, v, ctx.CommandOutput)
		}
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == expected {
				return nil
			}
		}
		return fmt.Errorf("expected %s to contain %q, got %v\nFull output:\n%s", fieldPath, expected, v, ctx.CommandOutput)
	default:
		// Check string representation
		if !strings.Contains(fmt.Sprintf("%v", value), expected) {
			return fmt.Errorf("expected %s to contain %q, got %v\nFull output:\n%s", fieldPath, expected, value, ctx.CommandOutput)
		}
	}
	return nil
}

func yamlFieldNotContains(ctx *internal.TestContext, fieldPath, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		// Field not found means it doesn't contain the value
		return nil
	}

	switch v := value.(type) {
	case string:
		if strings.Contains(v, expected) {
			return fmt.Errorf("expected %s NOT to contain %q, but it did", fieldPath, expected)
		}
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == expected {
				return fmt.Errorf("expected %s NOT to contain %q, but it did", fieldPath, expected)
			}
		}
	}
	return nil
}

func yamlFieldEmpty(ctx *internal.TestContext, fieldPath string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		// Field not found is considered empty
		return nil
	}

	switch v := value.(type) {
	case []interface{}:
		if len(v) != 0 {
			return fmt.Errorf("expected %s to be empty, got %v", fieldPath, v)
		}
	case string:
		if v != "" {
			return fmt.Errorf("expected %s to be empty, got %q", fieldPath, v)
		}
	case nil:
		// nil is empty
	default:
		return fmt.Errorf("expected %s to be empty, got %v", fieldPath, value)
	}
	return nil
}

func yamlFieldContainsExactly(ctx *internal.TestContext, fieldPath string, count int, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	var occurrences int
	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == expected {
				occurrences++
			}
		}
	case string:
		occurrences = strings.Count(v, expected)
	}

	if occurrences != count {
		return fmt.Errorf("expected %s to contain exactly %d occurrences of %q, got %d", fieldPath, count, expected, occurrences)
	}
	return nil
}

func yamlFieldUnderMs(ctx *internal.TestContext, fieldPath string, maxMs int) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	durationStr := fmt.Sprintf("%v", value)
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("failed to parse duration %q: %w", durationStr, err)
	}

	if duration.Milliseconds() > int64(maxMs) {
		return fmt.Errorf("expected %s under %dms, got %v", fieldPath, maxMs, duration)
	}
	return nil
}

func extractYAMLField(ctx *internal.TestContext, fieldPath string) (interface{}, error) {
	if cacheCtx.parsedYAML == nil {
		// Try to parse from raw output
		// Strip any non-YAML prefix (like "Detected devbox environment..." messages)
		output := ctx.CommandOutput
		yamlStart := findYAMLStart(output)
		if yamlStart > 0 {
			output = output[yamlStart:]
		}

		var data map[string]interface{}
		if err := yaml.Unmarshal([]byte(output), &data); err != nil {
			return nil, fmt.Errorf("failed to parse YAML output: %w\nOutput: %s", err, ctx.CommandOutput)
		}
		cacheCtx.parsedYAML = data
	}

	// Navigate nested path like "module_status.test-core.has_valid_ci"
	parts := strings.Split(fieldPath, ".")
	var current interface{} = cacheCtx.parsedYAML

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field %q not found at path %s", part, fieldPath)
			}
		case map[interface{}]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field %q not found at path %s", part, fieldPath)
			}
		default:
			return nil, fmt.Errorf("cannot navigate into %T at %s", current, part)
		}
	}

	return current, nil
}

// findYAMLStart finds the start of YAML content in output that may have prefix text.
// Returns the index where YAML content starts (0 if no prefix).
func findYAMLStart(output string) int {
	// Look for common YAML start patterns
	patterns := []string{"modules:", "is_fresh_build:", "directly_changed:", "base_sha:"}
	minIndex := len(output)

	for _, pattern := range patterns {
		idx := strings.Index(output, pattern)
		if idx >= 0 && idx < minIndex {
			// Find the start of the line containing this pattern
			lineStart := idx
			for lineStart > 0 && output[lineStart-1] != '\n' {
				lineStart--
			}
			if lineStart < minIndex {
				minIndex = lineStart
			}
		}
	}

	if minIndex < len(output) {
		return minIndex
	}
	return 0
}

func outputIndicatesWouldBuild(ctx *internal.TestContext, module string) error {
	patterns := []string{
		fmt.Sprintf("%s.*would be built", module),
		fmt.Sprintf("%s.*will build", module),
		fmt.Sprintf("building.*%s", module),
		fmt.Sprintf("%s.*needs rebuild", module),
		fmt.Sprintf("%s.*changed", module),
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, ctx.CommandOutput)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("output does not indicate %s would be built. Output: %s", module, ctx.CommandOutput)
}

func outputIndicatesSkipped(ctx *internal.TestContext, module string) error {
	patterns := []string{
		fmt.Sprintf("%s.*up-to-date", module),
		fmt.Sprintf("%s.*skipped", module),
		fmt.Sprintf("%s.*skip", module),
		fmt.Sprintf("skipping.*%s", module),
		fmt.Sprintf("%s.*cached", module),
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, ctx.CommandOutput)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("output does not indicate %s is up-to-date or would be skipped. Output: %s", module, ctx.CommandOutput)
}

func outputIndicatesWouldLint(ctx *internal.TestContext, module string) error {
	patterns := []string{
		fmt.Sprintf("%s.*would be linted", module),
		fmt.Sprintf("%s.*will lint", module),
		fmt.Sprintf("linting.*%s", module),
		fmt.Sprintf("%s.*needs linting", module),
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, ctx.CommandOutput)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("output does not indicate %s would be linted. Output: %s", module, ctx.CommandOutput)
}

func setupMockedCIStatus(ctx *internal.TestContext, table *godog.Table) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header
		}
		if len(row.Cells) < 3 {
			continue
		}

		module := row.Cells[0].Value
		lastSuccessSHA := row.Cells[1].Value
		hasFilesChanged, _ := strconv.ParseBool(row.Cells[2].Value)

		cacheCtx.mockedCIStatus[module] = mockedModuleCI{
			LastSuccessSHA:  lastSuccessSHA,
			HasFilesChanged: hasFilesChanged,
		}
	}

	return nil
}

func mockValidCIAtHead(ctx *internal.TestContext, module string) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	cacheCtx.mockedCIStatus[module] = mockedModuleCI{
		LastSuccessSHA:   cacheCtx.currentHeadSHA,
		HasFilesChanged:  false,
		HasValidCIAtHead: true,
	}
	return nil
}

func mockNoCIHistory(ctx *internal.TestContext, module string) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	cacheCtx.mockedCIStatus[module] = mockedModuleCI{
		NoHistory: true,
	}
	return nil
}

func mockCIAtDifferentSHA(ctx *internal.TestContext, module string) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	cacheCtx.mockedCIStatus[module] = mockedModuleCI{
		LastSuccessSHA:  "different-sha-123456",
		HasFilesChanged: true,
	}
	return nil
}

func lintModuleSuccessfully(ctx *internal.TestContext, module string) error {
	ctx.MustBeIsolated()

	// Set R2R_CONTAINER_ROOT for proper template loading
	origContainerRoot := os.Getenv("R2R_CONTAINER_ROOT")
	os.Setenv("R2R_CONTAINER_ROOT", ctx.OriginalRepoRoot)
	defer func() {
		if origContainerRoot == "" {
			os.Unsetenv("R2R_CONTAINER_ROOT")
		} else {
			os.Setenv("R2R_CONTAINER_ROOT", origContainerRoot)
		}
	}()

	// Load module registry to get file patterns
	reg, err := modules.LoadFromWorkspaceNoValidation(ctx.IsolatedDir)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	contract, ok := reg.Get(module)
	if !ok {
		return fmt.Errorf("module not found: %s", module)
	}

	// Expand glob patterns to get source files
	files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
	if err != nil {
		return fmt.Errorf("failed to expand patterns for %s: %w", module, err)
	}

	// Compute source hash
	sourceHash, err := hash.Files(ctx.IsolatedDir, files)
	if err != nil {
		return fmt.Errorf("failed to hash files for %s: %w", module, err)
	}

	// Create lint state showing module passed using workunit StateManager
	stateMgr := workunit.NewStateManager(ctx.IsolatedDir)
	if err := stateMgr.SaveModuleResult(workunit.ContextLint, module, true, sourceHash); err != nil {
		return fmt.Errorf("failed to update lint state for %s: %w", module, err)
	}

	return nil
}

func setLintStateFailed(ctx *internal.TestContext, module string) error {
	ctx.MustBeIsolated()

	// Set R2R_CONTAINER_ROOT for proper template loading
	origContainerRoot := os.Getenv("R2R_CONTAINER_ROOT")
	os.Setenv("R2R_CONTAINER_ROOT", ctx.OriginalRepoRoot)
	defer func() {
		if origContainerRoot == "" {
			os.Unsetenv("R2R_CONTAINER_ROOT")
		} else {
			os.Setenv("R2R_CONTAINER_ROOT", origContainerRoot)
		}
	}()

	// Load module registry to get file patterns
	reg, err := modules.LoadFromWorkspaceNoValidation(ctx.IsolatedDir)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	contract, ok := reg.Get(module)
	if !ok {
		return fmt.Errorf("module not found: %s", module)
	}

	// Expand glob patterns to get source files
	files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
	if err != nil {
		return fmt.Errorf("failed to expand patterns for %s: %w", module, err)
	}

	// Compute source hash
	sourceHash, err := hash.Files(ctx.IsolatedDir, files)
	if err != nil {
		return fmt.Errorf("failed to hash files for %s: %w", module, err)
	}

	// Create lint state showing module failed using workunit StateManager
	stateMgr := workunit.NewStateManager(ctx.IsolatedDir)
	if err := stateMgr.SaveModuleResult(workunit.ContextLint, module, false, sourceHash); err != nil {
		return fmt.Errorf("failed to update lint state for %s: %w", module, err)
	}

	return nil
}

func ensureNoModificationsInDir(ctx *internal.TestContext, dir string) error {
	ctx.MustBeIsolated()
	fullDir := filepath.Join(ctx.IsolatedDir, dir)
	cmd := exec.Command("git", "checkout", "--", fullDir)
	cmd.Dir = ctx.IsolatedDir
	_ = cmd.Run() // Ignore errors if nothing to reset
	return nil
}

func buildSpecificModules(ctx *internal.TestContext, mod1, mod2 string) error {
	ctx.MustBeIsolated()

	// Set R2R_CONTAINER_ROOT for proper template loading
	origContainerRoot := os.Getenv("R2R_CONTAINER_ROOT")
	os.Setenv("R2R_CONTAINER_ROOT", ctx.OriginalRepoRoot)
	defer func() {
		if origContainerRoot == "" {
			os.Unsetenv("R2R_CONTAINER_ROOT")
		} else {
			os.Setenv("R2R_CONTAINER_ROOT", origContainerRoot)
		}
	}()

	// Load module registry
	reg, err := modules.LoadFromWorkspaceNoValidation(ctx.IsolatedDir)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	// Save build state for the specified modules using workunit StateManager
	stateMgr := workunit.NewStateManager(ctx.IsolatedDir)
	moduleNames := []string{mod1, mod2}
	for _, moniker := range moduleNames {
		contract, ok := reg.Get(moniker)
		if !ok {
			return fmt.Errorf("module not found: %s", moniker)
		}
		// Expand glob patterns to get source files
		files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
		if err != nil {
			return fmt.Errorf("failed to expand patterns for %s: %w", moniker, err)
		}
		// Compute source hash
		sourceHash, err := hash.Files(ctx.IsolatedDir, files)
		if err != nil {
			return fmt.Errorf("failed to hash files for %s: %w", moniker, err)
		}
		// Save module state as passed build
		if err := stateMgr.SaveModuleResult(workunit.ContextBuild, moniker, true, sourceHash); err != nil {
			return fmt.Errorf("failed to update build state for %s: %w", moniker, err)
		}
	}

	return nil
}

func addNewModule(ctx *internal.TestContext, moniker, goRoot string) error {
	ctx.MustBeIsolated()
	// Read existing repository.yml
	repoPath := filepath.Join(ctx.IsolatedDir, ".r2r", "eac", "repository.yml")
	content, err := os.ReadFile(repoPath)
	if err != nil {
		return fmt.Errorf("failed to read repository.yml: %w", err)
	}

	// Append new module configuration (array format to match generateRepositoryYAML)
	newConfig := fmt.Sprintf(`  - moniker: %s
    name: %s Module
    description: Test module for %s
    components:
      go:
        root: %s
        patterns:
          source: ["**/*.go"]
`, moniker, moniker, moniker, goRoot)

	newContent := string(content) + newConfig
	if err := os.WriteFile(repoPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	// Create module directory and files
	goDir := filepath.Join(ctx.IsolatedDir, goRoot)
	if err := os.MkdirAll(goDir, 0o755); err != nil {
		return fmt.Errorf("failed to create go directory: %w", err)
	}

	mainPath := filepath.Join(goDir, "main.go")
	mainContent := fmt.Sprintf("package main\n\nfunc main() {\n\t// %s entry point\n}\n", moniker)
	if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}

func setupCircularDependency(ctx *internal.TestContext, table *godog.Table) error {
	ctx.MustBeIsolated()
	// Create a new repository.yml with circular dependencies
	var modules []moduleConfig
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header
		}
		if len(row.Cells) < 2 {
			continue
		}

		modules = append(modules, moduleConfig{
			Moniker:   row.Cells[0].Value,
			DependsOn: row.Cells[1].Value,
			GoRoot:    fmt.Sprintf("go/%s", row.Cells[0].Value),
		})
	}

	// Generate and write repository.yml
	repoYAML := generateRepositoryYAML(modules)
	repoPath := filepath.Join(ctx.IsolatedDir, ".r2r", "eac", "repository.yml")
	if err := os.WriteFile(repoPath, []byte(repoYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	// Create module directories
	for _, mod := range modules {
		goDir := filepath.Join(ctx.IsolatedDir, mod.GoRoot)
		if err := os.MkdirAll(goDir, 0o755); err != nil {
			return fmt.Errorf("failed to create go directory: %w", err)
		}

		mainPath := filepath.Join(goDir, "main.go")
		mainContent := "package main\n\nfunc main() {}\n"
		if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
			return fmt.Errorf("failed to write main.go: %w", err)
		}
	}

	return nil
}

func corruptBuildState(ctx *internal.TestContext, content string) error {
	ctx.MustBeIsolated()
	buildDir := filepath.Join(ctx.IsolatedDir, "out", "build")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return fmt.Errorf("failed to create build dir: %w", err)
	}

	statePath := filepath.Join(buildDir, ".build-state.json")
	if err := os.WriteFile(statePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write corrupted build state: %w", err)
	}

	return nil
}
