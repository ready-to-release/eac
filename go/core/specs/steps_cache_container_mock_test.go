// Package specs contains godog step implementations for core features.
//
// This file contains container registry mocking and container module setup
// functions for per-component container change detection BDD tests.
package specs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/environments"
	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

// containerModuleConfig describes a container component in the test repository.
type containerModuleConfig struct {
	Name string
	Root string
}

// containerRegistryMock is per-scenario state for mocked container registry responses.
type containerRegistryMock struct {
	// shas maps component name to last-build SHA (empty string = no tag).
	shas map[string]string
	// errors maps component name to error sentinel.
	errors map[string]bool
}

// containerMockCtx is per-scenario state, reset alongside cacheCtx.
var containerMockCtx containerRegistryMock

func resetContainerMockContext() {
	containerMockCtx = containerRegistryMock{
		shas:   make(map[string]string),
		errors: make(map[string]bool),
	}
}

// setupContainerModule creates a module with container components in the isolated repo.
// It adds container directories with Dockerfiles and updates repository.yml.
func setupContainerModule(ctx *eacgodog.TestContext, moduleName string, table *godog.Table) error {
	ctx.MustBeIsolated()

	var components []containerModuleConfig
	for i, row := range table.Rows {
		if i == 0 {
			continue // header
		}
		if len(row.Cells) < 2 {
			continue
		}
		components = append(components, containerModuleConfig{
			Name: row.Cells[0].Value,
			Root: row.Cells[1].Value,
		})
	}

	if len(components) == 0 {
		return fmt.Errorf("no container components specified")
	}

	// Create container directories with Dockerfiles
	for _, comp := range components {
		dir := filepath.Join(ctx.IsolatedDir, comp.Root)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create container dir %s: %w", comp.Root, err)
		}
		dockerfile := filepath.Join(dir, "Dockerfile")
		content := fmt.Sprintf("FROM alpine:3.18\nLABEL component=%s\nCOPY . /app\n", comp.Name)
		if err := os.WriteFile(dockerfile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write Dockerfile for %s: %w", comp.Name, err)
		}
		// Add a build script for file ownership testing
		scriptDir := filepath.Join(dir, "scripts")
		if err := os.MkdirAll(scriptDir, 0o755); err != nil {
			return fmt.Errorf("failed to create scripts dir: %w", err)
		}
		script := filepath.Join(scriptDir, "build.sh")
		if err := os.WriteFile(script, []byte("#!/bin/sh\necho building "+comp.Name+"\n"), 0o644); err != nil {
			return fmt.Errorf("failed to write build.sh for %s: %w", comp.Name, err)
		}
	}

	// Update repository.yml to add the container module
	repoPath := filepath.Join(ctx.IsolatedDir, ".eac", "repository.yml")
	content, err := os.ReadFile(repoPath)
	if err != nil {
		return fmt.Errorf("failed to read repository.yml: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  - moniker: %s\n", moduleName))
	sb.WriteString(fmt.Sprintf("    name: %s Module\n", moduleName))
	sb.WriteString("    description: Container module for testing\n")
	sb.WriteString("    components:\n")
	for _, comp := range components {
		sb.WriteString("      - type: dockerfile\n")
		sb.WriteString(fmt.Sprintf("        name: %s\n", comp.Name))
		sb.WriteString(fmt.Sprintf("        root: %s\n", comp.Root))
		sb.WriteString("        patterns:\n")
		sb.WriteString("          source: [\"**/*\"]\n")
	}

	newContent := string(content) + sb.String()
	if err := os.WriteFile(repoPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	// Commit the changes so git diff works correctly
	if err := coretesting.GitAddAll(ctx.IsolatedDir); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}
	if _, err := coretesting.GitCommit(ctx.IsolatedDir, "Add container module "+moduleName); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	// Re-capture HEAD SHA after the commit
	captureHeadSHA(ctx)

	// Invalidate cached registry since repository.yml changed
	cacheCtx.moduleRegistry = nil

	return nil
}

// mockContainerRegistryNoTags sets up an empty mock container registry (no previous builds).
func mockContainerRegistryNoTags() error {
	containerMockCtx.shas = make(map[string]string)
	containerMockCtx.errors = make(map[string]bool)
	return nil
}

// mockContainerRegistryFromTable sets up mock registry responses from a Gherkin table.
// The "last_build_sha" column supports a special value "HEAD_SHORT" which is replaced
// with the first 7 characters of the current HEAD SHA.
func mockContainerRegistryFromTable(table *godog.Table) error {
	for i, row := range table.Rows {
		if i == 0 {
			continue // header
		}
		if len(row.Cells) < 2 {
			continue
		}
		component := row.Cells[0].Value
		sha := row.Cells[1].Value

		// Replace HEAD_SHORT with actual short HEAD SHA
		if sha == "HEAD_SHORT" {
			if len(cacheCtx.currentHeadSHA) >= 7 {
				sha = cacheCtx.currentHeadSHA[:7]
			} else {
				sha = cacheCtx.currentHeadSHA
			}
		}

		containerMockCtx.shas[component] = sha
	}
	return nil
}

// mockContainerRegistryError marks a component to return an error from the registry.
func mockContainerRegistryError(component string) error {
	containerMockCtx.errors[component] = true
	return nil
}

