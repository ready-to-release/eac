// Command: show scan-summary
// Description: Generate pretty scan summary for a module
// Short: Generate pretty scan summary for a module
// Flag.scans: type=string, usage=Comma-separated list of scans that were run (e.g., sbom,vuln,secrets)
// Flag.failed-scans: type=string, usage=Space-separated list of scans that failed
// Flag.artifact-name: type=string, usage=Name of the artifact containing scan results
// Flag.status: type=string, usage=Overall status (success or failure)
// Long: The show scan-summary command generates a formatted security scan summary with status per scan.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent, attractive scan summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: The command reads from the scan manifest at out/scan/<module>/scan.manifest.json.
// Long: Status is derived from the manifest - success if all scans passed, failure otherwise.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted scan summary with emojis and styling
// Long: - Table showing each scan type with its pass/fail status
// Long: - Artifact name for results download
package show

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/manifest"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowScanSummary)
}

// ShowScanSummary generates a pretty scan summary
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

	// Load scan manifest
	moduleScanDir := filepath.Join(workspaceRoot, "out", "scan", module)
	mf, err := manifest.LoadScanManifest(moduleScanDir)
	if err != nil {
		log.Errorf("No scan manifest found at %s: %v", moduleScanDir, err)
		return 1
	}

	return generateScanSummaryFromManifest(module, mf, artifactName)
}

// generateScanSummaryFromManifest generates a scan summary from the scan manifest
func generateScanSummaryFromManifest(module string, mf *manifest.ScanManifest, artifactName string) int {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("## 🔒 Security Scans: %s\n\n", module))

	// Determine status from manifest
	status := "success"
	if !mf.AllPassed() {
		status = "failure"
	}

	// Status message
	if status == "success" {
		sb.WriteString("✅ **All scans passed**\n\n")
	} else {
		sb.WriteString("⚠️ **Some scans failed**\n\n")
	}

	// Build table using render.TableBuilder
	tb := render.NewTableBuilder().
		WithHeaders("Scan", "Status", "Duration")

	for scanType, result := range mf.Scans {
		duration := ""
		if result.DurationSeconds > 0 {
			duration = fmt.Sprintf("%.1fs", result.DurationSeconds)
		}

		switch result.Status {
		case manifest.ScanStatusPassed:
			tb.AddRow(scanType, "✅ Passed", duration)
		case manifest.ScanStatusFailed:
			if result.Error != "" {
				tb.AddRow(scanType, fmt.Sprintf("❌ Failed: %s", truncate(result.Error, 40)), duration)
			} else {
				tb.AddRow(scanType, "❌ Failed", duration)
			}
		case manifest.ScanStatusSkipped:
			tb.AddRow(scanType, "⏭️ Skipped", duration)
		default:
			tb.AddRow(scanType, result.Status, duration)
		}
	}

	sb.WriteString(tb.Build())
	sb.WriteString("\n")

	// Summary stats
	sb.WriteString(fmt.Sprintf("**Summary**: %d passed, %d failed, %d skipped (total: %d)\n",
		mf.Summary.Passed, mf.Summary.Failed, mf.Summary.Skipped, mf.Summary.Total))

	// Artifact info
	sb.WriteString(fmt.Sprintf("**Artifact**: `%s`\n", artifactName))

	// Git commit if available
	if mf.GitCommit != "" {
		sb.WriteString(fmt.Sprintf("**Commit**: `%s`\n", truncate(mf.GitCommit, 7)))
	}

	fmt.Print(sb.String())
	return 0
}

// truncate shortens a string to maxLen characters with "..." if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
