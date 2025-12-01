// Command: release changelog
// Short: Generate or update changelog from commits
// Long: Analyzes commits since the last release tag and generates changelog entries.
// Long:
// Long: The command reads conventional commit messages (feat, fix, refactor, etc.)
// Long: and categorizes them into changelog sections (Added, Fixed, Changed, etc.).
// Long:
// Long: By default, shows a preview of the changelog entries without writing.
// Long: Use --write to update the changelog file.
// Long:
// Long: Examples:
// Long:   release changelog src-cli              # Preview changelog entries
// Long:   release changelog src-cli --write      # Update release/src-cli/CHANGELOG.md
// Long:   release changelog src-cli --from v1.0  # From specific tag
// Flag.write: type=bool, usage=Write changes to changelog file (default: preview only)
// Flag.from: type=string, usage=Start from specific tag/ref (default: latest release tag)
// Flag.to: type=string, usage=End at specific tag/ref (default: HEAD)
// Flag.version: type=string, usage=Override calculated version number
// Flag.date: type=string, usage=Override release date (YYYY-MM-DD format)
// Flag.breaking: type=bool, usage=Force major version bump
package release

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/changelog"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/definitions"
	"github.com/ready-to-release/eac/src/core/git"
)

func init() {
	registry.Register(ReleaseChangelog)
}

func ReleaseChangelog() int {
	// Parse flags
	module := ""
	write := false
	fromRef := ""
	toRef := "HEAD"
	overrideVersion := ""
	overrideDate := ""
	forceBreaking := false

	args := os.Args[3:] // Skip binary, "release", "changelog"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--write":
			write = true
		case arg == "--breaking":
			forceBreaking = true
		case strings.HasPrefix(arg, "--from="):
			fromRef = strings.TrimPrefix(arg, "--from=")
		case arg == "--from" && i+1 < len(args):
			i++
			fromRef = args[i]
		case strings.HasPrefix(arg, "--to="):
			toRef = strings.TrimPrefix(arg, "--to=")
		case arg == "--to" && i+1 < len(args):
			i++
			toRef = args[i]
		case strings.HasPrefix(arg, "--version="):
			overrideVersion = strings.TrimPrefix(arg, "--version=")
		case arg == "--version" && i+1 < len(args):
			i++
			overrideVersion = args[i]
		case strings.HasPrefix(arg, "--date="):
			overrideDate = strings.TrimPrefix(arg, "--date=")
		case arg == "--date" && i+1 < len(args):
			i++
			overrideDate = args[i]
		default:
			if !strings.HasPrefix(arg, "--") && module == "" {
				module = arg
			}
		}
	}

	if module == "" {
		fmt.Fprintln(os.Stderr, "Error: module moniker required")
		fmt.Fprintln(os.Stderr, "Usage: release changelog <module> [--write] [--from <ref>] [--version <ver>]")
		return 1
	}

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load modules: %v\n", err)
		return 1
	}

	moduleContract, exists := moduleRegistry.Get(module)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: module '%s' not found\n", module)
		return 1
	}

	// Open git repository
	repo, err := git.Open("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open git repository: %v\n", err)
		return 1
	}

	// Determine changelog path from module contract
	changelogPath := moduleContract.GetChangelogPath()

	// Parse existing changelog or create new
	var existingChangelog *changelog.Changelog
	if _, err := os.Stat(changelogPath); err == nil {
		existingChangelog, err = changelog.Parse(changelogPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse existing changelog: %v\n", err)
		}
	}

	if existingChangelog == nil {
		existingChangelog = &changelog.Changelog{
			Module:      module,
			Title:       "Changelog",
			VersionType: changelog.Semver,
		}
	}

	// Determine version type from module or existing changelog
	versionType := existingChangelog.VersionType
	if module == "docs" {
		versionType = changelog.Calver
	}

	// Find latest tag for this module
	tagPattern := module + "/*"
	if fromRef == "" {
		latestTag, err := repo.LatestTag(tagPattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get latest tag: %v\n", err)
		}
		fromRef = latestTag
	}

	// Get commits since last release
	commits, err := repo.CommitsBetween(fromRef, toRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get commits: %v\n", err)
		return 1
	}

	if len(commits) == 0 {
		fmt.Println("No commits found since last release.")
		return 0
	}

	// Filter commits by module file patterns
	modulePatterns := moduleContract.GetGlobPatterns()
	var filteredCommits []*changelog.Commit
	for _, c := range commits {
		parsed := changelog.ParseCommitMessage(c.Message)
		parsed.SHA = c.ShortSHA
		parsed.Date = c.Date
		parsed.Files = c.Files

		// Filter by module patterns if available
		if len(modulePatterns) > 0 {
			if commitMatchesModule(c.Files, modulePatterns) {
				filteredCommits = append(filteredCommits, parsed)
			}
		} else {
			filteredCommits = append(filteredCommits, parsed)
		}
	}

	if len(filteredCommits) == 0 {
		fmt.Printf("No commits affecting module '%s' found since last release.\n", module)
		return 0
	}

	// Calculate new version
	currentVersion := existingChangelog.LatestVersionNumber()
	if currentVersion == "" {
		currentVersion = "0.0.1"
	}

	// Collect entries for bump calculation
	var entries []changelog.Entry
	for _, c := range filteredCommits {
		if c.IsConventionalCommit() {
			entry := c.ToEntry()
			if forceBreaking {
				entry.Breaking = true
			}
			entries = append(entries, entry)
		}
	}

	// Get existing version numbers for calver collision detection
	var existingVersions []string
	for _, v := range existingChangelog.Versions {
		existingVersions = append(existingVersions, v.Number)
	}

	// Load definitions for versioning constraints
	workspaceRoot, _ := registry.GetWorkspaceRoot()
	defs, err := definitions.Load(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load definitions: %v\n", err)
		defs = definitions.Default()
	}

	// Determine max bump based on constraints
	maxBump := changelog.BumpMajor // Default: unrestricted
	if defs.IsPatchOnly() {
		maxBump = changelog.BumpPatch
		if forceBreaking {
			fmt.Fprintln(os.Stderr, "Warning: --breaking ignored due to patch-only constraint in .r2r/definitions.yml")
			forceBreaking = false
		}
	}

	// Calculate next version
	var newVersion string
	if overrideVersion != "" {
		newVersion = overrideVersion
	} else {
		newVersion, err = changelog.CalculateNextVersionConstrained(
			currentVersion,
			versionType,
			entries,
			time.Now(),
			existingVersions,
			maxBump,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to calculate version: %v\n", err)
			return 1
		}
	}

	// Determine release date
	releaseDate := time.Now()
	if overrideDate != "" {
		releaseDate, err = time.Parse("2006-01-02", overrideDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid date format '%s' (use YYYY-MM-DD)\n", overrideDate)
			return 1
		}
	}

	// Create version entry from commits
	newVersionEntry := changelog.CommitsToVersion(filteredCommits, newVersion, releaseDate)

	// Display preview
	fmt.Printf("Module: %s\n", module)
	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("New version: %s\n", newVersion)
	if defs.IsPatchOnly() && versionType == changelog.Semver {
		fmt.Printf("Version constraint: patch-only (from .r2r/definitions.yml)\n")
	}
	fmt.Printf("Commits analyzed: %d\n", len(commits))
	fmt.Printf("Module commits: %d\n", len(filteredCommits))
	fmt.Println()

	if !newVersionEntry.HasEntries() {
		fmt.Println("No conventional commits found to generate changelog entries.")
		fmt.Println("Commits must follow format: type(scope): description")
		return 0
	}

	// Preview the changelog entry
	fmt.Println("--- Changelog Preview ---")
	previewChangelog := &changelog.Changelog{
		Module:      module,
		VersionType: versionType,
		Versions:    []changelog.Version{newVersionEntry},
	}
	fmt.Println(previewChangelog.String())
	fmt.Println("-------------------------")

	if !write {
		fmt.Printf("\nTo write changes, run: release changelog %s --write\n", module)
		return 0
	}

	// Add new version to changelog
	existingChangelog.AddVersion(newVersionEntry)
	existingChangelog.Module = module
	existingChangelog.VersionType = versionType

	// Extract repo URL from git remote
	if existingChangelog.RepoURL == "" {
		if remoteURL, err := repo.RemoteURL("origin"); err == nil {
			existingChangelog.RepoURL = normalizeGitHubURL(remoteURL)
		}
	}

	// Write changelog
	if err := existingChangelog.Write(changelogPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write changelog: %v\n", err)
		return 1
	}

	fmt.Printf("\n✅ Updated %s with version %s\n", changelogPath, newVersion)
	return 0
}

