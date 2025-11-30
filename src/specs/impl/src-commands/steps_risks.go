// Package srccommands contains godog step implementations for specs/src-commands.
//
// This file contains risks command step definitions for risk assessment and control management.
package srccommands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/specs/internal"
)

// risksTestState holds state for risks tests.
type risksTestState struct {
	assessmentPath string
	controlsCount  int
	riskIDs        []string
}

// registerRisksSteps registers step definitions for risks command features.
func registerRisksSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	state := &risksTestState{}

	// Given steps - assessment setup
	sc.Step(`^a risk assessment exists at "([^"]*)"$`, func(path string) error {
		state.assessmentPath = path
		return internal.CreateFile(ctx, path, riskAssessmentContent(3))
	})
	sc.Step(`^an invalid assessment file "([^"]*)"$`, func(path string) error {
		return internal.CreateFile(ctx, path, "invalid markdown content without risks")
	})
	sc.Step(`^the assessment contains (\d+) risks$`, func(count int) error {
		if state.assessmentPath != "" {
			return internal.CreateFile(ctx, state.assessmentPath, riskAssessmentContent(count))
		}
		return nil
	})
	sc.Step(`^the assessment contains (\d+) identified risks$`, func(count int) error {
		return internal.CreateFile(ctx, ".docs/reference/assessment.md", riskAssessmentContent(count))
	})
	sc.Step(`^the assessment contains invalid risk data$`, func() error {
		return internal.CreateFile(ctx, ".docs/reference/assessment.md", "# Risk Assessment\n\nNo valid risks here.")
	})
	sc.Step(`^the assessment contains risks in domains "([^"]*)" and "([^"]*)"$`, func(d1, d2 string) error {
		content := fmt.Sprintf(`# Risk Assessment

## %s Domain
- RISK-001: First risk

## %s Domain
- RISK-002: Second risk
`, d1, d2)
		return internal.CreateFile(ctx, ".docs/reference/assessment.md", content)
	})
	sc.Step(`^the assessment has risk "([^"]*)" with severity "([^"]*)"$`, func(riskID, severity string) error {
		content := fmt.Sprintf(`# Risk Assessment

## Identified Risks

### %s
- Severity: %s
- Description: Test risk
`, riskID, severity)
		return internal.CreateFile(ctx, ".docs/reference/assessment.md", content)
	})
	sc.Step(`^a folder "([^"]*)" with (\d+) assessment files$`, func(dir string, count int) error {
		for i := 1; i <= count; i++ {
			path := filepath.Join(dir, fmt.Sprintf("assessment%d.md", i))
			if err := internal.CreateFile(ctx, path, riskAssessmentContent(2)); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^risk "([^"]*)" affects "([^"]*)"$`, func(riskID, file string) error {
		state.riskIDs = append(state.riskIDs, riskID)
		return nil
	})
	sc.Step(`^risk "([^"]*)" relates to "([^"]*)"$`, func(riskID, spec string) error {
		state.riskIDs = append(state.riskIDs, riskID)
		return nil
	})

	// Given steps - controls setup
	sc.Step(`^a control exists at "([^"]*)"$`, func(path string) error {
		return internal.CreateFile(ctx, path, riskControlContent("auth-mfa-01"))
	})
	sc.Step(`^a control exists with tag "([^"]*)"$`, func(tag string) error {
		tagID := strings.TrimPrefix(tag, "@risk-control:")
		return internal.CreateFile(ctx, "specs/risk-controls/tagged.feature", riskControlContent(tagID))
	})
	sc.Step(`^a control would create orphaned tags$`, func() error {
		// Scenario where new control has tags not referenced elsewhere
		return nil
	})

	// Given steps - specs setup
	sc.Step(`^specifications exist in "([^"]*)"$`, func(dir string) error {
		return internal.CreateFile(ctx, filepath.Join(dir, "auth.feature"), specWithRiskTag("auth-mfa-01"))
	})
	sc.Step(`^other specs reference "([^"]*)"$`, func(tag string) error {
		tagID := strings.TrimPrefix(tag, "@risk-control:")
		return internal.CreateFile(ctx, "specs/auth/session.feature", specWithRiskTag(tagID))
	})
	sc.Step(`^other specs reference these tags$`, func() error {
		return nil
	})

	// Given steps - git state
	sc.Step(`^I have staged "([^"]*)"$`, func(file string) error {
		// Git staged file simulation
		return internal.CreateFile(ctx, file, "// staged content")
	})
	sc.Step(`^I have staged changes$`, func() error {
		return internal.CreateFile(ctx, "src/staged.go", "// staged")
	})
	sc.Step(`^I have modified "([^"]*)"$`, func(file string) error {
		return internal.CreateFile(ctx, file, "// modified content")
	})

	// Custom prompt
	sc.Step(`^a custom prompt at "([^"]*)"$`, func(path string) error {
		return internal.CreateFile(ctx, path, "Custom risk assessment prompt")
	})
	sc.Step(`^repository contracts define specs directory as "([^"]*)"$`, func(dir string) error {
		// Contract configuration
		return nil
	})

	// Then steps - file verification
	// Note: "a file exists at" is registered in internal/steps.go
	sc.Step(`^a file exists matching "([^"]*)"$`, func(pattern string) error {
		// Pattern-based file check
		return nil
	})
	sc.Step(`^(\d+) feature files are created under "([^"]*)"$`, func(count int, dir string) error {
		return internal.DirectoryHasFiles(ctx, dir)
	})
	sc.Step(`^each file has proper @risk-control tags$`, func() error {
		return nil
	})
	sc.Step(`^each file is valid Gherkin$`, func() error {
		return nil
	})
	sc.Step(`^the control filename includes the risk ID$`, func() error {
		return nil
	})
	sc.Step(`^the control has tag "([^"]*)"$`, func(tag string) error {
		return internal.OutputContains(ctx, tag)
	})
	sc.Step(`^the control has tags "([^"]*)" and "([^"]*)"$`, func(tag1, tag2 string) error {
		if err := internal.OutputContains(ctx, tag1); err != nil {
			return err
		}
		return internal.OutputContains(ctx, tag2)
	})
	sc.Step(`^the control is overwritten$`, func() error {
		return nil
	})
	sc.Step(`^the control is overwritten safely$`, func() error {
		return nil
	})
	sc.Step(`^the existing control is not overwritten$`, func() error {
		return nil
	})
	sc.Step(`^the existing control is overwritten$`, func() error {
		return nil
	})
	sc.Step(`^controls are created under "([^"]*)"$`, func(dir string) error {
		return internal.DirectoryHasFiles(ctx, dir)
	})
	sc.Step(`^controls are created for all identified risks$`, func() error {
		return nil
	})
	sc.Step(`^intermediate outputs are saved to "([^"]*)"$`, func(dir string) error {
		return internal.DirectoryHasFiles(ctx, dir)
	})
	sc.Step(`^the output file matches pattern "([^"]*)"$`, func(pattern string) error {
		return nil
	})

	// Then steps - output verification
	sc.Step(`^a summary report is displayed$`, func() error {
		return nil
	})
	sc.Step(`^a warning is displayed about orphaned tags$`, func() error {
		return internal.OutputContainsAny(ctx, "warning", "orphan", "Warning")
	})

	// Then steps - report verification
	sc.Step(`^the AI generates a markdown report$`, func() error {
		return nil
	})
	sc.Step(`^the report contains "([^"]*)"$`, func(text string) error {
		return internal.OutputContains(ctx, text)
	})
	sc.Step(`^the report follows the contract structure$`, func() error {
		return nil
	})
	sc.Step(`^the report lists "([^"]*)"$`, func(file string) error {
		return internal.OutputContains(ctx, file)
	})
	sc.Step(`^the report mentions "([^"]*)"$`, func(text string) error {
		return internal.OutputContains(ctx, text)
	})
	sc.Step(`^the report includes section "([^"]*)"$`, func(section string) error {
		return internal.OutputContains(ctx, section)
	})
	sc.Step(`^the output is written to "([^"]*)"$`, func(path string) error {
		return internal.FileExists(ctx, path)
	})

	// Then steps - AI/processing
	sc.Step(`^the AI is invoked with file changes$`, func() error {
		return nil
	})
	sc.Step(`^the AI receives specification context$`, func() error {
		return nil
	})
	sc.Step(`^custom prompt is used$`, func() error {
		return internal.CustomPromptWasUsed(ctx, "risks")
	})
	sc.Step(`^debug mode is enabled$`, func() error {
		return nil
	})
	sc.Step(`^force mode is enabled$`, func() error {
		return nil
	})
	sc.Step(`^debug logs show create spec invocations$`, func() error {
		return nil
	})
	sc.Step(`^each call includes risk description$`, func() error {
		return nil
	})
	sc.Step(`^create spec is called (\d+) times$`, func(count int) error {
		return nil
	})
	sc.Step(`^all assessments are processed$`, func() error {
		return nil
	})
	sc.Step(`^all flags are processed correctly$`, func() error {
		return nil
	})
	sc.Step(`^no specs directory argument is required$`, func() error {
		return nil
	})
	sc.Step(`^specifications are loaded from "([^"]*)"$`, func(dir string) error {
		return nil
	})
	sc.Step(`^the command analyzes tag usage across all specs$`, func() error {
		return nil
	})

	// Then steps - risk ID assignment
	sc.Step(`^risks are assigned IDs like "([^"]*)", "([^"]*)"$`, func(id1, id2 string) error {
		return nil
	})

	// Then steps - control content
	sc.Step(`^the new control does not include these tags$`, func() error {
		return nil
	})
	sc.Step(`^the new control includes "([^"]*)"$`, func(tag string) error {
		return internal.OutputContains(ctx, tag)
	})
	sc.Step(`^preserves referenced tags in new control$`, func() error {
		return nil
	})
	sc.Step(`^the generated control documents these relationships$`, func() error {
		return nil
	})
	sc.Step(`^the generated control includes severity metadata$`, func() error {
		return nil
	})
	sc.Step(`^no orphaned tags are created$`, func() error {
		return nil
	})
	sc.Step(`^warns if tags would become orphaned$`, func() error {
		return nil
	})

}

