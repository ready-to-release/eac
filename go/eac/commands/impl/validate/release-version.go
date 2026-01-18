// Command: validate release-version
// Short: Validate release version format
// Long: Validates that a version string is a valid semantic version.
// Long:
// Long: Checks:
// Long:   - Must be valid semver format: MAJOR.MINOR.PATCH (e.g., 1.2.3)
// Long:   - Must not have 'v' prefix (e.g., v1.2.3 is rejected)
// Long:   - Major, minor, and patch must be non-negative integers
// Long:   - No leading zeros allowed (except for 0 itself)
// Long:
// Long: Expected Output:
// Long:   Exit code 0 if valid semver format, exit code 1 if invalid.
// Long:   Error message displayed for invalid versions (v prefix, wrong format, leading zeros).
// Long:   No output on success (silent success).
// Long:
// Long: Examples:
// Long:   validate release-version 1.2.3        # Valid
// Long:   validate release-version 0.1.0        # Valid
// Long:   validate release-version v1.2.3       # Invalid: has 'v' prefix
// Long:   validate release-version 1.2          # Invalid: missing patch
// Long:   validate release-version 01.2.3       # Invalid: leading zero
package validate

import (
	"os"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

// Semver regex: matches MAJOR.MINOR.PATCH where each is a non-negative integer
// without leading zeros (except for 0 itself).
var semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

func init() {
	registry.Register(ValidateReleaseVersion)
}

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
	log.Info("Usage: r2r validate release-version <version>")
	log.Info("")
	log.Info("Validates that a version string is a valid semantic version.")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - Must be valid semver format: MAJOR.MINOR.PATCH")
	log.Info("  - Must not have 'v' prefix")
	log.Info("  - No leading zeros (except 0 itself)")
	log.Info("")
	log.Info("Examples:")
	log.Info("  r2r validate release-version 1.2.3      # Valid")
	log.Info("  r2r validate release-version 0.1.0      # Valid")
	log.Info("  r2r validate release-version v1.2.3     # Invalid: 'v' prefix")
	log.Info("  r2r validate release-version 1.2        # Invalid: missing patch")
}
