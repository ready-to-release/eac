// Command: release ext-eac
// Description: Create a release tag for ext-eac module
// Short: Create a git tag for releasing ext-eac using semver format
// Long: Creates a git tag in the format 'ext-eac/x.y.z' to trigger the release workflow.
// Long: The tag follows semantic versioning (semver) and will automatically trigger
// Long: the GitHub Actions workflow to retag and publish the container image.
// Long: The version must follow semver format (x.y.z) where x, y, z are non-negative integers.
// Long: IMPORTANT: This command requires --tag-direct flag to prevent accidental releases.
// Long: The preferred flow is: release this → commit → push → workflow creates tag.
// Long: Use --tag-direct only when you need to tag directly from devbox.
// Long:
// Long: Expected Output:
// Long:   - Git tag created in format ext-eac/x.y.z
// Long:   - Tag triggers the release workflow to retag and publish container image
// Long:
// Long: Example: release ext-eac --tag-direct 0.0.7
package release

import (
	"flag"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ReleaseExtEac)
}


func ReleaseExtEac() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	fs := flag.NewFlagSet("release ext-eac", flag.ExitOnError)
	tagDirect := fs.Bool("tag-direct", false, "Required flag to confirm direct tagging from devbox")
	dryRun := fs.Bool("dry-run", false, "Show what would be done without actually creating the tag")
	push := fs.Bool("push", true, "Push the tag to the remote repository after creation")

	// Parse flags from remaining args
	// When called via go/eac/commands dispatcher, os.Args contains: [binary, "release", "ext-eac", ...args]
	args := os.Args[1:]

	// Remove "release" and "ext-eac" from args to get actual command arguments
	for i, arg := range args {
		if arg == "release" {
			// Found "release", skip it and "ext-eac" (next arg)
			if i+1 < len(args) && args[i+1] == "ext-eac" {
				args = args[i+2:]
				break
			}
		}
	}

	if err := fs.Parse(args); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Get version from remaining args
	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		log.Errorf("Error: version required")
		log.Errorf("Usage: release ext-eac --tag-direct [--dry-run] [--push=true|false] <version>")
		log.Errorf("Example: release ext-eac --tag-direct 0.0.7")
		log.Errorf("Note: Flags must come before the version number")
		return 1
	}

	version := remainingArgs[0]

	// Require --tag-direct flag to prevent accidental releases
	if !*tagDirect && !*dryRun {
		log.Errorf("Error: --tag-direct flag required")
		log.Errorf("")
		log.Errorf("Direct tagging from devbox requires explicit confirmation.")
		log.Errorf("Preferred flow: release this → commit → push → workflow dispatch")
		log.Errorf("")
		log.Errorf("To tag directly: release ext-eac --tag-direct %s", version)
		return 1
	}

	// Validate module exists
	if err := validateModule("ext-eac"); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Validate semver format
	if err := validateSemver(version); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Build tag name
	tagName := buildTagName("ext-eac", version)

	// Check if tag already exists
	if tagExists(tagName) {
		log.Errorf("Error: tag '%s' already exists", tagName)
		return 1
	}

	if *dryRun {
		log.Infof("[DRY RUN] Would create tag: %s", tagName)
		if *push {
			log.Infof("[DRY RUN] Would push tag to remote")
		}
		return 0
	}

	// Create the git tag
	if err := createGitTag(tagName); err != nil {
		log.Errorf("Error: failed to create tag: %v", err)
		return 1
	}

	log.Infof("Created tag: %s", tagName)

	// Push the tag if requested
	if *push {
		if err := pushGitTag(tagName); err != nil {
			log.Errorf("Error: failed to push tag: %v", err)
			return 1
		}
		log.Infof("Pushed tag to remote: %s", tagName)
	}

	log.Infof("")
	log.Infof("✅ Release %s created successfully", tagName)
	if *push {
		log.Infof("The release workflow will be triggered automatically.")
	} else {
		log.Infof("Push the tag to trigger the release workflow: git push origin %s", tagName)
	}

	return 0
}
