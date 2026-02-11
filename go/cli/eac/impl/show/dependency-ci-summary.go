package show

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
)

type showDependencyCISummaryCommand struct{}

var _ core.SimpleCommandPort = (*showDependencyCISummaryCommand)(nil)

func (c *showDependencyCISummaryCommand) Name() string { return "show dependency-ci-summary" }

func (c *showDependencyCISummaryCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-dependency-ci-summary",
		Short:         "Generate dependency CI check summary",
		Long:          "The show dependency-ci-summary command generates a formatted summary of dependency CI check results.\nThis command is designed to be used in GitHub Actions workflows to create consistent CI check summaries.\nThe output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.\n\nExpected Output:\n- Markdown-formatted dependency CI summary with metrics table\n- Shows passed and skipped counts on success\n- Shows failure message on failure",
		Flags: []core.FlagSpec{
			{Name: "module", Type: "string", Usage: "Module name (required)"},
			{Name: "passed", Type: "int", DefaultValue: "0", Usage: "Number of dependencies that passed CI"},
			{Name: "skipped", Type: "int", DefaultValue: "0", Usage: "Number of dependencies skipped (no CI workflow)"},
			{Name: "status", Type: "string", DefaultValue: "success", Usage: "Overall status (success or failure)"},
		},
	}
}

func (c *showDependencyCISummaryCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowDependencyCISummary()
}

// ShowDependencyCISummary generates a dependency CI check summary.
func ShowDependencyCISummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "dependency-ci-summary"

	module := ""
	passed := 0
	skipped := 0
	status := "success"

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--module="):
			module = strings.TrimPrefix(arg, "--module=")
		case strings.HasPrefix(arg, "--passed="):
			if p, err := strconv.Atoi(strings.TrimPrefix(arg, "--passed=")); err == nil {
				passed = p
			}
		case strings.HasPrefix(arg, "--skipped="):
			if s, err := strconv.Atoi(strings.TrimPrefix(arg, "--skipped=")); err == nil {
				skipped = s
			}
		case strings.HasPrefix(arg, "--status="):
			status = strings.TrimPrefix(arg, "--status=")
		}
	}

	if module == "" {
		log.Errorf("Usage: show dependency-ci-summary --module=<name> [--passed=<n>] [--skipped=<n>] [--status=<status>]")
		return 1
	}

	return generateDependencyCISummary(module, passed, skipped, status)
}

func generateDependencyCISummary(module string, passed, skipped int, status string) int {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("## Dependency CI Check: %s\n\n", module))

	if status == "success" {
		sb.WriteString(":white_check_mark: All dependency CI checks passed\n\n")

		// Build metrics table
		tb := render.NewTableBuilder().
			WithHeaders("Metric", "Value")

		tb.AddRow("Dependencies Checked", fmt.Sprintf("%d", passed))
		if skipped > 0 {
			tb.AddRow("Skipped (no CI)", fmt.Sprintf("%d", skipped))
		}

		sb.WriteString(tb.Build())
	} else {
		sb.WriteString(":x: Dependency CI check failed\n\n")
		sb.WriteString("One or more dependencies have failing CI. See logs for details.\n")
	}

	fmt.Print(sb.String())
	return 0
}
