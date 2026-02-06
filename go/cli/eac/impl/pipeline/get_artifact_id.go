// Command: pipeline get-artifact-id
// Short: Get artifact ID from a workflow run
// Long: Queries GitHub API to find an artifact ID by name from a workflow run.
// Long:
// Long: This replaces the jq pattern:
// Long:   gh api ... --jq '.artifacts[] | select(.name == "X") | .id'
// Long:
// Long: SHA Detection:
// Long:   1. --sha flag (explicit)
// Long:   2. GITHUB_SHA env var (CI)
// Long:   3. origin/main (devbox)
// Long:
// Long: Example:
// Long:   pipeline get-artifact-id --workflow ci-r2r-cli.yaml --name build-artifacts
// Long:   pipeline get-artifact-id --run-id 12345 --name commands-binary
// Flag.workflow: type=string, usage=Workflow name to find run for
// Flag.run-id: type=string, usage=Specific run ID (alternative to workflow+sha)
// Flag.name: type=string, usage=Artifact name to find (required)
// Flag.sha: type=string, usage=Commit SHA (auto-detected if not provided)
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get"
	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(PipelineGetArtifactID)
}

// ArtifactInfo from GitHub API.
type ArtifactInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ArtifactsResponse struct {
	Artifacts []ArtifactInfo `json:"artifacts"`
}

func PipelineGetArtifactID() int {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Parse flags
	workflow := ""
	runID := ""
	artifactName := ""
	sha := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--workflow" && i+1 < len(os.Args):
			workflow = os.Args[i+1]
			i++
		case arg == "--run-id" && i+1 < len(os.Args):
			runID = os.Args[i+1]
			i++
		case arg == "--name" && i+1 < len(os.Args):
			artifactName = os.Args[i+1]
			i++
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		}
	}

	if artifactName == "" {
		fmt.Fprintln(os.Stderr, "Error: --name is required")
		return 1
	}

	// If no run-id, find it from workflow + sha
	if runID == "" {
		if workflow == "" {
			fmt.Fprintln(os.Stderr, "Error: either --run-id or --workflow is required")
			return 1
		}

		// Auto-detect SHA
		shaResult, err := get.DetectCurrentSHA(workspaceRoot, sha)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}

		// Find run ID
		foundRunID, err := findRunID(workflow, shaResult.SHA, workspaceRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		runID = foundRunID
	}

	// Get artifact ID from run
	artifactID, err := getArtifactIDFromRun(runID, artifactName, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println(artifactID)
	return 0
}

func findRunID(workflow, sha, workspaceRoot string) (string, error) {
	output, err := ghexec.Run(workspaceRoot, "run", "list",
		"--workflow", workflow,
		"--commit", sha,
		"--status", "success",
		"--json", "databaseId",
		"-q", ".[0].databaseId",
	)
	if err != nil {
		return "", fmt.Errorf("failed to find run: %w", err)
	}

	runID := strings.TrimSpace(string(output))
	if runID == "" {
		return "", fmt.Errorf("no successful run found for %s at %s", workflow, sha[:7])
	}

	return runID, nil
}

func getArtifactIDFromRun(runID, artifactName, workspaceRoot string) (string, error) {
	output, err := ghexec.Run(workspaceRoot, "api",
		fmt.Sprintf("repos/{owner}/{repo}/actions/runs/%s/artifacts", runID),
	)
	if err != nil {
		return "", fmt.Errorf("failed to query artifacts: %w", err)
	}

	var resp ArtifactsResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	for _, artifact := range resp.Artifacts {
		if artifact.Name == artifactName {
			return fmt.Sprintf("%d", artifact.ID), nil
		}
	}

	return "", fmt.Errorf("artifact '%s' not found in run %s", artifactName, runID)
}
