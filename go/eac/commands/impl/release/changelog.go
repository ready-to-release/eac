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
// Long: Expected Output:
// Long:   - Preview of changelog entries (default)
// Long:   - Updated release/<module>/CHANGELOG.md file if --write flag is specified
// Long:
// Long: Examples:
// Long:   release changelog r2r-cli              # Preview changelog entries
// Long:   release changelog r2r-cli --write      # Update release/r2r-cli/CHANGELOG.md
// Long:   release changelog r2r-cli --from v1.0  # From specific tag
// Flag.write: type=bool, usage=Write changes to changelog file (default: preview only)
// Flag.from: type=string, usage=Start from specific tag/ref (default: latest release tag)
// Flag.to: type=string, usage=End at specific tag/ref (default: HEAD)
// Flag.version: type=string, usage=Override calculated version number
// Flag.date: type=string, usage=Override release date (YYYY-MM-DD format)
// Flag.breaking: type=bool, usage=Force major version bump
package release

import (
	"os"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/changelog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// log is the package-level logger for release commands.
var log = logging.C()

func init() {
	registry.Register(ReleaseChangelog)
}

func ReleaseChangelog() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

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
		log.Error("module moniker required")
		log.Info("Usage: release changelog <module> [--write] [--from <ref>] [--version <ver>]")
		return 1
	}

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace("")
	if err != nil {
		log.Errorf("failed to load modules: %v", err)
		return 1
	}

	moduleContract, exists := moduleRegistry.Get(module)
	if !exists {
		log.Errorf("module '%s' not found", module)
		return 1
	}

	// Open git repository
	// Open git repository
	gitMgr := git.NewManager(logging.C().Zap())
	repo, err := gitMgr.Open("")
	if err != nil {
		log.Errorf("failed to open git repository: %v", err)
		return 1
	}

	// Determine changelog path from module contract
	changelogPath := moduleContract.GetChangelogPath()

	// Parse existing changelog or create new
	var existingChangelog *changelog.Changelog
	if _, err := os.Stat(changelogPath); err == nil {
		existingChangelog, err = changelog.Parse(changelogPath)
		if err != nil {
			log.Warnf("failed to parse existing changelog: %v", err)
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
			log.Warnf("failed to get latest tag: %v", err)
		}
		fromRef = latestTag
	}

	// Get commits since last release
	commits, err := repo.CommitsBetween(fromRef, toRef)
	if err != nil {
		log.Errorf("failed to get commits: %v", err)
		return 1
	}

	if len(commits) == 0 {
		log.Info("No commits found since last release.")
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
		log.Infof("No commits affecting module '%s' found since last release.", module)
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

	// Load config for versioning constraints
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	// Determine max bump based on constraints
	maxBump := changelog.BumpMajor // Default: unrestricted
	if cfg.Repository.Repository.Versioning.IsPatchOnly() {
		maxBump = changelog.BumpPatch
		if forceBreaking {
			log.Warn("--breaking ignored due to patch-only constraint in .r2r/eac/repository.yml")
			forceBreaking = false
		}
	}

	// Calculate next version
	// hasFileChanges ensures version bumps even without conventional commits
	hasFileChanges := len(filteredCommits) > 0

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
			hasFileChanges,
		)
		if err != nil {
			log.Errorf("failed to calculate version: %v", err)
			return 1
		}
	}

	// Determine release date
	releaseDate := time.Now()
	if overrideDate != "" {
		releaseDate, err = time.Parse("2006-01-02", overrideDate)
		if err != nil {
			log.Errorf("invalid date format '%s' (use YYYY-MM-DD)", overrideDate)
			return 1
		}
	}

	// Create version entry from commits
	newVersionEntry := changelog.CommitsToVersion(filteredCommits, newVersion, releaseDate)

	// Display preview
	log.Infof("Module: %s", module)
	log.Infof("Current version: %s", currentVersion)
	log.Infof("New version: %s", newVersion)
	if cfg.Repository.Repository.Versioning.IsPatchOnly() && versionType == changelog.Semver {
		log.Info("Version constraint: patch-only (from .r2r/eac/repository.yml)")
	}
	log.Infof("Commits analyzed: %d", len(commits))
	log.Infof("Module commits: %d", len(filteredCommits))
	log.Info("")

	if !newVersionEntry.HasEntries() {
		log.Info("No conventional commits found to generate changelog entries.")
		log.Info("Commits must follow format: type(scope): description")
		return 0
	}

	// Preview the changelog entry
	log.Info("--- Changelog Preview ---")
	previewChangelog := &changelog.Changelog{
		Module:      module,
		VersionType: versionType,
		Versions:    []changelog.Version{newVersionEntry},
	}
	log.Info(previewChangelog.String())
	log.Info("-------------------------")

	if !write {
		log.Infof("\nTo write changes, run: release changelog %s --write", module)
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
		log.Errorf("failed to write changelog: %v", err)
		return 1
	}

	log.Infof("\n✅ Updated %s with version %s", changelogPath, newVersion)
	return 0
}

// commitMatchesModule checks if any of the commit's files match module patterns.
func commitMatchesModule(files, patterns []string) bool {
	for _, file := range files {
		for _, pattern := range patterns {
			if matchChangelogPattern(file, pattern) {
				return true
			}
		}
	}
	return false
}

// matchChangelogPattern provides simple glob matching for changelog filtering.
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

// normalizeGitHubURL converts git remote URLs to HTTPS format.
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