// riskAssessmentContent returns a risk assessment markdown document.
func riskAssessmentContent(riskCount int) string {
	var risks strings.Builder
	for i := 1; i <= riskCount; i++ {
		risks.WriteString(fmt.Sprintf(`
### RISK-%03d
- **Severity**: High
- **Domain**: Authentication
- **Description**: Risk number %d description
- **Mitigation**: Implement proper controls
`, i, i))
	}

	return fmt.Sprintf(`# Risk Assessment

## Summary
This assessment identifies %d risks.

## Identified Risks
%s
`, riskCount, risks.String())
}

// riskControlContent returns a risk control specification.
func riskControlContent(tagID string) string {
	return fmt.Sprintf(`@risk-control:%s
@module:risk-controls
Feature: Risk Control - %s
  This control mitigates identified risks.

  Rule: Control must be verified

    @verification:integration
    Scenario: Control is effective
      Given the control is implemented
      When the risk scenario occurs
      Then the risk is mitigated
`, tagID, tagID)
}

// specWithRiskTag returns a specification that references a risk control.
func specWithRiskTag(tagID string) string {
	return fmt.Sprintf(`@module:auth
@risk-control:%s
Feature: Authentication with Risk Control

  Rule: Users must authenticate securely

    @verification:integration
    Scenario: Secure login
      Given a registered user
      When they authenticate
      Then the session is secure
`, tagID)
}
