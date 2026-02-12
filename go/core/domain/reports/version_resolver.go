package reports

import (
	"fmt"

	"github.com/ready-to-release/eac/go/core/changelog"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
)

// VersionInfo represents resolved version information for a module.
type VersionInfo struct {
	Module         string // Module moniker
	VersionNumber  string // Actual version number or "Unreleased"
	GitTag         string // Git tag for this version (e.g., "eac-ext/0.0.7")
	PreviousGitTag string // Git tag for previous version (empty if first release)
	IsUnreleased   bool   // True if this represents unreleased changes
	IsLatest       bool   // True if this is the latest released version
}

// ResolveVersion resolves a version string to concrete version information.
// Version string can be: "", "unreleased", "latest", or a specific version number.
// This function provides consistent version resolution logic for all show/get commands.
func ResolveVersion(workspaceRoot, module, versionStr string) (*VersionInfo, error) {
	// Load config to validate module exists
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	mod, ok := cfg.Repository.GetModule(module)
	if !ok {
		return nil, fmt.Errorf("module not found: %s", module)
	}

	// Implicit versioned modules don't have changelogs - they derive their version from parent modules
	// Modules without versioning config also default to Implicit
	if mod.Versioning == nil || mod.Versioning.Scheme == "Implicit" {
		return resolveImplicitVersion(module, versionStr)
	}

	// Parse changelog to get version history
	changelogPath := paths.ChangelogPath(workspaceRoot, module)
	log, err := changelog.Parse(changelogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse changelog: %w", err)
	}

	info := &VersionInfo{
		Module: module,
	}

	// Handle empty or "unreleased" - both mean unreleased changes
	if versionStr == "" || versionStr == "unreleased" {
		info.VersionNumber = "Unreleased"
		info.IsUnreleased = true

		// Previous tag is the latest release (if any)
		latestVer := log.LatestVersion()
		if latestVer != nil {
			info.GitTag = module + "/" + latestVer.Number
			info.PreviousGitTag = info.GitTag // For unreleased, compare since latest
		} else {
			info.GitTag = ""         // No previous release
			info.PreviousGitTag = "" // Compare from beginning
		}

		return info, nil
	}

	// Handle "latest" - most recent released version
	if versionStr == "latest" {
		latestVer := log.LatestVersion()
		if latestVer == nil {
			return nil, fmt.Errorf("no released versions found in changelog")
		}

		info.VersionNumber = latestVer.Number
		info.GitTag = module + "/" + latestVer.Number
		info.IsLatest = true

		// Find previous version for comparison
		if len(log.Versions) > 1 {
			info.PreviousGitTag = module + "/" + log.Versions[1].Number
		} else {
			info.PreviousGitTag = "" // First release
		}

		return info, nil
	}

	// Handle specific version number
	ver := log.GetVersion(versionStr)
	if ver == nil {
		return nil, fmt.Errorf("version not found: %s", versionStr)
	}

	info.VersionNumber = versionStr
	info.GitTag = module + "/" + versionStr
	info.IsLatest = (log.LatestVersion() != nil && log.LatestVersion().Number == versionStr)

	// Find previous version
	for i, v := range log.Versions {
		if v.Number == versionStr {
			if i+1 < len(log.Versions) {
				info.PreviousGitTag = module + "/" + log.Versions[i+1].Number
			} else {
				info.PreviousGitTag = "" // First release
			}
			break
		}
	}

	return info, nil
}

// ResolveBranchEndRef resolves a branch parameter to the git end-reference used
// when querying commits. The resolution rules are:
//   - If branch is non-empty and not "HEAD"/"current", use it directly.
//   - If branch is empty, fall back to the configured trunk branch (TrunkBranch),
//     then to "main" if TrunkBranch is also empty.
//   - If branch is "HEAD" or "current", resolve to "HEAD" (current branch).
//
// This is the single source of truth for branch resolution in reports;
// both GetSpecs and GetApprovalComments delegate here.
func ResolveBranchEndRef(branch, trunkBranch string) string {
	endRef := branch
	if endRef == "" {
		// Default to configured trunk branch (usually "main" or "master")
		endRef = trunkBranch
		if endRef == "" {
			endRef = "main" // Fallback if not configured
		}
	}
	// Special case: "HEAD" or "current" means use current branch
	if endRef == "HEAD" || endRef == "current" {
		endRef = "HEAD"
	}
	return endRef
}

// resolveImplicitVersion handles version resolution for implicit-versioned modules.
// Implicit modules don't have changelogs - they derive their version from parent modules.
// Only "unreleased" or "" are valid for implicit modules.
func resolveImplicitVersion(module, versionStr string) (*VersionInfo, error) {
	info := &VersionInfo{
		Module: module,
	}

	// Handle empty or "unreleased" - the only valid options for implicit modules
	if versionStr == "" || versionStr == "unreleased" {
		info.VersionNumber = "Unreleased"
		info.IsUnreleased = true
		info.GitTag = ""         // Implicit modules don't have tags
		info.PreviousGitTag = "" // Compare from beginning
		return info, nil
	}

	// "latest" and specific versions are not valid for implicit modules
	return nil, fmt.Errorf("implicit-versioned module %q does not support version %q; only 'unreleased' is valid", module, versionStr)
}

// ResolveVersionWithValidation resolves a version and validates that git tags exist.
// This provides better error messages when tags are missing due to shallow clones.
// deps provides injectable dependencies; pass nil to use zero-value defaults.
func ResolveVersionWithValidation(deps *ReportDeps, workspaceRoot, module, versionStr string) (*VersionInfo, error) {
	info, err := ResolveVersion(workspaceRoot, module, versionStr)
	if err != nil {
		return nil, err
	}

	// For unreleased, we don't strictly need the tag to exist
	if info.IsUnreleased {
		return info, nil
	}

	// Validate that the resolved git tag exists
	repo, err := deps.getVersionResolverRepo(workspaceRoot)
	if err != nil {
		return nil, err
	}

	if info.GitTag != "" {
		exists, err := repo.TagExists(info.GitTag)
		if err != nil {
			return nil, fmt.Errorf("failed to check tag existence: %w", err)
		}
		if !exists {
			// List available tags to help diagnose
			tags, tagErr := repo.TagsMatching(module + "/*")
			tagList := "none found"
			if tagErr != nil {
				tagList = "error listing tags"
			}
			if len(tags) > 0 {
				if len(tags) > 5 {
					tags = tags[:5]
				}
				tagList = fmt.Sprintf("%v", tags)
			}
			return nil, fmt.Errorf(
				"git tag %q not found in local repository\n"+
					"  Changelog version: %s\n"+
					"  Available %s tags: %s\n\n"+
					"Possible causes:\n"+
					"  - Shallow clone without tags (CI: add fetch-depth: 0 and fetch-tags: true)\n"+
					"  - Tag not pushed to remote (run: git push origin %s)\n"+
					"  - Version in changelog but release not completed",
				info.GitTag, info.VersionNumber, module, tagList, info.GitTag)
		}
	}

	return info, nil
}
