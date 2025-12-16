// Command: show changelog
// Short: Display changelog entries in human-readable format
// Long: Display changelog entries from CHANGELOG.md in a formatted markdown table.
// Long:
// Long: Shows all versions by default, or a specific version if provided.
// Long: Special keywords: "unreleased" for pending changes, "latest" for most recent release.
// Long:
// Long: Expected Output:
// Long: - Version header with number and date
// Long: - Markdown table with columns: Category, Description, Commit
// Long: - Grouped by change type (Added, Changed, Fixed, etc.)
// Long:
// Long: Example:
// Long:   show changelog ext-eac
// Long:   show changelog ext-eac 0.0.7
// Long:   show changelog ext-eac latest
// Long:   show changelog ext-eac unreleased
// Args: module [version]
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/changelog"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ShowChangelog)
}

func ShowChangelog() int {
	// Parse arguments - expect module after "show changelog"
	args := os.Args[1:]

	// Find where "show changelog" ends
	cmdIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "show" && args[i+1] == "changelog" {
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
		fmt.Fprintf(os.Stderr, "Usage: show changelog <module> [version]\n")
		return 1
	}

	module := positional[0]
	var version string
	if len(positional) > 1 {
		version = positional[1]
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get changelog data using core function
	report, err := reports.GetChangelog(workspaceRoot, module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	log := report.Changelog

	// Determine which versions to display
	var versionsToShow []*changelog.Version

	if version != "" {
		// Show specific version
		if version == "unreleased" {
			if log.Unreleased != nil && log.Unreleased.HasEntries() {
				versionsToShow = []*changelog.Version{log.Unreleased}
			} else {
				fmt.Printf("No unreleased changes for module: %s\n", module)
				return 0
			}
		} else if version == "latest" {
			latestVer := log.LatestVersion()
			if latestVer == nil {
				fmt.Printf("No released versions for module: %s\n", module)
				return 0
			}
			versionsToShow = []*changelog.Version{latestVer}
		} else {
			ver := log.GetVersion(version)
			if ver == nil {
				fmt.Fprintf(os.Stderr, "Error: version not found: %s\n", version)
				return 1
			}
			versionsToShow = []*changelog.Version{ver}
		}
	} else {
		// Show all versions
		if log.Unreleased != nil && log.Unreleased.HasEntries() {
			versionsToShow = append(versionsToShow, log.Unreleased)
		}
		for i := range log.Versions {
			versionsToShow = append(versionsToShow, &log.Versions[i])
		}
	}

	// Print header
	fmt.Printf("# Changelog: %s\n\n", module)

	// Display each version
	for _, ver := range versionsToShow {
		printChangelogVersion(ver)
	}

	return 0
}

// printChangelogVersion formats and prints a single changelog version
func printChangelogVersion(ver *changelog.Version) {
	// Print version header
	if ver.Number == "Unreleased" {
		fmt.Printf("## [Unreleased]\n\n")
	} else {
		dateStr := ""
		if !ver.Date.IsZero() {
			dateStr = fmt.Sprintf(" - %s", ver.Date.Format("2006-01-02"))
		}
		yankedStr := ""
		if ver.Yanked {
			yankedStr = " [YANKED]"
		}
		fmt.Printf("## [%s]%s%s\n\n", ver.Number, dateStr, yankedStr)
	}

	// Print entries by category
	printChangelogCategory("Added", ver.Added)
	printChangelogCategory("Changed", ver.Changed)
	printChangelogCategory("Deprecated", ver.Deprecated)
	printChangelogCategory("Removed", ver.Removed)
	printChangelogCategory("Fixed", ver.Fixed)
	printChangelogCategory("Security", ver.Security)

	fmt.Println()
}

// printChangelogCategory prints entries for a specific change category
func printChangelogCategory(category string, entries []changelog.Entry) {
	if len(entries) == 0 {
		return
	}

	fmt.Printf("### %s\n\n", category)

	tb := render.NewTableBuilder().
		WithHeaders("Description", "Type", "Scope", "Commit")

	for _, entry := range entries {
		commitType := entry.CommitType
		if commitType == "" {
			commitType = "-"
		}
		scope := entry.Scope
		if scope == "" {
			scope = "-"
		}
		commit := entry.CommitSHA
		if commit == "" {
			commit = "-"
		}

		description := entry.Description
		if entry.Breaking {
			description = "⚠️ BREAKING: " + description
		}

		tb.AddRow(description, commitType, scope, commit)
	}

	fmt.Println(tb.Build())
	fmt.Println()
}
