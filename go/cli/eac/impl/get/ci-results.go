package get

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type getCIResultsCommand struct{}

var _ core.SimpleCommandPort = (*getCIResultsCommand)(nil)

func (c *getCIResultsCommand) Name() string { return "get ci-results" }

func (c *getCIResultsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-ci-results",
		Short:         "Get CI workflow run results with job details and artifact links",
		Long: "Returns structured CI run data including job results, artifacts, and diagnostic links.\n" +
			"\n" +
			"Queries GitHub Actions for CI workflow runs at a given commit SHA or run ID,\n" +
			"then enriches each run with job-level details and downloadable artifact info.\n" +
			"\n" +
			"Input Detection:\n" +
			"  - 40-char hex or 7+ hex prefix: treated as commit SHA\n" +
			"  - Numeric value: treated as a specific run ID\n" +
			"  - Omitted: auto-detects SHA (GITHUB_SHA -> origin/main -> git HEAD)\n" +
			"\n" +
			"Output includes per-module:\n" +
			"  - Workflow run status and conclusion\n" +
			"  - Job-level results with durations\n" +
			"  - Artifact names and sizes\n" +
			"  - Diagnostic links (web URL, gh CLI commands)\n" +
			"\n" +
			"Example:\n" +
			"  get ci-results                          # Current HEAD\n" +
			"  get ci-results abc1234                   # Specific commit\n" +
			"  get ci-results abc1234 core clibase      # Specific modules\n" +
			"  get ci-results 12345678                   # Specific run ID\n" +
			"  get ci-results abc1234 --as-json          # JSON output",
		Args: "[sha-or-run-id] [module...]",
		Flags: []core.FlagSpec{
			{Name: "as-yaml", Type: "bool", Usage: "Output as YAML (default format)"},
			{Name: "as-json", Type: "bool", Usage: "Output as JSON"},
			{Name: "as-toml", Type: "bool", Usage: "Output as TOML"},
		},
	}
}

func (c *getCIResultsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetCIResults()
}

// CIResultsSummary is the top-level result for get ci-results.
type CIResultsSummary struct {
	HeadSHA      string       `json:"head_sha" yaml:"head_sha"`
	Orchestrator *CIRunResult `json:"orchestrator,omitempty" yaml:"orchestrator,omitempty"`
	Runs         []CIRunResult `json:"runs" yaml:"runs"`
	TotalRuns    int           `json:"total_runs" yaml:"total_runs"`
	Passed       int           `json:"passed" yaml:"passed"`
	Failed       int           `json:"failed" yaml:"failed"`
}

const orchestratorWorkflow = "change-trigger.yaml"

// CIRunResult represents a single CI workflow run with full job details.
type CIRunResult struct {
	Module     string        `json:"module" yaml:"module"`
	Workflow   string        `json:"workflow" yaml:"workflow"`
	RunID      int           `json:"run_id" yaml:"run_id"`
	HeadSHA    string        `json:"head_sha" yaml:"head_sha"`
	Status     string        `json:"status" yaml:"status"`
	Conclusion string        `json:"conclusion" yaml:"conclusion"`
	CreatedAt  string        `json:"created_at" yaml:"created_at"`
	Jobs       []CIJobResult `json:"jobs" yaml:"jobs"`
	Artifacts  []CIArtifact  `json:"artifacts" yaml:"artifacts"`
	Links      CIRunLinks    `json:"links" yaml:"links"`
}

// CIJobResult represents a job within a workflow run.
type CIJobResult struct {
	Name       string `json:"name" yaml:"name"`
	Status     string `json:"status" yaml:"status"`
	Conclusion string `json:"conclusion" yaml:"conclusion"`
	Duration   string `json:"duration" yaml:"duration"`
	HTMLURL    string `json:"html_url,omitempty" yaml:"html_url,omitempty"`
}

