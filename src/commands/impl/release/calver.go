// Command: release calver
// Short: Generate a calver tag for a module
// Long: Generates a calendar-versioned (calver) tag in the format prefix/YYYY.MM.DD.
// Long:
// Long: If a tag for the current date already exists, appends an incrementing suffix
// Long: (e.g., docs/2025.01.15.1, docs/2025.01.15.2).
// Long:
// Long: By default, only outputs the tag name. Use --create to create the git tag.
// Long:
// Long: Examples:
// Long:   release calver docs                    # Output: docs/2025.01.15
// Long:   release calver docs --create           # Create the tag locally
// Long:   release calver docs --create --push    # Create and push the tag
// Flag.create: type=bool, usage=Create the git tag (default: false, just output tag name)
// Flag.push: type=bool, usage=Push the tag to remote after creation (requires --create)
// Flag.dry-run: type=bool, usage=Show what would be done without creating/pushing
// HasSideEffects: true
package release

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/registry"
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
		fmt.Fprintln(os.Stderr, "Error: prefix required (e.g., 'docs')")
		fmt.Fprintln(os.Stderr, "Usage: release calver <prefix> [--create] [--push] [--dry-run]")
		return 1
	}

	// Generate base tag with current UTC date
	baseTag := fmt.Sprintf("%s/%s", prefix, time.Now().UTC().Format("2006.01.02"))

	// Find available tag (with suffix if needed)
	tag := findAvailableCalverTag(baseTag)

	// If not creating, just output the tag name
	if !create {
		fmt.Println(tag)
		return 0
	}

	// Creating the tag
	if dryRun {
		fmt.Printf("[DRY RUN] Would create tag: %s\n", tag)
		if push {
			fmt.Printf("[DRY RUN] Would push tag to remote\n")
		}
		return 0
	}

	// Create the git tag
	if err := createGitTag(tag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create tag: %v\n", err)
		return 1
	}
	fmt.Printf("Created tag: %s\n", tag)

	// Push if requested
	if push {
		if err := pushGitTag(tag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to push tag: %v\n", err)
			return 1
		}
		fmt.Printf("Pushed tag: %s\n", tag)
	}

	return 0
}

// findAvailableCalverTag finds an available calver tag, appending suffix if base exists
func findAvailableCalverTag(baseTag string) string {
	tag := baseTag
	suffix := 1

	for tagExists(tag) {
		tag = fmt.Sprintf("%s.%d", baseTag, suffix)
		suffix++
	}

	return tag
}

// getExistingCalverTags returns all existing tags matching the base pattern
func getExistingCalverTags(baseTag string) []string {
	// List all tags matching the pattern
	cmd := exec.Command("git", "tag", "-l", baseTag+"*")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var tags []string
	for _, line := range lines {
		if line != "" {
			tags = append(tags, line)
		}
	}
	return tags
}

// findHighestSuffix finds the highest suffix number from existing tags
func findHighestSuffix(baseTag string, existingTags []string) int {
	highest := 0

	for _, tag := range existingTags {
		if tag == baseTag {
			// Base tag exists, so we need at least .1
			if highest < 1 {
				highest = 0 // Will become 1 when we add 1
			}
			continue
		}

		// Check for suffix pattern: baseTag.N
		if strings.HasPrefix(tag, baseTag+".") {
			suffixStr := strings.TrimPrefix(tag, baseTag+".")
			if num, err := strconv.Atoi(suffixStr); err == nil && num > highest {
				highest = num
			}
		}
	}

	return highest
}
