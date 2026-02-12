package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/get"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/repository"
)

type pipelineDownloadEvidenceArtifactsCommand struct{}

var _ core.SimpleCommandPort = (*pipelineDownloadEvidenceArtifactsCommand)(nil)

func (c *pipelineDownloadEvidenceArtifactsCommand) Name() string {
	return "pipeline download-evidence-artifacts"
}

func (c *pipelineDownloadEvidenceArtifactsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "pipeline-download-evidence-artifacts",
		Short:         "Download test/scan artifacts for evidence building",
		Long:          "Downloads test results and scan artifacts from CI runs for a module and all its dependencies.\nUses the `get evidence-ci-runs` command to determine which CI runs to download from.\n\nFor each module in the dependency chain that has a ci-{module}.yaml workflow:\n  - Downloads test-results-{module}* artifacts to out/test/\n  - Downloads scan-results-{module} artifacts to out/scan/\n\nFails if any dependency with a CI workflow has no successful CI run.\n\nExample:\n  pipeline download-evidence-artifacts clie\n  pipeline download-evidence-artifacts eac-ext --output-dir custom/path",
		Args:          "module (required) - Module moniker to download evidence artifacts for",
		Flags: []core.FlagSpec{
			{Name: "output-dir", Type: "string", DefaultValue: "out", Usage: "Base output directory (test/ and scan/ subdirs will be created)"},
		},
	}
}

func (c *pipelineDownloadEvidenceArtifactsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return PipelineDownloadEvidenceArtifacts()
}

// DownloadResult tracks what was downloaded.
type DownloadResult struct {
	Module        string   `json:"module" yaml:"module"`
	RunID         int      `json:"run_id" yaml:"run_id"`
	TestArtifacts []string `json:"test_artifacts,omitempty" yaml:"test_artifacts,omitempty"`
	ScanArtifacts []string `json:"scan_artifacts,omitempty" yaml:"scan_artifacts,omitempty"`
	Error         string   `json:"error,omitempty" yaml:"error,omitempty"`
}

func PipelineDownloadEvidenceArtifacts() int {
	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	moniker, outputDir := parseDownloadArgs()

	if moniker == "" {
		fmt.Fprintln(os.Stderr, "Usage: pipeline download-evidence-artifacts <module> [--output-dir DIR]")
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Create output directories
	testDir := filepath.Join(outputDir, "test")
	scanDir := filepath.Join(outputDir, "scan")

	if err := os.MkdirAll(testDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating test directory: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(scanDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating scan directory: %v\n", err)
		return 1
	}

	// Get evidence CI runs
	fmt.Printf("=== Downloading evidence artifacts for: %s ===\n\n", moniker)

	ciRuns, skipped, err := getEvidenceCIRunsInternal(moniker, workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if len(skipped) > 0 {
		fmt.Printf("Skipped (no CI workflow): %s\n\n", strings.Join(skipped, ", "))
	}

	// Download artifacts from each CI run
	results, hasErrors := downloadAllRunArtifacts(ciRuns, testDir, scanDir, workspaceRoot)

	// Summary
	printDownloadSummary(results)

	if hasErrors {
		return 1
	}
	return 0
}

// parseDownloadArgs parses arguments for the download-evidence-artifacts command.
func parseDownloadArgs() (moniker, outputDir string) {
	args := os.Args[3:] // Skip program name, "pipeline", and "download-evidence-artifacts"
	outputDir = "out"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--output-dir" && i+1 < len(args):
			outputDir = args[i+1]
			i++
		case strings.HasPrefix(arg, "--output-dir="):
			outputDir = strings.TrimPrefix(arg, "--output-dir=")
		case !strings.HasPrefix(arg, "--") && moniker == "":
			moniker = arg
		}
	}

	return moniker, outputDir
}

// downloadAllRunArtifacts downloads test and scan artifacts from each CI run.
func downloadAllRunArtifacts(ciRuns []get.EvidenceCIRun, testDir, scanDir, workspaceRoot string) ([]DownloadResult, bool) {
	var results []DownloadResult
	hasErrors := false

	for _, run := range ciRuns {
		fmt.Printf("--- %s (run %d) ---\n", run.Module, run.RunID)

		result := DownloadResult{
			Module: run.Module,
			RunID:  run.RunID,
		}

		// Download test artifacts
		testArtifacts, err := downloadArtifacts(run.RunID, fmt.Sprintf("test-results-%s*", run.Module), testDir, workspaceRoot)
		if err != nil {
			fmt.Printf("  Warning: test artifacts: %v\n", err)
		} else if len(testArtifacts) > 0 {
			result.TestArtifacts = testArtifacts
			fmt.Printf("  Downloaded %d test artifact(s)\n", len(testArtifacts))
		} else {
			fmt.Printf("  No test artifacts found\n")
		}

		// Download scan artifacts
		scanArtifacts, err := downloadArtifacts(run.RunID, fmt.Sprintf("scan-results-%s", run.Module), scanDir, workspaceRoot)
		if err != nil {
			fmt.Printf("  Warning: scan artifacts: %v\n", err)
		} else if len(scanArtifacts) > 0 {
			result.ScanArtifacts = scanArtifacts
			fmt.Printf("  Downloaded %d scan artifact(s)\n", len(scanArtifacts))
		} else {
			fmt.Printf("  No scan artifacts found\n")
		}

		results = append(results, result)
		fmt.Println()
	}

	return results, hasErrors
}

// printDownloadSummary prints the final download summary.
func printDownloadSummary(results []DownloadResult) {
	fmt.Println("=== Download Summary ===")
	totalTests := 0
	totalScans := 0
	for _, r := range results {
		totalTests += len(r.TestArtifacts)
		totalScans += len(r.ScanArtifacts)
	}
	fmt.Printf("Modules processed: %d\n", len(results))
	fmt.Printf("Test artifacts: %d\n", totalTests)
	fmt.Printf("Scan artifacts: %d\n", totalScans)
}