// CIArtifact represents a downloadable artifact from the run.
type CIArtifact struct {
	Name      string `json:"name" yaml:"name"`
	SizeBytes int64  `json:"size_bytes" yaml:"size_bytes"`
	Expired   bool   `json:"expired" yaml:"expired"`
}

// CIRunLinks contains actionable URLs and CLI commands.
type CIRunLinks struct {
	WebURL         string `json:"web_url" yaml:"web_url"`
	ViewLogs       string `json:"view_logs" yaml:"view_logs"`
	ViewFailedLogs string `json:"view_failed_logs" yaml:"view_failed_logs"`
	DownloadAll    string `json:"download_all" yaml:"download_all"`
}

// inputType classifies the first positional argument.
type inputType int

const (
	inputSHA   inputType = iota
	inputRunID inputType = iota
	inputAuto  inputType = iota
)

var hexPattern = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

func GetCIResults() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse positional args and flags
	args := os.Args[3:] // Skip program, "get", "ci-results"
	var positional []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			continue
		}
		positional = append(positional, arg)
	}

	// Classify first arg
	iType, shaOrRunID := classifyInput(positional)
	modules := positional
	if len(positional) > 0 && iType != inputAuto {
		modules = positional[1:] // First arg was sha/run-id
	}

	// Resolve SHA
	var headSHA string
	var specificRunID int

	switch iType {
	case inputRunID:
		id, _ := strconv.Atoi(shaOrRunID)
		specificRunID = id
	case inputSHA:
		headSHA = shaOrRunID
	case inputAuto:
		result, err := DetectCurrentSHA(workspaceRoot, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to detect SHA: %v\n", err)
			return 1
		}
		headSHA = result.SHA
	}

	// Build GitHub API client
	api := github.NewGHClient(gh.New(tool.GlobalToolSystem(), workspaceRoot), workspaceRoot)

	// Get repo name for links
	repo := getRepoName(workspaceRoot)

	var summary *CIResultsSummary

	if specificRunID > 0 {
		summary, err = GetCIResultsForRunID(specificRunID, api, repo, workspaceRoot)
	} else {
		summary, err = GetCIResultsForSHA(headSHA, modules, api, repo, workspaceRoot)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return summary, nil
	})
}

// classifyInput determines if the first positional arg is a SHA, run ID, or absent.
func classifyInput(positional []string) (inputType, string) {
	if len(positional) == 0 {
		return inputAuto, ""
	}

	first := positional[0]

	// Pure numeric → run ID
	if _, err := strconv.Atoi(first); err == nil {
		return inputRunID, first
	}

	// Hex string 7-40 chars → SHA
	if hexPattern.MatchString(first) {
		return inputSHA, first
	}

	// Otherwise it's a module name → auto-detect SHA
	return inputAuto, ""
}

// getCIResultsForSHA queries CI runs across all module workflows at the given SHA.
func GetCIResultsForSHA(sha string, modules []string, api *github.GHClient, repo, workspaceRoot string) (*CIResultsSummary, error) {
	// Discover CI modules
	ciModules, err := discoverCIModules(modules, workspaceRoot)
	if err != nil {
		return nil, err
	}

	summary := &CIResultsSummary{
		HeadSHA: sha,
		Runs:    []CIRunResult{},
	}

	// Lookup the orchestrator run at this SHA
	orchRun, err := api.FindRunBySHA(orchestratorWorkflow, sha, 20)
	if err == nil {
		orch, _ := enrichRunResult("ci-orchestrator", orchestratorWorkflow, orchRun, api, repo)
		if orch != nil {
			summary.Orchestrator = orch
		}
	}

	for _, module := range ciModules {
		workflow := fmt.Sprintf("ci-%s.yaml", module)

		// Find run at this SHA
		run, err := api.FindRunBySHA(workflow, sha, 20)
		if err != nil {
			// No run found for this module - skip
			continue
		}

		result, err := enrichRunResult(module, workflow, run, api, repo)
		if err != nil {
			// Partial failure - include what we have
			result = &CIRunResult{
				Module:     module,
				Workflow:   workflow,
				RunID:      run.ID,
				HeadSHA:    run.HeadSHA,
				Status:     run.Status,
				Conclusion: run.Conclusion,
				CreatedAt:  run.CreatedAt.Format(time.RFC3339),
				Links:      buildRunLinks(run.ID, repo),
			}
		}

		summary.Runs = append(summary.Runs, *result)
	}

	// Compute totals
	summary.TotalRuns = len(summary.Runs)
	for _, r := range summary.Runs {
		switch r.Conclusion {
		case "success":
			summary.Passed++
		case "failure":
			summary.Failed++
		}
	}

	return summary, nil
}

