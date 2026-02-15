package work

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/work/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
)

type showWorkspacesCommand struct{}

var _ core.SimpleCommandPort = (*showWorkspacesCommand)(nil)

func (c *showWorkspacesCommand) Name() string { return "show workspaces" }

func (c *showWorkspacesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "show-workspaces",
		Short:         "List all workspaces and their status",
		Long:          "Lists all git worktrees (workspaces) in a formatted table showing their path, branch, and status.\n\nThe status column indicates whether the worktree has uncommitted changes:\n  - clean: No uncommitted changes\n  - dirty: Has uncommitted changes\n\nUse --verbose to see additional information including commit SHA.\nUse --debug to enable detailed logging to out/commands.log.\n\nExample:\n  show workspaces\n  show workspaces --verbose\n  show workspaces -v\n  show workspaces --debug",
		Flags: []core.FlagSpec{
			{Name: "verbose", Type: "bool", Shorthand: "v", DefaultValue: "false", Usage: "Show detailed information including commit SHA"},
			{Name: "debug", Type: "bool", Shorthand: "d", DefaultValue: "false", Usage: "Enable debug logging"},
		},
	}
}

func (c *showWorkspacesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ShowWorkspaces()
}

// ShowWorkspaces displays all git worktrees in a formatted table.
func ShowWorkspaces() int {
	cmdStart := time.Now()

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Phase 1: Parse configuration
	phase1Start := time.Now()
	config, err := parseListConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	defer func() { _ = config.base.Logger.Sync() }() //nolint:errcheck // best-effort logger sync

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

	displayWorktrees(worktrees, config.verbose)

	config.base.Logger.Debug("Phase 4: Completed",
		zap.String("phase", "phase4"),
		zap.Duration("duration", time.Since(phase4Start)))

	config.base.Logger.Debug("Work list command completed successfully",
		zap.Duration("totalDuration", time.Since(cmdStart)))
	return 0
}

// listConfig holds configuration for the list command.
type listConfig struct {
	base    *internal.BaseConfig
	verbose bool
}

// parseListConfig parses command line arguments.
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

// displayWorktrees formats and displays worktrees in a table.
func displayWorktrees(worktrees []internal.Worktree, verbose bool) {
	if len(worktrees) == 0 {
		log.Debugf("No worktrees to display")
		log.Info("No worktrees found")
		return
	}

	// Phase 4.1: Build table
	phase41Start := time.Now()
	log.Debugf("Phase 4.1: Starting table building: phase=phase4.1, worktreeCount=%d, verbose=%v", len(worktrees), verbose)

	table := buildWorktreeTable(worktrees, verbose)

	log.Debugf("Phase 4.1: Completed: phase=phase4.1, tableLength=%d, duration=%v", len(table), time.Since(phase41Start))

	// Phase 4.2: Display results
	phase42Start := time.Now()
	log.Debugf("Phase 4.2: Starting result display: phase=phase4.2")

	log.Info(table)
	log.Info("")

	// Display summary
	if len(worktrees) == 1 {
		log.Info("1 worktree total")
	} else {
		log.Infof("%d worktrees total", len(worktrees))
	}

	log.Debugf("Phase 4.2: Completed: phase=phase4.2, duration=%v", time.Since(phase42Start))
}

// buildWorktreeTable creates a formatted table of worktrees.
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

// FormatWorktreeStatus returns a human-readable status string.
func FormatWorktreeStatus(clean bool) string {
	if clean {
		return "clean"
	}
	return "dirty"
}

// ShortenSHA returns the first 7 characters of a SHA.
func ShortenSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// FormatBranch formats branch name, handling detached HEAD.
func FormatBranch(branch string) string {
	if branch == "" || strings.TrimSpace(branch) == "" {
		return "(detached)"
	}
	return branch
}
