// Command: release cleanup
// Short: Clean up orphaned tags and partial releases after failure
// Long: Removes orphaned release artifacts when a release workflow fails.
// Long:
// Long: This command deletes:
// Long:   - Partial GitHub releases (if any exist for the tag)
// Long:   - Git tags from the remote repository
// Long:
// Long: Use this in release workflow cleanup jobs to prevent orphaned tags
// Long: from blocking future release attempts.
// Long:
// Long: Example:
// Long:   release cleanup --tag r2r-cli/1.0.0
// Long:   release cleanup --tag ext-eac/2.0.0
// Flag.tag: type=string, usage=Tag name to clean up (e.g., r2r-cli/1.0.0)
package release

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ReleaseCleanup)
}

func ReleaseCleanup() int {
	// Parse flags
	tagName := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--tag":
			if i+1 < len(os.Args) {
				tagName = os.Args[i+1]
				i++
			}
		}
	}

	if tagName == "" {
		log.Errorf("Error: --tag is required")
		return 1
	}

	log.Infof("Cleaning up failed release: %s", tagName)
	log.Infof("")

	// Delete partial release if it exists
	if releaseExists(tagName) {
		log.Infof("Deleting partial release...")
		if err := deleteRelease(tagName); err != nil {
			log.Errorf("Warning: failed to delete release: %v", err)
		} else {
			log.Infof("✓ Partial release deleted")
		}
	} else {
		log.Infof("No partial release found")
	}

	// Delete tag from remote
	log.Infof("")
	log.Infof("Deleting tag from remote...")
	if err := deleteTag(tagName); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Reference does not exist") {
			log.Infof("Tag not found or already deleted")
		} else {
			log.Errorf("Warning: failed to delete tag: %v", err)
		}
	} else {
		log.Infof("✓ Tag deleted: %s", tagName)
	}

	log.Infof("")
	log.Infof("Cleanup complete. You can re-run the release after fixing the issue.")

	return 0
}

// releaseExists checks if a GitHub release exists for the given tag
func releaseExists(tagName string) bool {
	cmd := exec.Command("gh", "release", "view", tagName)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

// deleteRelease deletes a GitHub release
func deleteRelease(tagName string) error {
	cmd := exec.Command("gh", "release", "delete", tagName, "--yes")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// deleteTag deletes a tag from the remote repository using gh api
func deleteTag(tagName string) error {
	// Get repository from gh
	repoCmd := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	repoOutput, err := repoCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}
	repo := strings.TrimSpace(string(repoOutput))

	// Delete tag via API
	apiPath := fmt.Sprintf("repos/%s/git/refs/tags/%s", repo, tagName)
	cmd := exec.Command("gh", "api", "--method", "DELETE", apiPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}
