package release

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/changelog"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

type releaseGetVersionCommand struct{}

var _ core.SimpleCommandPort = (*releaseGetVersionCommand)(nil)

func (c *releaseGetVersionCommand) Name() string { return "release get-version" }

func (c *releaseGetVersionCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-get-version",
		Short:         "Extract latest version from changelog",
		Long:          "Reads the CHANGELOG.md file for a module and outputs the latest version.\n\nThis command is designed for use in CI/CD pipelines where the changelog\nis the source of truth for versioning.\n\nExpected Output:\n  - Version string (e.g., 0.0.14) by default\n  - Tag format (e.g., clie/0.0.14) if --tag flag is specified\n\nExamples:\n  release get-version clie              # Output: 0.0.14\n  release get-version docs                 # Output: 2025.12.01\n  release get-version clie --tag        # Output: clie/0.0.14\n  release get-version clie --json       # Output JSON format",
		Flags: []core.FlagSpec{
			{Name: "tag", Type: "bool", Usage: "Output as git tag format (module/version)"},
			{Name: "json", Type: "bool", Usage: "Output in JSON format"},
			{Name: "path", Type: "string", Usage: "Override changelog path (default: release/<module>/CHANGELOG.md)"},
		},
	}
}

func (c *releaseGetVersionCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseGetVersion()
}

var getVersionLog = logging.C("release")

// VersionInfo contains version information for JSON output.
type VersionInfo struct {
	Module      string `json:"module"`
	Version     string `json:"version"`
	Tag         string `json:"tag"`
	Date        string `json:"date,omitempty"`
	VersionType string `json:"version_type"`
}

func ReleaseGetVersion() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
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
	// NOTE: Output to stdout (fmt.Print), not logger, so it can be captured by shell scripts
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
		fmt.Println(string(output))
	} else if asTag {
		fmt.Println(tag)
	} else {
		fmt.Println(version)
	}

	return 0
}
