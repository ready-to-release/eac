package reports

import (
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/github"
)

// ReportDeps holds injectable dependencies for report functions.
// This replaces global mutable state with explicit constructor injection,
// making dependencies visible and tests safe for parallel execution.
type ReportDeps struct {
	// GitHubCLI is a high-level interface for fetching PR data.
	// When set, it takes priority over CLIExecutor.
	// Typically used in tests with a mock implementation.
	GitHubCLI GitHubCLI

	// CLIExecutor is the low-level GitHub CLI command executor.
	// Used to build a realGitHubCLI when GitHubCLI is nil.
	// Set during CLI bootstrap via DefaultReportDeps.
	CLIExecutor github.CLIExecutor

	// GitRepo is the git repository used for commit queries.
	// When nil, a new repository is opened from the workspace root.
	GitRepo git.GitRepository

	// VersionResolverRepo is the git repository used for tag validation
	// in ResolveVersionWithValidation. When nil, a new repository is
	// opened from the workspace root.
	VersionResolverRepo git.GitRepository
}

// defaultDeps holds the production ReportDeps, set once during CLI bootstrap.
// This is the only package-level state; report functions receive deps explicitly,
// so tests never touch this variable.
var defaultDeps *ReportDeps

// InitDeps sets the package-level production dependencies.
// Called once during CLI bootstrap (main.go). Not used in tests.
func InitDeps(deps *ReportDeps) {
	defaultDeps = deps
}

// Deps returns the production ReportDeps set during bootstrap.
// CLI command handlers call this to obtain the deps they pass to report functions.
// Returns a zero-value ReportDeps if InitDeps was never called.
func Deps() *ReportDeps {
	if defaultDeps != nil {
		return defaultDeps
	}
	return &ReportDeps{}
}

// DefaultReportDeps creates a ReportDeps with production defaults.
// The CLIExecutor should be provided by the CLI bootstrap layer.
func DefaultReportDeps(executor github.CLIExecutor) *ReportDeps {
	return &ReportDeps{
		CLIExecutor: executor,
	}
}

// getGitHubCLI returns the GitHub CLI to use for PR queries.
func (d *ReportDeps) getGitHubCLI() GitHubCLI {
	if d != nil && d.GitHubCLI != nil {
		return d.GitHubCLI
	}
	var executor github.CLIExecutor
	if d != nil {
		executor = d.CLIExecutor
	}
	return &realGitHubCLI{executor: executor}
}

// getGitRepo returns the git repository, opening one at workspaceRoot if needed.
func (d *ReportDeps) getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if d != nil && d.GitRepo != nil {
		return d.GitRepo, nil
	}
	return git.NewManager(nil).Open(workspaceRoot)
}

// getVersionResolverRepo returns the git repository for tag validation,
// opening one at workspaceRoot if needed.
func (d *ReportDeps) getVersionResolverRepo(workspaceRoot string) (git.GitRepository, error) {
	if d != nil && d.VersionResolverRepo != nil {
		return d.VersionResolverRepo, nil
	}
	return git.NewManager(nil).Open(workspaceRoot)
}
