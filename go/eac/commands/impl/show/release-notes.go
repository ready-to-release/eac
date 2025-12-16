// Command: show release-notes
// Short: Display release notes in human-readable format
// Long: Display release notes from RELEASE-NOTES.md in formatted markdown.
// Long:
// Long: Shows the latest version by default, or a specific version if provided.
// Long: Special keyword: "latest" for most recent release (same as default behavior).
// Long: Displays all sections for the version (Conclusion on Fitness, Impact on Business Process, etc.)
// Long:
// Long: Expected Output:
// Long: - Version header (## [version] - date) matching changelog format
// Long: - All sections with headers (###) and content formatted as markdown
// Long:
// Long: Example:
// Long:   show release-notes ext-eac
// Long:   show release-notes ext-eac latest
// Long:   show release-notes ext-eac 0.0.7
// Args: module [version]
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/releasenotes"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowReleaseNotes)
}

func ShowReleaseNotes() int {
	// Parse arguments - expect module after "show release-notes"
	args := os.Args[1:]

	// Find where "show release-notes" ends
	cmdIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "show" && args[i+1] == "release-notes" {
			cmdIdx = i + 2
			break
		}
	}

	// Collect positional arguments (non-flag arguments after command)
	var positional []string
	if cmdIdx != -1 && cmdIdx < len(args) {
		for i := cmdIdx; i < len(args); i++ {
			if args[i][0] != '-' {
				positional = append(positional, args[i])
			}
		}
	}

	if len(positional) < 1 {
		fmt.Fprintf(os.Stderr, "Error: module argument required\n")
		fmt.Fprintf(os.Stderr, "Usage: show release-notes <module> [version]\n")
		return 1
	}

	module := positional[0]
	var versionNum string
	if len(positional) > 1 {
		versionNum = positional[1]
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get release notes data using core function
	report, err := reports.GetReleaseNotes(workspaceRoot, module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	rn := report.ReleaseNotes

	// Determine which version to show
	var ver *releasenotes.ReleaseNotesVersion

	if versionNum != "" {
		// Resolve version using common helper
		versionInfo, err := reports.ResolveVersion(workspaceRoot, module, versionNum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}

		// Get specific version
		v, err := rn.GetVersion(versionInfo.VersionNumber)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		ver = v
	} else {
		// Show latest version by default
		if len(rn.Versions) == 0 {
			fmt.Printf("No release notes found for module: %s\n", module)
			return 0
		}
		ver = &rn.Versions[0]
	}

	// Print release notes
	printReleaseNotesVersion(ver)

	return 0
}

// printReleaseNotesVersion formats and prints a single release notes version
func printReleaseNotesVersion(ver *releasenotes.ReleaseNotesVersion) {
	// Print header
	dateStr := ""
	if !ver.Date.IsZero() {
		dateStr = fmt.Sprintf(" - %s", ver.Date.Format("2006-01-02"))
	}
	fmt.Printf("## [%s]%s\n\n", ver.Number, dateStr)

	// Print all sections
	for _, section := range ver.Sections {
		fmt.Printf("### %s\n\n", section.Header)
		fmt.Printf("%s\n\n", section.Content)
	}
}
