package release

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

type releaseGetModuleCalverCommand struct{}

var _ core.SimpleCommandPort = (*releaseGetModuleCalverCommand)(nil)

func (c *releaseGetModuleCalverCommand) Name() string { return "release get-module-calver" }

func (c *releaseGetModuleCalverCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-get-module-calver",
		Short:         "Generate a calver tag for a module",
		Long: "Generates a calendar-versioned (calver) tag in the format prefix/YYYY.MMDD.HHMM.\n\nFormat follows CD Model CalVer specification:\n  - YYYY: Four-digit year\n  - MMDD: Month and day (packed, no separator)\n  - HHMM: Hour and minute in UTC (ensures uniqueness)\n  - Patch: Omitted (inferred as 0 for main branch commits)\n\nBy default, only outputs the tag name. Use --create to create the git tag.",
		Notes: "Expected Output:\n  - Tag name in format prefix/YYYY.MMDD.HHMM (default behavior)\n  - Git tag created if --create flag is specified",
		Examples: []string{
			"eac release get-module-calver docs                  # Output: docs/2025.1214.1630",
			"eac release get-module-calver docs --create         # Create the tag locally",
			"eac release get-module-calver docs --create --push  # Create and push the tag",
		},
		Flags: []core.FlagSpec{
			{Name: "create", Type: "bool", Usage: "Create the git tag (default: false, just output tag name)"},
			{Name: "push", Type: "bool", Usage: "Push the tag to remote after creation (requires --create)"},
			{Name: "dry-run", Type: "bool", Usage: "Show what would be done without creating/pushing"},
		},
	}
}

func (c *releaseGetModuleCalverCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseCalver()
}

func ReleaseCalver() int {
	s, exitCode := newReleaseScaffold()
	if s == nil {
		return exitCode
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
		log.Errorf("Usage: release get-module-calver <prefix> [--create] [--push] [--dry-run]")
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

// GenerateCalverTag generates a calendar-versioned tag in format prefix/YYYY.MMDD.HHMM.
func GenerateCalverTag(prefix string, t time.Time) string {
	return fmt.Sprintf("%s/%s.%s", prefix, t.Format("2006.0102"), t.Format("1504"))
}

// ParseCalverTag parses a calver tag and returns the prefix and timestamp
// Returns prefix, time, and error if parsing fails.
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

// IsValidCalverVersion checks if a version string matches calver format YYYY.MMDD.HHMM.
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
