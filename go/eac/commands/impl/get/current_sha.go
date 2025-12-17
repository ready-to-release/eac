// Command: get current-sha
// Short: Get current commit SHA with auto-detection
// Long: Returns the current commit SHA using smart detection.
// Long:
// Long: Detection order:
// Long:   1. --sha flag (explicit override)
// Long:   2. GITHUB_SHA environment variable (GitHub Actions CI)
// Long:   3. origin/main HEAD after fetch (local devbox)
// Long:
// Long: Output formats:
// Long:   default: Just the SHA
// Long:   --format shell: SHA="..." SOURCE="ci|devbox|explicit"
// Long:
// Long: Example:
// Long:   get current-sha                    # Auto-detect
// Long:   get current-sha --sha abc123       # Explicit
// Long:   get current-sha --format shell     # For eval
// Flag.sha: type=string, usage=Override with explicit SHA
// Flag.format: type=string, usage=Output format (default, shell)
package get

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetCurrentSHA)
}

// SHASource indicates where the SHA was detected from
type SHASource string

const (
	SHASourceExplicit SHASource = "explicit"
	SHASourceCI       SHASource = "ci"
	SHASourceDevbox   SHASource = "devbox"
)

// SHAResult holds the detected SHA and its source
type SHAResult struct {
	SHA    string
	Source SHASource
}

// DetectCurrentSHA finds the current SHA using smart detection
func DetectCurrentSHA(workspaceRoot, explicitSHA string) (*SHAResult, error) {
	log := logging.C()

	// 1. Explicit SHA provided
	if explicitSHA != "" {
		log.Infof("Using explicit SHA: %s", shortSHA(explicitSHA))
		return &SHAResult{SHA: explicitSHA, Source: SHASourceExplicit}, nil
	}

	// 2. GitHub Actions CI environment
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		log.Infof("Detected CI environment (GITHUB_SHA)")
		log.Infof("Defaulted to: %s", shortSHA(sha))
		return &SHAResult{SHA: sha, Source: SHASourceCI}, nil
	}

	// 3. Local devbox - fetch and use origin/main
	log.Infof("Detected devbox environment (no GITHUB_SHA)")

	// Fetch latest from origin
	fetchCmd := exec.Command("git", "fetch", "origin", "main", "--quiet")
	fetchCmd.Dir = workspaceRoot
	_ = fetchCmd.Run() // Ignore errors, ref might already exist

	// Get origin/main SHA
	cmd := exec.Command("git", "rev-parse", "origin/main")
	cmd.Dir = workspaceRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get origin/main: %w", err)
	}

	sha := strings.TrimSpace(string(output))
	log.Infof("Defaulted to: %s (origin/main)", shortSHA(sha))
	return &SHAResult{SHA: sha, Source: SHASourceDevbox}, nil
}

func GetCurrentSHA() int {
	log := logging.C()

	// Get workspace root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Parse flags
	sha := ""
	format := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--sha" && i+1 < len(os.Args):
			sha = os.Args[i+1]
			i++
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		}
	}

	result, err := DetectCurrentSHA(workspaceRoot, sha)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	if format == "shell" {
		fmt.Printf("SHA=\"%s\"\n", result.SHA)
		fmt.Printf("SOURCE=\"%s\"\n", result.Source)
	} else {
		fmt.Println(result.SHA)
	}

	return 0
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
