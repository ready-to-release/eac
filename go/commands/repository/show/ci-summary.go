package show

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
)

type showCISummaryCommand struct{}

var _ core.SimpleCommandPort = (*showCISummaryCommand)(nil)

func (c *showCISummaryCommand) Name() string { return "show ci-summary" }

func (c *showCISummaryCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-ci-summary",
		Short:         "Generate CI workflow summary for a module",
		Long:          "The show ci-summary command generates a formatted CI workflow summary with job results.\nThis command is designed to be used in GitHub Actions workflows to create consistent CI summaries.\nThe output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.\n\nExpected Output:\n- Markdown-formatted CI summary with job results table\n- Shows build, test (Linux/Windows/macOS), container test, and scan results\n- Supports both container and binary module types",
		Flags: []core.FlagSpec{
			{Name: "build", Type: "string", Usage: "Build job result (success/failure/skipped)"},
			{Name: "container", Type: "bool", DefaultValue: "false", Usage: "Whether this is a container module"},
			{Name: "container-test", Type: "string", Usage: "Container test result (for container modules)"},
			{Name: "container-test-enabled", Type: "bool", DefaultValue: "false", Usage: "Whether container test was enabled"},
			{Name: "test-linux", Type: "string", Usage: "Linux test result (for binary modules)"},
			{Name: "test-windows", Type: "string", Usage: "Windows test result (for binary modules)"},
			{Name: "test-macos", Type: "string", Usage: "macOS test result (for binary modules)"},
			{Name: "test-on-windows", Type: "bool", DefaultValue: "false", Usage: "Whether Windows tests were enabled"},
			{Name: "test-on-macos", Type: "bool", DefaultValue: "false", Usage: "Whether macOS tests were enabled"},
			{Name: "scan", Type: "string", Usage: "Security scan result"},
			{Name: "scans-enabled", Type: "bool", DefaultValue: "false", Usage: "Whether scans were enabled"},
			{Name: "evidence", Type: "string", Usage: "Evidence build result"},
			{Name: "evidence-enabled", Type: "bool", DefaultValue: "false", Usage: "Whether evidence building was enabled"},
		},
	}
}

func (c *showCISummaryCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowCISummary()
}

// ShowCISummary generates a CI workflow summary.
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
	evidenceResult string, evidenceEnabled bool,
) int {
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

// formatJobResult formats a job result with appropriate emoji.
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

// formatScanResult formats a scan result (warnings instead of failures).
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
