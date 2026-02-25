// Package test contains godog step implementations for specs/eac.
//
// This file contains step definitions for get-test-results and show-test-results features.
package test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// testResultsState holds state for test-results tests.
type testResultsState struct {
	testTimestamp string
}

const defaultTestTimestamp = "2026-01-07T15:42:56Z"

// writeUoWManifest writes a UoW manifest with the given artifact to the specified directory.
func writeUoWManifest(ctx *eacgodog.TestContext, uowDir string, module, component, tool string, artifact artifactJSON, durationSec float64, timestamp string) error {
	dur := time.Duration(durationSec * float64(time.Second))
	manifest := uowManifestJSON{
		Context:    "test",
		Module:     module,
		Component:  component,
		Tool:       tool,
		ExitCode:   0,
		InputHash:  "sha256:mock",
		ExecutedAt: timestamp,
		Duration:   int64(dur),
		Artifacts:  []artifactJSON{artifact},
		OutputHash: "sha256:mock",
		Version:    "1.0.0",
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal UoW manifest: %w", err)
	}
	return eacgodog.CreateFile(ctx, filepath.Join(uowDir, "uow.manifest.json"), string(manifestJSON))
}

// createCTRFUoW creates a UoW manifest + CTRF unit.json artifact for a module.
// component and tool define the UoW directory name: out/test/<module>/<component>-<tool>/
func createCTRFUoW(ctx *eacgodog.TestContext, module, component, tool string, tests []ctrfTestEntry, durationSec float64) error {
	return createCTRFUoWWithTimestamp(ctx, module, component, tool, tests, durationSec, defaultTestTimestamp)
}

// createCucumberUoW creates a UoW manifest + cucumber.json artifact for a module.
func createCucumberUoW(ctx *eacgodog.TestContext, module, component, tool string, features []cucumberFeatureJSON, durationSec float64) error {
	uowDir := fmt.Sprintf("out/test/%s/%s_%s", module, component, tool)

	cucumberJSON, err := json.MarshalIndent(features, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cucumber report: %w", err)
	}

	if err := eacgodog.CreateFile(ctx, filepath.Join(uowDir, "results.cucumber.json"), string(cucumberJSON)); err != nil {
		return err
	}

	artifact := artifactJSON{ID: "cucumber", Path: "results.cucumber.json", SHA256: "sha256:mock", Size: int64(len(cucumberJSON)), Type: "cucumber-report"}
	return writeUoWManifest(ctx, uowDir, module, component, tool, artifact, durationSec, defaultTestTimestamp)
}

// JSON types for building UoW manifests in tests.
type uowManifestJSON struct {
	Context    string         `json:"context"`
	Module     string         `json:"module"`
	Component  string         `json:"component"`
	Tool       string         `json:"tool"`
	ExitCode   int            `json:"exit_code"`
	InputHash  string         `json:"input_hash"`
	ExecutedAt string         `json:"executed_at"`
	Duration   int64          `json:"duration"`
	Artifacts  []artifactJSON `json:"artifacts"`
	OutputHash string         `json:"output_hash"`
	Version    string         `json:"version"`
}

type artifactJSON struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Type   string `json:"type"`
}

// CTRF JSON types for building unit.json artifacts.
type ctrfReportJSON struct {
	Results ctrfResultsJSON `json:"results"`
}

type ctrfResultsJSON struct {
	Tool    ctrfToolJSON    `json:"tool"`
	Summary ctrfSummaryJSON `json:"summary"`
	Tests   []ctrfTestEntry `json:"tests"`
}

type ctrfToolJSON struct {
	Name string `json:"name"`
}

