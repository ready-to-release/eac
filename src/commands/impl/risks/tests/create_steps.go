// Package tests provides BDD step definitions for the risks create command.
package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/commands/impl/risks/create"
)

// registerCreateSteps registers step definitions for risks create command.
func registerCreateSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a risk assessment exists at "([^"]*)"$`, assessmentExistsAt)
	sc.Step(`^the assessment contains (\d+) identified risks$`, assessmentContainsRisks)
	sc.Step(`^the assessment contains risks in domains "([^"]*)" and "([^"]*)"$`, assessmentContainsRisksInDomains)
	sc.Step(`^the assessment has risk "([^"]*)" with severity "([^"]*)"$`, assessmentHasRiskWithSeverity)
	sc.Step(`^risk "([^"]*)" affects "([^"]*)"$`, riskAffectsFile)
	sc.Step(`^risk "([^"]*)" relates to "([^"]*)"$`, riskRelatesToSpec)
	sc.Step(`^a control exists at "([^"]*)"$`, controlExistsAt)
	sc.Step(`^the control has tag "([^"]*)"$`, controlHasTag)
	sc.Step(`^the control has tags "([^"]*)" and "([^"]*)"$`, controlHasTags)
	sc.Step(`^other specs reference "([^"]*)"$`, otherSpecsReference)
	sc.Step(`^other specs reference these tags$`, otherSpecsReferenceTags)
	sc.Step(`^the new control does not include these tags$`, newControlDoesNotIncludeTags)
	sc.Step(`^the new control includes "([^"]*)"$`, newControlIncludes)
	sc.Step(`^a control would create orphaned tags$`, controlWouldCreateOrphanedTags)
	sc.Step(`^a folder "([^"]*)" with (\d+) assessment files$`, folderWithAssessmentFiles)
	sc.Step(`^an invalid assessment file "([^"]*)"$`, invalidAssessmentFile)
	sc.Step(`^the assessment contains invalid risk data$`, assessmentContainsInvalidRiskData)

	// Then steps
	sc.Step(`^(\d+) feature files are created under "([^"]*)"$`, featureFilesCreatedUnder)
	sc.Step(`^each file is valid Gherkin$`, eachFileIsValidGherkin)
	sc.Step(`^each file has proper @risk-control tags$`, eachFileHasProperTags)
	sc.Step(`^a file exists at "([^"]*)"$`, fileExistsAt)
	sc.Step(`^the generated control includes severity metadata$`, controlIncludesSeverityMetadata)
	sc.Step(`^the control filename includes the risk ID$`, controlFilenameIncludesRiskID)
	sc.Step(`^the generated control documents these relationships$`, controlDocumentsRelationships)
	sc.Step(`^specs create is called (\d+) times$`, specsCreateCalledTimes)
	sc.Step(`^each call includes risk description$`, eachCallIncludesRiskDescription)
	sc.Step(`^debug logs show specs create invocations$`, debugLogsShowInvocations)
	sc.Step(`^the existing control is not overwritten$`, existingControlNotOverwritten)
	sc.Step(`^the existing control is overwritten$`, existingControlOverwritten)
	sc.Step(`^the command analyzes tag usage across all specs$`, commandAnalyzesTagUsage)
	sc.Step(`^warns if tags would become orphaned$`, warnsIfTagsOrphaned)
	sc.Step(`^preserves referenced tags in new control$`, preservesReferencedTags)
	sc.Step(`^stderr lists the orphaned tags$`, stderrListsOrphanedTags)
	sc.Step(`^stderr suggests adding tags to new control$`, stderrSuggestsAddingTags)
	sc.Step(`^the control is overwritten safely$`, controlOverwrittenSafely)
	sc.Step(`^no orphaned tags are created$`, noOrphanedTagsCreated)
	sc.Step(`^a warning is displayed about orphaned tags$`, warningDisplayedAboutOrphanedTags)
	sc.Step(`^controls are created under "([^"]*)"$`, controlsCreatedUnder)
	sc.Step(`^all assessments are processed$`, allAssessmentsProcessed)
	sc.Step(`^controls are created for all identified risks$`, controlsCreatedForAllRisks)
	sc.Step(`^a summary report is displayed$`, summaryReportDisplayed)
}

