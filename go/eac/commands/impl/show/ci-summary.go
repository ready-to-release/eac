// Command: show ci-summary
// Description: Generate CI workflow summary for a module
// Short: Generate CI workflow summary for a module
// Flag.build: type=string, usage=Build job result (success/failure/skipped)
// Flag.container: type=bool, default=false, usage=Whether this is a container module
// Flag.container-test: type=string, usage=Container test result (for container modules)
// Flag.container-test-enabled: type=bool, default=false, usage=Whether container test was enabled
// Flag.test-linux: type=string, usage=Linux test result (for binary modules)
// Flag.test-windows: type=string, usage=Windows test result (for binary modules)
// Flag.test-macos: type=string, usage=macOS test result (for binary modules)
// Flag.test-on-windows: type=bool, default=false, usage=Whether Windows tests were enabled
// Flag.test-on-macos: type=bool, default=false, usage=Whether macOS tests were enabled
// Flag.scan: type=string, usage=Security scan result
// Flag.scans-enabled: type=bool, default=false, usage=Whether scans were enabled
// Flag.evidence: type=string, usage=Evidence build result
// Flag.evidence-enabled: type=bool, default=false, usage=Whether evidence building was enabled
// Long: The show ci-summary command generates a formatted CI workflow summary with job results.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent CI summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted CI summary with job results table
// Long: - Shows build, test (Linux/Windows/macOS), container test, and scan results
// Long: - Supports both container and binary module types

package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ShowCISummary)
}

// ShowCISummary generates a CI workflow summary
func ShowCISummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "ci-summary"

	buildResult := ""
	isContainer := false
	containerTestResult := ""
	containerTestEnabled := false
	testLinuxResult := ""
	testWindowsResult := ""
	testMacOSResult := ""
	testOnWindows := false
	testOnMacOS := false
	scanResult := ""
	scansEnabled := false
	evidenceResult := ""
	evidenceEnabled := false

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--build="):
			buildResult = strings.TrimPrefix(arg, "--build=")
		case arg == "--container" || arg == "--container=true":
			isContainer = true
		case arg == "--container=false":
			isContainer = false
		case strings.HasPrefix(arg, "--container-test="):
			containerTestResult = strings.TrimPrefix(arg, "--container-test=")
		case arg == "--container-test-enabled" || arg == "--container-test-enabled=true":
			containerTestEnabled = true
		case strings.HasPrefix(arg, "--test-linux="):
			testLinuxResult = strings.TrimPrefix(arg, "--test-linux=")
		case strings.HasPrefix(arg, "--test-windows="):
			testWindowsResult = strings.TrimPrefix(arg, "--test-windows=")
		case strings.HasPrefix(arg, "--test-macos="):
			testMacOSResult = strings.TrimPrefix(arg, "--test-macos=")
		case arg == "--test-on-windows" || arg == "--test-on-windows=true":
			testOnWindows = true
		case arg == "--test-on-macos" || arg == "--test-on-macos=true":
			testOnMacOS = true
		case strings.HasPrefix(arg, "--scan="):
			scanResult = strings.TrimPrefix(arg, "--scan=")
		case arg == "--scans-enabled" || arg == "--scans-enabled=true":
			scansEnabled = true
		case strings.HasPrefix(arg, "--evidence="):
			evidenceResult = strings.TrimPrefix(arg, "--evidence=")
		case arg == "--evidence-enabled" || arg == "--evidence-enabled=true":
			evidenceEnabled = true
		}
	}

	return generateCISummary(buildResult, isContainer, containerTestResult, containerTestEnabled,
		testLinuxResult, testWindowsResult, testMacOSResult, testOnWindows, testOnMacOS, scanResult, scansEnabled,
		evidenceResult, evidenceEnabled)
}

func generateCISummary(buildResult string, isContainer bool, containerTestResult string, containerTestEnabled bool,
	testLinuxResult, testWindowsResult, testMacOSResult string, testOnWindows, testOnMacOS bool, scanResult string, scansEnabled bool,
	evidenceResult string, evidenceEnabled bool) int {

	var sb strings.Builder

	// Determine overall status
	overall := "success"
	if buildResult != "success" {
		overall = "failure"
	}

	// Check test results based on module type
	if isContainer {
		// Container modules use container-test
		if containerTestEnabled && containerTestResult != "success" && containerTestResult != "skipped" {
			overall = "failure"
		}
	} else {
		// Binary modules use test-linux (and optionally test-windows and test-macos)
		if testLinuxResult != "success" && testLinuxResult != "skipped" {
			overall = "failure"
		}
		if testOnWindows && testWindowsResult != "success" && testWindowsResult != "skipped" {
			overall = "failure"
		}
		if testOnMacOS && testMacOSResult != "success" && testMacOSResult != "skipped" {
			overall = "failure"
		}
	}
	// Scan failures don't fail overall (controlled by scan-fail-mode)

	// Header
	if overall == "success" {
		sb.WriteString("### :white_check_mark: All checks passed\n\n")
	} else {
		sb.WriteString("### :x: Some checks failed\n\n")
	}

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("Job", "Result")

	// Build row
	tb.AddRow("Build", formatJobResult(buildResult))

	// Container test (for container modules)
	if isContainer && containerTestEnabled {
		tb.AddRow("Container Test", formatJobResult(containerTestResult))
	}

	// Test Linux (for binary modules)
	if !isContainer {
		tb.AddRow("Test (Linux)", formatJobResult(testLinuxResult))
	}

	// Test Windows (if enabled, binary modules only)
	if !isContainer && testOnWindows {
		tb.AddRow("Test (Windows)", formatJobResult(testWindowsResult))
	}

	// Test macOS (if enabled, binary modules only)
	if !isContainer && testOnMacOS {
		tb.AddRow("Test (macOS)", formatJobResult(testMacOSResult))
	}

	// Scan (if enabled)
	if scansEnabled {
		tb.AddRow("Security Scan", formatScanResult(scanResult))
	}

	// Evidence (if enabled)
	if evidenceEnabled {
		tb.AddRow("Evidence", formatJobResult(evidenceResult))
	}

	sb.WriteString(tb.Build())

	fmt.Print(sb.String())
	return 0
}

// formatJobResult formats a job result with appropriate emoji
func formatJobResult(result string) string {
	switch result {
	case "success":
		return ":white_check_mark:"
	case "skipped":
		return ":grey_question: skipped"
	default:
		if result == "" {
			return ":grey_question: unknown"
		}
		return fmt.Sprintf(":x: %s", result)
	}
}

// formatScanResult formats a scan result (warnings instead of failures)
func formatScanResult(result string) string {
	switch result {
	case "success":
		return ":white_check_mark:"
	case "skipped":
		return ":grey_question: skipped"
	default:
		if result == "" {
			return ":grey_question: unknown"
		}
		return fmt.Sprintf(":warning: %s", result)
	}
}
