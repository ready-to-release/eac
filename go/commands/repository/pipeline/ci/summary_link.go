package ci

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
)

type pipelineCISummaryLinkCommand struct{}

var _ core.SimpleCommandPort = (*pipelineCISummaryLinkCommand)(nil)

func (c *pipelineCISummaryLinkCommand) Name() string { return "pipeline ci summary-link" }

func (c *pipelineCISummaryLinkCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "pipeline-ci-summary-link",
		Short:         "Generate diagnostic markdown for CI summaries",
		Long:          "Generate a markdown code block with gh CLI commands for diagnosing CI failures.\n\nThis command outputs markdown that can be piped directly into $GITHUB_STEP_SUMMARY.\nThe generated commands use the actual run ID and repository so they can be\ncopy-pasted directly.\n\nExpected Output:\n  - Markdown code block with gh CLI diagnostic commands\n  - Commands use actual run ID and repository\n  - Suitable for piping to $GITHUB_STEP_SUMMARY\n\nExample:\n  pipeline ci summary-link 12345678                    # Basic diagnostic link\n  pipeline ci summary-link 12345678 --type test       # Include artifact download\n  pipeline ci summary-link 12345678 --artifact results # Specific artifact name\n  pipeline ci summary-link 12345678 --type container   # Container-specific diagnostics\n\nIn a workflow:\n  go run ./go/cli/eac pipeline ci summary-link ${{ github.run_id }} >> $GITHUB_STEP_SUMMARY",
		Flags: []core.FlagSpec{
			{Name: "type", Type: "string", Usage: "Failure type: build, test, container, release, docs (default: build)"},
			{Name: "artifact", Type: "string", Usage: "Artifact name to include in download command"},
			{Name: "image", Type: "string", Usage: "Container image for container-type diagnostics"},
			{Name: "workflow", Type: "string", Usage: "CI workflow name for release-type diagnostics"},
			{Name: "commit", Type: "string", Usage: "Commit SHA for release-type diagnostics"},
		},
	}
}

func (c *pipelineCISummaryLinkCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return PipelineCISummaryLink()
}

func PipelineCISummaryLink() int {
	// Validate flags before parsing (args start at index 5 for "pipeline ci summary-link <run-id>")
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}
	// Parse arguments (args start at index 4 for "pipeline ci summary-link")
	if len(os.Args) < 5 {
		log.Error("Error: run ID required")
		log.Error("Usage: pipeline ci summary-link <run-id> [--type <type>] [--artifact <name>]")
		return 1
	}

	runID := os.Args[4]
	failureType := "build"
	artifactName := ""
	imageName := ""
	workflowName := ""
	commitSHA := ""

	// Parse flags
	for i := 5; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--type" && i+1 < len(os.Args) {
			failureType = os.Args[i+1]
			i++
		} else if arg == "--artifact" && i+1 < len(os.Args) {
			artifactName = os.Args[i+1]
			i++
		} else if arg == "--image" && i+1 < len(os.Args) {
			imageName = os.Args[i+1]
			i++
		} else if arg == "--workflow" && i+1 < len(os.Args) {
			workflowName = os.Args[i+1]
			i++
		} else if arg == "--commit" && i+1 < len(os.Args) {
			commitSHA = os.Args[i+1]
			i++
		}
	}

	// Get repository from environment or git
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		repo = getRepoFromGit()
	}
	if repo == "" {
		repo = "{owner}/{repo}" // fallback placeholder
	}

	// Generate markdown based on type
	var sb strings.Builder

	// Add clickable link to the run
	sb.WriteString("### Diagnose Locally\n\n")
	sb.WriteString(fmt.Sprintf("[View Run #%s](https://github.com/%s/actions/runs/%s)\n\n", runID, repo, runID))
	sb.WriteString("```bash\n")

	switch failureType {
	case "test":
		sb.WriteString("# View failed step logs\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s --log-failed\n", runID, repo))
		sb.WriteString("\n")
		if artifactName != "" {
			sb.WriteString("# Download test results artifact\n")
			sb.WriteString(fmt.Sprintf("gh run download %s --repo %s -n %s\n", runID, repo, artifactName))
		} else {
			sb.WriteString("# Download all artifacts\n")
			sb.WriteString(fmt.Sprintf("gh run download %s --repo %s\n", runID, repo))
		}

	case "container":
		sb.WriteString("# View failed step logs\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s --log-failed\n", runID, repo))
		sb.WriteString("\n")
		if imageName != "" {
			sb.WriteString("# Pull and test the CI image locally\n")
			sb.WriteString(fmt.Sprintf("docker pull %s\n", imageName))
			sb.WriteString(fmt.Sprintf("docker run --rm %s extension-meta\n", imageName))
		}

	case "release":
		sb.WriteString("# View failed step logs\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s --log-failed\n", runID, repo))
		sb.WriteString("\n")
		if workflowName != "" && commitSHA != "" {
			sb.WriteString("# Check CI status for this commit\n")
			sb.WriteString(fmt.Sprintf("gh run list --repo %s --workflow %s --commit %s\n", repo, workflowName, commitSHA))
			sb.WriteString("\n")
		}
		sb.WriteString("# Check if release/tag already exists\n")
		sb.WriteString(fmt.Sprintf("gh release list --repo %s --limit 5\n", repo))
		sb.WriteString("git ls-remote --tags origin\n")

	case "docs":
		sb.WriteString("# View failed step logs\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s --log-failed\n", runID, repo))
		sb.WriteString("\n")
		if artifactName != "" {
			sb.WriteString("# Download build artifacts to inspect\n")
			sb.WriteString(fmt.Sprintf("gh run download %s --repo %s -n %s\n", runID, repo, artifactName))
		}
		sb.WriteString("\n")
		sb.WriteString("# Check GitHub Pages settings\n")
		sb.WriteString(fmt.Sprintf("gh api repos/%s/pages --jq '.status'\n", repo))

	case "deviation":
		// For cron-full-trigger deviation detection
		sb.WriteString("# View the failed full rebuild run\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s\n", runID, repo))
		sb.WriteString("\n")
		sb.WriteString("# View failed step logs from the full rebuild\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s --log-failed\n", runID, repo))
		sb.WriteString("\n")
		sb.WriteString("# Compare with recent incremental CI runs\n")
		sb.WriteString(fmt.Sprintf("gh run list --repo %s --workflow change-trigger.yaml --branch main --limit 5\n", repo))
		sb.WriteString("\n")
		sb.WriteString("# Check which modules were built in full vs incremental\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s --json jobs --jq '.jobs[] | {name, conclusion}'\n", runID, repo))

	default: // "build" or unknown
		sb.WriteString("# View failed step logs\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s --log-failed\n", runID, repo))
		sb.WriteString("\n")
		sb.WriteString("# Or download full logs\n")
		sb.WriteString(fmt.Sprintf("gh run view %s --repo %s --log\n", runID, repo))
	}

	sb.WriteString("```\n")

	log.Info(sb.String())
	return 0
}

// getRepoFromGit tries to get the repository from git remote.
func getRepoFromGit() string {
	// Try common environment variables first
	if repo := os.Getenv("GH_REPO"); repo != "" {
		return repo
	}

	// This is a simple implementation - in CI, GITHUB_REPOSITORY should be set
	return ""
}