// getCIResultsForRunID queries a specific run ID.
func GetCIResultsForRunID(runID int, api *github.GHClient, repo, workspaceRoot string) (*CIResultsSummary, error) {
	// Query the run directly via gh API
	output, err := api.Exec("api",
		fmt.Sprintf("repos/{owner}/{repo}/actions/runs/%d", runID),
		"--jq", ".head_sha,.status,.conclusion,.created_at,.name,.path",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query run %d: %w", runID, err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 4 {
		return nil, fmt.Errorf("unexpected response for run %d", runID)
	}

	sha := lines[0]
	status := lines[1]
	conclusion := lines[2]
	createdAt := lines[3]
	workflowName := ""
	if len(lines) >= 6 {
		// path is like .github/workflows/ci-core.yaml
		workflowName = filepath.Base(lines[5])
	}

	// Derive module from workflow name
	module := ""
	if strings.HasPrefix(workflowName, "ci-") && strings.HasSuffix(workflowName, ".yaml") {
		module = strings.TrimSuffix(strings.TrimPrefix(workflowName, "ci-"), ".yaml")
	}

	run := &github.WorkflowRun{
		ID:         runID,
		HeadSHA:    sha,
		Status:     status,
		Conclusion: conclusion,
	}

	result, err := enrichRunResult(module, workflowName, run, api, repo)
	if err != nil {
		result = &CIRunResult{
			Module:     module,
			Workflow:   workflowName,
			RunID:      runID,
			HeadSHA:    sha,
			Status:     status,
			Conclusion: conclusion,
			CreatedAt:  createdAt,
			Links:      buildRunLinks(runID, repo),
		}
	}

	summary := &CIResultsSummary{
		HeadSHA:   sha,
		Runs:      []CIRunResult{*result},
		TotalRuns: 1,
	}
	if conclusion == "success" {
		summary.Passed = 1
	} else if conclusion == "failure" {
		summary.Failed = 1
	}

	return summary, nil
}

// enrichRunResult adds job details and artifacts to a run result.
func enrichRunResult(module, workflow string, run *github.WorkflowRun, api *github.GHClient, repo string) (*CIRunResult, error) {
	result := &CIRunResult{
		Module:     module,
		Workflow:   workflow,
		RunID:      run.ID,
		HeadSHA:    run.HeadSHA,
		Status:     run.Status,
		Conclusion: run.Conclusion,
		CreatedAt:  run.CreatedAt.Format(time.RFC3339),
		Links:      buildRunLinks(run.ID, repo),
	}

	// Fetch jobs
	jobs, err := fetchRunJobs(run.ID, api)
	if err == nil {
		result.Jobs = jobs
	}

	// Fetch artifacts
	artifacts, err := fetchRunArtifacts(run.ID, api)
	if err == nil {
		result.Artifacts = artifacts
	}

	return result, nil
}

// jobAPIResponse maps the GitHub Actions jobs API response.
type jobAPIResponse struct {
	Jobs []struct {
		Name        string    `json:"name"`
		Status      string    `json:"status"`
		Conclusion  string    `json:"conclusion"`
		StartedAt   time.Time `json:"started_at"`
		CompletedAt time.Time `json:"completed_at"`
		HTMLURL     string    `json:"html_url"`
	} `json:"jobs"`
}

