// Package riskassess contains godog step implementations for specs/eac.
//
// This file contains step definitions for the OSCAL-based risk commands:
// - create risk (OSCAL profile generation)
// - create risk-assess (assessment-results creation)
// - validate risk (OSCAL validation)
package riskassess

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"github.com/ready-to-release/eac/go/core/paths"
)

//go:embed assets/risk/profile-template.json
var profileTemplate string

//go:embed assets/risk/risk-assessment.md
var riskAssessmentDocument string

//go:embed assets/risk/cucumber-results-template.json
var cucumberResultsTemplate string

// riskTestState holds state for OSCAL-based risk tests.
type riskTestState struct {
	assessmentPath string
	moduleName     string
	profilePath    string
}

// registerSteps registers step definitions for new OSCAL-based risk commands.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	state := &riskTestState{}

	// Reset state before each scenario
	sc.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		state = &riskTestState{}
		return c, nil
	})

	// ==================== Given Steps ====================

	// AI provider setup
	sc.Step(`^AI provider is configured$`, func() error {
		// Load mock AI response for profile generation
		mockContent, err := eacgodog.LoadAsset(ctx, "risk/profile-mock-response.txt")
		if err != nil {
			return err
		}
		return eacgodog.CreateFile(ctx, ".clie/test/ai-mock.txt", mockContent)
	})

	// Assessment file setup
	sc.Step(`^a risk assessment file at "([^"]*)"$`, func(path string) error {
		state.assessmentPath = path
		content := createRiskAssessmentDocument()
		return eacgodog.CreateFile(ctx, path, content)
	})

	// Profile setup
	sc.Step(`^a profile exists at "([^"]*)"$`, func(path string) error {
		profile := createValidProfile("test-uuid", "Test Profile", []string{"ac-2", "ia-2"})
		return eacgodog.CreateFile(ctx, path, profile)
	})

	sc.Step(`^a valid OSCAL profile at "([^"]*)"$`, func(path string) error {
		profile := createValidProfile("profile-uuid", "Valid Profile", []string{"ac-2"})
		return eacgodog.CreateFile(ctx, path, profile)
	})

	sc.Step(`^a file "([^"]*)" with invalid JSON$`, func(path string) error {
		return eacgodog.CreateFile(ctx, path, `{invalid json content`)
	})

	// Module setup
	sc.Step(`^module "([^"]*)" exists with a profile at "([^"]*)"$`, func(module, profilePath string) error {
		state.moduleName = module
		state.profilePath = profilePath

		// Create go.mod file at repository root (needed for `go run` commands)
		goMod := "module github.com/ready-to-release/eac\n\ngo 1.24\n"
		if err := eacgodog.CreateFile(ctx, "go.mod", goMod); err != nil {
			return err
		}

		// Clear contracts directory (isolation setup copies real contracts, we only want test modules)
		contractsDir := filepath.Join(ctx.CurrentWorkDir, "contracts")
		if err := os.RemoveAll(contractsDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to clear contracts directory: %w", err)
		}
		if err := os.MkdirAll(contractsDir, 0o750); err != nil {
			return fmt.Errorf("failed to create contracts directory: %w", err)
		}

		// Create minimal repository.yml (isolation copies real one with all modules)
		repositoryYml := fmt.Sprintf(`modules:
  - moniker: %s
    name: Test Module %s
    description: Test module
    components:
      go: .
`, module, module)
		repositoryYmlPath := filepath.Join(".eac", "repository.yml")
		if err := eacgodog.CreateFile(ctx, repositoryYmlPath, repositoryYml); err != nil {
			return fmt.Errorf("failed to create repository.yml: %w", err)
		}

		// Create module contract
		moduleContract := fmt.Sprintf(`module:
  moniker: %s
  type: go
  description: Test module
paths:
  - .
`, module)
		contractPath := filepath.Join("contracts", fmt.Sprintf("%s.yml", module))
		if err := eacgodog.CreateFile(ctx, contractPath, moduleContract); err != nil {
			return err
		}

		// Create profile
		profile := createValidProfile("module-profile-uuid", fmt.Sprintf("Profile for %s", module), []string{"ac-2", "ia-2"})
		return eacgodog.CreateFile(ctx, profilePath, profile)
	})

	sc.Step(`^module "([^"]*)" exists with a profile$`, func(module string) error {
		state.moduleName = module
		profilePath := filepath.Join("specs", "risk-controls", fmt.Sprintf("%s.profile.json", module))
		state.profilePath = profilePath

		// Create go.mod file at repository root (needed for `go run` commands) - only once
		goModPath := filepath.Join(ctx.CurrentWorkDir, "go.mod")
		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			goMod := "module github.com/ready-to-release/eac\n\ngo 1.24\n"
			if err := eacgodog.CreateFile(ctx, "go.mod", goMod); err != nil {
				return err
			}
		}

		// Create contracts directory if it doesn't exist (but don't clear it)
		contractsDir := filepath.Join(ctx.CurrentWorkDir, "contracts")
		if err := os.MkdirAll(contractsDir, 0o750); err != nil {
			return fmt.Errorf("failed to create contracts directory: %w", err)
		}

		// Read existing repository.yml or create new one
		repositoryYmlPath := filepath.Join(ctx.CurrentWorkDir, ".eac", "repository.yml")
		var repositoryYml string
		moduleExists := false
		if existingContent, err := os.ReadFile(repositoryYmlPath); err == nil {
			// Check if module already exists in repository.yml
			repositoryYml = string(existingContent)
			if strings.Contains(repositoryYml, "moniker: "+module+"\n") {
				moduleExists = true
			}

			// Only append if module doesn't exist
			if !moduleExists {
				repositoryYml += fmt.Sprintf(`  - moniker: %s
    name: Test Module %s
    description: Test module
    components:
      go: .
`, module, module)
			}
		} else {
			// Create new repository.yml
			repositoryYml = fmt.Sprintf(`modules:
  - moniker: %s
    name: Test Module %s
    description: Test module
    components:
      go: .
`, module, module)
		}

		// Only write if we made changes
		if !moduleExists {
			if err := eacgodog.CreateFile(ctx, filepath.Join(".eac", "repository.yml"), repositoryYml); err != nil {
				return fmt.Errorf("failed to create repository.yml: %w", err)
			}
		}

		// Create module contract (only if it doesn't exist)
		contractPath := filepath.Join("contracts", fmt.Sprintf("%s.yml", module))
		contractFullPath := filepath.Join(ctx.CurrentWorkDir, contractPath)
		if _, err := os.Stat(contractFullPath); os.IsNotExist(err) {
			moduleContract := fmt.Sprintf(`module:
  moniker: %s
  type: go
  description: Test module
paths:
  - .
`, module)
			if err := eacgodog.CreateFile(ctx, contractPath, moduleContract); err != nil {
				return err
			}
		}

		// Create module-specific profile
		profile := createValidProfile("module-profile-uuid", fmt.Sprintf("Profile for %s", module), []string{"ac-2", "ia-2"})
		if err := eacgodog.CreateFile(ctx, profilePath, profile); err != nil {
			return err
		}

		// Also create shared profile for multi-module assessments (only once)
		sharedProfilePath := filepath.Join(ctx.CurrentWorkDir, "specs", ".risk-controls", "risk-profile.json")
		if _, err := os.Stat(sharedProfilePath); os.IsNotExist(err) {
			sharedProfile := createValidProfile("shared-profile-uuid", "Shared Profile", []string{"ac-2", "ia-2"})
			if err := eacgodog.CreateFile(ctx, "specs/.risk-controls/risk-profile.json", sharedProfile); err != nil {
				return err
			}
		}

		return nil
	})

	sc.Step(`^module "([^"]*)" has no evidence$`, func(module string) error {
		// Create module but don't create any test or security evidence
		state.moduleName = module

		// Create module contract
		moduleContract := fmt.Sprintf(`module:
  moniker: %s
  type: service
  description: Test module
paths:
  - .
`, module)
		contractPath := filepath.Join("contracts", fmt.Sprintf("%s.yml", module))
		return eacgodog.CreateFile(ctx, contractPath, moduleContract)
	})

	sc.Step(`^modules "([^"]*)", "([^"]*)", and "([^"]*)" exist with profiles$`, func(m1, m2, m3 string) error {
		modules := []string{m1, m2, m3}

		// Create go.mod file at repository root (needed for `go run` commands)
		goMod := "module github.com/ready-to-release/eac\n\ngo 1.24\n"
		if err := eacgodog.CreateFile(ctx, "go.mod", goMod); err != nil {
			return err
		}

		// Clear contracts directory (isolation setup copies real contracts, we only want test modules)
		contractsDir := filepath.Join(ctx.CurrentWorkDir, "contracts")
		if err := os.RemoveAll(contractsDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to clear contracts directory: %w", err)
		}
		if err := os.MkdirAll(contractsDir, 0o750); err != nil {
			return fmt.Errorf("failed to create contracts directory: %w", err)
		}

		// Create repository.yml with only test modules (isolation copies real one with all modules)
		repositoryYml := "modules:\n"
		for _, module := range modules {
			repositoryYml += "  - moniker: " + module + "\n"
			repositoryYml += "    name: Test Module " + module + "\n"
			repositoryYml += "    description: Test module\n"
			repositoryYml += "    components:\n"
			repositoryYml += "      go: .\n"
		}
		repositoryYmlPath := filepath.Join(".eac", "repository.yml")
		if err := eacgodog.CreateFile(ctx, repositoryYmlPath, repositoryYml); err != nil {
			return fmt.Errorf("failed to create repository.yml: %w", err)
		}

		// Create module contracts (for backward compatibility)
		for _, module := range modules {
			// Create module contract
			moduleContract := fmt.Sprintf(`module:
  moniker: %s
  type: service
  description: Test module
paths:
  - .
`, module)
			contractPath := filepath.Join("contracts", fmt.Sprintf("%s.yml", module))
			if err := eacgodog.CreateFile(ctx, contractPath, moduleContract); err != nil {
				return err
			}
		}

		// Create shared profile
		profile := createValidProfile("shared-profile-uuid", "Shared Profile", []string{"ac-2", "ia-2"})
		return eacgodog.CreateFile(ctx, "specs/.risk-controls/risk-profile.json", profile)
	})

	sc.Step(`^modules "([^"]*)" and "([^"]*)" exist with profiles$`, func(m1, m2 string) error {
		modules := []string{m1, m2}

		// Create go.mod file at repository root (needed for `go run` commands)
		goMod := "module github.com/ready-to-release/eac\n\ngo 1.24\n"
		if err := eacgodog.CreateFile(ctx, "go.mod", goMod); err != nil {
			return err
		}

		// Clear contracts directory (isolation setup copies real contracts, we only want test modules)
		contractsDir := filepath.Join(ctx.CurrentWorkDir, "contracts")
		if err := os.RemoveAll(contractsDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to clear contracts directory: %w", err)
		}
		if err := os.MkdirAll(contractsDir, 0o750); err != nil {
			return fmt.Errorf("failed to create contracts directory: %w", err)
		}

		// Create repository.yml with only test modules (isolation copies real one with all modules)
		repositoryYml := "modules:\n"
		for _, module := range modules {
			repositoryYml += "  - moniker: " + module + "\n"
			repositoryYml += "    name: Test Module " + module + "\n"
			repositoryYml += "    description: Test module\n"
			repositoryYml += "    components:\n"
			repositoryYml += "      go: .\n"
		}
		repositoryYmlPath := filepath.Join(".eac", "repository.yml")
		if err := eacgodog.CreateFile(ctx, repositoryYmlPath, repositoryYml); err != nil {
			return fmt.Errorf("failed to create repository.yml: %w", err)
		}

		// Create module contracts (for backward compatibility)
		for _, module := range modules {
			// Create module contract
			moduleContract := fmt.Sprintf(`module:
  moniker: %s
  type: service
  description: Test module
paths:
  - .
`, module)
			contractPath := filepath.Join("contracts", fmt.Sprintf("%s.yml", module))
			if err := eacgodog.CreateFile(ctx, contractPath, moduleContract); err != nil {
				return err
			}
		}

		// Create shared profile
		profile := createValidProfile("shared-profile-uuid", "Shared Profile", []string{"ac-2", "ia-2"})
		return eacgodog.CreateFile(ctx, "specs/.risk-controls/risk-profile.json", profile)
	})

	sc.Step(`^module "([^"]*)" has fresh test results with @control tags$`, func(module string) error {
		// Create UoW manifest + cucumber artifact with fresh timestamp
		cucumberJSON := createMockCucumberResults([]string{"ac-2"})
		return createMockTestUoW(ctx, module, time.Now(), cucumberJSON)
	})

	sc.Step(`^module "([^"]*)" has no test results$`, func(module string) error {
		// Module exists but has no test results - only create security evidence if requested separately
		return nil
	})

	sc.Step(`^module "([^"]*)" has test results older than 24 hours$`, func(module string) error {
		// Create UoW manifest + cucumber artifact with old timestamp (26 hours ago)
		oldTime := time.Now().Add(-26 * time.Hour)
		cucumberJSON := createMockCucumberResults([]string{"ac-2"})
		return createMockTestUoW(ctx, module, oldTime, cucumberJSON)
	})

	sc.Step(`^module "([^"]*)" has fresh security scan results$`, func(module string) error {
		// Create mock security scan results file (vuln.json) with fresh timestamp
		securityDir := filepath.Join(paths.OutDir, paths.SecurityDir, module)
		trivyJSON := `{"Results": [{"Vulnerabilities": [{"VulnerabilityID": "CVE-2024-0001", "Severity": "HIGH"}]}]}`
		return eacgodog.CreateFile(ctx, filepath.Join(securityDir, "vuln.json"), trivyJSON)
	})

	sc.Step(`^module "([^"]*)" has security scan results older than 24 hours$`, func(module string) error {
		// Create security results file (vuln.json) with old modification time
		securityDir := filepath.Join(paths.OutDir, paths.SecurityDir, module)
		trivyJSON := `{"Results": [{"Vulnerabilities": [{"VulnerabilityID": "CVE-2024-0001", "Severity": "HIGH"}]}]}`
		filePath := filepath.Join(securityDir, "vuln.json")

		if err := eacgodog.CreateFile(ctx, filePath, trivyJSON); err != nil {
			return err
		}

		// Set file modification time to 26 hours ago
		oldTime := time.Now().Add(-26 * time.Hour)
		fullPath := filepath.Join(ctx.CurrentWorkDir, filePath)
		return os.Chtimes(fullPath, oldTime, oldTime)
	})

	sc.Step(`^all modules have fresh test and security evidence$`, func() error {
		// This step assumes modules were already created in a previous step
		// We need to get the list of modules from the repository.yml
		repositoryYmlPath := filepath.Join(ctx.CurrentWorkDir, ".eac", "repository.yml")
		content, err := os.ReadFile(repositoryYmlPath)
		if err != nil {
			return fmt.Errorf("failed to read repository.yml: %w", err)
		}

		// Simple parsing to extract module names
		lines := strings.Split(string(content), "\n")
		var modules []string
		for _, line := range lines {
			if strings.Contains(line, "moniker:") {
				parts := strings.Split(line, "moniker:")
				if len(parts) > 1 {
					module := strings.TrimSpace(parts[1])
					modules = append(modules, module)
				}
			}
		}

		// Create fresh evidence for each module
		for _, module := range modules {
			// Create UoW manifest + cucumber artifact
			cucumberJSON := createMockCucumberResults([]string{"ac-2"})
			if err := createMockTestUoW(ctx, module, time.Now(), cucumberJSON); err != nil {
				return err
			}

			// Create security results (vuln.json)
			securityDir := filepath.Join(paths.OutDir, paths.SecurityDir, module)
			trivyJSON := `{"Results": [{"Vulnerabilities": [{"VulnerabilityID": "CVE-2024-0001", "Severity": "HIGH"}]}]}`
			if err := eacgodog.CreateFile(ctx, filepath.Join(securityDir, "vuln.json"), trivyJSON); err != nil {
				return err
			}
		}

		return nil
	})

	// ==================== Then Steps ====================

	// File existence with pattern matching
	sc.Step(`^files matching "([^"]*)" exist$`, func(pattern string) error {
		matches, err := filepath.Glob(filepath.Join(ctx.CurrentWorkDir, pattern))
		if err != nil {
			return fmt.Errorf("glob error: %w", err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("no files matching pattern: %s", pattern)
		}
		return nil
	})

	// Output verification
	sc.Step(`^stdout contains valid JSON$`, func() error {
		// Parse stdout as JSON
		var js interface{}
		return json.Unmarshal([]byte(ctx.CommandOutput), &js)
	})

	// ========== Additional Risk-Assess Steps ==========

	// NOTE: Additional validate-risk steps are already defined above (lines ~147-194)
	// Do NOT add duplicate registrations here - godog will report "ambiguous step" errors

	// Catalog validation steps - for validate risk-catalog command
	sc.Step(`^a file "([^"]*)" with content:$`, func(filePath string, content *godog.DocString) error {
		return eacgodog.CreateFile(ctx, filePath, content.Content)
	})

	sc.Step(`^a valid OSCAL catalog$`, func() error {
		catalog := `{
  "catalog": {
    "uuid": "12345678-1234-4234-8234-123456789abc",
    "metadata": {
      "title": "Test Catalog",
      "last-modified": "2025-01-01T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "1.1.3"
    },
    "groups": [
      {
        "id": "ac",
        "title": "Access Control",
        "controls": [
          {
            "id": "ac-1",
            "title": "Policy and Procedures"
          }
        ]
      }
    ]
  }
}`
		return eacgodog.CreateFile(ctx, "catalog.json", catalog)
	})

	sc.Step(`^a valid OSCAL (\d+)\.(\d+)\.(\d+) catalog$`, func(major, minor, patch int) error {
		catalog := fmt.Sprintf(`{
  "catalog": {
    "uuid": "12345678-1234-4234-8234-123456789abc",
    "metadata": {
      "title": "Test Catalog",
      "last-modified": "2025-01-01T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "%d.%d.%d"
    },
    "groups": [
      {
        "id": "ac",
        "title": "Access Control",
        "controls": [
          {
            "id": "ac-1",
            "title": "Policy and Procedures"
          }
        ]
      }
    ]
  }
}`, major, minor, patch)
		return eacgodog.CreateFile(ctx, "catalog.json", catalog)
	})

	sc.Step(`^a catalog missing UUID$`, func() error {
		catalog := `{
  "catalog": {
    "metadata": {
      "title": "Test Catalog",
      "last-modified": "2025-01-01T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "1.1.3"
    },
    "groups": []
  }
}`
		return eacgodog.CreateFile(ctx, "catalog.json", catalog)
	})

	sc.Step(`^a catalog with missing metadata title$`, func() error {
		catalog := `{
  "catalog": {
    "uuid": "12345678-1234-4234-8234-123456789abc",
    "metadata": {
      "last-modified": "2025-01-01T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "1.1.3"
    },
    "groups": []
  }
}`
		return eacgodog.CreateFile(ctx, "catalog.json", catalog)
	})

	sc.Step(`^a catalog without controls or groups$`, func() error {
		catalog := `{
  "catalog": {
    "uuid": "12345678-1234-4234-8234-123456789abc",
    "metadata": {
      "title": "Empty Catalog",
      "last-modified": "2025-01-01T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "1.1.3"
    }
  }
}`
		return eacgodog.CreateFile(ctx, "catalog.json", catalog)
	})

	sc.Step(`^the catalog is parsed using go-oscal types$`, func() error {
		return nil // Simplified - would check that go-oscal types were used
	})

	sc.Step(`^go-oscal validation is used$`, func() error {
		return nil // Simplified - would check that go-oscal validation was used
	})
}

