package ci

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/gh"
	"github.com/ready-to-release/eac/go/commands/repository/get"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type pipelineCIScheduleCommand struct{}

var _ core.SimpleCommandPort = (*pipelineCIScheduleCommand)(nil)

func (c *pipelineCIScheduleCommand) Name() string { return "pipeline ci schedule" }

func (c *pipelineCIScheduleCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "pipeline-ci-schedule",
		Short:         "Schedule and dispatch CI workflows with concurrency limits",
		Long:          "Replaces wave-based dispatch with a pull-based scheduler that:\n  1. Filters modules (CI cache check, same as get ci-dispatch)\n  2. Builds dependency graph from repository config\n  3. Dispatches modules as capacity allows and deps are satisfied\n  4. Polls for completion, dispatches next ready modules\n  5. Exits 0 when all complete successfully, 1 on any failure\n\nThis command is the CI-level analog of the local DependencyScheduler.\nIt handles the full dispatch lifecycle: filter, dispatch, poll, report.\n\nExample:\n  pipeline ci schedule --directly-changed \"core\" --invalidated \"eac docs\" --head-sha abc123 --max-concurrent 6\n  pipeline ci schedule --directly-changed \"core\" --max-concurrent 20 --timeout 3600",
		Flags: []core.FlagSpec{
			{Name: "directly-changed", Type: "string", Usage: "Space-separated list of directly changed modules"},
			{Name: "invalidated", Type: "string", Usage: "Space-separated list of invalidated (dependent) modules"},
			{Name: "head-sha", Type: "string", Usage: "Commit SHA to dispatch CI for"},
			{Name: "dispatch-ref", Type: "string", Usage: "Git ref to dispatch workflows on (default: current branch)"},
			{Name: "max-concurrent", Type: "int", DefaultValue: "6", Usage: "Maximum number of concurrent CI dispatches"},
			{Name: "timeout", Type: "int", DefaultValue: "3600", Usage: "Maximum time in seconds to wait for all CI"},
			{Name: "poll-interval", Type: "int", DefaultValue: "10", Usage: "How often to check for completed workflows (seconds)"},
			{Name: "trigger-run-id", Type: "string", Usage: "Run ID of the triggering workflow (for artifact download)"},
			{Name: "force-all-containers", Type: "bool", Usage: "Pass force-all-containers=true to dispatched workflows"},
			{Name: "mock", Type: "string", Usage: "Mock CI cache status (JSON format) for testing"},
		},
	}
}

func (c *pipelineCIScheduleCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return PipelineCISchedule()
}

// PipelineCISchedule is the command entry point for `pipeline ci schedule`.
func PipelineCISchedule() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Parse flags (args start at index 4 for "pipeline ci schedule").
	cfg := CISchedulerConfig{
		MaxConcurrent: 6,
		Timeout:       3600 * time.Second,
		PollInterval:  10 * time.Second,
		WorkspaceRoot: workspaceRoot,
	}

	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--directly-changed" && i+1 < len(os.Args):
			cfg.DirectlyChanged = os.Args[i+1]
			i++
		case arg == "--invalidated" && i+1 < len(os.Args):
			cfg.Invalidated = os.Args[i+1]
			i++
		case arg == "--head-sha" && i+1 < len(os.Args):
			cfg.HeadSHA = os.Args[i+1]
			i++
		case arg == "--dispatch-ref" && i+1 < len(os.Args):
			cfg.DispatchRef = os.Args[i+1]
			i++
		case arg == "--max-concurrent" && i+1 < len(os.Args):
			if v, parseErr := strconv.Atoi(os.Args[i+1]); parseErr == nil {
				cfg.MaxConcurrent = v
			}
			i++
		case arg == "--timeout" && i+1 < len(os.Args):
			if v, parseErr := strconv.Atoi(os.Args[i+1]); parseErr == nil {
				cfg.Timeout = time.Duration(v) * time.Second
			}
			i++
		case arg == "--poll-interval" && i+1 < len(os.Args):
			if v, parseErr := strconv.Atoi(os.Args[i+1]); parseErr == nil {
				cfg.PollInterval = time.Duration(v) * time.Second
			}
			i++
		case arg == "--trigger-run-id" && i+1 < len(os.Args):
			cfg.TriggerRunID = os.Args[i+1]
			i++
		case arg == "--force-all-containers":
			cfg.ForceAllContainers = true
		case arg == "--mock" && i+1 < len(os.Args):
			cfg.MockJSON = os.Args[i+1]
			i++
		case arg == "--help" || arg == "-h":
			printScheduleUsage()
			return 0
		}
	}

	// Detect HEAD SHA if not provided.
	if cfg.HeadSHA == "" {
		shaResult, shaErr := get.DetectCurrentSHA(workspaceRoot, "")
		if shaErr != nil {
			log.Errorf("Error: failed to detect SHA: %v", shaErr)
			return 1
		}
		cfg.HeadSHA = shaResult.SHA
	}

	// Filter modules via get.FilterCIDispatch logic.
	dispatchResult, filterErr := filterModulesForSchedule(cfg)
	if filterErr != nil {
		log.Errorf("Error filtering modules: %v", filterErr)
		return 1
	}

	// Create dispatcher.
	dispatcher := NewGHWorkflowDispatcher(workspaceRoot)

	// Build scheduler.
	scheduler := NewCIScheduler(cfg, dispatcher)
	scheduler.SetDispatchList(
		dispatchResult.Dispatch,
		dispatchResult.CIDependencies,
		dispatchResult.Skipped,
	)

	// Run.
	ctx := context.Background()
	result, schedErr := scheduler.Schedule(ctx)

	// Report results.
	printScheduleResult(result)
	writeScheduleSummary(result)

	if schedErr != nil {
		log.Errorf("CI Scheduler failed: %v", schedErr)
		return 1
	}

	log.Infof("CI Scheduler completed successfully in %v", result.TotalTime.Round(time.Second))
	return 0
}

