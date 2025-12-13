// Command: release pending
// Short: Check if module has pending changes for release
// Long: Analyzes commits since the last release and outputs release decision data.
// Long:
// Long: This command checks the git history and changelog to determine if there
// Long: are unreleased changes that warrant a new version.
// Long:
// Long: Output includes:
// Long:   - has_changes: whether there are releasable changes
// Long:   - current_version: the current released version
// Long:   - next_version: the calculated next version
// Long:   - change_counts: breakdown by change type (added, fixed, changed, etc.)
// Long:
// Long: This is designed for CI/CD pipelines to determine if a release is needed.
// Long:
// Long: Expected Output:
// Long:   - JSON object containing:
// Long:     - has_changes: boolean indicating if there are releasable changes
// Long:     - current_version: the current released version
// Long:     - next_version: the calculated next version
// Long:     - change_counts: breakdown by change type (added, fixed, changed, etc.)
// Long:
// Long: Examples:
// Long:   release pending r2r-cli           # Check r2r-cli for pending changes
// Long:   release pending r2r-cli --quiet   # Exit code only (0=changes, 1=no changes)
// Long:   release pending --all             # Check all releasable modules
// Flag.quiet: type=bool, usage=Suppress output, use exit code only (0=has changes, 1=no changes)
// Flag.all: type=bool, usage=Check all modules with changelogs
// Args: modules
package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/changelog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/git"
)

func init() {
	registry.Register(ReleasePending)
}

// PendingRelease contains release decision data for CI/CD
type PendingRelease struct {
	Module         string        `json:"module"`
	HasChanges     bool          `json:"has_changes"`
	CurrentVersion string        `json:"current_version"`
	NextVersion    string        `json:"next_version"`
	VersionType    string        `json:"version_type"`
	Constraint     string        `json:"constraint"`
	Tag            string        `json:"tag"`
	CommitsTotal   int           `json:"commits_total"`
	CommitsModule  int           `json:"commits_module"`
	ChangeSummary  ChangeSummary `json:"change_summary"`
}

// ChangeSummary breaks down changes by type
type ChangeSummary struct {
	Added      int `json:"added"`
	Changed    int `json:"changed"`
	Deprecated int `json:"deprecated"`
	Removed    int `json:"removed"`
	Fixed      int `json:"fixed"`
	Security   int `json:"security"`
}

// PendingResult is the overall result for one or more modules
type PendingResult struct {
	Modules      []PendingRelease `json:"modules"`
	HasAnyChange bool             `json:"has_any_change"`
}

func ReleasePending() int {
	// Parse flags
	module := ""
	quiet := false
	checkAll := false

	args := os.Args[3:] // Skip binary, "release", "pending"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--quiet" || arg == "-q":
			quiet = true
		case arg == "--all":
			checkAll = true
		default:
			if !strings.HasPrefix(arg, "--") && module == "" {
				module = arg
			}
		}
	}

	if !checkAll && module == "" {
		log.Error("module moniker required (or use --all)")
		log.Info("Usage: release pending <module> [--quiet]")
		log.Info("       release pending --all")
		return 1
	}

	// Load workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		log.Errorf("failed to get workspace root: %v", err)
		return 1
	}

	// Load config for versioning constraints
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}
	versioning := cfg.Repository.Repository.Versioning

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace("")
	if err != nil {
		log.Errorf("failed to load modules: %v", err)
		return 1
	}

	// Open git repository
	repo, err := git.Open("")
	if err != nil {
		log.Errorf("failed to open git repository: %v", err)
		return 1
	}

	// Determine which modules to check
	var modulesToCheck []string
	if checkAll {
		// Find all modules with changelogs
		modulesToCheck = findModulesWithChangelogs(workspaceRoot)
	} else {
		modulesToCheck = []string{module}
	}

	// Check each module
	result := PendingResult{
		Modules:      make([]PendingRelease, 0, len(modulesToCheck)),
		HasAnyChange: false,
	}

	for _, mod := range modulesToCheck {
		pending, err := checkModulePending(mod, moduleRegistry, repo, versioning, workspaceRoot)
		if err != nil {
			log.Warnf("failed to check module '%s': %v", mod, err)
			continue
		}
		result.Modules = append(result.Modules, pending)
		if pending.HasChanges {
			result.HasAnyChange = true
		}
	}

	// Output results
	if quiet {
		if result.HasAnyChange {
			return 0 // Has changes
		}
		return 1 // No changes
	}

	// JSON output
	var output interface{}
	if len(result.Modules) == 1 {
		output = result.Modules[0]
	} else {
		output = result
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Errorf("failed to marshal JSON: %v", err)
		return 1
	}
	log.Info(string(jsonBytes))

	return 0
}