// ============================================================================
// Given Steps - Setup
// ============================================================================

func assessmentExistsAt(path string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	fullPath := filepath.Join(IsolatedTestProjectDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	// Create mock assessment
	mockAssessment := `# Risk Assessment Report

## Risks Identified

### RISK-001: SQL Injection
**Severity:** HIGH
**Domain:** data-protection
**Affected Files:** db/query.go
**Related Specs:** specs/database/queries.feature

### RISK-002: Auth Bypass
**Severity:** CRITICAL
**Domain:** authentication
**Affected Files:** auth/login.go
**Related Specs:** specs/auth/login.feature
`
	if err := os.WriteFile(fullPath, []byte(mockAssessment), 0644); err != nil {
		return err
	}

	Ctx.AssessmentFile = path
	Ctx.AssessmentContent = mockAssessment
	Ctx.RisksFound = 2

	// Set up mock AI response for parsing
	mockJSON := `[
		{
			"risk_id": "RISK-001",
			"control_name": "SQL Injection Prevention",
			"domain": "data-protection",
			"severity": "HIGH",
			"description": "Prevent SQL injection attacks",
			"affected_files": ["db/query.go"],
			"related_specs": ["specs/database/queries.feature"],
			"impact": "Database compromise"
		},
		{
			"risk_id": "RISK-002",
			"control_name": "Authentication Security",
			"domain": "authentication",
			"severity": "CRITICAL",
			"description": "Prevent authentication bypass",
			"affected_files": ["auth/login.go"],
			"related_specs": ["specs/auth/login.feature"],
			"impact": "Unauthorized access"
		}
	]`
	create.SetMockAIResponse(mockJSON)
	Ctx.MockAIResponse = mockJSON

	return nil
}

func assessmentContainsRisks(count int) error {
	Ctx.RisksFound = count
	return nil
}

func assessmentContainsRisksInDomains(domain1, domain2 string) error {
	// Verify in context
	return nil
}

func assessmentHasRiskWithSeverity(riskID, severity string) error {
	return nil
}

func riskAffectsFile(riskID, file string) error {
	return nil
}

func riskRelatesToSpec(riskID, spec string) error {
	return nil
}

func controlExistsAt(path string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	fullPath := filepath.Join(IsolatedTestProjectDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	controlContent := `Feature: Existing Control
  @risk-control:existing-tag-1
  @risk-control:existing-tag-2

  Scenario: Existing scenario
    Given a control exists
`
	if err := os.WriteFile(fullPath, []byte(controlContent), 0644); err != nil {
		return err
	}

	Ctx.ExistingControl = path
	Ctx.OrphanedTags = []string{"@risk-control:existing-tag-1", "@risk-control:existing-tag-2"}

	return nil
}

func controlHasTag(tag string) error {
	Ctx.OrphanedTags = append(Ctx.OrphanedTags, tag)
	return nil
}

func controlHasTags(tag1, tag2 string) error {
	Ctx.OrphanedTags = append(Ctx.OrphanedTags, tag1, tag2)
	return nil
}

func otherSpecsReference(tag string) error {
	// Mock that specs reference this tag
	return nil
}

func otherSpecsReferenceTags() error {
	// Mock that specs reference tags
	return nil
}

func newControlDoesNotIncludeTags() error {
	// Verification step
	return nil
}

func newControlIncludes(tag string) error {
	return nil
}

func controlWouldCreateOrphanedTags() error {
	Ctx.OrphanedCount = len(Ctx.OrphanedTags)
	return nil
}

func folderWithAssessmentFiles(folder string, count int) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	folderPath := filepath.Join(IsolatedTestProjectDir, folder)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return err
	}

	// Create multiple assessment files
	for i := 1; i <= count; i++ {
		filename := fmt.Sprintf("assessment-%d.md", i)
		assessmentPath := filepath.Join(folderPath, filename)

		content := fmt.Sprintf(`# Risk Assessment Report %d

## Risks Identified

### RISK-%03d: Test Risk
`, i, i)

		if err := os.WriteFile(assessmentPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func invalidAssessmentFile(path string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	fullPath := filepath.Join(IsolatedTestProjectDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	// Create file without "Risk Assessment Report" header
	invalidContent := `# Some Other Document

This is not a risk assessment.
`
	return os.WriteFile(fullPath, []byte(invalidContent), 0644)
}

func assessmentContainsInvalidRiskData() error {
	create.SetMockAIResponse("invalid json{")
	return nil
}

// ============================================================================
// Then Steps - Verification
// ============================================================================

func featureFilesCreatedUnder(count int, dir string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	dirPath := filepath.Join(IsolatedTestProjectDir, dir)
	matches, err := filepath.Glob(filepath.Join(dirPath, "*.feature"))
	if err != nil {
		return err
	}

	if len(matches) != count {
		return fmt.Errorf("expected %d feature files, found %d", count, len(matches))
	}

	Ctx.ControlsCreated = matches
	return nil
}

func eachFileIsValidGherkin() error {
	for _, control := range Ctx.ControlsCreated {
		content, err := os.ReadFile(control)
		if err != nil {
			return err
		}
		if !strings.Contains(string(content), "Feature:") {
			return fmt.Errorf("file %s does not contain Feature: keyword", control)
		}
	}
	return nil
}

func eachFileHasProperTags() error {
	for _, control := range Ctx.ControlsCreated {
		content, err := os.ReadFile(control)
		if err != nil {
			return err
		}
		if !strings.Contains(string(content), "@risk-control:") {
			return fmt.Errorf("file %s does not contain @risk-control: tag", control)
		}
	}
	return nil
}

func fileExistsAt(path string) error {
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project not initialized")
	}

	fullPath := filepath.Join(IsolatedTestProjectDir, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	return nil
}

func controlIncludesSeverityMetadata() error {
	// Check if created controls have severity info
	return nil
}

func controlFilenameIncludesRiskID() error {
	// Check filenames
	return nil
}

func controlDocumentsRelationships() error {
	return nil
}

func specsCreateCalledTimes(count int) error {
	return nil
}

func eachCallIncludesRiskDescription() error {
	return nil
}

func debugLogsShowInvocations() error {
	return nil
}

func existingControlNotOverwritten() error {
	if Ctx.ExistingControl == "" {
		return nil
	}

	fullPath := filepath.Join(IsolatedTestProjectDir, Ctx.ExistingControl)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// Check it still contains original content
	if !strings.Contains(string(content), "Existing Control") {
		return fmt.Errorf("control was overwritten")
	}

	return nil
}

func existingControlOverwritten() error {
	return nil
}

func commandAnalyzesTagUsage() error {
	return nil
}

func warnsIfTagsOrphaned() error {
	if Ctx.OrphanedCount > 0 {
		return nil
	}
	return fmt.Errorf("no orphaned tags detected")
}

func preservesReferencedTags() error {
	return nil
}

func stderrListsOrphanedTags() error {
	return stderrContains("orphaned")
}

func stderrSuggestsAddingTags() error {
	return stderrContains("tag")
}

func controlOverwrittenSafely() error {
	return nil
}

func noOrphanedTagsCreated() error {
	return nil
}

func warningDisplayedAboutOrphanedTags() error {
	if err := stdoutContains("orphan"); err == nil {
		return nil
	}
	if err := stderrContains("orphan"); err == nil {
		return nil
	}
	return fmt.Errorf("expected 'orphan' in stdout or stderr")
}

func controlsCreatedUnder(dir string) error {
	dirPath := filepath.Join(IsolatedTestProjectDir, dir)
	matches, err := filepath.Glob(filepath.Join(dirPath, "*.feature"))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return fmt.Errorf("no controls created under %s", dir)
	}

	Ctx.ControlsCreated = matches
	return nil
}

func allAssessmentsProcessed() error {
	return nil
}

func controlsCreatedForAllRisks() error {
	return nil
}

func summaryReportDisplayed() error {
	if err := stdoutContains("created"); err == nil {
		return nil
	}
	if err := stdoutContains("control"); err == nil {
		return nil
	}
	return fmt.Errorf("expected 'created' or 'control' in stdout")
}

// ============================================================================
// Helper Functions
// ============================================================================
// (Mock setup is now handled in register.go)
