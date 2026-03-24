package release

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
)

type releaseClieCommand struct{}

var _ core.SimpleCommandPort = (*releaseClieCommand)(nil)

func (c *releaseClieCommand) Name() string { return "release clie" }

func (c *releaseClieCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-clie",
		Short:         "Create a git tag for releasing clie using semver format",
		Long: "Creates a git tag in the format 'clie/x.y.z' to trigger the release workflow.\nThe tag follows semantic versioning (semver) and will automatically trigger\nthe GitHub Actions workflow to build and release binaries for multiple platforms.\nThe version must follow semver format (x.y.z) where x, y, z are non-negative integers.\nIMPORTANT: This command requires --tag-direct flag to prevent accidental releases.\nThe preferred flow is: release this → commit → push → workflow creates tag.\nUse --tag-direct only when you need to tag directly from devbox.",
		Notes: "Expected Output:\n  - Git tag created in format clie/x.y.z\n  - Tag triggers the release workflow to build and publish binaries\n\nExample: release clie --tag-direct 1.0.0",
	}
}

func (c *releaseClieCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseSrcCli()
}

var reSemverStrict = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

func ReleaseSrcCli() int {
	s, exitCode := newReleaseScaffold()
	if s == nil {
		return exitCode
	}

	fs := flag.NewFlagSet("release clie", flag.ExitOnError)
	tagDirect := fs.Bool("tag-direct", false, "Required flag to confirm direct tagging from devbox")
	dryRun := fs.Bool("dry-run", false, "Show what would be done without actually creating the tag")
	push := fs.Bool("push", true, "Push the tag to the remote repository after creation")

	// Parse flags from remaining args
	// When called via go/cli/eac dispatcher, os.Args contains: [binary, "release", "clie", ...args]
	args := os.Args[1:]

	// Remove "release" and "clie" from args to get actual command arguments
	for i, arg := range args {
		if arg == "release" {
			// Found "release", skip it and "clie" (next arg)
			if i+1 < len(args) && args[i+1] == "clie" {
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
		log.Errorf("Usage: release clie --tag-direct [--dry-run] [--push=true|false] <version>")
		log.Errorf("Example: release clie --tag-direct 1.0.0")
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
		log.Errorf("To tag directly: release clie --tag-direct %s", version)
		return 1
	}

	// Validate module exists
	if err := validateModule("clie"); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Validate semver format
	if err := validateSemver(version); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Build tag name
	tagName := buildTagName("clie", version)

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

// validateModule checks if a module exists in the repository.
func validateModule(moduleName string) error {
	// Try to load the module contracts
	registry, err := modules.LoadFromWorkspace("")
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}

	// Check if module exists
	module, exists := registry.Get(moduleName)
	if !exists {
		return fmt.Errorf("module '%s' not found in repository", moduleName)
	}

	// Verify module is not nil
	if module == nil {
		return fmt.Errorf("module '%s' exists but contract is invalid", moduleName)
	}

	return nil
}

// validateSemver validates that a version string follows semantic versioning (x.y.z).
func validateSemver(version string) error {
	// Remove leading 'v' if present - we don't want it
	if strings.HasPrefix(version, "v") {
		return fmt.Errorf("invalid semver format: remove leading 'v'. Use x.y.z (e.g., 1.0.0)")
	}

	semverPattern := reSemverStrict

	if !semverPattern.MatchString(version) {
		return fmt.Errorf("invalid semver format. Use x.y.z where x, y, z are non-negative integers (e.g., 1.0.0)")
	}

	// Additional validation: ensure all components are valid integers
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid semver format. Use x.y.z (e.g., 1.0.0)")
	}

	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 {
			componentNames := []string{"major", "minor", "patch"}
			return fmt.Errorf("invalid semver format: %s version must be a non-negative integer", componentNames[i])
		}
	}

	return nil
}

// buildTagName constructs a tag name in the format "module/version".
func buildTagName(moduleName, version string) string {
	return fmt.Sprintf("%s/%s", moduleName, version)
}

// tagExists checks if a git tag already exists.
func tagExists(tagName string) bool {
	ts := tool.GlobalToolSystem()
	if ts == nil {
		return false
	}
	output, err := ts.RunTool(context.Background(), "git", ".", "tag", "-l", tagName)
	if err != nil {
		return false
	}

	// If output is not empty, tag exists
	return strings.TrimSpace(string(output)) != ""
}

// createGitTag creates a git tag with the given name.
func createGitTag(tagName string) error {
	ts := tool.GlobalToolSystem()
	if ts == nil {
		return fmt.Errorf("tool system not initialized")
	}
	_, err := ts.RunTool(context.Background(), "git", ".", "tag", "-a", tagName, "-m", fmt.Sprintf("Release %s", tagName))
	return err
}

// pushGitTag pushes a git tag to the remote repository.
func pushGitTag(tagName string) error {
	ts := tool.GlobalToolSystem()
	if ts == nil {
		return fmt.Errorf("tool system not initialized")
	}
	_, err := ts.RunTool(context.Background(), "git", ".", "push", "origin", tagName)
	return err
}
