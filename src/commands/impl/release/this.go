// Command: release this
// Short: Finalize changelog and prepare module for release
// Long: Updates the changelog with all commits since the last release and prepares
// Long: the module for release.
// Long:
// Long: This command:
// Long:   1. Analyzes commits since the last release tag
// Long:   2. Generates changelog entries from conventional commits
// Long:   3. Merges any manual entries from [Unreleased] section
// Long:   4. Calculates the next version (respecting constraints from .r2r/definitions.yml)
// Long:   5. Adds a new version section to the changelog
// Long:   6. Clears the [Unreleased] section
// Long:   7. Writes the updated changelog
// Long:
// Long: Manual entries in [Unreleased] are preserved and merged with auto-generated
// Long: entries. This allows team members to add changelog entries that aren't tied
// Long: to specific commits.
// Long:
// Long: After running this command, commit the changelog and create a PR. Once merged,
// Long: the release-auto workflow will detect the new version and create the git tag,
// Long: which triggers the module's release workflow.
// Long:
// Long: Examples:
// Long:   release this src-cli              # Update changelog
// Long:   release this src-cli --dry-run    # Preview without writing
// Long:   release this src-cli --json       # Output result as JSON
// Flag.dry-run: type=bool, usage=Preview changes without writing to changelog
// Flag.json: type=bool, usage=Output result in JSON format
// Flag.date: type=string, usage=Override release date (YYYY-MM-DD format)
// Args: modules
package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/changelog"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/definitions"
	"github.com/ready-to-release/eac/src/core/git"
)

func init() {
	registry.Register(ReleaseThis)
}

// ReleaseResult contains the result of a release operation
type ReleaseResult struct {
	Module          string `json:"module"`
	Success         bool   `json:"success"`
	PreviousVersion string `json:"previous_version"`
	NewVersion      string `json:"new_version"`
	Tag             string `json:"tag"`
	VersionType     string `json:"version_type"`
	Constraint      string `json:"constraint"`
	ChangelogPath   string `json:"changelog_path"`
	EntriesAdded    int    `json:"entries_added"`
	DryRun          bool   `json:"dry_run,omitempty"`
	Error           string `json:"error,omitempty"`
}

func ReleaseThis() int {
	// Parse flags
	module := ""
	dryRun := false
	asJSON := false
	overrideDate := ""

	args := os.Args[3:] // Skip binary, "release", "this"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--dry-run":
			dryRun = true
		case arg == "--json":
			asJSON = true
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
		log.Error("module moniker required")
		log.Info("Usage: release this <module> [--dry-run] [--json]")
		return 1
	}

	result := performRelease(module, dryRun, overrideDate)

	if asJSON {
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Errorf("failed to marshal JSON: %v", err)
			return 1
		}
		log.Info(string(jsonBytes))
	} else {
		if result.Error != "" {
			log.Errorf("%s", result.Error)
			return 1
		}

		if dryRun {
			log.Info("=== DRY RUN - No changes written ===")
			log.Info("")
		}

		log.Infof("Module: %s", result.Module)
		log.Infof("Previous version: %s", result.PreviousVersion)
		log.Infof("New version: %s", result.NewVersion)
		log.Infof("Tag: %s", result.Tag)
		log.Infof("Entries added: %d", result.EntriesAdded)

		if !dryRun {
			log.Info("")
			log.Infof("Updated: %s", result.ChangelogPath)
			log.Info("")
			log.Info("Next steps:")
			log.Infof("  1. git add %s", result.ChangelogPath)
			log.Infof("  2. git commit -m \"release(%s): %s\"", result.Module, result.NewVersion)
			log.Info("  3. Create PR and merge to main")
			log.Info("  4. release-auto workflow will create tag and trigger release")
		}
	}

	if !result.Success {
		return 1
	}
	return 0
}

