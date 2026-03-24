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

type showApproveSummaryCommand struct{}

var _ core.SimpleCommandPort = (*showApproveSummaryCommand)(nil)

func (c *showApproveSummaryCommand) Name() string { return "show approve-summary" }

func (c *showApproveSummaryCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-approve-summary",
		Short:         "Generate release approval summary",
		Long: "The show approve-summary command generates a formatted release approval summary.\nThis command is designed to be used in GitHub Actions workflows to create consistent approval summaries.\nThe output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.",
		Notes: "Expected Output:\n- Markdown-formatted approval summary with check status table\n- Shows version, tag, commit, changelog, existing release, and CI check status\n- On failure, can output diagnostic link via pipeline ci summary-link",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", Usage: "Module name (required)"},
			{Name: "version", Type: "string", Usage: "Release version (required)"},
			{Name: "tag", Type: "string", Usage: "Full tag name (required)"},
			{Name: "commit", Type: "string", Usage: "Commit SHA (required)"},
			{Name: "version-type", Type: "string", DefaultValue: "semver", Usage: "Version type (semver or calver)"},
			{Name: "ci-skipped", Type: "bool", DefaultValue: "false", Usage: "Whether CI check was skipped"},
			{Name: "status", Type: "string", DefaultValue: "success", Usage: "Overall status (success or failure)"},
			{Name: "run-id", Type: "string", Usage: "Run ID for diagnostic links on failure"},
		},
	}
}

func (c *showApproveSummaryCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowApproveSummary()
}

// ShowApproveSummary generates a release approval summary.
func ShowApproveSummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

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