// fetchRunJobs queries the jobs for a workflow run.
func fetchRunJobs(runID int, api *github.GHClient) ([]CIJobResult, error) {
	output, err := api.Exec("api",
		fmt.Sprintf("repos/{owner}/{repo}/actions/runs/%d/jobs", runID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}

	var resp jobAPIResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse jobs response: %w", err)
	}

	var jobs []CIJobResult
	for _, j := range resp.Jobs {
		duration := ""
		if !j.CompletedAt.IsZero() && !j.StartedAt.IsZero() {
			d := j.CompletedAt.Sub(j.StartedAt)
			if d > 0 {
				duration = formatDuration(d)
			}
		}

		jobs = append(jobs, CIJobResult{
			Name:       j.Name,
			Status:     j.Status,
			Conclusion: j.Conclusion,
			Duration:   duration,
			HTMLURL:    j.HTMLURL,
		})
	}

	return jobs, nil
}

// artifactAPIResponse maps the GitHub Actions artifacts API response.
type artifactAPIResponse struct {
	Artifacts []struct {
		Name               string `json:"name"`
		SizeInBytes        int64  `json:"size_in_bytes"`
		Expired            bool   `json:"expired"`
	} `json:"artifacts"`
}

// fetchRunArtifacts queries the artifacts for a workflow run.
func fetchRunArtifacts(runID int, api *github.GHClient) ([]CIArtifact, error) {
	output, err := api.Exec("api",
		fmt.Sprintf("repos/{owner}/{repo}/actions/runs/%d/artifacts", runID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artifacts: %w", err)
	}

	var resp artifactAPIResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse artifacts response: %w", err)
	}

	var artifacts []CIArtifact
	for _, a := range resp.Artifacts {
		artifacts = append(artifacts, CIArtifact{
			Name:      a.Name,
			SizeBytes: a.SizeInBytes,
			Expired:   a.Expired,
		})
	}

	return artifacts, nil
}

// buildRunLinks generates actionable links for a run.
func buildRunLinks(runID int, repo string) CIRunLinks {
	return CIRunLinks{
		WebURL:         fmt.Sprintf("https://github.com/%s/actions/runs/%d", repo, runID),
		ViewLogs:       fmt.Sprintf("gh run view %d --repo %s --log", runID, repo),
		ViewFailedLogs: fmt.Sprintf("gh run view %d --repo %s --log-failed", runID, repo),
		DownloadAll:    fmt.Sprintf("gh run download %d --repo %s", runID, repo),
	}
}

// discoverCIModules returns CI module names. If modules is non-empty, validates they exist.
// If empty, discovers all from ci-*.yaml workflow files.
func discoverCIModules(modules []string, workspaceRoot string) ([]string, error) {
	validModules, err := getValidCIModules(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to discover CI modules: %w", err)
	}

	if len(modules) == 0 {
		// Return all
		result := make([]string, 0, len(validModules))
		for m := range validModules {
			result = append(result, m)
		}
		// Sort for deterministic output
		sortStrings(result)
		return result, nil
	}

	// Validate requested modules
	for _, m := range modules {
		if !validModules[m] {
			return nil, fmt.Errorf("no CI workflow found for module %q (expected ci-%s.yaml)", m, m)
		}
	}
	return modules, nil
}

// sortStrings sorts a string slice in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", m, s)
}

// getRepoName returns the GitHub owner/repo string.
func getRepoName(workspaceRoot string) string {
	// Try environment variable first (CI)
	if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
		return repo
	}
	if repo := os.Getenv("GH_REPO"); repo != "" {
		return repo
	}
	// Try gh api
	output, err := tool.GlobalToolSystem().RunTool(context.Background(), "gh", workspaceRoot, "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err == nil {
		name := strings.TrimSpace(string(output))
		if name != "" {
			return name
		}
	}
	return "{owner}/{repo}"
}
