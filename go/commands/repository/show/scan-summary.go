package show

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/repository"
)

type showScanSummaryCommand struct{}

var _ core.SimpleCommandPort = (*showScanSummaryCommand)(nil)

func (c *showScanSummaryCommand) Name() string { return "show scan-summary" }

func (c *showScanSummaryCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-scan-summary",
		Short:         "Generate pretty scan summary for a module",
		Long: "The show scan-summary command generates a formatted security scan summary with status per scan.\nThis command is designed to be used in GitHub Actions workflows to create consistent, attractive scan summaries.\nThe output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.\n\nThe command reads from UoW manifests at out/scan/<module>/*/uow.manifest.json.\nStatus is derived from exit codes - success if all zero, failure otherwise.",
		Notes: "Expected Output:\n- Markdown-formatted scan summary with emojis and styling\n- Table showing each scan type with its pass/fail status\n- Artifact name for results download",
		Flags: []core.FlagSpec{
			{Name: "artifact-name", Type: "string", Usage: "Name of the artifact containing scan results"},
		},
	}
}

func (c *showScanSummaryCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowScanSummary()
}

// ShowScanSummary generates a pretty scan summary.
func ShowScanSummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "scan-summary"

	if len(args) == 0 {
		log.Errorf("Usage: show scan-summary <module> [--artifact-name=<name>]")
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
		artifactName = fmt.Sprintf("scan-results-%s", module)
	}

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Failed to find repository root: %v", err)
		return 1
	}

	// Load scan UoW manifests
	reader := coreoutput.NewReader(workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionScan, module)
	if err != nil || len(manifests) == 0 {
		// No manifests is not a fatal error — the scan may not have produced
		// UoW manifests (e.g., container modules). Output an informational
		// summary and exit successfully to avoid failing the CI job.
		fmt.Printf("## 🔒 Security Scans: %s\n\n", module)
		fmt.Println("No scan manifests found for this module.")
		fmt.Printf("**Artifact**: `%s`\n", artifactName)
		return 0
	}

	return generateScanSummaryFromUoWs(module, manifests, artifactName)
}

// generateScanSummaryFromUoWs generates a scan summary from UoW manifests.
func generateScanSummaryFromUoWs(module string, manifests []*coreoutput.UoWManifest, artifactName string) int {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("## 🔒 Security Scans: %s\n\n", module))

	// Determine overall status
	allPassed := true
	passed := 0
	failed := 0
	skipped := 0

	for _, m := range manifests {
		if m.ExitCode > 0 {
			allPassed = false
			failed++
		} else if m.ExitCode < 0 {
			skipped++
		} else {
			passed++
		}
	}

	// Status message
	if allPassed {
		sb.WriteString("✅ **All scans passed**\n\n")
	} else {
		sb.WriteString("⚠️ **Some scans failed**\n\n")
	}

	// Build table using render.TableBuilder
	tb := render.NewTableBuilder().
		WithHeaders("Scan", "Status", "Duration")

	for _, m := range manifests {
		scanName := m.Tool
		if m.Component != "" && m.Component != m.Tool {
			scanName = m.Component + "/" + m.Tool
		}

		duration := ""
		if m.Duration.Seconds() > 0 {
			duration = fmt.Sprintf("%.1fs", m.Duration.Seconds())
		}

		switch {
		case m.ExitCode == 0:
			tb.AddRow(scanName, "✅ Passed", duration)
		case m.ExitCode < 0:
			tb.AddRow(scanName, "⏭️ Cached", duration)
		default:
			tb.AddRow(scanName, "❌ Failed", duration)
		}
	}

	sb.WriteString(tb.Build())
	sb.WriteString("\n")

	// Summary stats
	total := passed + failed + skipped
	sb.WriteString(fmt.Sprintf("**Summary**: %d passed, %d failed, %d skipped (total: %d)\n",
		passed, failed, skipped, total))

	// Artifact info
	sb.WriteString(fmt.Sprintf("**Artifact**: `%s`\n", artifactName))

	fmt.Print(sb.String())
	return 0
}

// truncate shortens a string to maxLen characters with "..." if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
