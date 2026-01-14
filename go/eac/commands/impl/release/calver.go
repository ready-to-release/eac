// Command: release generate-module-calver
// Short: Generate a calver tag for a module
// Long: Generates a calendar-versioned (calver) tag in the format prefix/YYYY.MMDD.HHMM.
// Long:
// Long: Format follows CD Model CalVer specification:
// Long:   - YYYY: Four-digit year
// Long:   - MMDD: Month and day (packed, no separator)
// Long:   - HHMM: Hour and minute in UTC (ensures uniqueness)
// Long:   - Patch: Omitted (inferred as 0 for main branch commits)
// Long:
// Long: By default, only outputs the tag name. Use --create to create the git tag.
// Long:
// Long: Expected Output:
// Long:   - Tag name in format prefix/YYYY.MMDD.HHMM (default behavior)
// Long:   - Git tag created if --create flag is specified
// Long:
// Long: Examples:
// Long:   release generate-module-calver docs                    # Output: docs/2025.1214.1630
// Long:   release generate-module-calver docs --create           # Create the tag locally
// Long:   release generate-module-calver docs --create --push    # Create and push the tag
// Flag.create: type=bool, usage=Create the git tag (default: false, just output tag name)
// Flag.push: type=bool, usage=Push the tag to remote after creation (requires --create)
// Flag.dry-run: type=bool, usage=Show what would be done without creating/pushing
package release

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ReleaseCalver)
}

func ReleaseCalver() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags manually (consistent with other commands)
	prefix := ""
	create := false
	push := false
	dryRun := false

	args := os.Args[3:] // Skip binary, "release", "calver"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--create":
			create = true
		case "--push":
			push = true
		case "--dry-run":
			dryRun = true
		default:
			if !strings.HasPrefix(arg, "--") && prefix == "" {
				prefix = arg
			}
		}
	}

	if prefix == "" {
		log.Errorf("Error: prefix required (e.g., 'docs')")
		log.Errorf("Usage: release generate-module-calver <prefix> [--create] [--push] [--dry-run]")
		return 1
	}

	// Generate tag with current UTC timestamp: prefix/YYYY.MMDD.HHMM
	tag := GenerateCalverTag(prefix, time.Now().UTC())

	// If not creating, just output the tag name
	if !create {
		log.Info(tag)
		return 0
	}

	// Creating the tag
	if dryRun {
		log.Infof("[DRY RUN] Would create tag: %s", tag)
		if push {
			log.Infof("[DRY RUN] Would push tag to remote")
		}
		return 0
	}

	// Create the git tag
	if err := createGitTag(tag); err != nil {
		log.Errorf("Error: failed to create tag: %v", err)
		return 1
	}
	log.Infof("Created tag: %s", tag)

	// Push if requested
	if push {
		if err := pushGitTag(tag); err != nil {
			log.Errorf("Error: failed to push tag: %v", err)
			return 1
		}
		log.Infof("Pushed tag: %s", tag)
	}

	return 0
}

// GenerateCalverTag generates a calendar-versioned tag in format prefix/YYYY.MMDD.HHMM
func GenerateCalverTag(prefix string, t time.Time) string {
	return fmt.Sprintf("%s/%s.%s", prefix, t.Format("2006.0102"), t.Format("1504"))
}

// ParseCalverTag parses a calver tag and returns the prefix and timestamp
// Returns prefix, time, and error if parsing fails
func ParseCalverTag(tag string) (string, time.Time, error) {
	parts := strings.SplitN(tag, "/", 2)
	if len(parts) != 2 {
		return "", time.Time{}, fmt.Errorf("invalid calver tag format: %s", tag)
	}

	prefix := parts[0]
	version := parts[1]

	// Parse YYYY.MMDD.HHMM format
	versionParts := strings.Split(version, ".")
	if len(versionParts) != 3 {
		return "", time.Time{}, fmt.Errorf("invalid calver version format: %s", version)
	}

	// Reconstruct as parseable time string: YYYY-MM-DD HH:MM
	year := versionParts[0]
	mmdd := versionParts[1]
	hhmm := versionParts[2]

	if len(mmdd) != 4 || len(hhmm) != 4 {
		return "", time.Time{}, fmt.Errorf("invalid calver components: %s", version)
	}

	timeStr := fmt.Sprintf("%s-%s-%s %s:%s", year, mmdd[:2], mmdd[2:], hhmm[:2], hhmm[2:])
	t, err := time.Parse("2006-01-02 15:04", timeStr)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse calver time: %w", err)
	}

	return prefix, t.UTC(), nil
}

// IsValidCalverVersion checks if a version string matches calver format YYYY.MMDD.HHMM
func IsValidCalverVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	// Check year: 4 digits, reasonable range
	if len(parts[0]) != 4 {
		return false
	}

	// Check MMDD: 4 digits
	if len(parts[1]) != 4 {
		return false
	}

	// Check HHMM: 4 digits
	if len(parts[2]) != 4 {
		return false
	}

	// Try parsing to validate
	_, _, err := ParseCalverTag("test/" + version)
	return err == nil
}