func checkModulePending(module string, moduleRegistry *modules.Registry, repo git.GitRepository, versioning config.VersioningConfig, workspaceRoot string) (PendingRelease, error) {
	pending := PendingRelease{
		Module:     module,
		HasChanges: false,
	}

	// Get module contract
	moduleContract, exists := moduleRegistry.Get(module)
	if !exists {
		return pending, fmt.Errorf("module '%s' not found", module)
	}

	// Determine changelog path from module contract
	changelogPath := moduleContract.GetChangelogPath()
	fullChangelogPath := filepath.Join(workspaceRoot, changelogPath)

	// Parse existing changelog
	var existingChangelog *changelog.Changelog
	if _, err := os.Stat(fullChangelogPath); err == nil {
		existingChangelog, err = changelog.Parse(fullChangelogPath)
		if err != nil {
			return pending, fmt.Errorf("failed to parse changelog: %v", err)
		}
	} else {
		existingChangelog = &changelog.Changelog{
			Module:      module,
			VersionType: changelog.Semver,
		}
	}

	// Get version info
	pending.CurrentVersion = existingChangelog.LatestVersionNumber()
	if pending.CurrentVersion == "" {
		pending.CurrentVersion = "0.0.1"
	}

	versionType := existingChangelog.VersionType
	if module == "docs" {
		versionType = changelog.Calver
	}
	pending.VersionType = versionType.String()

	// Get constraint
	if versioning.IsPatchOnly() {
		pending.Constraint = "patch-only"
	} else if versioning.IsCalverOnly() {
		pending.Constraint = "calver-only"
	} else {
		pending.Constraint = "unrestricted"
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
		return pending, fmt.Errorf("failed to get commits: %v", err)
	}

	pending.CommitsTotal = len(commits)

	if len(commits) == 0 {
		pending.Tag = fmt.Sprintf("%s/%s", module, pending.CurrentVersion)
		return pending, nil
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

	pending.CommitsModule = len(filteredCommits)

	// HasChanges is based SOLELY on whether module files changed (file ownership)
	// NOT on whether commits follow conventional commit format
	pending.HasChanges = len(filteredCommits) > 0

	if len(filteredCommits) == 0 {
		pending.NextVersion = pending.CurrentVersion
		pending.Tag = fmt.Sprintf("%s/%s", module, pending.CurrentVersion)
		return pending, nil
	}

	// Collect entries and count changes from conventional commits
	// This is for changelog content and summary display, not for release decision
	var entries []changelog.Entry
	for _, c := range filteredCommits {
		if c.IsConventionalCommit() {
			entry := c.ToEntry()
			entries = append(entries, entry)

			// Count by change type
			switch changelog.CommitTypeToChangeType(c.Type) {
			case changelog.Added:
				pending.ChangeSummary.Added++
			case changelog.Changed:
				pending.ChangeSummary.Changed++
			case changelog.Deprecated:
				pending.ChangeSummary.Deprecated++
			case changelog.Removed:
				pending.ChangeSummary.Removed++
			case changelog.Fixed:
				pending.ChangeSummary.Fixed++
			case changelog.Security:
				pending.ChangeSummary.Security++
			}
		}
	}

	// Calculate next version
	var existingVersions []string
	for _, v := range existingChangelog.Versions {
		existingVersions = append(existingVersions, v.Number)
	}

	maxBump := changelog.BumpMajor
	if versioning.IsPatchOnly() {
		maxBump = changelog.BumpPatch
	}

	// Pass hasFileChanges (len(filteredCommits) > 0) to ensure version bumps
	// even when commits don't follow conventional format
	hasFileChanges := len(filteredCommits) > 0
	nextVersion, err := changelog.CalculateNextVersionConstrained(
		pending.CurrentVersion,
		versionType,
		entries,
		time.Now(),
		existingVersions,
		maxBump,
		hasFileChanges,
	)
	if err != nil {
		return pending, fmt.Errorf("failed to calculate version: %v", err)
	}

	pending.NextVersion = nextVersion
	pending.Tag = fmt.Sprintf("%s/%s", module, nextVersion)

	return pending, nil
}

func findModulesWithChangelogs(workspaceRoot string) []string {
	releaseDir := filepath.Join(workspaceRoot, "release")
	entries, err := os.ReadDir(releaseDir)
	if err != nil {
		return nil
	}

	var modules []string
	for _, entry := range entries {
		if entry.IsDir() {
			changelogPath := filepath.Join(releaseDir, entry.Name(), "CHANGELOG.md")
			if _, err := os.Stat(changelogPath); err == nil {
				modules = append(modules, entry.Name())
			}
		}
	}
	return modules
}
