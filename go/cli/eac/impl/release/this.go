package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/changelog"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/releasenotes"
	"github.com/ready-to-release/eac/go/core/repository"
)

type releaseThisCommand struct{}

var _ core.SimpleCommandPort = (*releaseThisCommand)(nil)

func (c *releaseThisCommand) Name() string { return "release this" }

func (c *releaseThisCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-this",
		Short:         "Finalize changelog and prepare module for release",
		Long:          "Updates the changelog with all commits since the last release and prepares\nthe module for release.\n\nThis command:\n  1. Analyzes commits since the last release tag\n  2. Generates changelog entries from conventional commits\n  3. Merges any manual entries from [Unreleased] section\n  4. Calculates the next version (respecting constraints from .eac/repository.yml)\n  5. Adds a new version section to the changelog\n  6. Clears the [Unreleased] section\n  7. Writes the updated changelog\n\nManual entries in [Unreleased] are preserved and merged with auto-generated\nentries. This allows team members to add changelog entries that aren't tied\nto specific commits.\n\nAfter running this command, commit the changelog and create a PR. Once merged,\nthe release-auto workflow will detect the new version and create the git tag,\nwhich triggers the module's release workflow.\n\nExpected Output:\n  - Updated changelog with new version section added\n  - [Unreleased] section cleared\n  - Changelog file ready to be committed and submitted in a pull request\n\nExamples:\n  release this clie                      # Update single module changelog\n  release this clie eac-ext              # Update multiple modules\n  release this clie eac-ext --dry-run    # Preview without writing\n  release this clie --json               # Output result as JSON",
		Args:          "modules",
		Flags: []core.FlagSpec{
			{Name: "dry-run", Type: "bool", Usage: "Preview changes without writing to changelog"},
			{Name: "json", Type: "bool", Usage: "Output result in JSON format"},
			{Name: "date", Type: "string", Usage: "Override release date (YYYY-MM-DD format)"},
		},
	}
}

func (c *releaseThisCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseThis()
}

// ReleaseResult contains the result of a release operation.
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
	AlreadyPending  bool   `json:"already_pending,omitempty"`
	Error           string `json:"error,omitempty"`
}

func ReleaseThis() int {
	s, exitCode := newReleaseScaffold()
	if s == nil {
		return exitCode
	}

	// Parse flags
	var moduleList []string
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
			if !strings.HasPrefix(arg, "--") {
				moduleList = append(moduleList, arg)
			}
		}
	}

	if len(moduleList) == 0 {
		log.Error("module moniker required")
		log.Info("Usage: release this <module> [<module>...] [--dry-run] [--json]")
		return 1
	}

	// Process all modules
	var results []ReleaseResult
	for _, module := range moduleList {
		result := performRelease(module, dryRun, overrideDate)
		results = append(results, result)
	}

	// Output results
	if asJSON {
		jsonBytes, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Errorf("failed to marshal JSON: %v", err)
			return 1
		}
		log.Info(string(jsonBytes))
	} else {
		if dryRun {
			log.Info("=== DRY RUN - No changes written ===")
			log.Info("")
		}

		for i := range results {
			result := &results[i]
			if i > 0 {
				log.Info("---")
				log.Info("")
			}

			if result.Error != "" {
				log.Errorf("Module %s: %s", result.Module, result.Error)
				continue
			}

			log.Infof("Module: %s", result.Module)
			log.Infof("Previous version: %s", result.PreviousVersion)
			log.Infof("New version: %s", result.NewVersion)
			log.Infof("Tag: %s", result.Tag)

			if result.AlreadyPending {
				log.Info("")
				log.Info("Already pending release - no changes made")
			} else {
				log.Infof("Entries added: %d", result.EntriesAdded)
				if !dryRun {
					log.Info("")
					log.Infof("Updated: %s", result.ChangelogPath)
				}
			}
		}

		if !dryRun && len(results) > 0 {
			// Collect modules that were actually modified (not already pending)
			var modifiedModules []string
			var changelogPaths []string
			for i := range results {
				r := &results[i]
				if r.Success && !r.AlreadyPending {
					modifiedModules = append(modifiedModules, r.Module)
					changelogPaths = append(changelogPaths, r.ChangelogPath)
				}
			}

			if len(modifiedModules) > 0 {
				log.Info("")
				log.Info("Next steps:")
				log.Infof("  1. git add %s", strings.Join(changelogPaths, " "))
				if len(modifiedModules) == 1 {
					// Find the version for this module
					var version string
					for i := range results {
						r := &results[i]
						if r.Module == modifiedModules[0] {
							version = r.NewVersion
							break
						}
					}
					log.Infof("  2. git commit -m \"release(%s): %s\"", modifiedModules[0], version)
				} else {
					log.Infof("  2. git commit -m \"release: %s\"", strings.Join(modifiedModules, ", "))
				}
				log.Info("  3. Create PR and merge to main")
				log.Info("  4. release-auto workflow will create tags and trigger releases")
			}
		}
	}

	// Return failure if any module failed
	for i := range results {
		result := &results[i]
		if !result.Success {
			return 1
		}
	}
	return 0
}

