// Command: show workspaces
// Short: List all workspaces and their status
// Long: Lists all git worktrees (workspaces) in a formatted table showing their path, branch, and status.
// Long:
// Long: The status column indicates whether the worktree has uncommitted changes:
// Long:   - clean: No uncommitted changes
// Long:   - dirty: Has uncommitted changes
// Long:
// Long: Use --verbose to see additional information including commit SHA.
// Long: Use --debug to enable detailed logging to out/logs/work/.
// Long:
// Long: Example:
// Long:   show workspaces
// Long:   show workspaces --verbose
// Long:   show workspaces -v
// Long:   show workspaces --debug
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed information including commit SHA
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
package work

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/impl/work/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

func init() {
	registry.Register(ShowWorkspaces)
}

// Intent: List all workspaces (git worktrees) with their status
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear flow: validate → get worktrees → format → display
//   - Table-based output for easy scanning
//   - Status indicators (clean/dirty) are self-explanatory
//
// Easy to change:
//   - Formatting isolated in buildWorktreeTable function
//   - Verbose mode easily extended with additional columns
//   - Uses shared worktree utilities
//
// Hard to break:
//   - Validates git repo before proceeding
//   - Handles empty worktree list
//   - Clear error messages for failure cases

// ShowWorkspaces displays all git worktrees in a formatted table
func ShowWorkspaces() int {
	cmdStart := time.Now()

	// Phase 1: Parse configuration
	phase1Start := time.Now()
	config, err := parseListConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer config.base.Logger.Sync()

	config.base.Logger.Debug("Phase 1: Starting configuration parsing",
		zap.String("phase", "phase1"))

	config.base.Logger.Debug("Starting work list command",
		zap.Bool("verbose", config.verbose),
		zap.Bool("debug", config.base.Debug),
		zap.String("repoRoot", config.base.RepoRoot))

	config.base.Logger.Debug("Phase 1: Completed",
		zap.String("phase", "phase1"),
		zap.Duration("duration", time.Since(phase1Start)))

	// Phase 2: Validate environment
	phase2Start := time.Now()
	config.base.Logger.Debug("Phase 2: Starting environment validation",
		zap.String("phase", "phase2"))

	if err := internal.EnsureInGitRepo(); err != nil {
		config.base.Logger.Debug("Phase 2: Failed",
			zap.String("phase", "phase2"),
			zap.Error(err),
			zap.Duration("duration", time.Since(phase2Start)))
		config.base.Logger.Error(fmt.Sprintf("Not in a git repository: %v", err))
		return 1
	}

	config.base.Logger.Debug("Phase 2: Completed",
		zap.String("phase", "phase2"),
		zap.Duration("duration", time.Since(phase2Start)))

	// Phase 3: Get worktrees
	phase3Start := time.Now()
	config.base.Logger.Debug("Phase 3: Starting worktree retrieval",
		zap.String("phase", "phase3"),
		zap.String("repoRoot", config.base.RepoRoot))

	worktrees, err := internal.GetWorktrees(config.base.RepoRoot)
	if err != nil {
		config.base.Logger.Debug("Phase 3: Failed",
			zap.String("phase", "phase3"),
			zap.Error(err),
			zap.Duration("duration", time.Since(phase3Start)))
		config.base.Logger.Error(fmt.Sprintf("Failed to get worktrees: %v", err))
		return 1
	}

	config.base.Logger.Debug("Phase 3: Completed",
		zap.String("phase", "phase3"),
		zap.Int("count", len(worktrees)),
		zap.Duration("duration", time.Since(phase3Start)))

	// Phase 4: Display worktrees
	phase4Start := time.Now()
	config.base.Logger.Debug("Phase 4: Starting worktree display",
		zap.String("phase", "phase4"),
		zap.Int("count", len(worktrees)),
		zap.Bool("verbose", config.verbose))

	displayWorktrees(config.base.Logger, worktrees, config.verbose)

	config.base.Logger.Debug("Phase 4: Completed",
		zap.String("phase", "phase4"),
		zap.Duration("duration", time.Since(phase4Start)))

	config.base.Logger.Debug("Work list command completed successfully",
		zap.Duration("totalDuration", time.Since(cmdStart)))
	return 0
}

// listConfig holds configuration for the list command
type listConfig struct {
	base    *internal.BaseConfig
	verbose bool
}

// parseListConfig parses command line arguments
func parseListConfig() (*listConfig, error) {
	args := os.Args[3:] // Skip program name, "work", "list"

	// Parse base config (debug flag, repo root, logger, git ops)
	baseConfig, err := internal.ParseBaseConfig(args)
	if err != nil {
		return nil, err
	}

	config := &listConfig{
		base:    baseConfig,
		verbose: false,
	}

	// Parse verbose flag
	config.verbose = flags.HasFlag(args, "--verbose", "-v")

	return config, nil
}

// displayWorktrees formats and displays worktrees in a table
func displayWorktrees(logger *logging.Logger, worktrees []internal.Worktree, verbose bool) {
	if len(worktrees) == 0 {
		logger.Debug("No worktrees to display")
		logger.Info("No worktrees found")
		return
	}

	// Phase 4.1: Build table
	phase41Start := time.Now()
	logger.Debug("Phase 4.1: Starting table building",
		zap.String("phase", "phase4.1"),
		zap.Int("worktreeCount", len(worktrees)),
		zap.Bool("verbose", verbose))

	table := buildWorktreeTable(worktrees, verbose)

	logger.Debug("Phase 4.1: Completed",
		zap.String("phase", "phase4.1"),
		zap.Int("tableLength", len(table)),
		zap.Duration("duration", time.Since(phase41Start)))

	// Phase 4.2: Display results
	phase42Start := time.Now()
	logger.Debug("Phase 4.2: Starting result display",
		zap.String("phase", "phase4.2"))

	logger.Info(table)
	logger.Info("")

	// Display summary
	if len(worktrees) == 1 {
		logger.Info("1 worktree total")
	} else {
		logger.Info(fmt.Sprintf("%d worktrees total", len(worktrees)))
	}

	logger.Debug("Phase 4.2: Completed",
		zap.String("phase", "phase4.2"),
		zap.Duration("duration", time.Since(phase42Start)))
}

// buildWorktreeTable creates a formatted table of worktrees
func buildWorktreeTable(worktrees []internal.Worktree, verbose bool) string {
	// Determine headers based on verbose mode
	var headers []string
	if verbose {
		headers = []string{"Path", "Branch", "SHA", "Status"}
	} else {
		headers = []string{"Path", "Branch", "Status"}
	}

	tb := render.NewTableBuilder().WithHeaders(headers...)

	// Add rows
	for _, wt := range worktrees {
		status := "clean"
		if !wt.Clean {
			status = "dirty"
		}

		branch := wt.Branch
		if branch == "" {
			branch = "(detached)"
		}

		if verbose {
			// Show short SHA (first 7 chars)
			shortSHA := wt.SHA
			if len(shortSHA) > 7 {
				shortSHA = shortSHA[:7]
			}
			tb.AddRow(wt.Path, branch, shortSHA, status)
		} else {
			tb.AddRow(wt.Path, branch, status)
		}
	}

	return tb.Build()
}

// FormatWorktreeStatus returns a human-readable status string
func FormatWorktreeStatus(clean bool) string {
	if clean {
		return "clean"
	}
	return "dirty"
}

// ShortenSHA returns the first 7 characters of a SHA
func ShortenSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// FormatBranch formats branch name, handling detached HEAD
func FormatBranch(branch string) string {
	if branch == "" || strings.TrimSpace(branch) == "" {
		return "(detached)"
	}
	return branch
}
