package reports

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/git"
)

// ApprovalComment represents a single PR approval review
type ApprovalComment struct {
	PRNumber     int       `json:"pr_number" yaml:"pr_number" toml:"pr_number"`
	PRTitle      string    `json:"pr_title" yaml:"pr_title" toml:"pr_title"`
	PRAuthor     string    `json:"pr_author" yaml:"pr_author" toml:"pr_author"`
	PRBody       string    `json:"pr_body" yaml:"pr_body" toml:"pr_body"`                   // PR description/body
	MergeMessage string    `json:"merge_message" yaml:"merge_message" toml:"merge_message"` // Merge commit message
	Reviewer     string    `json:"reviewer" yaml:"reviewer" toml:"reviewer"`
	ReviewState  string    `json:"review_state" yaml:"review_state" toml:"review_state"` // APPROVED, CHANGES_REQUESTED
	ReviewedAt   time.Time `json:"reviewed_at" yaml:"reviewed_at" toml:"reviewed_at"`
	SpecFiles    []string  `json:"spec_files" yaml:"spec_files" toml:"spec_files"` // .feature files in PR
	MergedAt     time.Time `json:"merged_at" yaml:"merged_at" toml:"merged_at"`
}

// ApprovalCommentsReport represents all PR approvals for a module and version
type ApprovalCommentsReport struct {
	Module         string            `json:"module" yaml:"module" toml:"module"`
	Version        string            `json:"version" yaml:"version" toml:"version"`
	PRs            []PRData          `json:"prs,omitempty" yaml:"prs,omitempty" toml:"prs,omitempty"` // All PRs found (with spec files)
	Approvals      []ApprovalComment `json:"approvals" yaml:"approvals" toml:"approvals"`
	TotalPRs       int               `json:"total_prs" yaml:"total_prs" toml:"total_prs"`
	TotalApprovals int               `json:"total_approvals" yaml:"total_approvals" toml:"total_approvals"`
}

// PRData represents GitHub PR information
type PRData struct {
	Number             int
	Title              string
	Author             string
	Body               string // PR description/body
	MergedAt           time.Time
	MergeCommitMessage string // Merge commit message
	Files              []string
	Reviews            []ReviewData
}

// ReviewData represents a single review on a PR
type ReviewData struct {
	Author      string
	State       string // APPROVED, CHANGES_REQUESTED, COMMENTED
	SubmittedAt time.Time
}

// GitHubCLI interface for fetching PR data (allows mock injection for testing)
type GitHubCLI interface {
	GetPR(workspaceRoot string, prNumber int) (*PRData, error)
}

// ghCLI holds the GitHub CLI executor for testing (allows mock injection)
var ghCLI GitHubCLI

// getGitHubCLI returns the GitHub CLI executor
func getGitHubCLI() GitHubCLI {
	if ghCLI != nil {
		return ghCLI
	}
	return &realGitHubCLI{}
}

// SetGitHubCLI allows tests to inject a mock GitHub CLI
func SetGitHubCLI(cli GitHubCLI) {
	ghCLI = cli
}

// realGitHubCLI implements GitHubCLI using actual gh command
type realGitHubCLI struct{}

