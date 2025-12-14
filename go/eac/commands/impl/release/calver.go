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

	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ReleaseCalver)
}

func ReleaseCalver() int {
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
	now := time.Now().UTC()
	tag := fmt.Sprintf("%s/%s.%s", prefix, now.Format("2006.0102"), now.Format("1504"))

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
