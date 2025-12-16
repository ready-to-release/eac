package reports

import (
	"fmt"

	"github.com/ready-to-release/eac/go/eac/core/changelog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// VersionInfo represents resolved version information for a module
type VersionInfo struct {
	Module         string // Module moniker
	VersionNumber  string // Actual version number or "Unreleased"
	GitTag         string // Git tag for this version (e.g., "ext-eac/0.0.7")
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

	_, ok := cfg.Repository.GetModule(module)
	if !ok {
		return nil, fmt.Errorf("module not found: %s", module)
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
			return nil, fmt.Errorf("no released versions found")
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