// filterModulesForSchedule filters modules using the same logic as `get ci-dispatch`.
func filterModulesForSchedule(cfg CISchedulerConfig) (*get.CIDispatchResult, error) {
	// Build CICacheChecker.
	var checker *cache.CICacheChecker
	if cfg.MockJSON != "" {
		checker = get.MockCheckerFromJSON(cfg.MockJSON, cfg.HeadSHA)
	} else {
		api := github.NewGHClient(gh.New(tool.GlobalToolSystem(), cfg.WorkspaceRoot), cfg.WorkspaceRoot)
		querier := get.NewGHCIRunQuerier(api)
		checker = cache.NewCICacheChecker(querier, nil)
	}

	return get.FilterCIDispatch(cfg.DirectlyChanged, cfg.Invalidated, cfg.HeadSHA, checker, cfg.WorkspaceRoot)
}

func printScheduleResult(result *CIScheduleResult) {
	if result == nil {
		return
	}

	log.Info("")
	log.Info("=== CI Schedule Results ===")
	if len(result.Completed) > 0 {
		log.Infof("  Completed: %s", strings.Join(result.Completed, " "))
	}
	if len(result.Failed) > 0 {
		log.Warnf("  Failed: %s", strings.Join(result.Failed, " "))
	}
	if len(result.CascadeFailed) > 0 {
		log.Warnf("  Cascade-skipped: %s", strings.Join(result.CascadeFailed, " "))
	}
	if len(result.Cached) > 0 {
		log.Infof("  Cached (not dispatched): %s", strings.Join(result.Cached, " "))
	}
	log.Infof("  Total time: %v", result.TotalTime.Round(time.Second))
}

func writeScheduleSummary(result *CIScheduleResult) {
	if result == nil {
		return
	}

	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		return
	}

	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintln(f, "## CI Scheduler Results")
	fmt.Fprintln(f, "")

	totalDispatched := len(result.Dispatched)
	fmt.Fprintf(f, "Dispatched **%d** module(s) with concurrency-limited scheduling.\n\n", totalDispatched)

	if len(result.Completed) > 0 {
		modules := make([]string, len(result.Completed))
		for i, m := range result.Completed {
			modules[i] = fmt.Sprintf("`%s`", m)
		}
		fmt.Fprintf(f, "**Completed**: %s\n\n", strings.Join(modules, ", "))
	}

	if len(result.Failed) > 0 {
		modules := make([]string, len(result.Failed))
		for i, m := range result.Failed {
			modules[i] = fmt.Sprintf("`%s`", m)
		}
		fmt.Fprintf(f, "**Failed**: %s\n\n", strings.Join(modules, ", "))
	}

	if len(result.CascadeFailed) > 0 {
		modules := make([]string, len(result.CascadeFailed))
		for i, m := range result.CascadeFailed {
			modules[i] = fmt.Sprintf("`%s`", m)
		}
		fmt.Fprintf(f, "**Cascade-skipped**: %s\n\n", strings.Join(modules, ", "))
	}

	if len(result.Cached) > 0 {
		sorted := make([]string, len(result.Cached))
		copy(sorted, result.Cached)
		sort.Strings(sorted)
		modules := make([]string, len(sorted))
		for i, m := range sorted {
			modules[i] = fmt.Sprintf("`%s`", m)
		}
		fmt.Fprintf(f, "**Cached** (valid CI at HEAD): %s\n\n", strings.Join(modules, ", "))
	}

	fmt.Fprintf(f, "**Total time**: %v\n", result.TotalTime.Round(time.Second))
}

func printScheduleUsage() {
	log.Info("Schedule and dispatch CI workflows with concurrency limits")
	log.Info("")
	log.Info("Usage: pipeline ci schedule [flags]")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --directly-changed <modules>  Space-separated list of directly changed modules")
	log.Info("  --invalidated <modules>       Space-separated list of invalidated modules")
	log.Info("  --head-sha <sha>              Commit SHA to dispatch CI for (auto-detected)")
	log.Info("  --dispatch-ref <ref>          Git ref for workflow dispatch (default: current branch)")
	log.Info("  --max-concurrent <n>          Max concurrent dispatches (default: 6)")
	log.Info("  --timeout <seconds>           Max wait time (default: 3600)")
	log.Info("  --poll-interval <seconds>     Poll interval (default: 10)")
	log.Info("  --trigger-run-id <id>         Triggering workflow run ID")
	log.Info("  --force-all-containers        Pass force-all-containers=true to dispatched workflows")
	log.Info("  --mock <json>                 Mock CI cache status for testing")
	log.Info("  -h, --help                    Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  pipeline ci schedule --directly-changed \"core\" --invalidated \"eac docs\" --head-sha abc123")
	log.Info("  pipeline ci schedule --directly-changed \"core\" --max-concurrent 20 --timeout 3600")
}
