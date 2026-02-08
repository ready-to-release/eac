// Package specs contains godog step implementations for core features.
//
// This file contains step definitions for the cache invalidation tests.
// These tests verify that the build cache correctly detects when modules need rebuilding.
package specs

import (
	"context"
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

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/output"
)

// cacheContext holds test state for cache invalidation scenarios.
type cacheContext struct {
	parsedYAML     map[string]interface{}
	mockedCIStatus map[string]mockedModuleCI
	changedFiles   []string
	currentHeadSHA string
}

type mockedModuleCI struct {
	LastSuccessSHA   string
	HasFilesChanged  bool
	HasValidCIAtHead bool
	NoHistory        bool
}

var cacheCtx cacheContext

func resetCacheContext() {
	cacheCtx = cacheContext{
		mockedCIStatus: make(map[string]mockedModuleCI),
	}
}

// registerCacheSteps registers step definitions for cache invalidation feature specs.
func registerCacheSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	resetCacheContext()

	// Hook to capture HEAD SHA after isolation is set up
	sc.After(func(ctx2 context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		resetCacheContext()
		return ctx2, nil
	})

	// Background / Setup - use hook to capture HEAD SHA when isolation is set up
	sc.StepContext().Before(func(ctx2 context.Context, st *godog.Step) (context.Context, error) {
		if strings.Contains(st.Text, "isolated test repository") && ctx.IsolatedDir != "" {
			captureHeadSHA(ctx)
		}
		return ctx2, nil
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

	// Exit code
	sc.Step(`^the exit code is not (\d+)$`, func(notExpected int) error {
		if ctx.ExitCode == notExpected {
			return fmt.Errorf("expected exit code NOT to be %d, but it was\nOutput: %s", notExpected, ctx.CommandOutput)
		}
		return nil
	})

	// YAML output assertions
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

	// Output assertions
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
		return eacgodog.OutputContains(ctx, expected)
	})

	// CI mocking
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