func performRelease(module string, dryRun bool, overrideDate string) ReleaseResult {
	result := ReleaseResult{
		Module: module,
		DryRun: dryRun,
	}

	// Load workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		result.Error = fmt.Sprintf("failed to get workspace root: %v", err)
		return result
	}

	// Load config for versioning constraints
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		result.Error = fmt.Sprintf("failed to load config: %v", err)
		return result
	}
	versioning := cfg.Repository.Repository.Versioning

	// Set constraint info
	if versioning.IsPatchOnly() {
		result.Constraint = "patch-only"
	} else if versioning.IsCalverOnly() {
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
	// Open git repository
	gitMgr := git.NewManager(logging.C().Zap())
	repo, err := gitMgr.Open("")
	if err != nil {
		result.Error = fmt.Sprintf("failed to open git repository: %v", err)
		return result
	}

	// Determine changelog path from module contract
	changelogPath := moduleContract.GetChangelogPath()
	if changelogPath == "" {
		result.Error = fmt.Sprintf("module '%s' does not have a changelog configured (requires versioning.scheme: SemVer or explicit versioning.changelog path)", module)
		return result
	}
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

	// Get current version from changelog
	result.PreviousVersion = existingChangelog.LatestVersionNumber()
	if result.PreviousVersion == "" {
		result.PreviousVersion = "0.0.0" // Start from 0.0.0 so first release is 0.0.1
	}

	// Find latest tag for this module
	tagPattern := module + "/*"
	fromRef, err := repo.LatestTag(tagPattern)
	if err != nil {
		fromRef = ""
	}

	// Idempotency check: if changelog has a version without a corresponding tag,
	// that version is already pending release - don't bump again
	if result.PreviousVersion != "0.0.0" {
		expectedTag := fmt.Sprintf("%s/%s", module, result.PreviousVersion)
		tagExists, tagErr := repo.TagExists(expectedTag)
		if tagErr == nil && !tagExists {
			// Changelog has a version that hasn't been tagged yet - already pending
			result.NewVersion = result.PreviousVersion
			result.Tag = expectedTag
			result.AlreadyPending = true
			result.Success = true
			return result
		}
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
	for i := range commits {
		c := &commits[i]
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
	for i := range existingChangelog.Versions {
		v := &existingChangelog.Versions[i]
		existingVersions = append(existingVersions, v.Number)
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

	// Calculate next version with constraints
	maxBump := changelog.BumpMajor
	if versioning.IsPatchOnly() {
		maxBump = changelog.BumpPatch
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
	existingChangelog.AddVersion(&newVersionEntry)
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
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		result.Error = fmt.Sprintf("failed to create release directory: %v", err)
		return result
	}

	// If dry run, skip validation and changelog writing
	if dryRun {
		result.Success = true
		return result
	}

	// Validate RELEASE-NOTES.md before writing changelog
	releaseNotesPath := moduleContract.GetReleaseNotesPath()
	fullReleaseNotesPath := filepath.Join(workspaceRoot, releaseNotesPath)

	// Check if RELEASE-NOTES.md exists
	if _, err := os.Stat(fullReleaseNotesPath); os.IsNotExist(err) {
		// Generate from template using repository config
		if err := releasenotes.GenerateTemplate(workspaceRoot, cfg.Repository, fullReleaseNotesPath, module, newVersion, releaseDate); err != nil {
			result.Error = fmt.Sprintf("failed to generate release notes: %v", err)
			return result
		}

		result.Error = fmt.Sprintf(
			"RELEASE-NOTES.md created at %s\n\n"+
				"Please:\n"+
				"  1. Edit the file and fill in required sections\n"+
				"  2. Run 'release this %s' again to complete the release",
			releaseNotesPath, module)
		return result // HALT - user must edit
	}

	// Parse and validate release notes
	rn, err := releasenotes.Parse(fullReleaseNotesPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to parse release notes: %v", err)
		return result
	}

	if err := rn.ValidateVersion(newVersion); err != nil {
		result.Error = fmt.Sprintf(
			"RELEASE-NOTES.md validation failed: %v\n\n"+
				"Please add version [%s] to %s\n\n"+
				"Expected format:\n\n"+
				"  ## [%s] - %s\n\n"+
				"  ### Summary\n\n"+
				"  Your release description here...",
			err, newVersion, releaseNotesPath, newVersion, releaseDate.Format("2006-01-02"))
		return result // HALT - version missing or empty
	}

	// Validation passed - write changelog
	if err := existingChangelog.Write(fullChangelogPath); err != nil {
		result.Error = fmt.Sprintf("failed to write changelog: %v", err)
		return result
	}

	result.Success = true
	return result
}
