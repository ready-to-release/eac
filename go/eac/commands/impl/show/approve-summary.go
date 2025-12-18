// Command: show approve-summary
// Description: Generate release approval summary
// Short: Generate release approval summary
// Long: The show approve-summary command generates a formatted release approval summary.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent approval summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted approval summary with check status table
// Long: - Shows version, tag, commit, changelog, existing release, and CI check status
// Long: - On failure, can output diagnostic link via pipeline ci summary-link
// Long:
// Flag.module: type=string, usage=Module name (required)
// Flag.version: type=string, usage=Release version (required)
// Flag.tag: type=string, usage=Full tag name (required)
// Flag.commit: type=string, usage=Commit SHA (required)
// Flag.version-type: type=string, default=semver, usage=Version type (semver or calver)
// Flag.ci-skipped: type=bool, default=false, usage=Whether CI check was skipped
// Flag.status: type=string, default=success, usage=Overall status (success or failure)
// Flag.run-id: type=string, usage=Run ID for diagnostic links on failure
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ShowApproveSummary)
}

// ShowApproveSummary generates a release approval summary
func ShowApproveSummary() int {
	args := os.Args[3:] // Skip program name, "show", and "approve-summary"

	module := ""
	version := ""
	tag := ""
	commit := ""
	versionType := "semver"
	ciSkipped := false
	status := "success"
	runID := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--module="):
			module = strings.TrimPrefix(arg, "--module=")
		case strings.HasPrefix(arg, "--version="):
			version = strings.TrimPrefix(arg, "--version=")
		case strings.HasPrefix(arg, "--tag="):
			tag = strings.TrimPrefix(arg, "--tag=")
		case strings.HasPrefix(arg, "--commit="):
			commit = strings.TrimPrefix(arg, "--commit=")
		case strings.HasPrefix(arg, "--version-type="):
			versionType = strings.TrimPrefix(arg, "--version-type=")
		case arg == "--ci-skipped" || arg == "--ci-skipped=true":
			ciSkipped = true
		case arg == "--ci-skipped=false":
			ciSkipped = false
		case strings.HasPrefix(arg, "--status="):
			status = strings.TrimPrefix(arg, "--status=")
		case strings.HasPrefix(arg, "--run-id="):
			runID = strings.TrimPrefix(arg, "--run-id=")
		}
	}

	if module == "" || version == "" || tag == "" || commit == "" {
		log.Errorf("Usage: show approve-summary --module=<name> --version=<ver> --tag=<tag> --commit=<sha> [options]")
		return 1
	}

	return generateApproveSummary(module, version, tag, commit, versionType, ciSkipped, status, runID)
}

func generateApproveSummary(module, version, tag, commit, versionType string, ciSkipped bool, status, runID string) int {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("## Release Approval: %s\n\n", module))

	// Shorten commit for display
	shortCommit := commit
	if len(commit) > 7 {
		shortCommit = commit[:7]
	}

	// Build table using render.TableBuilder
	tb := render.NewTableBuilder().
		WithHeaders("Check", "Status")

	// Version info rows
	tb.AddRow("Version", fmt.Sprintf("`%s` (%s)", version, versionType))
	tb.AddRow("Tag", fmt.Sprintf("`%s`", tag))
	tb.AddRow("Commit", fmt.Sprintf("`%s`", shortCommit))

	// Show success status only if job succeeded
	if status == "success" {
		// Changelog check
		if versionType == "semver" {
			tb.AddRow("Changelog", ":white_check_mark: Updated")
		} else {
			tb.AddRow("Changelog", ":white_check_mark: N/A (calver)")
		}

		// Existing release check
		tb.AddRow("Existing Release", ":white_check_mark: None")

		// CI check
		if ciSkipped {
			tb.AddRow("CI Check", ":warning: Skipped")
		} else {
			tb.AddRow("CI Check", ":white_check_mark: Passed (strict)")
		}
	}

	sb.WriteString(tb.Build())

	// On failure, add note about diagnostic link (caller adds via pipeline ci summary-link)
	if status == "failure" && runID != "" {
		sb.WriteString("\n")
		sb.WriteString("*See diagnostic link below for CI details.*\n")
	}

	fmt.Print(sb.String())
	return 0
}