func performRelease(module string, dryRun bool, overrideDate string) ReleaseResult {
	result := ReleaseResult{
		Module: module,
		DryRun: dryRun,
	}

	// Load workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		result.Error = fmt.Sprintf("failed to get workspace root: %v", err)
		return result
	}

	// Load definitions for versioning constraints
	defs, err := definitions.Load(workspaceRoot)
	if err != nil {
		defs = definitions.Default()
	}

	// Set constraint info
	if defs.IsPatchOnly() {
		result.Constraint = "patch-only"
	} else if defs.IsCalverOnly() {
		result.Constraint = "calver-only"
	} else {
		result.Constraint = "unrestricted"
	}

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace("")
	if err != nil {
		result.Error = fmt.Sprintf("failed to load modules: %v", err)
		return result
	}

	moduleContract, exists := moduleRegistry.Get(module)
	if !exists {
		result.Error = fmt.Sprintf("module '%s' not found", module)
		return result
	}

	// Open git repository
	repo, err := git.Open("")
	if err != nil {
		result.Error = fmt.Sprintf("failed to open git repository: %v", err)
		return result
	}

	// Determine changelog path from module contract
	changelogPath := moduleContract.GetChangelogPath()
	fullChangelogPath := filepath.Join(workspaceRoot, changelogPath)
	result.ChangelogPath = changelogPath

	// Parse existing changelog or create new
	var existingChangelog *changelog.Changelog
	if _, err := os.Stat(fullChangelogPath); err == nil {
		existingChangelog, err = changelog.Parse(fullChangelogPath)
		if err != nil {
			result.Error = fmt.Sprintf("failed to parse existing changelog: %v", err)
			return result
		}
	} else {
		existingChangelog = &changelog.Changelog{
			Module:      module,
			Title:       "Changelog",
			VersionType: changelog.Semver,
		}
	}

	// Determine version type
	versionType := existingChangelog.VersionType
	if module == "docs" {
		versionType = changelog.Calver
	}
	result.VersionType = versionType.String()

	// Get current version
	result.PreviousVersion = existingChangelog.LatestVersionNumber()
	if result.PreviousVersion == "" {
		result.PreviousVersion = "0.0.1"
	}

	// Find latest tag for this module
	tagPattern := module + "/*"
	fromRef, err := repo.LatestTag(tagPattern)
	if err != nil {
		fromRef = ""
	}

	// Get commits since last release
	commits, err := repo.CommitsBetween(fromRef, "HEAD")
	if err != nil {
		result.Error = fmt.Sprintf("failed to get commits: %v", err)
		return result
	}

	if len(commits) == 0 {
		result.Error = "no commits found since last release"
		return result
	}

	// Filter commits by module file patterns
	modulePatterns := moduleContract.GetGlobPatterns()
	var filteredCommits []*changelog.Commit
	for _, c := range commits {
		parsed := changelog.ParseCommitMessage(c.Message)
		parsed.SHA = c.ShortSHA
		parsed.Date = c.Date
		parsed.Files = c.Files

		if len(modulePatterns) > 0 {
			if commitMatchesModule(c.Files, modulePatterns) {
				filteredCommits = append(filteredCommits, parsed)
			}
		} else {
			filteredCommits = append(filteredCommits, parsed)
		}
	}

	if len(filteredCommits) == 0 {
		result.Error = fmt.Sprintf("no commits affecting module '%s' found since last release", module)
		return result
	}

	// Collect entries for version calculation from conventional commits
	// Note: Release decision is based on file changes, not conventional commit format
	var entries []changelog.Entry
	for _, c := range filteredCommits {
		if c.IsConventionalCommit() {
			entry := c.ToEntry()
			entries = append(entries, entry)
		}
	}

	// File changes determine if release is needed, not conventional commit format
	// Changelog may be empty if no conventional commits, but release still proceeds
	result.EntriesAdded = len(entries)

	// Get existing version numbers for calver collision detection
	var existingVersions []string
	for _, v := range existingChangelog.Versions {
		existingVersions = append(existingVersions, v.Number)
	}

	// Calculate next version with constraints
	maxBump := changelog.BumpMajor
	if defs.IsPatchOnly() {
		maxBump = changelog.BumpPatch
	}

	// Determine release date
	releaseDate := time.Now()
	if overrideDate != "" {
		releaseDate, err = time.Parse("2006-01-02", overrideDate)
		if err != nil {
			result.Error = fmt.Sprintf("invalid date format '%s' (use YYYY-MM-DD)", overrideDate)
			return result
		}
	}

	// Pass hasFileChanges to ensure version bumps even without conventional commits
	hasFileChanges := len(filteredCommits) > 0
	newVersion, err := changelog.CalculateNextVersionConstrained(
		result.PreviousVersion,
		versionType,
		entries,
		releaseDate,
		existingVersions,
		maxBump,
		hasFileChanges,
	)
	if err != nil {
		result.Error = fmt.Sprintf("failed to calculate version: %v", err)
		return result
	}

	result.NewVersion = newVersion
	result.Tag = fmt.Sprintf("%s/%s", module, newVersion)

	// If dry run, stop here
	if dryRun {
		result.Success = true
		return result
	}

	// Create version entry from commits
	newVersionEntry := changelog.CommitsToVersion(filteredCommits, newVersion, releaseDate)

	// Merge any manual entries from [Unreleased] section
	if existingChangelog.Unreleased != nil && existingChangelog.Unreleased.HasEntries() {
		newVersionEntry.Added = append(existingChangelog.Unreleased.Added, newVersionEntry.Added...)
		newVersionEntry.Changed = append(existingChangelog.Unreleased.Changed, newVersionEntry.Changed...)
		newVersionEntry.Deprecated = append(existingChangelog.Unreleased.Deprecated, newVersionEntry.Deprecated...)
		newVersionEntry.Removed = append(existingChangelog.Unreleased.Removed, newVersionEntry.Removed...)
		newVersionEntry.Fixed = append(existingChangelog.Unreleased.Fixed, newVersionEntry.Fixed...)
		newVersionEntry.Security = append(existingChangelog.Unreleased.Security, newVersionEntry.Security...)
	}
	// Clear Unreleased (will be output as empty placeholder)
	existingChangelog.Unreleased = nil

	// Add new version to changelog
	existingChangelog.AddVersion(newVersionEntry)
	existingChangelog.Module = module
	existingChangelog.VersionType = versionType

	// Extract repo URL from git remote if not set
	if existingChangelog.RepoURL == "" {
		if remoteURL, err := repo.RemoteURL("origin"); err == nil {
			existingChangelog.RepoURL = normalizeGitHubURL(remoteURL)
		}
	}

	// Create release directory if needed
	releaseDir := filepath.Dir(fullChangelogPath)
	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create release directory: %v", err)
		return result
	}

	// Write changelog
	if err := existingChangelog.Write(fullChangelogPath); err != nil {
		result.Error = fmt.Sprintf("failed to write changelog: %v", err)
		return result
	}

	result.Success = true
	return result
}