func (r *realGitHubCLI) GetPR(workspaceRoot string, prNumber int) (*PRData, error) {
	// Execute: gh pr view <number> --json number,title,author,body,mergedAt,mergeCommit,files,reviews
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(prNumber),
		"--json", "number,title,author,body,mergedAt,mergeCommit,files,reviews")
	cmd.Dir = workspaceRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	// Parse JSON response
	var result struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		MergedAt    time.Time `json:"mergedAt"`
		MergeCommit struct {
			MessageHeadline string `json:"messageHeadline"`
			MessageBody     string `json:"messageBody"`
		} `json:"mergeCommit"`
		Files []struct {
			Path string `json:"path"`
		} `json:"files"`
		Reviews []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			State       string    `json:"state"`
			SubmittedAt time.Time `json:"submittedAt"`
		} `json:"reviews"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	// Combine merge commit headline and body
	mergeMessage := result.MergeCommit.MessageHeadline
	if result.MergeCommit.MessageBody != "" {
		mergeMessage += "\n\n" + result.MergeCommit.MessageBody
	}

	prData := &PRData{
		Number:             result.Number,
		Title:              result.Title,
		Author:             result.Author.Login,
		Body:               result.Body,
		MergedAt:           result.MergedAt,
		MergeCommitMessage: mergeMessage,
		Files:              make([]string, len(result.Files)),
		Reviews:            make([]ReviewData, len(result.Reviews)),
	}

	for i, f := range result.Files {
		prData.Files[i] = f.Path
	}

	for i, r := range result.Reviews {
		prData.Reviews[i] = ReviewData{
			Author:      r.Author.Login,
			State:       r.State,
			SubmittedAt: r.SubmittedAt,
		}
	}

	return prData, nil
}

// GetApprovalComments loads PR approval comments for a module and version
// Version can be: empty (unreleased), "latest", "unreleased", or specific version number
// For bundle/container modules with dependencies, approvals are aggregated from all dependent modules
// If includeAllReviews is true, includes all review states (APPROVED, CHANGES_REQUESTED, COMMENTED)
// If includeAllReviews is false, only includes APPROVED reviews
// Branch specifies which branch to query (default "main"). Use "HEAD" or "current" for current branch.
func GetApprovalComments(workspaceRoot, module, version string, includeAllReviews bool, branch string) (*ApprovalCommentsReport, error) {
	// Load config to get module information
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	modContract, ok := cfg.Repository.GetModule(module)
	if !ok {
		return nil, fmt.Errorf("module not found: %s", module)
	}

	// Resolve version using common helper
	versionInfo, err := ResolveVersion(workspaceRoot, module, version)
	if err != nil {
		return nil, err
	}

	// Get git repository using the standard interface pattern
	repo, err := getGitRepo(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get commits for the version range
	var commits []git.CommitInfo

	// Determine end reference based on branch parameter and version
	endRef := branch
	if endRef == "" {
		// Default to configured trunk branch (usually "main" or "master")
		endRef = cfg.Repository.Repository.TrunkBranch
		if endRef == "" {
			endRef = "main" // Fallback if not configured
		}
	}
	// Special case: "HEAD" or "current" means use current branch
	if endRef == "HEAD" || endRef == "current" {
		endRef = "HEAD"
	}
	// Try origin/branch if local branch doesn't exist
	// (resolveRef in git layer handles this automatically)

	// For specific versions (not unreleased), use the version's tag as end point
	if !versionInfo.IsUnreleased && versionInfo.GitTag != "" {
		endRef = versionInfo.GitTag
	}

	if versionInfo.PreviousGitTag != "" {
		commits, err = repo.CommitsBetween(versionInfo.PreviousGitTag, endRef)
	} else {
		commits, err = repo.CommitsBetween("", endRef)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	// Determine which modules to query (bundle support)
	var modulesToQuery []string
	if len(modContract.DependsOn) > 0 {
		modulesToQuery = make([]string, len(modContract.DependsOn))
		copy(modulesToQuery, modContract.DependsOn)
		modulesToQuery = append(modulesToQuery, module)
	} else {
		modulesToQuery = []string{module}
	}

	// Build report aggregating approvals from all modules
	report := &ApprovalCommentsReport{
		Module:    versionInfo.Module,
		Version:   versionInfo.VersionNumber,
		Approvals: []ApprovalComment{},
	}

	// Extract PR numbers from commits
	prNumbers := extractPRNumbers(commits)

	// For each PR, get review data
	cli := getGitHubCLI()
	for _, prNum := range prNumbers {
		prData, err := cli.GetPR(workspaceRoot, prNum)
		if err != nil {
			// Log warning but continue (PR might not exist or gh CLI unavailable)
			continue
		}

		// Check if PR contains spec files from any of the queried modules
		specFiles := filterSpecFiles(prData.Files, modulesToQuery)
		if len(specFiles) == 0 {
			continue // Skip PRs without spec files
		}

		report.TotalPRs++

		// Add PR to the list of all found PRs (even if it has no reviews)
		report.PRs = append(report.PRs, *prData)

		// Extract reviews from PR
		for _, review := range prData.Reviews {
			// Filter by review state based on includeAllReviews flag
			if !includeAllReviews && review.State != "APPROVED" {
				continue
			}

			approval := ApprovalComment{
				PRNumber:     prNum,
				PRTitle:      prData.Title,
				PRAuthor:     prData.Author,
				PRBody:       prData.Body,
				MergeMessage: prData.MergeCommitMessage,
				Reviewer:     review.Author,
				ReviewState:  review.State,
				ReviewedAt:   review.SubmittedAt,
				SpecFiles:    specFiles,
				MergedAt:     prData.MergedAt,
			}
			report.Approvals = append(report.Approvals, approval)
			report.TotalApprovals++
		}
	}

	return report, nil
}

// extractPRNumbers extracts PR numbers from commit messages
// Matches patterns like "(#123)" or "Merge pull request #123"
func extractPRNumbers(commits []git.CommitInfo) []int {
	prMap := make(map[int]bool)
	prRegex := regexp.MustCompile(`(?:Merge pull request #(\d+)|#(\d+)\))`)

	for _, commit := range commits {
		matches := prRegex.FindAllStringSubmatch(commit.Message, -1)
		for _, match := range matches {
			for i := 1; i < len(match); i++ {
				if match[i] != "" {
					if num, err := strconv.Atoi(match[i]); err == nil {
						prMap[num] = true
					}
				}
			}
		}
	}

	// Convert map to sorted slice
	prNumbers := make([]int, 0, len(prMap))
	for num := range prMap {
		prNumbers = append(prNumbers, num)
	}
	sort.Ints(prNumbers)
	return prNumbers
}

// filterSpecFiles returns spec files that belong to any of the queried modules
func filterSpecFiles(files []string, modules []string) []string {
	var specFiles []string
	for _, file := range files {
		normalizedFile := filepath.ToSlash(file)

		// Only process .feature files
		if !strings.HasSuffix(normalizedFile, ".feature") {
			continue
		}

		// Check if file is in specs/<module>/ for any module
		for _, mod := range modules {
			modulePrefix := "specs/" + mod + "/"
			if strings.HasPrefix(normalizedFile, modulePrefix) {
				specFiles = append(specFiles, file)
				break
			}
		}
	}
	return specFiles
}