// commitChangeToFile creates a new commit with a modification to the specified file.
func commitChangeToFile(ctx *eacgodog.TestContext, filePath string) error {
	ctx.MustBeIsolated()

	fullPath := filepath.Join(ctx.IsolatedDir, filePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
	}

	// Append a change to the file (or create it if it doesn't exist)
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	if _, err := f.WriteString("\n# change for test\n"); err != nil {
		f.Close()
		return fmt.Errorf("failed to write to %s: %w", filePath, err)
	}
	f.Close()

	if err := coretesting.GitAddAll(ctx.IsolatedDir); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}
	if _, err := coretesting.GitCommit(ctx.IsolatedDir, "Modify "+filePath); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	// Re-capture HEAD SHA
	captureHeadSHA(ctx)
	// Invalidate cached registry
	cacheCtx.moduleRegistry = nil

	return nil
}

// writeContainerRegistryMock writes the mock registry JSON file and sets the env var.
func writeContainerRegistryMock(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()

	// Build the mock JSON: {"component": "sha", ...}
	// For errors, use "__ERROR__" sentinel that the mock querier treats as an error.
	mockData := make(map[string]string)
	for comp, sha := range containerMockCtx.shas {
		mockData[comp] = sha
	}
	for comp := range containerMockCtx.errors {
		mockData[comp] = "__ERROR__"
	}

	mockJSON, err := json.Marshal(mockData)
	if err != nil {
		return fmt.Errorf("failed to marshal mock registry: %w", err)
	}

	mockDir := filepath.Join(ctx.IsolatedDir, "out", "test-mocks")
	if err := os.MkdirAll(mockDir, 0o755); err != nil {
		return fmt.Errorf("failed to create mock dir: %w", err)
	}

	mockPath := filepath.Join(mockDir, "container-registry.json")
	if err := os.WriteFile(mockPath, mockJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write mock registry file: %w", err)
	}

	ctx.SetMockOverride(environments.EnvEACMockContainerRegistry, mockPath)
	return nil
}

// runCommandWithMockedContainerRegistry writes mock data and runs the command.
func runCommandWithMockedContainerRegistry(ctx *eacgodog.TestContext, cmdLine string) error {
	if err := writeContainerRegistryMock(ctx); err != nil {
		return err
	}

	// Also set the HEAD SHA mock for commands that need it
	ctx.SetMockOverride(environments.EnvCLIEMockHeadSHA, cacheCtx.currentHeadSHA)

	if err := ctx.RunCommand(cmdLine); err != nil {
		return err
	}

	parseYAMLOutput(ctx)
	return nil
}

// yamlFieldHasNEntries asserts that a YAML array field has exactly N entries.
func yamlFieldHasNEntries(ctx *eacgodog.TestContext, fieldPath string, n int) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		if n == 0 {
			return nil // field not found = 0 entries
		}
		return err
	}

	switch v := value.(type) {
	case []interface{}:
		if len(v) != n {
			return fmt.Errorf("expected %s to have %d entries, got %d: %v\nFull output:\n%s",
				fieldPath, n, len(v), v, ctx.CommandOutput)
		}
	case nil:
		if n != 0 {
			return fmt.Errorf("expected %s to have %d entries, got nil", fieldPath, n)
		}
	default:
		return fmt.Errorf("expected %s to be an array, got %T", fieldPath, value)
	}
	return nil
}

// yamlFieldContainsComponent asserts that a YAML array contains an object with name=component.
func yamlFieldContainsComponent(ctx *eacgodog.TestContext, fieldPath, componentName string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	items, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected %s to be an array, got %T", fieldPath, value)
	}

	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := m["name"]; ok && fmt.Sprintf("%v", name) == componentName {
			return nil
		}
	}

	return fmt.Errorf("expected %s to contain component %q, got %v\nFull output:\n%s",
		fieldPath, componentName, items, ctx.CommandOutput)
}

// yamlArrayAllHaveReason asserts that all entries in a YAML array have the specified reason.
func yamlArrayAllHaveReason(ctx *eacgodog.TestContext, fieldPath, expectedReason string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	items, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected %s to be an array, got %T", fieldPath, value)
	}

	if len(items) == 0 {
		return fmt.Errorf("expected %s to have entries, got empty array", fieldPath)
	}

	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		reason, ok := m["reason"]
		if !ok {
			return fmt.Errorf("entry in %s has no 'reason' field", fieldPath)
		}
		if fmt.Sprintf("%v", reason) != expectedReason {
			return fmt.Errorf("expected all entries in %s to have reason %q, but found %q",
				fieldPath, expectedReason, reason)
		}
	}
	return nil
}

// yamlComponentReasonContains asserts that a specific component's reason contains a substring.
func yamlComponentReasonContains(ctx *eacgodog.TestContext, fieldPath, componentName, expectedSubstring string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	items, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected %s to be an array, got %T", fieldPath, value)
	}

	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := m["name"]
		if !ok || fmt.Sprintf("%v", name) != componentName {
			continue
		}
		reason, ok := m["reason"]
		if !ok {
			return fmt.Errorf("component %q in %s has no 'reason' field", componentName, fieldPath)
		}
		if !strings.Contains(fmt.Sprintf("%v", reason), expectedSubstring) {
			return fmt.Errorf("expected component %q reason to contain %q, got %q",
				componentName, expectedSubstring, reason)
		}
		return nil
	}

	return fmt.Errorf("component %q not found in %s\nFull output:\n%s",
		componentName, fieldPath, ctx.CommandOutput)
}