// commitMatchesModule checks if any of the commit's files match module patterns
func commitMatchesModule(files []string, patterns []string) bool {
	for _, file := range files {
		for _, pattern := range patterns {
			if matchChangelogPattern(file, pattern) {
				return true
			}
		}
	}
	return false
}

// matchChangelogPattern provides simple glob matching for changelog filtering
func matchChangelogPattern(path, pattern string) bool {
	// Normalize separators
	path = strings.ReplaceAll(path, "\\", "/")
	pattern = strings.ReplaceAll(pattern, "\\", "/")

	// Handle ** pattern
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			if suffix != "" {
				remaining := strings.TrimPrefix(path, prefix)
				return strings.HasSuffix(remaining, suffix) || strings.Contains(remaining, suffix)
			}
			return true
		}
	}

	// Simple prefix match
	return strings.HasPrefix(path, pattern) || path == pattern
}

// normalizeGitHubURL converts git remote URLs to HTTPS format
func normalizeGitHubURL(remoteURL string) string {
	// Handle SSH format: git@github.com:org/repo.git
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		path := strings.TrimPrefix(remoteURL, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		return "https://github.com/" + path
	}

	// Handle HTTPS format: https://github.com/org/repo.git
	if strings.HasPrefix(remoteURL, "https://github.com/") {
		return strings.TrimSuffix(remoteURL, ".git")
	}

	return remoteURL
}
