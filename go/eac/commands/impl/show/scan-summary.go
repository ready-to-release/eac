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

	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
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
		log.Errorf("Usage: show scan-summary <module> --scans=<list> [--failed-scans=<list>] [--artifact-name=<name>] [--status=<status>]")
		return 1
	}

	module := args[0]
	scans := ""
	failedScans := ""
	artifactName := ""
	status := "success"

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--scans="):
			scans = strings.TrimPrefix(arg, "--scans=")
		case strings.HasPrefix(arg, "--failed-scans="):
			failedScans = strings.TrimPrefix(arg, "--failed-scans=")
		case strings.HasPrefix(arg, "--artifact-name="):
			artifactName = strings.TrimPrefix(arg, "--artifact-name=")
		case strings.HasPrefix(arg, "--status="):
			status = strings.TrimPrefix(arg, "--status=")
		}
	}

	if scans == "" {
		log.Errorf("--scans is required")
		return 1
	}

	// Default artifact name
	if artifactName == "" {
		artifactName = fmt.Sprintf("scan-results-%s", module)
	}

	return generateScanSummary(module, scans, failedScans, artifactName, status)
}

func generateScanSummary(module, scans, failedScans, artifactName, status string) int {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("## 🔒 Security Scans: %s\n\n", module))

	// Status message
	if status == "success" {
		sb.WriteString("✅ **All scans passed**\n\n")
	} else {
		sb.WriteString("⚠️ **Some scans failed**\n\n")
	}

	// Parse scan lists
	scanList := strings.Split(scans, ",")
	failedSet := make(map[string]bool)
	for _, s := range strings.Fields(failedScans) {
		failedSet[strings.TrimSpace(s)] = true
	}

	// Build table using render.TableBuilder
	tb := render.NewTableBuilder().
		WithHeaders("Scan", "Status")

	for _, scan := range scanList {
		scan = strings.TrimSpace(scan)
		if scan == "" {
			continue
		}

		if failedSet[scan] {
			tb.AddRow(scan, "❌ Failed")
		} else {
			// Check for output files
			outputDir := filepath.Join("out", "security", module, scan)
			fileCount := countFiles(outputDir)
			if fileCount > 0 {
				tb.AddRow(scan, fmt.Sprintf("✅ Complete (%d files)", fileCount))
			} else {
				tb.AddRow(scan, "⚠️ No output")
			}
		}
	}

	sb.WriteString(tb.Build())
	sb.WriteString("\n")

	// Artifact info
	sb.WriteString(fmt.Sprintf("**Artifact**: `%s`\n", artifactName))

	fmt.Print(sb.String())
	return 0
}

// countFiles counts files in a directory
func countFiles(dir string) int {
	count := 0
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