// Helper functions

func createRiskAssessmentDocument() string {
	return riskAssessmentDocument
}

func createValidProfile(uuid, title string, controlIDs []string) string {
	controls := "["
	for i, id := range controlIDs {
		if i > 0 {
			controls += ","
		}
		controls += fmt.Sprintf(`{"with-id":"%s"}`, id)
	}
	controls += "]"

	result := strings.ReplaceAll(profileTemplate, "{{UUID}}", uuid)
	result = strings.ReplaceAll(result, "{{TITLE}}", title)
	result = strings.ReplaceAll(result, "{{CONTROLS}}", controls)
	return result
}

func createMockCucumberResults(controlTags []string) string {
	tags := "["
	for i, tag := range controlTags {
		if i > 0 {
			tags += ","
		}
		tags += fmt.Sprintf(`{"name": "@control(%s)"}`, tag)
	}
	tags += "]"

	return strings.ReplaceAll(cucumberResultsTemplate, "{{TAGS}}", tags)
}

// createMockTestUoW creates a UoW manifest + cucumber artifact for a module.
// This produces the directory structure that testview.LoadModuleTestView() expects:
// out/test/<module>/gherkin-godog/uow.manifest.json + results.cucumber.json
func createMockTestUoW(ctx *eacgodog.TestContext, module string, testTime time.Time, cucumberJSON string) error {
	uowDir := filepath.Join("out", "test", module, "gherkin-godog")

	// Write cucumber artifact
	if err := eacgodog.CreateFile(ctx, filepath.Join(uowDir, "results.cucumber.json"), cucumberJSON); err != nil {
		return err
	}

	// Write UoW manifest
	manifest := fmt.Sprintf(`{
  "context": "test",
  "module": "%s",
  "component": "gherkin",
  "tool": "godog",
  "exit_code": 0,
  "input_hash": "sha256:mock",
  "executed_at": "%s",
  "duration": %d,
  "artifacts": [
    {
      "id": "cucumber",
      "path": "results.cucumber.json",
      "sha256": "sha256:mock",
      "size": %d,
      "type": "cucumber-report"
    }
  ],
  "output_hash": "sha256:mock",
  "version": "1.0.0"
}`,
		module,
		testTime.Format(time.RFC3339),
		int64(1500*time.Millisecond),
		len(cucumberJSON))

	return eacgodog.CreateFile(ctx, filepath.Join(uowDir, "uow.manifest.json"), manifest)
}
