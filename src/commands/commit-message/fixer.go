package commitmessage

import (
	"bytes"
	"fmt"
)

// FixWithFeedback attempts to fix validation errors by feeding them back to Claude
func FixWithFeedback(agentFilePath string, originalPrompt string, commitMessage string, validationErrors []ValidationError, apiCaller func(string, string) (string, error)) (string, error) {
	if len(validationErrors) == 0 {
		return commitMessage, nil
	}

	// Build feedback prompt
	feedbackPrompt := buildValidationFeedback(commitMessage, validationErrors)

	// Build conversation with feedback
	// Format: <original prompt> + <validation feedback>
	conversationWithFeedback := fmt.Sprintf("%s\n\n---\n\nYou generated:\n\n%s\n\n---\n\n%s",
		originalPrompt,
		commitMessage,
		feedbackPrompt)

	// Call API to fix the message
	fixedMessage, err := apiCaller(agentFilePath, conversationWithFeedback)
	if err != nil {
		return "", fmt.Errorf("API call failed during fix: %w", err)
	}

	return fixedMessage, nil
}

// buildValidationFeedback creates a feedback prompt from validation errors
func buildValidationFeedback(commitMessage string, errors []ValidationError) string {
	var feedback bytes.Buffer

	feedback.WriteString("⚠️  CONTRACT VALIDATION FAILED\n\n")
	feedback.WriteString("The commit message you generated has the following violations:\n\n")

	// Group by severity
	var errorList, warningList []ValidationError
	for _, err := range errors {
		if err.Severity == "error" {
			errorList = append(errorList, err)
		} else {
			warningList = append(warningList, err)
		}
	}

	if len(errorList) > 0 {
		feedback.WriteString(fmt.Sprintf("## ERRORS (%d)\n\n", len(errorList)))
		for _, err := range errorList {
			feedback.WriteString(fmt.Sprintf("- [%s] %s\n", err.Code, err.Message))
			if err.Line > 0 {
				feedback.WriteString(fmt.Sprintf("  Line %d\n", err.Line))
			}
		}
		feedback.WriteString("\n")
	}

	if len(warningList) > 0 {
		feedback.WriteString(fmt.Sprintf("## WARNINGS (%d)\n\n", len(warningList)))
		for _, err := range warningList {
			feedback.WriteString(fmt.Sprintf("- [%s] %s\n", err.Code, err.Message))
			if err.Line > 0 {
				feedback.WriteString(fmt.Sprintf("  Line %d\n", err.Line))
			}
		}
		feedback.WriteString("\n")
	}

	feedback.WriteString("---\n\n")
	feedback.WriteString("**YOUR TASK**: Fix ALL errors (warnings are optional).\n\n")
	feedback.WriteString("**CRITICAL RULES**:\n")
	feedback.WriteString("- Output ONLY the corrected commit message\n")
	feedback.WriteString("- Start directly with `# <title>`\n")
	feedback.WriteString("- NO preamble like \"Here's the fixed version:\"\n")
	feedback.WriteString("- NO explanations of what you changed\n")
	feedback.WriteString("- Just output the pure, corrected commit message\n\n")
	feedback.WriteString("Generate the corrected commit message now:\n")

	return feedback.String()
}
