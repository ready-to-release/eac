package domain

import "fmt"

// FormatValidationErrors formats validation errors for display.
func FormatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var result string
	errorCount := 0
	warningCount := 0

	for _, verr := range errors {
		if !verr.IsWarning() {
			errorCount++
		} else {
			warningCount++
		}
	}

	if errorCount > 0 {
		result += fmt.Sprintf("❌ Found %d contract violation(s):\n\n", errorCount)
	}
	if warningCount > 0 {
		result += fmt.Sprintf("⚠️  Found %d warning(s):\n\n", warningCount)
	}

	for _, verr := range errors {
		icon := "❌"
		if verr.IsWarning() {
			icon = "⚠️ "
		}
		result += fmt.Sprintf("%s %s\n", icon, verr.Error())
	}

	return result
}

// CountCriticalErrors returns the number of errors that are not warnings.
func CountCriticalErrors(errors []ValidationError) int {
	count := 0
	for _, err := range errors {
		if !err.IsWarning() {
			count++
		}
	}
	return count
}
