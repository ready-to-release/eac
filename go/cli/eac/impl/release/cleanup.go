package release

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/ghexec"
)

type releaseCleanupCommand struct{}

var _ core.SimpleCommandPort = (*releaseCleanupCommand)(nil)

func (c *releaseCleanupCommand) Name() string { return "release cleanup" }

func (c *releaseCleanupCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-cleanup",
		Short:         "Clean up orphaned tags and partial releases after failure",
		Long:          "Removes orphaned release artifacts when a release workflow fails.\n\nThis command deletes:\n  - Partial GitHub releases (if any exist for the tag)\n  - Git tags from the remote repository\n\nUse this in release workflow cleanup jobs to prevent orphaned tags\nfrom blocking future release attempts.\n\nExpected Output:\n  - Deletion of partial GitHub releases (if any exist for the tag)\n  - Deletion of git tags from the remote repository\n\nExample:\n  release cleanup --tag clie/1.0.0\n  release cleanup --tag eac-ext/2.0.0",
		Flags: []core.FlagSpec{
			{Name: "tag", Type: "string", Usage: "Tag name to clean up (e.g., clie/1.0.0)"},
		},
	}
}

func (c *releaseCleanupCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseCleanup()
}

func ReleaseCleanup() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

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

// releaseExists checks if a GitHub release exists for the given tag.
func releaseExists(tagName string) bool {
	_, exitCode, err := ghexec.RunCombined(context.Background(), ".", "release", "view", tagName)
	return err == nil && exitCode == 0
}

// deleteRelease deletes a GitHub release.
func deleteRelease(tagName string) error {
	output, exitCode, err := ghexec.RunCombined(context.Background(), ".", "release", "delete", tagName, "--yes")
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("%s (exit %d)", strings.TrimSpace(string(output)), exitCode)
	}
	return nil
}

// deleteTag deletes a tag from the remote repository using gh api.
func deleteTag(tagName string) error {
	// Get repository from gh
	repoOutput, err := ghexec.Run(".", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}
	repo := strings.TrimSpace(string(repoOutput))

	// Delete tag via API
	apiPath := fmt.Sprintf("repos/%s/git/refs/tags/%s", repo, tagName)
	output, exitCode, err := ghexec.RunCombined(context.Background(), ".", "api", "--method", "DELETE", apiPath)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("%s (exit %d)", strings.TrimSpace(string(output)), exitCode)
	}
	return nil
}
