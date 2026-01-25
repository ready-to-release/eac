// Command: show lint-summary
// Short: Generate pretty lint summary for a module
// Flag.artifact-name: type=string, usage=Name of the artifact containing lint results
// Long: The show lint-summary command generates a formatted lint summary with status.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent, attractive lint summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: The command reads from the lint manifest at out/lint/<module>/lint.manifest.json.
// Long: Status is derived from the manifest - success if no issues, failure otherwise.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted lint summary with emojis and styling
// Long: - Issue count and duration metrics
// Long: - List of providers that were run
// Long: - Artifact name for results download

package show

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowLintSummary)
}

// LintManifest represents the lint manifest structure.
// This is duplicated from the lint package to avoid import cycles.
type LintManifest struct {
	Moniker         string    `json:"moniker"`
	ModuleType      string    `json:"module_type"`
	GitCommit       string    `json:"git_commit"`
	RunTime         time.Time `json:"run_time"`
	DurationSeconds float64   `json:"duration_seconds"`
	Success         bool      `json:"success"`
	IssueCount      int       `json:"issue_count"`
	FixedCount      int       `json:"fixed_count,omitempty"`
	Providers       []string  `json:"providers,omitempty"`
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

	// Load lint manifest
	manifestPath := filepath.Join(workspaceRoot, "out", "lint", module, "lint.manifest.json")
	mf, err := loadLintManifest(manifestPath)
	if err != nil {
		log.Errorf("No lint manifest found at %s: %v", manifestPath, err)
		return 1
	}

	return generateLintSummaryFromManifest(module, mf, artifactName)
}

// loadLintManifest loads a lint manifest from the given path.
func loadLintManifest(path string) (*LintManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var mf LintManifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, err
	}

	return &mf, nil
}

// generateLintSummaryFromManifest generates a lint summary from the lint manifest.
func generateLintSummaryFromManifest(module string, mf *LintManifest, artifactName string) int {
	var sb strings.Builder

	// Header with emoji based on status
	if mf.Success {
		sb.WriteString(fmt.Sprintf("## ✅ Lint: %s\n\n", module))
		if mf.IssueCount == 0 {
			sb.WriteString("**No lint issues found**\n\n")
		} else if mf.FixedCount > 0 {
			sb.WriteString(fmt.Sprintf("**%d issues auto-fixed**\n\n", mf.FixedCount))
		}
	} else {
		sb.WriteString(fmt.Sprintf("## ❌ Lint: %s\n\n", module))
		sb.WriteString(fmt.Sprintf("**%d lint issues found**\n\n", mf.IssueCount))
	}

	// Build metrics table
	tb := render.NewTableBuilder().
		WithHeaders("Metric", "Value")

	tb.AddRow("Duration", fmt.Sprintf("%.2fs", mf.DurationSeconds))
	tb.AddRow("Issues Found", fmt.Sprintf("%d", mf.IssueCount))
	if mf.FixedCount > 0 {
		tb.AddRow("Issues Fixed", fmt.Sprintf("%d", mf.FixedCount))
	}
	if len(mf.Providers) > 0 {
		tb.AddRow("Providers", strings.Join(mf.Providers, ", "))
	}
	if mf.ModuleType != "" {
		tb.AddRow("Module Type", mf.ModuleType)
	}

	sb.WriteString(tb.Build())
	sb.WriteString("\n")

	// Artifact info
	sb.WriteString(fmt.Sprintf("**Artifact**: `%s`\n", artifactName))

	// Git commit if available
	if mf.GitCommit != "" {
		sb.WriteString(fmt.Sprintf("**Commit**: `%s`\n", truncateLint(mf.GitCommit, 7)))
	}

	fmt.Print(sb.String())
	return 0
}

// truncateLint shortens a string to maxLen characters with "..." if needed.
func truncateLint(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
