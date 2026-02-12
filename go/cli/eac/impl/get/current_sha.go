package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getCurrentSHACommand struct{}

var _ core.SimpleCommandPort = (*getCurrentSHACommand)(nil)

func (c *getCurrentSHACommand) Name() string { return "get current-sha" }

func (c *getCurrentSHACommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-current-sha",
		Short:         "Get current commit SHA with auto-detection",
		Long: "Returns the current commit SHA using smart detection.\n" +
			"\n" +
			"Detection order:\n" +
			"  1. --sha flag (explicit override)\n" +
			"  2. GITHUB_SHA environment variable (GitHub Actions CI)\n" +
			"  3. origin/main HEAD after fetch (local devbox)\n" +
			"\n" +
			"Output formats:\n" +
			"  default: Just the SHA\n" +
			"  --format shell: SHA=\"...\" SOURCE=\"ci|devbox|explicit\"\n" +
			"\n" +
			"Example:\n" +
			"  get current-sha                    # Auto-detect\n" +
			"  get current-sha --sha abc123       # Explicit\n" +
			"  get current-sha --format shell     # For eval",
		Flags: []core.FlagSpec{
			{Name: "sha", Type: "string", Usage: "Override with explicit SHA"},
			{Name: "format", Type: "string", Usage: "Output format (default, shell)"},
		},
	}
}

func (c *getCurrentSHACommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetCurrentSHA()
}

// SHASource indicates where the SHA was detected from.
type SHASource string

const (
	SHASourceExplicit SHASource = "explicit"
	SHASourceCI       SHASource = "ci"
	SHASourceDevbox   SHASource = "devbox"
)

// SHAResult holds the detected SHA and its source.
type SHAResult struct {
	SHA    string
	Source SHASource
}

// DetectCurrentSHA finds the current SHA using smart detection.
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

	// Fetch latest from origin (best-effort, ref might already exist)
	ts := tool.GlobalToolSystem()
	_, _ = ts.RunTool(context.Background(), "git", workspaceRoot, "fetch", "origin", "main", "--quiet") //nolint:errcheck

	// Get origin/main SHA
	output, err := ts.RunTool(context.Background(), "git", workspaceRoot, "rev-parse", "origin/main")
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
