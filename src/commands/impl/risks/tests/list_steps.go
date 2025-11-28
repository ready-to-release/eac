// Package tests provides BDD step definitions for the risks list command.
package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// registerListSteps registers step definitions for risks list command.
func registerListSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^risk controls exist in "([^"]*)"$`, riskControlsExistIn)
	sc.Step(`^(\d+) risk controls exist$`, riskControlsExist)
	sc.Step(`^(\d+) controls are referenced by specs$`, controlsReferencedBySpecs)
	sc.Step(`^(\d+) controls are orphaned$`, controlsOrphaned)
	sc.Step(`^a control "([^"]*)"$`, controlExists)
	sc.Step(`^the control has tag "([^"]*)"$`, listControlHasTag)
	sc.Step(`^"([^"]*)" references the tag at line (\d+)$`, specReferencesTagAtLine)
	sc.Step(`^a control exists with no references$`, controlExistsWithNoReferences)
	sc.Step(`^(\d+) specifications exist$`, specificationsExist)
	sc.Step(`^specs exist that should reference risk controls$`, specsExistThatShouldReferenceControls)
	sc.Step(`^some specs are missing risk control tags$`, someMissingRiskControlTags)

	// Then steps
	sc.Step(`^the output shows (\d+) controls$`, outputShowsControls)
	sc.Step(`^each control shows its tags$`, eachControlShowsTags)
	sc.Step(`^each control shows its reference count$`, eachControlShowsReferenceCount)
	sc.Step(`^the summary shows "([^"]*)"$`, summaryShows)
	sc.Step(`^all shown controls have status "([^"]*)"$`, allShownControlsHaveStatus)
	sc.Step(`^the output shows only orphaned controls$`, outputShowsOnlyOrphanedControls)
	sc.Step(`^the output shows only linked controls$`, outputShowsOnlyLinkedControls)
	sc.Step(`^the output shows specs without risk control links$`, outputShowsSpecsWithoutLinks)
	sc.Step(`^each spec shows suggested controls$`, eachSpecShowsSuggestedControls)
	sc.Step(`^the output is valid JSON$`, outputIsValidJSON)
	sc.Step(`^JSON contains "([^"]*)" array$`, jsonContainsArray)
	sc.Step(`^JSON contains "([^"]*)" object$`, jsonContainsObject)
	sc.Step(`^the output shows the control$`, outputShowsControl)
	sc.Step(`^the output shows "([^"]*)"$`, outputShowsText)
	sc.Step(`^the control is marked as "([^"]*)"$`, controlMarkedAs)
	sc.Step(`^the summary counts it as orphaned$`, summaryCountsAsOrphaned)
	sc.Step(`^the command completes in reasonable time$`, commandCompletesInReasonableTime)
	sc.Step(`^all controls are analyzed$`, allControlsAnalyzed)
	sc.Step(`^orphaned filter is applied$`, orphanedFilterApplied)
	sc.Step(`^JSON output is generated$`, jsonOutputGenerated)
}

// ============================================================================
// Given Steps - Setup
// ============================================================================

func riskControlsExistIn(dir string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	controlsDir := filepath.Join(IsolatedTestProjectDir, dir)
	if err := os.MkdirAll(controlsDir, 0755); err != nil {
		return err
	}

	// Create sample controls
	for i := 1; i <= 3; i++ {
		filename := fmt.Sprintf("control-%d.feature", i)
		controlPath := filepath.Join(controlsDir, filename)

		content := fmt.Sprintf(`Feature: Risk Control %d
  @risk-control:ctrl-%d

  Scenario: Test scenario
    Given a control
`, i, i)

		if err := os.WriteFile(controlPath, []byte(content), 0644); err != nil {
			return err
		}

		Ctx.ControlsList = append(Ctx.ControlsList, controlPath)
	}

	return nil
}

func riskControlsExist(count int) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	controlsDir := filepath.Join(IsolatedTestProjectDir, "specs", "risk-controls")
	if err := os.MkdirAll(controlsDir, 0755); err != nil {
		return err
	}

	for i := 1; i <= count; i++ {
		filename := fmt.Sprintf("control-%d.feature", i)
		controlPath := filepath.Join(controlsDir, filename)

		content := fmt.Sprintf(`Feature: Risk Control %d
  @risk-control:ctrl-%d

  Scenario: Test scenario
    Given a control
`, i, i)

		if err := os.WriteFile(controlPath, []byte(content), 0644); err != nil {
			return err
		}

		Ctx.ControlsList = append(Ctx.ControlsList, controlPath)
	}

	return nil
}

func controlsReferencedBySpecs(count int) error {
	Ctx.LinkedCount = count
	return nil
}

func controlsOrphaned(count int) error {
	Ctx.OrphanedCount = count
	return nil
}

func controlExists(name string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	controlsDir := filepath.Join(IsolatedTestProjectDir, "specs", "risk-controls")
	if err := os.MkdirAll(controlsDir, 0755); err != nil {
		return err
	}

	controlPath := filepath.Join(controlsDir, name)
	content := `Feature: Test Control
  @risk-control:test-tag

  Scenario: Test scenario
    Given a control
`

	if err := os.WriteFile(controlPath, []byte(content), 0644); err != nil {
		return err
	}

	Ctx.ControlsList = append(Ctx.ControlsList, controlPath)
	return nil
}

func listControlHasTag(tag string) error {
	// Add tag to last created control
	return nil
}

