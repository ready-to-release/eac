package testing

import (
	"fmt"
	"strings"
)

// GenerateTraceabilityMatrix generates risk traceability matrix
func GenerateTraceabilityMatrix(tests []TestReference) string {
	var report strings.Builder

	report.WriteString("# Risk Traceability Matrix\n\n")
	report.WriteString("| Test | Risk Control | Type |\n")
	report.WriteString("|------|--------------|------|\n")

	for _, test := range tests {
		for _, rc := range test.RiskControls {
			testType := "Automated"
			if test.IsManual {
				testType = "Manual"
			}

			report.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
				test.TestName,
				rc,
				testType,
			))
		}
	}

	return report.String()
}

// GenerateGxPReport generates GxP implementation report
func GenerateGxPReport(tests []TestReference) string {
	var report strings.Builder

	report.WriteString("# GxP Implementation Report\n\n")

	// Filter GxP tests
	gxpTests := GetGxPTests(tests)

	// URS/FS section
	report.WriteString("## Requirements Specifications (URS/FS)\n\n")
	features := getUniqueFeatures(gxpTests)
	report.WriteString(fmt.Sprintf("Total GxP Features: %d\n\n", len(features)))

	// Test Summary
	report.WriteString("## Test Summary\n\n")
	report.WriteString(fmt.Sprintf("- Total Tests: %d\n", len(tests)))
	report.WriteString(fmt.Sprintf("- GxP Tests: %d\n", len(gxpTests)))
	report.WriteString(fmt.Sprintf("- Critical Aspects: %d\n", countCriticalAspects(gxpTests)))
	report.WriteString(fmt.Sprintf("- Manual Tests: %d\n", countManualTests(gxpTests)))

	// Risk Traceability Matrix
	report.WriteString("\n## Risk Traceability Matrix\n\n")
	report.WriteString(GenerateTraceabilityMatrix(gxpTests))

	return report.String()
}

// getUniqueFeatures extracts unique feature file paths
func getUniqueFeatures(tests []TestReference) []string {
	featuresMap := make(map[string]bool)
	for _, test := range tests {
		featuresMap[test.FilePath] = true
	}

	features := make([]string, 0, len(featuresMap))
	for feature := range featuresMap {
		features = append(features, feature)
	}
	return features
}

// countCriticalAspects counts tests with @critical-aspect
func countCriticalAspects(tests []TestReference) int {
	count := 0
	for _, test := range tests {
		if test.IsCriticalAspect {
			count++
		}
	}
	return count
}

// countManualTests counts tests with @Manual
func countManualTests(tests []TestReference) int {
	count := 0
	for _, test := range tests {
		if test.IsManual {
			count++
		}
	}
	return count
}
