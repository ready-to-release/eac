// Command: release check-exists
// Short: Check if a release/tag already exists
// Long: Checks GitHub for an existing release with the given tag.
// Long:
// Long: Exit codes:
// Long:   0 - Release does NOT exist (safe to create)
// Long:   1 - Release EXISTS or error
// Long:
// Long: Output formats:
// Long:   default: Human readable message
// Long:   --format shell: EXISTS="true" TAG="..." ERROR="..."
// Long:
// Long: Example:
// Long:   release check-exists --tag r2r-cli/1.0.0
// Long:   eval $(release check-exists --tag r2r-cli/1.0.0 --format shell)
// Flag.tag: type=string, usage=Tag name to check (required)
// Flag.format: type=string, usage=Output format (default, shell)
package release

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(ReleaseCheckExists)
}

func ReleaseCheckExists() int {
	// Parse flags
	tag := ""
	format := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--tag" && i+1 < len(os.Args):
			tag = os.Args[i+1]
			i++
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		case !strings.HasPrefix(arg, "--") && tag == "":
			tag = arg
		}
	}

	if tag == "" {
		fmt.Fprintln(os.Stderr, "Error: --tag is required")
		fmt.Fprintln(os.Stderr, "Usage: release check-exists --tag <tag>")
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		if format == "shell" {
			fmt.Printf("EXISTS=\"\"\n")
			fmt.Printf("TAG=\"%s\"\n", tag)
			fmt.Printf("ERROR=\"failed to find repository\"\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return 1
	}

	// Check if release exists
	exists, err := checkReleaseExistsRemote(tag, workspaceRoot)

	if format == "shell" {
		fmt.Printf("EXISTS=\"%t\"\n", exists)
		fmt.Printf("TAG=\"%s\"\n", tag)
		if err != nil {
			fmt.Printf("ERROR=\"%s\"\n", err.Error())
		} else {
			fmt.Printf("ERROR=\"\"\n")
		}
	} else {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking release: %v\n", err)
			return 1
		}
		if exists {
			fmt.Printf("Release exists for tag: %s\n", tag)
			fmt.Println("")
			fmt.Println("A GitHub release has already been created for this version.")
			fmt.Println("To re-release:")
			fmt.Printf("  1. Delete release: gh release delete %s\n", tag)
			fmt.Printf("  2. Delete tag: git push --delete origin %s\n", tag)
			fmt.Println("  3. Re-run workflow")
		} else {
			fmt.Printf("No existing release for %s\n", tag)
		}
	}

	// Exit 0 if release does NOT exist (safe to proceed)
	// Exit 1 if release EXISTS (should not proceed)
	if exists {
		return 1
	}
	return 0
}

func checkReleaseExistsRemote(tag, workspaceRoot string) (bool, error) {
	_, exitCode, err := ghexec.RunCombined(context.Background(), workspaceRoot, "release", "view", tag)
	if err != nil {
		return false, err
	}

	// Exit code 0 = release exists, non-zero = doesn't exist
	return exitCode == 0, nil
}
