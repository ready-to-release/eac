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
)

type releasePendingCommand struct{}

var _ core.SimpleCommandPort = (*releasePendingCommand)(nil)

func (c *releasePendingCommand) Name() string { return "release pending" }

func (c *releasePendingCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-pending",
		Short:         "Check if module has pending changes for release",
		Long:          "Analyzes commits since the last release and outputs release decision data.\n\nThis command checks the git history and changelog to determine if there\nare unreleased changes that warrant a new version.\n\nOutput includes:\n  - has_changes: whether there are releasable changes\n  - current_version: the current released version\n  - next_version: the calculated next version\n  - change_counts: breakdown by change type (added, fixed, changed, etc.)\n\nThis is designed for CI/CD pipelines to determine if a release is needed.\n\nExpected Output:\n  - JSON object containing:\n    - has_changes: boolean indicating if there are releasable changes\n    - current_version: the current released version\n    - next_version: the calculated next version\n    - change_counts: breakdown by change type (added, fixed, changed, etc.)\n\nExamples:\n  release pending clie            # Check clie for pending changes\n  release pending clie --quiet    # Exit code only (0=changes, 1=no changes)\n  release pending --all              # Check all releasable modules\n  release pending --published        # Check only published modules\n  release pending --internal         # Check only internal modules",
		Args: "modules",
		Flags: []core.FlagSpec{
			{Name: "quiet", Type: "bool", Usage: "Suppress output, use exit code only (0=has changes, 1=no changes)"},
			{Name: "all", Type: "bool", Usage: "Check all modules with changelogs"},
			{Name: "published", Type: "bool", Usage: "Check only published modules"},
			{Name: "internal", Type: "bool", Usage: "Check only internal modules"},
		},
	}
}

func (c *releasePendingCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleasePending()
}

// PendingRelease contains release decision data for CI/CD.
type PendingRelease struct {
	Module         string        `json:"module"`
	ReleaseType    string        `json:"release_type"`    // published | internal | bundle | none
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

// ChangeSummary breaks down changes by type.
type ChangeSummary struct {
	Added      int `json:"added"`
	Changed    int `json:"changed"`
	Deprecated int `json:"deprecated"`
	Removed    int `json:"removed"`
	Fixed      int `json:"fixed"`
	Security   int `json:"security"`
}

// PendingResult is the overall result for one or more modules.
type PendingResult struct {
	Modules      []PendingRelease `json:"modules"`
	HasAnyChange bool             `json:"has_any_change"`
}

func ReleasePending() int {
	s, exitCode := newReleaseScaffold(withModules(), withGit(), withConfig())
	if s == nil {
		return exitCode
	}

	// Parse flags
	module := ""
	quiet := false
	checkAll := false
	filterPublished := false
	filterInternal := false

	args := os.Args[3:] // Skip binary, "release", "pending"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--quiet", "-q":
			quiet = true
		case "--all":
			checkAll = true
		case "--published":
			filterPublished = true
			checkAll = true // Implies --all
		case "--internal":
			filterInternal = true
			checkAll = true // Implies --all
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

	workspaceRoot := s.WorkspaceRoot
	versioning := s.Config.Repository.Versioning
	moduleRegistry := s.ModuleRegistry
	repo := s.Repo

	// Determine which modules to check
	var modulesToCheck []string
	if checkAll {
		// Find all modules with changelogs (both published and internal)
		modulesToCheck = findModulesWithChangelogsAll(workspaceRoot)

		// Apply release type filter if specified
		if filterPublished {
			modulesToCheck = filterModulesByReleaseType(workspaceRoot, modulesToCheck, "published")
		} else if filterInternal {
			modulesToCheck = filterModulesByReleaseType(workspaceRoot, modulesToCheck, "internal")
		}
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
		Module:      module,
		HasChanges:  false,
		ReleaseType: getModuleReleaseType(workspaceRoot, module),
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
			return pending, fmt.Errorf("failed to parse changelog: %w", err)
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
		return pending, fmt.Errorf("failed to get commits: %w", err)
	}

	pending.CommitsTotal = len(commits)

	if len(commits) == 0 {
		pending.Tag = fmt.Sprintf("%s/%s", module, pending.CurrentVersion)
		return pending, nil
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
	for i := range existingChangelog.Versions {
		v := &existingChangelog.Versions[i]
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
		return pending, fmt.Errorf("failed to calculate version: %w", err)
	}

	pending.NextVersion = nextVersion
	pending.Tag = fmt.Sprintf("%s/%s", module, nextVersion)

	return pending, nil
}

func findModulesWithChangelogs(workspaceRoot string) []string {
	// NOTE: config.RepositoryConfig.ReleasePathAbs() is now available but this
	// function doesn't have access to config. Uses the same conventional "release/" path.
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

// findModulesWithChangelogsAll finds all modules that have changelogs (both published and internal).
func findModulesWithChangelogsAll(workspaceRoot string) []string {
	// Load module registry to get all modules with versioning
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil
	}

	var modulesWithChangelogs []string
	for _, moniker := range moduleRegistry.AllMonikers() {
		moduleContract, exists := moduleRegistry.Get(moniker)
		if !exists || moduleContract.Versioning == nil {
			continue
		}

		// Check if changelog file exists
		changelogPath := moduleContract.GetChangelogPath()
		if changelogPath == "" {
			continue
		}

		fullPath := filepath.Join(workspaceRoot, changelogPath)
		if _, err := os.Stat(fullPath); err == nil {
			modulesWithChangelogs = append(modulesWithChangelogs, moniker)
		}
	}

	return modulesWithChangelogs
}

// filterModulesByReleaseType filters modules by their release type.
// filterType can be "published", "internal", "bundle", "none", or "" (all).
func filterModulesByReleaseType(workspaceRoot string, moduleMonikers []string, filterType string) []string {
	if filterType == "" {
		return moduleMonikers
	}

	// Load module registry
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil
	}

	var filtered []string
	for _, moniker := range moduleMonikers {
		moduleContract, exists := moduleRegistry.Get(moniker)
		if !exists || moduleContract.Versioning == nil {
			continue
		}

		releaseType := moduleContract.Versioning.ReleaseType
		if releaseType == filterType {
			filtered = append(filtered, moniker)
		}
	}

	return filtered
}

// getModuleReleaseType returns the release type for a module.
func getModuleReleaseType(workspaceRoot string, module string) string {
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return ""
	}

	moduleContract, exists := moduleRegistry.Get(module)
	if !exists || moduleContract.Versioning == nil {
		return ""
	}

	return moduleContract.Versioning.ReleaseType
}