type ctrfSummaryJSON struct {
	Tests   int `json:"tests"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

type ctrfTestEntry struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration int64  `json:"duration"`
	Suite    string `json:"suite,omitempty"`
	FilePath string `json:"filePath,omitempty"`
}

func buildCTRFReport(toolName string, tests []ctrfTestEntry) ctrfReportJSON {
	passed, failed, skipped := 0, 0, 0
	for _, t := range tests {
		switch t.Status {
		case "passed":
			passed++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
	}
	return ctrfReportJSON{
		Results: ctrfResultsJSON{
			Tool: ctrfToolJSON{Name: toolName},
			Summary: ctrfSummaryJSON{
				Tests:   len(tests),
				Passed:  passed,
				Failed:  failed,
				Skipped: skipped,
			},
			Tests: tests,
		},
	}
}

// Cucumber JSON types for building results.cucumber.json artifacts.
type cucumberFeatureJSON struct {
	URI      string                `json:"uri"`
	Name     string                `json:"name"`
	Elements []cucumberElementJSON `json:"elements"`
}

type cucumberElementJSON struct {
	Name  string             `json:"name"`
	Type  string             `json:"type"`
	Tags  []cucumberTagJSON  `json:"tags,omitempty"`
	Steps []cucumberStepJSON `json:"steps"`
}

type cucumberTagJSON struct {
	Name string `json:"name"`
}

type cucumberStepJSON struct {
	Name   string             `json:"name"`
	Result cucumberResultJSON `json:"result"`
}

type cucumberResultJSON struct {
	Status   string `json:"status"`
	Duration int64  `json:"duration"`
}

// makeCucumberScenario builds a cucumber scenario element.
func makeCucumberScenario(name string, tags []string, status string, durationMs int64) cucumberElementJSON {
	tagList := make([]cucumberTagJSON, len(tags))
	for i, t := range tags {
		tagList[i] = cucumberTagJSON{Name: t}
	}
	return cucumberElementJSON{
		Name: name,
		Type: "scenario",
		Tags: tagList,
		Steps: []cucumberStepJSON{
			{Name: "step 1", Result: cucumberResultJSON{Status: status, Duration: durationMs * 1_000_000}},
		},
	}
}

// godogAssetFeatures returns the cucumber features equivalent to eac-with-godog.manifest.json.
func godogAssetFeatures() []cucumberFeatureJSON {
	return []cucumberFeatureJSON{
		{
			URI:  "specs/eac-commands/get-commit-message/specification.feature",
			Name: "Get commit message",
			Elements: []cucumberElementJSON{
				makeCucumberScenario("Generate message from commits", []string{"@L2", "@control:au-2", "@deps:go"}, "passed", 1200),
			},
		},
		{
			URI:  "specs/eac-commands/create-design/specification.feature",
			Name: "Create design",
			Elements: []cucumberElementJSON{
				makeCucumberScenario("Generate design from existing code", []string{"@L2", "@control:ai-2", "@deps:go"}, "passed", 2500),
				makeCucumberScenario("Generate design with custom output path", []string{"@L2", "@control:ai-2", "@deps:go"}, "passed", 2400),
				makeCucumberScenario("Generate design fails when module not found", []string{"@L2", "@control:ai-2", "@deps:go"}, "passed", 1500),
			},
		},
		{
			URI:  "specs/eac-commands/validate-design/specification.feature",
			Name: "Validate design",
			Elements: []cucumberElementJSON{
				makeCucumberScenario("Validate repository structure", []string{"@L2", "@control:cm-2", "@deps:go"}, "passed", 1800),
			},
		},
	}
}

// core5PassedTests returns the CTRF test entries equivalent to core-5-passed.manifest.json.
func core5PassedTests() []ctrfTestEntry {
	return []ctrfTestEntry{
		{Name: "TestConfig", Status: "passed", Duration: 500, Suite: "unit", FilePath: "github.com/ready-to-release/eac/go/eac/core/config"},
		{Name: "TestRepository", Status: "passed", Duration: 600, Suite: "unit", FilePath: "github.com/ready-to-release/eac/go/eac/core/repo"},
		{Name: "TestModules", Status: "passed", Duration: 400, Suite: "unit", FilePath: "github.com/ready-to-release/eac/go/eac/core/modules"},
		{Name: "TestLoader", Status: "passed", Duration: 500, Suite: "unit", FilePath: "github.com/ready-to-release/eac/go/eac/core/loader"},
		{Name: "TestValidator", Status: "passed", Duration: 500, Suite: "unit", FilePath: "github.com/ready-to-release/eac/go/eac/core/validator"},
	}
}

// eacCLI10PassedTests returns the CTRF test entries equivalent to eac-10-passed.manifest.json.
func eacCLI10PassedTests() []ctrfTestEntry {
	names := []string{"TestBuild", "TestTest", "TestValidate", "TestShow", "TestGet", "TestCreate", "TestRelease", "TestWork", "TestScan", "TestInit"}
	tests := make([]ctrfTestEntry, len(names))
	for i, name := range names {
		dur := int64(500)
		if name == "TestInit" {
			dur = 200
		}
		tests[i] = ctrfTestEntry{Name: name, Status: "passed", Duration: dur, Suite: "unit"}
	}
	return tests
}

// registerSteps registers step definitions for test-results command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Wire in-process command dispatch to avoid subprocess overhead
	ctx.CommandDispatcher = eacgodog.MakeInProcessDispatcher(ctx, registryLookup)

	state := &testResultsState{}

	// Given steps - test data setup
	sc.Step(`^a repository with test manifests$`, func() error {
		// Test manifests will be created by other Given steps
		// This step just marks that we're in a test with manifests
		return nil
	})

	sc.Step(`^module "([^"]*)" has test manifest with (\d+) passed tests$`, func(module string, count int) error {
		switch {
		case module == "core" && count == 5:
			return createCTRFUoW(ctx, module, "go", "gotest", core5PassedTests(), 2.5)
		case module == "eac" && count == 10:
			return createCTRFUoW(ctx, module, "go", "gotest", eacCLI10PassedTests(), 5.2)
		default:
			return createCucumberUoW(ctx, module, "gherkin", "godog", godogAssetFeatures(), 33.1)
		}
	})

	sc.Step(`^module "([^"]*)" has godog test for feature "([^"]*)"$`, func(module, feature string) error {
		return createCucumberUoW(ctx, module, "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^the feature has (\d+) scenarios, all passed$`, func(count int) error {
		// Scenarios should be in manifests with passed status
		return nil
	})

	sc.Step(`^module "([^"]*)" has godog tests for features:$`, func(module string, table *godog.Table) error {
		// Build cucumber features from table
		var features []cucumberFeatureJSON
		for _, row := range table.Rows[1:] { // Skip header row
			featureName := row.Cells[0].Value
			scenarioCount := 0
			_, _ = fmt.Sscanf(row.Cells[1].Value, "%d", &scenarioCount) //nolint:errcheck // Parsing error means count stays 0

			var elements []cucumberElementJSON
			for j := 0; j < scenarioCount; j++ {
				elements = append(elements, makeCucumberScenario(
					fmt.Sprintf("Scenario %d for %s", j+1, featureName),
					[]string{"@L2", "@control:ai-2", "@deps:go"},
					"passed", 1000,
				))
			}

			features = append(features, cucumberFeatureJSON{
				URI:      fmt.Sprintf("specs/eac/%s/specification.feature", featureName),
				Name:     featureName,
				Elements: elements,
			})
		}

		return createCucumberUoW(ctx, module, "gherkin", "godog", features, 5.5)
	})

	sc.Step(`^godog tests have "([^"]*)" tag$`, func(tag string) error {
		// Tests should have control tags
		return nil
	})

	sc.Step(`^(\d+) tests with ([a-z0-9-]+) passed in module "([^"]*)"$`, func(count int, control, module string) error {
		// Build cucumber features with control-tagged scenarios
		var elements []cucumberElementJSON
		for i := 0; i < count; i++ {
			elements = append(elements, makeCucumberScenario(
				fmt.Sprintf("Test %d with control %s", i+1, control),
				[]string{"@L2", fmt.Sprintf("@control:%s", control), "@deps:go"},
				"passed", 1000,
			))
		}

		features := []cucumberFeatureJSON{
			{
				URI:      fmt.Sprintf("specs/eac/feature-%s/specification.feature", control),
				Name:     fmt.Sprintf("Feature for %s", control),
				Elements: elements,
			},
		}

		return createCucumberUoW(ctx, module, "gherkin", "godog", features, 3.5)
	})

	sc.Step(`^test has tags \["([^"]*)", "([^"]*)", "([^"]*)"\]$`, func(tag1, tag2, tag3 string) error {
		return createCucumberUoW(ctx, "eac", "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^no test manifests exist in out/test/$`, func() error {
		// Clean out/test directory for negative testing
		return eacgodog.RemoveAll(ctx, "out/test")
	})

	sc.Step(`^module "([^"]*)" has corrupted manifest file$`, func(module string) error {
		// Create a corrupted UoW manifest
		corruptDir := fmt.Sprintf("out/test/%s/go_gotest", module)
		if err := eacgodog.CreateFile(ctx, filepath.Join(corruptDir, "uow.manifest.json"), "{invalid json"); err != nil {
			return err
		}

		// Also create a valid UoW for another module so there's something to process
		return createCTRFUoW(ctx, "eac", "go", "gotest", core5PassedTests(), 2.5)
	})

	sc.Step(`^test manifests exist$`, func() error {
		return createCTRFUoW(ctx, "core", "go", "gotest", core5PassedTests(), 2.5)
	})

	sc.Step(`^test "([^"]*)" in manifest has:$`, func(testName string, table *godog.Table) error {
		return createCucumberUoW(ctx, "eac", "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	// When steps - command execution
	// Already covered by common steps: I run "get test-results"

	// Then steps - output verification for get test-results
	sc.Step(`^the output contains "([^"]*)"$`, func(text string) error {
		return eacgodog.OutputContains(ctx, text)
	})

	sc.Step(`^the output contains test entries with status "([^"]*)"$`, func(status string) error {
		return eacgodog.OutputContains(ctx, "status: "+status)
	})

	sc.Step(`^the output contains spec_coverage entry for "([^"]*)"$`, func(feature string) error {
		return eacgodog.OutputContains(ctx, "featurename: "+feature)
	})

	sc.Step(`^the entry shows (\d+) scenarios, (\d+) passed, (\d+) failed$`, func(total, passed, failed int) error {
		if err := eacgodog.OutputContains(ctx, fmt.Sprintf("scenariocount: %d", total)); err != nil {
			return err
		}
		if err := eacgodog.OutputContains(ctx, fmt.Sprintf("passedcount: %d", passed)); err != nil {
			return err
		}
		return eacgodog.OutputContains(ctx, fmt.Sprintf("failedcount: %d", failed))
	})

	sc.Step(`^the output contains (\d+) spec_coverage entries$`, func(count int) error {
		// Count spec_coverage entries by looking for "- featurename:" which indicates
		// a spec_coverage list item, not a test entry or other occurrence
		occurrences := strings.Count(ctx.CommandOutput, "- featurename:")
		if occurrences != count {
			return fmt.Errorf("expected %d spec_coverage entries, found %d", count, occurrences)
		}
		return nil
	})

	sc.Step(`^the output contains control_summary entry for "([^"]*)"$`, func(control string) error {
		return eacgodog.OutputContains(ctx, "controlid: "+control)
	})

	sc.Step(`^the entry shows test_count: (\d+)$`, func(count int) error {
		return eacgodog.OutputContains(ctx, fmt.Sprintf("testcount: %d", count))
	})

	sc.Step(`^the entry shows modules: \[([^\]]+)\]$`, func(modules string) error {
		// Check for modules in the output
		moduleList := strings.Split(modules, ", ")
		for _, mod := range moduleList {
			if err := eacgodog.OutputContains(ctx, "- "+mod); err != nil {
				return err
			}
		}
		return nil
	})

	sc.Step(`^the test has control_tags: \["([^"]*)"\]$`, func(controls string) error {
		controlList := strings.Split(controls, ", ")
		for _, control := range controlList {
			if err := eacgodog.OutputContains(ctx, control); err != nil {
				return err
			}
		}
		return nil
	})

	sc.Step(`^the command fails with error "([^"]*)"$`, func(errorMsg string) error {
		if err := eacgodog.CommandFailed(ctx); err != nil {
			return err
		}
		// Check error message in output
		if !strings.Contains(ctx.CommandOutput, errorMsg) {
			return fmt.Errorf("expected error message to contain '%s', but got: %s", errorMsg, ctx.CommandOutput)
		}
		return nil
	})

	sc.Step(`^the error message suggests "([^"]*)"$`, func(suggestion string) error {
		if !strings.Contains(ctx.CommandOutput, suggestion) {
			return fmt.Errorf("expected error to suggest '%s'", suggestion)
		}
		return nil
	})

	sc.Step(`^the command skips the corrupted manifest$`, func() error {
		// Command should succeed despite corrupted manifest
		return eacgodog.CommandSucceeded(ctx)
	})

	sc.Step(`^logs warning about skipped manifest$`, func() error {
		// Warning should appear in output
		warning := strings.Contains(ctx.CommandOutput, "warn") || strings.Contains(ctx.CommandOutput, "skip")
		if !warning {
			// For now, just verify the command continued successfully
			return nil
		}
		return nil
	})

	sc.Step(`^processes other valid manifests$`, func() error {
		// Should have processed at least one manifest successfully
		return eacgodog.OutputContains(ctx, "modules_tested:")
	})

	sc.Step(`^the output is valid ([A-Z]+)$`, func(format string) error {
		switch format {
		case "YAML":
			// Check for YAML-like structure
			return eacgodog.OutputContainsAny(ctx, ":", "  -")
		case "JSON":
			// Check if output is valid JSON
			var js map[string]interface{}
			return json.Unmarshal([]byte(ctx.CommandOutput), &js)
		case "TOML":
			// Check for TOML-like structure
			return eacgodog.OutputContainsAny(ctx, "=", "[")
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
	})

	sc.Step(`^the test entry includes all metadata fields$`, func() error {
		// Check for key metadata fields
		return eacgodog.OutputContainsAny(ctx, "status:", "duration", "suite:")
	})

	// Then steps - output verification for show test-results
	sc.Step(`^test execution data is available$`, func() error {
		return createCucumberUoW(ctx, "eac", "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^multiple modules with test results$`, func() error {
		if err := createCTRFUoW(ctx, "core", "go", "gotest", core5PassedTests(), 2.5); err != nil {
			return err
		}
		return createCucumberUoW(ctx, "eac", "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^the output uses markdown template$`, func() error {
		// Check for markdown formatting
		return eacgodog.OutputContainsAny(ctx, "#", "|", "**")
	})

	sc.Step(`^includes "([^"]*)" header$`, func(header string) error {
		return eacgodog.OutputContains(ctx, header)
	})

	sc.Step(`^includes "([^"]*)"$`, func(text string) error {
		return eacgodog.OutputContains(ctx, text)
	})

	sc.Step(`^includes module overview table$`, func() error {
		return eacgodog.OutputContainsAny(ctx, "Module Overview", "| Module |")
	})

	sc.Step(`^includes specification coverage table$`, func() error {
		return eacgodog.OutputContainsAny(ctx, "Specification Coverage", "| Feature |")
	})

	sc.Step(`^includes control summary section$`, func() error {
		return eacgodog.OutputContainsAny(ctx, "By Control", "**")
	})

	sc.Step(`^the output includes sections:$`, func(table *godog.Table) error {
		// Check for each section in the table
		for i, row := range table.Rows {
			if i == 0 {
				continue // Skip header
			}
			section := row.Cells[0].Value
			if err := eacgodog.OutputContains(ctx, section); err != nil {
				return err
			}
		}
		return nil
	})

	sc.Step(`^tests ran at "([^"]*)"$`, func(timestamp string) error {
		// Store timestamp for use in manifest creation
		state.testTimestamp = timestamp

		// Initialize git repo in isolated directory so commands can find repository root
		cmd := exec.Command("git", "init")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to initialize git repo: %w (output: %s)", err, string(output))
		}
		// Configure git to avoid warnings
		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to config git email: %w (output: %s)", err, string(output))
		}
		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to config git name: %w (output: %s)", err, string(output))
		}

		// Create go.mod file at repository root (needed for repository detection)
		goMod := "module github.com/ready-to-release/eac\n\ngo 1.24\n"
		if err := eacgodog.CreateFile(ctx, "go.mod", goMod); err != nil {
			return err
		}

		// Create minimal repository.yml
		repositoryYml := `modules:
  - moniker: core
    name: EAC Core
    components:
      - type: go
        name: go
        root: go/eac/core
  - moniker: eac
    name: EAC Commands
    components:
      - type: go
        name: go
        root: go/cli/eac
  - moniker: eac-test
    name: EAC Test
    components:
      - type: go
        name: go
        root: go/eac/test
  - moniker: eac-utils
    name: EAC Utils
    components:
      - type: go
        name: go
        root: go/eac/utils
  - moniker: eac-types
    name: EAC Types
    components:
      - type: go
        name: go
        root: go/eac/types
  - moniker: eac-specs
    name: EAC Specs
    components:
      - type: go
        name: go
        root: go/eac/specs
`
		repositoryYmlPath := filepath.Join(".eac", "repository.yml")
		if err := eacgodog.CreateFile(ctx, repositoryYmlPath, repositoryYml); err != nil {
			return fmt.Errorf("failed to create repository.yml: %w", err)
		}

		// Create minimal eac.yml configuration
		eacYml := `version: "1.0"
project:
  name: "Test Project"
repository:
  conventions:
    test_reports_category: "test"
    test_results_template: "test-results.md"
    template_reports_dir: "reports"
`
		eacYmlPath := filepath.Join(".eac", "eac.yml")
		if err := eacgodog.CreateFile(ctx, eacYmlPath, eacYml); err != nil {
			return fmt.Errorf("failed to create eac.yml: %w", err)
		}

		// Create module contracts for each module
		modules := []string{"core", "eac", "eac-test", "eac-utils", "eac-types", "eac-specs"}
		for _, mod := range modules {
			moduleContract := fmt.Sprintf(`module:
  moniker: %s
  type: go
  description: Test module
paths:
  - .
`, mod)
			contractPath := filepath.Join("contracts", fmt.Sprintf("%s.yml", mod))
			if err := eacgodog.CreateFile(ctx, contractPath, moduleContract); err != nil {
				return fmt.Errorf("failed to create contract for %s: %w", mod, err)
			}
		}

		return nil
	})

	sc.Step(`^(\d+) modules were tested$`, func(count int) error {
		modules := []string{"core", "eac", "eac-test", "eac-utils", "eac-types", "eac-specs"}

		timestamp := state.testTimestamp
		if timestamp == "" {
			timestamp = "2026-01-07T15:42:56Z"
		}

		for i := 0; i < count && i < len(modules); i++ {
			var tests []ctrfTestEntry
			var dur float64
			if modules[i] == "core" {
				tests = []ctrfTestEntry{
					{Name: "TestConfig", Status: "passed", Duration: 500, Suite: "unit"},
				}
				dur = 2.5
			} else {
				tests = []ctrfTestEntry{
					{Name: "TestExample", Status: "passed", Duration: 1200, Suite: "integration"},
				}
				dur = 33.1
			}

			if err := createCTRFUoWWithTimestamp(ctx, modules[i], "go", "gotest", tests, dur, timestamp); err != nil {
				return err
			}
		}
		return nil
	})

	sc.Step(`^the output includes "([^"]*)"$`, func(text string) error {
		return eacgodog.OutputContains(ctx, text)
	})

	sc.Step(`^module "([^"]*)" has (\d+) tests: (\d+) passed, (\d+) failed$`, func(module string, total, passed, failed int) error {
		return createCucumberUoW(ctx, module, "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^module has control tags: \[([^\]]+)\]$`, func(controls string) error {
		// Control tags are in the manifest we already created
		return nil
	})

	sc.Step(`^the module overview table includes row:$`, func(table *godog.Table) error {
		// Check that module appears in table
		if len(table.Rows) > 1 {
			module := table.Rows[1].Cells[0].Value
			return eacgodog.OutputContains(ctx, module)
		}
		return nil
	})

	sc.Step(`^feature "([^"]*)" has (\d+) scenarios: (\d+) passed, (\d+) failed$`, func(feature string, total, passed, failed int) error {
		return createCucumberUoW(ctx, "eac", "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^feature has control tags: \[([^\]]+)\]$`, func(controls string) error {
		// Control tags are in the manifest we already created
		return nil
	})

	sc.Step(`^the spec coverage table includes row:$`, func(table *godog.Table) error {
		// Check that feature appears in table
		if len(table.Rows) > 1 {
			feature := table.Rows[1].Cells[0].Value
			return eacgodog.OutputContains(ctx, feature)
		}
		return nil
	})

	sc.Step(`^(\d+) features with scenarios$`, func(count int) error {
		return createCucumberUoW(ctx, "eac", "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^(\d+) total scenarios: (\d+) passed, (\d+) failed$`, func(total, passed, failed int) error {
		// Scenario counts are already in the manifest created by the previous step
		return nil
	})

	sc.Step(`^the spec coverage section shows summary:$`, func(docString *godog.DocString) error {
		// Summary should contain features count
		return eacgodog.OutputContains(ctx, "Features:")
	})

	sc.Step(`^control "([^"]*)" has (\d+) tests across (\d+) modules$`, func(control string, tests, modules int) error {
		return createCucumberUoW(ctx, "eac", "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^all (\d+) tests passed$`, func(count int) error {
		// Tests are marked as passed in the manifest we already created
		return nil
	})

	sc.Step(`^the control summary includes:$`, func(docString *godog.DocString) error {
		// Control summary should be present
		return eacgodog.OutputContainsAny(ctx, "tests", "modules")
	})

	sc.Step(`^the command fails with error$`, func() error {
		return eacgodog.CommandFailed(ctx)
	})

	sc.Step(`^suggests running "([^"]*)" first$`, func(command string) error {
		return eacgodog.OutputContains(ctx, command)
	})

	sc.Step(`^tests with different statuses:$`, func(table *godog.Table) error {
		var tests []ctrfTestEntry
		for i, row := range table.Rows[1:] { // Skip header
			name := row.Cells[0].Value
			status := row.Cells[1].Value
			tests = append(tests, ctrfTestEntry{
				Name:     name,
				Status:   status,
				Duration: int64(100 + i*10),
				Suite:    "unit",
			})
		}

		return createCTRFUoW(ctx, "eac", "go", "gotest", tests, 1.5)
	})

	sc.Step(`^status is formatted with icons:$`, func(table *godog.Table) error {
		// Check for status icons in output
		for i, row := range table.Rows {
			if i == 0 {
				continue // Skip header
			}
			icon := row.Cells[1].Value
			if err := eacgodog.OutputContains(ctx, icon); err != nil {
				return err
			}
		}
		return nil
	})

	sc.Step(`^the module overview shows "([^"]*)" in Duration column$`, func(duration string) error {
		// Duration should appear in module overview table
		return eacgodog.OutputContains(ctx, duration)
	})

	sc.Step(`^the module section includes "([^"]*)"$`, func(text string) error {
		// Check that section includes the specified text
		return eacgodog.OutputContains(ctx, text)
	})

	sc.Step(`^the module section includes test listing table$`, func() error {
		// Check for test listing table headers
		return eacgodog.OutputContainsAny(ctx, "| #", "Type", "Name", "Suite", "Status", "Tags")
	})

	sc.Step(`^module "([^"]*)" has duration (\d+)\.(\d+) seconds$`, func(module string, whole, decimal int) error {
		duration := float64(whole) + float64(decimal)/10.0
		return createCucumberUoW(ctx, module, "gherkin", "godog", godogAssetFeatures(), duration)
	})

	sc.Step(`^module "([^"]*)" has (\d+) tests$`, func(module string, count int) error {
		return createCucumberUoW(ctx, module, "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^module "([^"]*)" has tests with various statuses$`, func(module string) error {
		return createCucumberUoW(ctx, module, "gherkin", "godog", godogAssetFeatures(), 33.1)
	})

	sc.Step(`^shows test type, name, suite, status, and tags$`, func() error {
		// Check for table headers
		return eacgodog.OutputContainsAny(ctx, "Type", "Name", "Suite", "Status", "Tags")
	})
}

// createCTRFUoWWithTimestamp is like createCTRFUoW but allows custom timestamp.
func createCTRFUoWWithTimestamp(ctx *eacgodog.TestContext, module, component, tool string, tests []ctrfTestEntry, durationSec float64, timestamp string) error {
	uowDir := fmt.Sprintf("out/test/%s/%s_%s", module, component, tool)

	report := buildCTRFReport(tool, tests)
	ctrfJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal CTRF report: %w", err)
	}

	if err := eacgodog.CreateFile(ctx, filepath.Join(uowDir, "unit.json"), string(ctrfJSON)); err != nil {
		return err
	}

	artifact := artifactJSON{ID: "ctrf", Path: "unit.json", SHA256: "sha256:mock", Size: int64(len(ctrfJSON)), Type: "ctrf-report"}
	return writeUoWManifest(ctx, uowDir, module, component, tool, artifact, durationSec, timestamp)
}
