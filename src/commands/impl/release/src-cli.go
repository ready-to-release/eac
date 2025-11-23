// Command: release src-cli
// Description: Create a release tag for src-cli module
// Short: Create a git tag for releasing src-cli using semver format
// Long: Creates a git tag in the format 'src-cli/x.y.z' to trigger the release workflow.
// Long: The tag follows semantic versioning (semver) and will automatically trigger
// Long: the GitHub Actions workflow to build and release binaries for multiple platforms.
// Long: The version must follow semver format (x.y.z) where x, y, z are non-negative integers.
// Long: Example: release src-cli 1.0.0
// HasSideEffects: true
package release

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

func init() {
	registry.Register(ReleaseSrcCli)
}

func ReleaseSrcCli() int {
	fs := flag.NewFlagSet("release src-cli", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be done without actually creating the tag")
	push := fs.Bool("push", true, "Push the tag to the remote repository after creation")

	// Parse flags from remaining args
	// When called via src/commands dispatcher, os.Args contains: [binary, "release", "src-cli", ...args]
	args := os.Args[1:]

	// Remove "release" and "src-cli" from args to get actual command arguments
	for i, arg := range args {
		if arg == "release" {
			// Found "release", skip it and "src-cli" (next arg)
			if i+1 < len(args) && args[i+1] == "src-cli" {
				args = args[i+2:]
				break
			}
		}
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Get version from remaining args
	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		fmt.Fprintf(os.Stderr, "Error: version required\n")
		fmt.Fprintf(os.Stderr, "Usage: release src-cli [--dry-run] [--push=true|false] <version>\n")
		fmt.Fprintf(os.Stderr, "Example: release src-cli --dry-run 1.0.0\n")
		fmt.Fprintf(os.Stderr, "Note: Flags must come before the version number\n")
		return 1
	}

	version := remainingArgs[0]

	// Validate module exists
	if err := validateModule("src-cli"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Validate semver format
	if err := validateSemver(version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Build tag name
	tagName := buildTagName("src-cli", version)

	// Check if tag already exists
	if tagExists(tagName) {
		fmt.Fprintf(os.Stderr, "Error: tag '%s' already exists\n", tagName)
		return 1
	}

	if *dryRun {
		fmt.Printf("[DRY RUN] Would create tag: %s\n", tagName)
		if *push {
			fmt.Printf("[DRY RUN] Would push tag to remote\n")
		}
		return 0
	}

	// Create the git tag
	if err := createGitTag(tagName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create tag: %v\n", err)
		return 1
	}

	fmt.Printf("Created tag: %s\n", tagName)

	// Push the tag if requested
	if *push {
		if err := pushGitTag(tagName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to push tag: %v\n", err)
			return 1
		}
		fmt.Printf("Pushed tag to remote: %s\n", tagName)
	}

	fmt.Printf("\n✅ Release %s created successfully\n", tagName)
	if *push {
		fmt.Printf("The release workflow will be triggered automatically.\n")
	} else {
		fmt.Printf("Push the tag to trigger the release workflow: git push origin %s\n", tagName)
	}

	return 0
}

// validateModule checks if a module exists in the repository
func validateModule(moduleName string) error {
	// Try to load the module contracts
	registry, err := modules.LoadFromWorkspace("", "0.1.0")
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

// validateSemver validates that a version string follows semantic versioning (x.y.z)
func validateSemver(version string) error {
	// Remove leading 'v' if present - we don't want it
	if strings.HasPrefix(version, "v") {
		return fmt.Errorf("invalid semver format: remove leading 'v'. Use x.y.z (e.g., 1.0.0)")
	}

	// Regex pattern for semantic versioning: x.y.z where x, y, z are non-negative integers
	semverPattern := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

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

// buildTagName constructs a tag name in the format "module/version"
func buildTagName(moduleName, version string) string {
	return fmt.Sprintf("%s/%s", moduleName, version)
}

// tagExists checks if a git tag already exists
func tagExists(tagName string) bool {
	cmd := exec.Command("git", "tag", "-l", tagName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If output is not empty, tag exists
	return strings.TrimSpace(string(output)) != ""
}

// createGitTag creates a git tag with the given name
func createGitTag(tagName string) error {
	cmd := exec.Command("git", "tag", "-a", tagName, "-m", fmt.Sprintf("Release %s", tagName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// pushGitTag pushes a git tag to the remote repository
func pushGitTag(tagName string) error {
	cmd := exec.Command("git", "push", "origin", tagName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