// captureHeadSHA captures the current HEAD SHA after isolation is set up.
func captureHeadSHA(ctx *eacgodog.TestContext) {
	if ctx.IsolatedDir == "" {
		return
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.Output()
	if err == nil {
		cacheCtx.currentHeadSHA = strings.TrimSpace(string(output))
	}
}

func setupMultiModuleStructure(ctx *eacgodog.TestContext, table *godog.Table) error {
	ctx.MustBeIsolated()

	clieDir := filepath.Join(ctx.IsolatedDir, ".eac")
	if err := os.MkdirAll(clieDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .clie directory: %w", err)
	}

	var mods []cacheModuleConfig
	for i, row := range table.Rows {
		if i == 0 {
			continue
		}
		if len(row.Cells) < 3 {
			continue
		}
		mods = append(mods, cacheModuleConfig{
			Moniker:   row.Cells[0].Value,
			DependsOn: row.Cells[1].Value,
			GoRoot:    row.Cells[2].Value,
		})
	}

	repoYAML := generateCacheRepositoryYAML(mods)
	repoPath := filepath.Join(clieDir, "repository.yml")
	if err := os.WriteFile(repoPath, []byte(repoYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	for _, mod := range mods {
		goDir := filepath.Join(ctx.IsolatedDir, mod.GoRoot)
		if err := os.MkdirAll(goDir, 0o755); err != nil {
			return fmt.Errorf("failed to create go directory: %w", err)
		}

		mainPath := filepath.Join(goDir, "main.go")
		mainContent := fmt.Sprintf("package main\n\nfunc main() {\n\t// %s entry point\n}\n", mod.Moniker)
		if err := os.WriteFile(mainPath, []byte(mainContent), 0o644); err != nil {
			return fmt.Errorf("failed to write main.go: %w", err)
		}

		helperPath := filepath.Join(goDir, "helper.go")
		helperContent := fmt.Sprintf("package main\n\n// Helper functions for %s\n", mod.Moniker)
		if err := os.WriteFile(helperPath, []byte(helperContent), 0o644); err != nil {
			return fmt.Errorf("failed to write helper.go: %w", err)
		}

		internalDir := filepath.Join(goDir, "internal")
		if err := os.MkdirAll(internalDir, 0o755); err != nil {
			return fmt.Errorf("failed to create internal directory: %w", err)
		}

		handlerPath := filepath.Join(internalDir, "handler.go")
		handlerContent := fmt.Sprintf("package internal\n\n// Handler for %s\n", mod.Moniker)
		if err := os.WriteFile(handlerPath, []byte(handlerContent), 0o644); err != nil {
			return fmt.Errorf("failed to write handler.go: %w", err)
		}

		changelogPath := filepath.Join(goDir, "CHANGELOG.md")
		changelogContent := fmt.Sprintf("# Changelog for %s\n\n## v1.0.0\n- Initial release\n", mod.Moniker)
		if err := os.WriteFile(changelogPath, []byte(changelogContent), 0o644); err != nil {
			return fmt.Errorf("failed to write CHANGELOG.md: %w", err)
		}
	}

	workflowsDir := filepath.Join(ctx.IsolatedDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}
	for _, mod := range mods {
		workflowPath := filepath.Join(workflowsDir, fmt.Sprintf("ci-%s.yaml", mod.Moniker))
		workflowContent := fmt.Sprintf("name: CI %s\non: [push]\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n", mod.Moniker)
		if err := os.WriteFile(workflowPath, []byte(workflowContent), 0o644); err != nil {
			return fmt.Errorf("failed to write CI workflow for %s: %w", mod.Moniker, err)
		}
	}

	gitignorePath := filepath.Join(ctx.IsolatedDir, ".gitignore")
	gitignoreContent := "out/\n"
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

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

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.Output()
	if err == nil {
		cacheCtx.currentHeadSHA = strings.TrimSpace(string(output))
	}

	return nil
}

type cacheModuleConfig struct {
	Moniker   string
	DependsOn string
	GoRoot    string
}

func generateCacheRepositoryYAML(mods []cacheModuleConfig) string {
	var sb strings.Builder
	sb.WriteString("# Auto-generated repository.yml for cache invalidation tests\n\n")
	sb.WriteString("repository:\n")
	sb.WriteString("  type: mono\n")
	sb.WriteString("  remote:\n")
	sb.WriteString("    owner: test-owner\n")
	sb.WriteString("    repo: test-repo\n\n")
	sb.WriteString("modules:\n")

	for _, mod := range mods {
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

func ensureNoBuildState(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()
	buildDir := filepath.Join(ctx.IsolatedDir, "out", "build")
	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("failed to remove build state directory: %w", err)
	}
	return nil
}

func buildAllModules(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()

	origContainerRoot := os.Getenv(environments.EnvCLIEContainerRoot)
	os.Setenv(environments.EnvCLIEContainerRoot, ctx.OriginalRepoRoot)
	defer func() {
		if origContainerRoot == "" {
			os.Unsetenv(environments.EnvCLIEContainerRoot)
		} else {
			os.Setenv(environments.EnvCLIEContainerRoot, origContainerRoot)
		}
	}()

	reg, err := modules.LoadFromWorkspaceNoValidation(ctx.IsolatedDir)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	// Create UoW manifests for each module (simulating a successful build)
	for _, contract := range reg.All() {
		files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
		if err != nil {
			return fmt.Errorf("failed to expand patterns for %s: %w", contract.Moniker, err)
		}
		sourceHash, err := hash.Files(ctx.IsolatedDir, files)
		if err != nil {
			return fmt.Errorf("failed to hash files for %s: %w", contract.Moniker, err)
		}

		// Create UoW manifest in the expected location
		// Format: out/build/<module>/<component>-<tool>/uow.manifest.json
		manifest := &output.UoWManifest{
			Action:     core.ActionBuild,
			Module:     contract.Moniker,
			Component:  "go",  // Default component for Go modules
			Tool:       "go",  // Default tool
			ExitCode:   0,
			InputHash:  sourceHash,
			ExecutedAt: time.Now().UTC().Truncate(time.Second),
			Duration:   time.Second,
			Artifacts:  []output.Artifact{},
			OutputHash: "sha256:output-" + contract.Moniker,
			Version:    "1.0.0",
		}

		manifestDir := filepath.Join(ctx.IsolatedDir, "out", "build", contract.Moniker, "go-go")
		if err := os.MkdirAll(manifestDir, 0755); err != nil {
			return fmt.Errorf("failed to create manifest dir for %s: %w", contract.Moniker, err)
		}

		manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
		data, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal manifest for %s: %w", contract.Moniker, err)
		}
		if err := os.WriteFile(manifestPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write manifest for %s: %w", contract.Moniker, err)
		}
	}

	return nil
}

func ensureNoModifications(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()
	cmd := exec.Command("git", "checkout", "--", ".")
	cmd.Dir = ctx.IsolatedDir
	_ = cmd.Run()
	return nil
}

func appendToFile(ctx *eacgodog.TestContext, content, filePath string) error {
	ctx.MustBeIsolated()
	fullPath := filepath.Join(ctx.IsolatedDir, filePath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

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

func deleteFile(ctx *eacgodog.TestContext, filePath string) error {
	ctx.MustBeIsolated()
	fullPath := filepath.Join(ctx.IsolatedDir, filePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func parseYAMLOutput(ctx *eacgodog.TestContext) {
	cacheCtx.parsedYAML = nil
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(ctx.CommandOutput), &parsed); err == nil {
		cacheCtx.parsedYAML = parsed
	}
}

func runCommandWithMockedCI(ctx *eacgodog.TestContext, cmdLine string) error {
	ctx.MustBeIsolated()

	mockJSON, err := json.Marshal(cacheCtx.mockedCIStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal mock CI status: %w", err)
	}

	mockDir := filepath.Join(ctx.IsolatedDir, "out", "test-mocks")
	if err := os.MkdirAll(mockDir, 0o755); err != nil {
		return fmt.Errorf("failed to create mock dir: %w", err)
	}

	mockPath := filepath.Join(mockDir, "ci-status.json")
	if err := os.WriteFile(mockPath, mockJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write mock file: %w", err)
	}

	ctx.SetMockOverride("CLIE_MOCK_CI_STATUS", mockPath)
	ctx.SetMockOverride("CLIE_MOCK_HEAD_SHA", cacheCtx.currentHeadSHA)
	ctx.SetMockOverride("CLIE_MOCK_CHANGED_FILES", strings.Join(cacheCtx.changedFiles, ","))

	if err := ctx.RunCommand(cmdLine); err != nil {
		return err
	}

	parseYAMLOutput(ctx)

	return nil
}

func yamlFieldEquals(ctx *eacgodog.TestContext, fieldPath, expected string) error {
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

func yamlFieldContains(ctx *eacgodog.TestContext, fieldPath, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
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
		if !strings.Contains(fmt.Sprintf("%v", value), expected) {
			return fmt.Errorf("expected %s to contain %q, got %v\nFull output:\n%s", fieldPath, expected, value, ctx.CommandOutput)
		}
	}
	return nil
}

func yamlFieldNotContains(ctx *eacgodog.TestContext, fieldPath, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
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

func yamlFieldEmpty(ctx *eacgodog.TestContext, fieldPath string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
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

func yamlFieldContainsExactly(ctx *eacgodog.TestContext, fieldPath string, count int, expected string) error {
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

func yamlFieldUnderMs(ctx *eacgodog.TestContext, fieldPath string, maxMs int) error {
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

func extractYAMLField(ctx *eacgodog.TestContext, fieldPath string) (interface{}, error) {
	if cacheCtx.parsedYAML == nil {
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

func findYAMLStart(output string) int {
	patterns := []string{"modules:", "is_fresh_build:", "directly_changed:", "base_sha:"}
	minIndex := len(output)

	for _, pattern := range patterns {
		idx := strings.Index(output, pattern)
		if idx >= 0 && idx < minIndex {
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

func outputIndicatesWouldBuild(ctx *eacgodog.TestContext, module string) error {
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

func outputIndicatesSkipped(ctx *eacgodog.TestContext, module string) error {
	patterns := []string{
		fmt.Sprintf("%s.*up-to-date", module),
		fmt.Sprintf("%s.*skipped", module),
		fmt.Sprintf("%s.*skip", module),
		fmt.Sprintf("skipping.*%s", module),
		fmt.Sprintf("%s.*cached", module),
		fmt.Sprintf("cached.*%s", module),
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, ctx.CommandOutput)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("output does not indicate %s is up-to-date or would be skipped. Output: %s", module, ctx.CommandOutput)
}

func outputIndicatesWouldLint(ctx *eacgodog.TestContext, module string) error {
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

func setupMockedCIStatus(ctx *eacgodog.TestContext, table *godog.Table) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	for i, row := range table.Rows {
		if i == 0 {
			continue
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

func mockValidCIAtHead(ctx *eacgodog.TestContext, module string) error {
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

func mockNoCIHistory(ctx *eacgodog.TestContext, module string) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	cacheCtx.mockedCIStatus[module] = mockedModuleCI{
		NoHistory: true,
	}
	return nil
}

func mockCIAtDifferentSHA(ctx *eacgodog.TestContext, module string) error {
	if cacheCtx.mockedCIStatus == nil {
		cacheCtx.mockedCIStatus = make(map[string]mockedModuleCI)
	}

	cacheCtx.mockedCIStatus[module] = mockedModuleCI{
		LastSuccessSHA:  "different-sha-123456",
		HasFilesChanged: true,
	}
	return nil
}

func lintModuleSuccessfully(ctx *eacgodog.TestContext, module string) error {
	ctx.MustBeIsolated()

	origContainerRoot := os.Getenv(environments.EnvCLIEContainerRoot)
	os.Setenv(environments.EnvCLIEContainerRoot, ctx.OriginalRepoRoot)
	defer func() {
		if origContainerRoot == "" {
			os.Unsetenv(environments.EnvCLIEContainerRoot)
		} else {
			os.Setenv(environments.EnvCLIEContainerRoot, origContainerRoot)
		}
	}()

	reg, err := modules.LoadFromWorkspaceNoValidation(ctx.IsolatedDir)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	contract, ok := reg.Get(module)
	if !ok {
		return fmt.Errorf("module not found: %s", module)
	}

	files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
	if err != nil {
		return fmt.Errorf("failed to expand patterns for %s: %w", module, err)
	}

	sourceHash, err := hash.Files(ctx.IsolatedDir, files)
	if err != nil {
		return fmt.Errorf("failed to hash files for %s: %w", module, err)
	}

	// Create UoW manifest for successful lint
	manifest := &output.UoWManifest{
		Action:     core.ActionLint,
		Module:     module,
		Component:  "go",
		Tool:       "golangci-lint",
		ExitCode:   0,
		InputHash:  sourceHash,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   time.Second,
		Artifacts:  []output.Artifact{},
		OutputHash: "sha256:lint-output-" + module,
		Version:    "1.0.0",
	}

	manifestDir := filepath.Join(ctx.IsolatedDir, "out", "lint", module, "go-golangci-lint")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest dir for %s: %w", module, err)
	}

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for %s: %w", module, err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest for %s: %w", module, err)
	}

	return nil
}

func setLintStateFailed(ctx *eacgodog.TestContext, module string) error {
	ctx.MustBeIsolated()

	origContainerRoot := os.Getenv(environments.EnvCLIEContainerRoot)
	os.Setenv(environments.EnvCLIEContainerRoot, ctx.OriginalRepoRoot)
	defer func() {
		if origContainerRoot == "" {
			os.Unsetenv(environments.EnvCLIEContainerRoot)
		} else {
			os.Setenv(environments.EnvCLIEContainerRoot, origContainerRoot)
		}
	}()

	reg, err := modules.LoadFromWorkspaceNoValidation(ctx.IsolatedDir)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	contract, ok := reg.Get(module)
	if !ok {
		return fmt.Errorf("module not found: %s", module)
	}

	files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
	if err != nil {
		return fmt.Errorf("failed to expand patterns for %s: %w", module, err)
	}

	sourceHash, err := hash.Files(ctx.IsolatedDir, files)
	if err != nil {
		return fmt.Errorf("failed to hash files for %s: %w", module, err)
	}

	// Create UoW manifest for failed lint (exit_code: 1)
	manifest := &output.UoWManifest{
		Action:     core.ActionLint,
		Module:     module,
		Component:  "go",
		Tool:       "golangci-lint",
		ExitCode:   1, // Failed
		InputHash:  sourceHash,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   time.Second,
		Artifacts:  []output.Artifact{},
		OutputHash: "sha256:lint-output-" + module,
		Version:    "1.0.0",
	}

	manifestDir := filepath.Join(ctx.IsolatedDir, "out", "lint", module, "go-golangci-lint")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest dir for %s: %w", module, err)
	}

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest for %s: %w", module, err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest for %s: %w", module, err)
	}

	return nil
}

func ensureNoModificationsInDir(ctx *eacgodog.TestContext, dir string) error {
	ctx.MustBeIsolated()
	fullDir := filepath.Join(ctx.IsolatedDir, dir)
	cmd := exec.Command("git", "checkout", "--", fullDir)
	cmd.Dir = ctx.IsolatedDir
	_ = cmd.Run()
	return nil
}

func buildSpecificModules(ctx *eacgodog.TestContext, mod1, mod2 string) error {
	ctx.MustBeIsolated()

	origContainerRoot := os.Getenv(environments.EnvCLIEContainerRoot)
	os.Setenv(environments.EnvCLIEContainerRoot, ctx.OriginalRepoRoot)
	defer func() {
		if origContainerRoot == "" {
			os.Unsetenv(environments.EnvCLIEContainerRoot)
		} else {
			os.Setenv(environments.EnvCLIEContainerRoot, origContainerRoot)
		}
	}()

	reg, err := modules.LoadFromWorkspaceNoValidation(ctx.IsolatedDir)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	moduleNames := []string{mod1, mod2}
	for _, moniker := range moduleNames {
		contract, ok := reg.Get(moniker)
		if !ok {
			return fmt.Errorf("module not found: %s", moniker)
		}
		files, err := hash.ExpandGlobPatterns(ctx.IsolatedDir, contract.GetGlobPatterns())
		if err != nil {
			return fmt.Errorf("failed to expand patterns for %s: %w", moniker, err)
		}
		sourceHash, err := hash.Files(ctx.IsolatedDir, files)
		if err != nil {
			return fmt.Errorf("failed to hash files for %s: %w", moniker, err)
		}

		// Create UoW manifest for successful build
		manifest := &output.UoWManifest{
			Action:     core.ActionBuild,
			Module:     moniker,
			Component:  "go",
			Tool:       "go",
			ExitCode:   0,
			InputHash:  sourceHash,
			ExecutedAt: time.Now().UTC().Truncate(time.Second),
			Duration:   time.Second,
			Artifacts:  []output.Artifact{},
			OutputHash: "sha256:output-" + moniker,
			Version:    "1.0.0",
		}

		manifestDir := filepath.Join(ctx.IsolatedDir, "out", "build", moniker, "go-go")
		if err := os.MkdirAll(manifestDir, 0755); err != nil {
			return fmt.Errorf("failed to create manifest dir for %s: %w", moniker, err)
		}

		manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
		data, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal manifest for %s: %w", moniker, err)
		}
		if err := os.WriteFile(manifestPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write manifest for %s: %w", moniker, err)
		}
	}

	return nil
}

func addNewModule(ctx *eacgodog.TestContext, moniker, goRoot string) error {
	ctx.MustBeIsolated()
	repoPath := filepath.Join(ctx.IsolatedDir, ".eac", "repository.yml")
	content, err := os.ReadFile(repoPath)
	if err != nil {
		return fmt.Errorf("failed to read repository.yml: %w", err)
	}

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

func setupCircularDependency(ctx *eacgodog.TestContext, table *godog.Table) error {
	ctx.MustBeIsolated()
	var mods []cacheModuleConfig
	for i, row := range table.Rows {
		if i == 0 {
			continue
		}
		if len(row.Cells) < 2 {
			continue
		}

		mods = append(mods, cacheModuleConfig{
			Moniker:   row.Cells[0].Value,
			DependsOn: row.Cells[1].Value,
			GoRoot:    fmt.Sprintf("go/%s", row.Cells[0].Value),
		})
	}

	repoYAML := generateCacheRepositoryYAML(mods)
	repoPath := filepath.Join(ctx.IsolatedDir, ".eac", "repository.yml")
	if err := os.WriteFile(repoPath, []byte(repoYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	for _, mod := range mods {
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

func corruptBuildState(ctx *eacgodog.TestContext, content string) error {
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