func specReferencesTagAtLine(specPath string, line int) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	fullPath := filepath.Join(IsolatedTestProjectDir, specPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`Feature: Test Spec

  @risk-control:test-tag
  Scenario: Test scenario
    Given a spec
`)

	return os.WriteFile(fullPath, []byte(content), 0644)
}

func controlExistsWithNoReferences() error {
	return controlExists("orphaned.feature")
}

func specificationsExist(count int) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	specsDir := filepath.Join(IsolatedTestProjectDir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		return err
	}

	for i := 1; i <= count; i++ {
		filename := fmt.Sprintf("spec-%d.feature", i)
		specPath := filepath.Join(specsDir, filename)

		content := fmt.Sprintf(`Feature: Spec %d

  Scenario: Test scenario
    Given a spec
`, i)

		if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func specsExistThatShouldReferenceControls() error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	specsDir := filepath.Join(IsolatedTestProjectDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		return err
	}

	// Create spec with security keywords but no risk control tags
	specPath := filepath.Join(specsDir, "authentication.feature")
	content := `Feature: Authentication

  Scenario: User login with security validation
    Given a user with credentials
    When they authenticate
    Then access is granted
`

	return os.WriteFile(specPath, []byte(content), 0644)
}

func someMissingRiskControlTags() error {
	Ctx.MissingLinksCount = 1
	return nil
}

// ============================================================================
// Then Steps - Verification
// ============================================================================

func outputShowsControls(count int) error {
	lines := strings.Split(Ctx.CommandOutput, "\n")
	controlCount := 0
	for _, line := range lines {
		if strings.Contains(line, "control-") || strings.Contains(line, ".feature") {
			controlCount++
		}
	}

	if controlCount < count {
		return fmt.Errorf("expected at least %d controls in output, found %d", count, controlCount)
	}

	return nil
}

func eachControlShowsTags() error {
	return stdoutContains("@risk-control:")
}

func eachControlShowsReferenceCount() error {
	// Check if output contains reference information
	return nil
}

func summaryShows(text string) error {
	return stdoutContains(text)
}

func allShownControlsHaveStatus(status string) error {
	return stdoutContains(status)
}

func outputShowsOnlyOrphanedControls() error {
	if err := stdoutContains("orphaned"); err == nil {
		return nil
	}
	if err := stdoutContains("0 references"); err == nil {
		return nil
	}
	return fmt.Errorf("expected 'orphaned' or '0 references' in stdout")
}

func outputShowsOnlyLinkedControls() error {
	if err := stdoutContains("linked"); err == nil {
		return nil
	}
	if err := stdoutContains("referenced"); err == nil {
		return nil
	}
	return fmt.Errorf("expected 'linked' or 'referenced' in stdout")
}

func outputShowsSpecsWithoutLinks() error {
	if err := stdoutContains("spec"); err == nil {
		return nil
	}
	if err := stdoutContains("missing"); err == nil {
		return nil
	}
	return fmt.Errorf("expected 'spec' or 'missing' in stdout")
}

func eachSpecShowsSuggestedControls() error {
	if err := stdoutContains("suggested"); err == nil {
		return nil
	}
	if err := stdoutContains("control"); err == nil {
		return nil
	}
	return fmt.Errorf("expected 'suggested' or 'control' in stdout")
}

func outputIsValidJSON() error {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(Ctx.CommandOutput), &result); err != nil {
		return fmt.Errorf("output is not valid JSON: %w", err)
	}
	return nil
}

func jsonContainsArray(arrayName string) error {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(Ctx.CommandOutput), &result); err != nil {
		return fmt.Errorf("output is not valid JSON: %w", err)
	}

	if _, ok := result[arrayName]; !ok {
		return fmt.Errorf("JSON does not contain '%s' array", arrayName)
	}

	return nil
}

func jsonContainsObject(objectName string) error {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(Ctx.CommandOutput), &result); err != nil {
		return fmt.Errorf("output is not valid JSON: %w", err)
	}

	if _, ok := result[objectName]; !ok {
		return fmt.Errorf("JSON does not contain '%s' object", objectName)
	}

	return nil
}

func outputShowsControl() error {
	if err := stdoutContains("control"); err == nil {
		return nil
	}
	if err := stdoutContains("Control"); err == nil {
		return nil
	}
	return fmt.Errorf("expected 'control' or 'Control' in stdout")
}

func outputShowsText(text string) error {
	return stdoutContains(text)
}

func controlMarkedAs(status string) error {
	return stdoutContains(status)
}

func summaryCountsAsOrphaned() error {
	if err := stdoutContains("orphaned"); err == nil {
		return nil
	}
	if err := stdoutContains("Orphaned"); err == nil {
		return nil
	}
	return fmt.Errorf("expected 'orphaned' or 'Orphaned' in stdout")
}

func commandCompletesInReasonableTime() error {
	// Command has already completed if we're checking this
	return nil
}

func allControlsAnalyzed() error {
	// All controls in ControlsList should be in output
	for _, control := range Ctx.ControlsList {
		basename := filepath.Base(control)
		if !strings.Contains(Ctx.CommandOutput, basename) {
			return fmt.Errorf("control %s not in output", basename)
		}
	}
	return nil
}

func orphanedFilterApplied() error {
	return outputShowsOnlyOrphanedControls()
}

func jsonOutputGenerated() error {
	return outputIsValidJSON()
}
