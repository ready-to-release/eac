package release

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/tool"
)

type releaseCheckExistsCommand struct{}

var _ core.SimpleCommandPort = (*releaseCheckExistsCommand)(nil)

func (c *releaseCheckExistsCommand) Name() string { return "release check-exists" }

func (c *releaseCheckExistsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "release-check-exists",
		Short:         "Check if a release/tag already exists",
		Long:          "Checks GitHub for an existing release with the given tag.\n\nExit codes:\n  0 - Release does NOT exist (safe to create)\n  1 - Release EXISTS or error\n\nOutput formats:\n  default: Human readable message\n  --format shell: EXISTS=\"true\" TAG=\"...\" ERROR=\"...\"\n\nExample:\n  release check-exists --tag clie/1.0.0\n  eval $(release check-exists --tag clie/1.0.0 --format shell)",
		Flags: []core.FlagSpec{
			{Name: "tag", Type: "string", Usage: "Tag name to check (required)"},
			{Name: "format", Type: "string", Usage: "Output format (default, shell)"},
		},
	}
}

func (c *releaseCheckExistsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ReleaseCheckExists()
}

func ReleaseCheckExists() int {
	// Parse flags first (before scaffold, since format affects error output)
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

	s, exitCode := newReleaseScaffoldNoFlags()
	if s == nil {
		if format == "shell" {
			fmt.Printf("EXISTS=\"\"\n")
			fmt.Printf("TAG=\"%s\"\n", tag)
			fmt.Printf("ERROR=\"failed to find repository\"\n")
		}
		return exitCode
	}
	workspaceRoot := s.WorkspaceRoot

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
	_, exitCode, err := tool.GlobalToolSystem().RunToolCombined(context.Background(), "gh", workspaceRoot, "release", "view", tag)
	if err != nil {
		return false, err
	}

	// Exit code 0 = release exists, non-zero = doesn't exist
	return exitCode == 0, nil
}
