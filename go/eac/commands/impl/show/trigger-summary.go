// Command: show trigger-summary
// Description: Generate release trigger summary
// Short: Generate release trigger summary
// Flag.module: type=string, usage=Module name (required)
// Flag.workflow: type=string, usage=Workflow filename (required)
// Flag.workflow-desc: type=string, usage=Workflow description
// Flag.version: type=string, usage=Version being released (optional)
// Flag.trigger-run-id: type=string, usage=Original trigger workflow run ID
// Flag.ci-run-id: type=string, usage=CI workflow run ID (required)
// Flag.branch: type=string, usage=Git branch name (required)
// Flag.commit: type=string, usage=Git commit SHA (required)
// Long: The show trigger-summary command generates a formatted summary when a release workflow is triggered.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent trigger summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted trigger summary with workflow and property tables
// Long: - Shows workflow name and description
// Long: - Shows version, run IDs, branch, and commit information

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
	registry.Register(ShowTriggerSummary)
}

// ShowTriggerSummary generates a release trigger summary
func ShowTriggerSummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "trigger-summary"

	module := ""
	workflow := ""
	workflowDesc := "Release module"
	version := ""
	triggerRunID := ""
	ciRunID := ""
	branch := ""
	commit := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--module="):
			module = strings.TrimPrefix(arg, "--module=")
		case strings.HasPrefix(arg, "--workflow="):
			workflow = strings.TrimPrefix(arg, "--workflow=")
		case strings.HasPrefix(arg, "--workflow-desc="):
			workflowDesc = strings.TrimPrefix(arg, "--workflow-desc=")
		case strings.HasPrefix(arg, "--version="):
			version = strings.TrimPrefix(arg, "--version=")
		case strings.HasPrefix(arg, "--trigger-run-id="):
			triggerRunID = strings.TrimPrefix(arg, "--trigger-run-id=")
		case strings.HasPrefix(arg, "--ci-run-id="):
			ciRunID = strings.TrimPrefix(arg, "--ci-run-id=")
		case strings.HasPrefix(arg, "--branch="):
			branch = strings.TrimPrefix(arg, "--branch=")
		case strings.HasPrefix(arg, "--commit="):
			commit = strings.TrimPrefix(arg, "--commit=")
		}
	}

	if module == "" || workflow == "" || ciRunID == "" || branch == "" || commit == "" {
		log.Errorf("Usage: show trigger-summary --module=<name> --workflow=<file> --ci-run-id=<id> --branch=<branch> --commit=<sha> [options]")
		return 1
	}

	return generateTriggerSummary(module, workflow, workflowDesc, version, triggerRunID, ciRunID, branch, commit)
}

func generateTriggerSummary(module, workflow, workflowDesc, version, triggerRunID, ciRunID, branch, commit string) int {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("## Release Triggered: %s\n\n", module))
	sb.WriteString("Triggered release workflow:\n\n")

	// Workflow table
	wfTable := render.NewTableBuilder().
		WithHeaders("Workflow", "Description")
	wfTable.AddRow(fmt.Sprintf("`%s`", workflow), workflowDesc)
	sb.WriteString(wfTable.Build())
	sb.WriteString("\n")

	// Properties table
	propTable := render.NewTableBuilder().
		WithHeaders("Property", "Value")

	if version != "" {
		propTable.AddRow("Version", fmt.Sprintf("`%s`", version))
	}

	if triggerRunID != "" {
		propTable.AddRow("Trigger Run ID", fmt.Sprintf("`%s`", triggerRunID))
	} else {
		propTable.AddRow("Trigger Run ID", "`none (standalone)`")
	}

	propTable.AddRow("CI Run ID", fmt.Sprintf("`%s`", ciRunID))
	propTable.AddRow("Branch", fmt.Sprintf("`%s`", branch))
	propTable.AddRow("Commit", fmt.Sprintf("`%s`", commit))

	sb.WriteString(propTable.Build())

	fmt.Print(sb.String())
	return 0
}
