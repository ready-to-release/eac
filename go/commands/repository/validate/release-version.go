package validate

import (
	"context"
	"os"
	"regexp"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
)

type validateReleaseVersionCommand struct{}

var _ core.SimpleCommandPort = (*validateReleaseVersionCommand)(nil)

func (c *validateReleaseVersionCommand) Name() string { return "validate release-version" }

func (c *validateReleaseVersionCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-release-version",
		Short:         "Validate release version format",
		Long:          "Validates that a version string is a valid semantic version.\n\nChecks:\n  - Must be valid semver format: MAJOR.MINOR.PATCH (e.g., 1.2.3)\n  - Must not have 'v' prefix (e.g., v1.2.3 is rejected)\n  - Major, minor, and patch must be non-negative integers\n  - No leading zeros allowed (except for 0 itself)\n\nExpected Output:\n  Exit code 0 if valid semver format, exit code 1 if invalid.\n  Error message displayed for invalid versions (v prefix, wrong format, leading zeros).\n  No output on success (silent success).\n\nExamples:\n  validate release-version 1.2.3        # Valid\n  validate release-version 0.1.0        # Valid\n  validate release-version v1.2.3       # Invalid: has 'v' prefix\n  validate release-version 1.2          # Invalid: missing patch\n  validate release-version 01.2.3       # Invalid: leading zero",
	}
}

func (c *validateReleaseVersionCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateReleaseVersion()
}

// Semver regex: matches MAJOR.MINOR.PATCH where each is a non-negative integer
// without leading zeros (except for 0 itself).
var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

// ValidateReleaseVersion validates a release version string.
func ValidateReleaseVersion() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip binary, "validate", "release-version"

	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printReleaseVersionUsage()
		if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
			return 0
		}
		return 1
	}

	version := args[0]

	// Check for 'v' prefix
	if strings.HasPrefix(version, "v") || strings.HasPrefix(version, "V") {
		log.Errorf("Error: version should not have 'v' prefix: %s", version)
		log.Errorf("Use: %s", strings.TrimPrefix(strings.TrimPrefix(version, "v"), "V"))
		return 1
	}

	// Validate semver format
	if !semverRegex.MatchString(version) {
		log.Errorf("Error: invalid semver format: %s", version)
		log.Errorf("Expected format: MAJOR.MINOR.PATCH (e.g., 1.2.3)")
		return 1
	}

	// Version is valid - exit silently with code 0
	return 0
}

func printReleaseVersionUsage() {
	log.Info("Validate release version format")
	log.Info("")
	log.Info("Usage: clie validate release-version <version>")
	log.Info("")
	log.Info("Validates that a version string is a valid semantic version.")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - Must be valid semver format: MAJOR.MINOR.PATCH")
	log.Info("  - Must not have 'v' prefix")
	log.Info("  - No leading zeros (except 0 itself)")
	log.Info("")
	log.Info("Examples:")
	log.Info("  clie validate release-version 1.2.3      # Valid")
	log.Info("  clie validate release-version 0.1.0      # Valid")
	log.Info("  clie validate release-version v1.2.3     # Invalid: 'v' prefix")
	log.Info("  clie validate release-version 1.2        # Invalid: missing patch")
}
