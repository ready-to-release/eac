// Command: release get-version
// Short: Extract latest version from changelog
// Long: Reads the CHANGELOG.md file for a module and outputs the latest version.
// Long:
// Long: This command is designed for use in CI/CD pipelines where the changelog
// Long: is the source of truth for versioning.
// Long:
// Long: Expected Output:
// Long:   - Version string (e.g., 0.0.14) by default
// Long:   - Tag format (e.g., r2r-cli/0.0.14) if --tag flag is specified
// Long:
// Long: Examples:
// Long:   release get-version r2r-cli              # Output: 0.0.14
// Long:   release get-version docs                 # Output: 2025.12.01
// Long:   release get-version r2r-cli --tag        # Output: r2r-cli/0.0.14
// Long:   release get-version r2r-cli --json       # Output JSON format
// Flag.tag: type=bool, usage=Output as git tag format (module/version)
// Flag.json: type=bool, usage=Output in JSON format
// Flag.path: type=string, usage=Override changelog path (default: release/<module>/CHANGELOG.md)
package release

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/changelog"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

var getVersionLog = logging.C("release")

func init() {
	registry.Register(ReleaseGetVersion)
}

// VersionInfo contains version information for JSON output
type VersionInfo struct {
	Module      string `json:"module"`
	Version     string `json:"version"`
	Tag         string `json:"tag"`
	Date        string `json:"date,omitempty"`
	VersionType string `json:"version_type"`
}

var getVersionFlags = []flags.FlagDefinition{
	{Name: "--tag", HasValue: false, ValueType: "bool"},
	{Name: "--json", HasValue: false, ValueType: "bool"},
	{Name: "--path", HasValue: true, ValueType: "string"},
}

func ReleaseGetVersion() int {
	// Validate flags before parsing
	if err := flags.ValidateFlags(os.Args[3:], getVersionFlags); err != nil {
		getVersionLog.Errorf("%v", err)
		return 1
	}

	// Parse flags
	module := ""
	asTag := false
	asJSON := false
	customPath := ""

	args := os.Args[3:] // Skip binary, "release", "get-version"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--tag":
			asTag = true
		case arg == "--json":
			asJSON = true
		case strings.HasPrefix(arg, "--path="):
			customPath = strings.TrimPrefix(arg, "--path=")
		case arg == "--path" && i+1 < len(args):
			i++
			customPath = args[i]
		default:
			if !strings.HasPrefix(arg, "--") && module == "" {
				module = arg
			}
		}
	}

	if module == "" {
		getVersionLog.Error("module moniker required")
		getVersionLog.Info("Usage: release get-version <module> [--tag] [--json]")
		return 1
	}

	// Determine changelog path
	changelogPath := customPath
	if changelogPath == "" {
		changelogPath = paths.ChangelogPath(".", module)
	}

	// Check if file exists
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		getVersionLog.Errorf("changelog not found at %s", changelogPath)
		return 1
	}

	// Parse changelog
	cl, err := changelog.Parse(changelogPath)
	if err != nil {
		getVersionLog.Errorf("failed to parse changelog: %v", err)
		return 1
	}

	// Get latest version
	latestVersion := cl.LatestVersion()
	if latestVersion == nil {
		getVersionLog.Error("no versions found in changelog")
		return 1
	}

	// Prepare output
	version := latestVersion.Number
	tag := fmt.Sprintf("%s/%s", module, version)
	versionType := "semver"
	if cl.VersionType == changelog.Calver {
		versionType = "calver"
	}

	// Output based on format
	if asJSON {
		info := VersionInfo{
			Module:      module,
			Version:     version,
			Tag:         tag,
			VersionType: versionType,
		}
		if !latestVersion.Date.IsZero() {
			info.Date = latestVersion.Date.Format("2006-01-02")
		}

		output, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			getVersionLog.Errorf("failed to marshal JSON: %v", err)
			return 1
		}
		getVersionLog.Info(string(output))
	} else if asTag {
		getVersionLog.Info(tag)
	} else {
		getVersionLog.Info(version)
	}

	return 0
}
