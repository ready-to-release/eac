// Command: show lint-summary
// Short: Generate pretty lint summary for a module
// Flag.artifact-name: type=string, usage=Name of the artifact containing lint results
// Long: The show lint-summary command generates a formatted lint summary with status.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent, attractive lint summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: The command reads from UoW manifests at out/lint/<module>/*/uow.manifest.json.
// Long: Status is derived from exit codes - success if all zero, failure otherwise.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted lint summary with emojis and styling
// Long: - Issue count and duration metrics
// Long: - List of providers that were run
// Long: - Artifact name for results download

package show

import (
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/clibase/registry"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(ShowLintSummary)
}

// ShowLintSummary generates a pretty lint summary.
func ShowLintSummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "lint-summary"

	if len(args) == 0 {
		log.Errorf("Usage: show lint-summary <module> [--artifact-name=<name>]")
		return 1
	}

	module := args[0]
	artifactName := ""

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--artifact-name=") {
			artifactName = strings.TrimPrefix(arg, "--artifact-name=")
		}
	}

	// Default artifact name
	if artifactName == "" {
		artifactName = fmt.Sprintf("lint-results-%s", module)
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Failed to find repository root: %v", err)
		return 1
	}

	// Load lint UoW manifests
	reader := coreoutput.NewReader(workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionLint, module)
	if err != nil || len(manifests) == 0 {
		log.Errorf("No lint manifests found for module %s (run lint first)", module)
		return 1
	}

	// Aggregate data from UoW manifests
	success := true
	var totalDuration float64
	failedCount := 0
	providers := []string{}
	seen := map[string]bool{}

	for _, m := range manifests {
		totalDuration += m.Duration.Seconds()
		if m.ExitCode > 0 {
			success = false
			failedCount++
		}
		if m.Tool != "" && !seen[m.Tool] {
			providers = append(providers, m.Tool)
			seen[m.Tool] = true
		}
	}

	return generateLintSummary(module, success, totalDuration, failedCount, providers, artifactName)
}

// generateLintSummary generates a lint summary from aggregated UoW data.
func generateLintSummary(module string, success bool, durationSec float64, failedCount int, providers []string, artifactName string) int {
	var sb strings.Builder

	// Header with emoji based on status
	if success {
		sb.WriteString(fmt.Sprintf("## ✅ Lint: %s\n\n", module))
		sb.WriteString("**No lint issues found**\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("## ❌ Lint: %s\n\n", module))
		sb.WriteString(fmt.Sprintf("**%d lint provider(s) reported issues**\n\n", failedCount))
	}

	// Build metrics table
	tb := render.NewTableBuilder().
		WithHeaders("Metric", "Value")

	tb.AddRow("Duration", fmt.Sprintf("%.2fs", durationSec))
	if len(providers) > 0 {
		tb.AddRow("Providers", strings.Join(providers, ", "))
	}

	sb.WriteString(tb.Build())
	sb.WriteString("\n")

	// Artifact info
	sb.WriteString(fmt.Sprintf("**Artifact**: `%s`\n", artifactName))

	fmt.Print(sb.String())
	return 0
}
