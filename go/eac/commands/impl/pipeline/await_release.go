// Command: pipeline await-release
// Short: Wait for release workflows to complete for a specific commit
// Long: Wait for release workflows (release-*.yaml) to complete for a specific commit SHA.
// Long:
// Long: This command polls GitHub Actions for in_progress or queued release workflow
// Long: runs that match the specified SHA and waits until all complete.
// Long:
// Long: SHA Detection (in order of precedence):
// Long:   1. --sha flag (explicit override)
// Long:   2. GITHUB_SHA environment variable (GitHub Actions)
// Long:   3. origin/main HEAD after fetch (local devbox)
// Long:
// Long: Expected Output:
// Long:   - Live progress display showing active workflow count
// Long:   - Exit code 0 when all workflows complete successfully
// Long:   - Exit code 1 on timeout or failure
// Long:
// Long: Example:
// Long:   pipeline await-release                              # Auto-detect SHA
// Long:   pipeline await-release --sha abc123                 # Explicit SHA
// Long:   pipeline await-release --timeout 300                # 5 minute timeout
// Long:   pipeline await-release --exclude r2r-eac-bundle     # Exclude bundle workflow
// Flag.sha: type=string, usage=Commit SHA to filter runs (auto-detected if not provided)
// Flag.timeout: type=int, default=600, usage=Maximum wait time in seconds (default: 600)
// Flag.interval: type=int, default=30, usage=Poll interval in seconds (default: 30)
// Flag.pattern: type=string, default=release-*.yaml, usage=Workflow file pattern to match
// Flag.exclude: type=string, usage=Workflow name substring to exclude (e.g., r2r-eac-bundle)
package pipeline

import (
	"os"
	"strconv"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(PipelineAwaitRelease)
}

func PipelineAwaitRelease() int {
	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Parse flags
	sha := ""
	timeout := 600 // 10 minutes default
	interval := 30 // 30 seconds default
	pattern := "release-*.yaml"
	exclude := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		case arg == "--timeout" && i+1 < len(os.Args):
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				timeout = v
			}
			i++
		case arg == "--interval" && i+1 < len(os.Args):
			if v, err := strconv.Atoi(os.Args[i+1]); err == nil {
				interval = v
			}
			i++
		case arg == "--pattern" && i+1 < len(os.Args):
			pattern = os.Args[i+1]
			i++
		case arg == "--exclude" && i+1 < len(os.Args):
			exclude = os.Args[i+1]
			i++
		}
	}

	// Auto-detect SHA using shared detection logic
	result, err := get.DetectCurrentSHA(workspaceRoot, sha)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	sha = result.SHA

	return awaitWorkflows(workspaceRoot, pattern, exclude, sha, timeout, interval, "release")
}
