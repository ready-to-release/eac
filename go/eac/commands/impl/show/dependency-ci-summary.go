// Command: show dependency-ci-summary
// Short: Generate dependency CI check summary
// Flag.module: type=string, usage=Module name (required)
// Flag.passed: type=int, default=0, usage=Number of dependencies that passed CI
// Flag.skipped: type=int, default=0, usage=Number of dependencies skipped (no CI workflow)
// Flag.status: type=string, default=success, usage=Overall status (success or failure)
// Long: The show dependency-ci-summary command generates a formatted summary of dependency CI check results.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent CI check summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted dependency CI summary with metrics table
// Long: - Shows passed and skipped counts on success
// Long: - Shows failure message on failure

package show

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ShowDependencyCISummary)
}

// ShowDependencyCISummary generates a dependency CI check summary
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
			passed, _ = strconv.Atoi(strings.TrimPrefix(arg, "--passed="))
		case strings.HasPrefix(arg, "--skipped="):
			skipped, _ = strconv.Atoi(strings.TrimPrefix(arg, "--skipped="))
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
