// Command: pipeline ci schedule
// Short: Schedule and dispatch CI workflows with concurrency limits
// Long: Replaces wave-based dispatch with a pull-based scheduler that:
// Long:   1. Filters modules (CI cache check, same as get ci-dispatch)
// Long:   2. Builds dependency graph from repository config
// Long:   3. Dispatches modules as capacity allows and deps are satisfied
// Long:   4. Polls for completion, dispatches next ready modules
// Long:   5. Exits 0 when all complete successfully, 1 on any failure
// Long:
// Long: This command is the CI-level analog of the local DependencyScheduler.
// Long: It handles the full dispatch lifecycle: filter, dispatch, poll, report.
// Long:
// Long: Example:
// Long:   pipeline ci schedule --directly-changed "core" --invalidated "eac docs" --head-sha abc123 --max-concurrent 6
// Long:   pipeline ci schedule --directly-changed "core" --max-concurrent 20 --timeout 3600
// Flag.directly-changed: type=string, usage=Space-separated list of directly changed modules
// Flag.invalidated: type=string, usage=Space-separated list of invalidated (dependent) modules
// Flag.head-sha: type=string, usage=Commit SHA to dispatch CI for
// Flag.dispatch-ref: type=string, usage=Git ref to dispatch workflows on (default: current branch)
// Flag.max-concurrent: type=int, default=6, usage=Maximum number of concurrent CI dispatches
// Flag.timeout: type=int, default=3600, usage=Maximum time in seconds to wait for all CI
// Flag.poll-interval: type=int, default=30, usage=How often to check for completed workflows (seconds)
// Flag.trigger-run-id: type=string, usage=Run ID of the triggering workflow (for artifact download)
// Flag.mock: type=string, usage=Mock CI cache status (JSON format) for testing
package ci

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/ghexec"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/github"
	"github.com/ready-to-release/eac/go/core/repository"
)

// CISchedulerConfig holds configuration for the CI scheduler.
type CISchedulerConfig struct {
	MaxConcurrent   int
	HeadSHA         string
	DispatchRef     string
	Timeout         time.Duration
	PollInterval    time.Duration
	TriggerRunID    string
	DirectlyChanged string
	Invalidated     string
	WorkspaceRoot   string
	MockJSON        string
}

// CIModuleStatus represents the status of a module in the scheduler.
type CIModuleStatus int

const (
	ciModulePending   CIModuleStatus = iota // Not yet dispatched
	ciModuleActive                          // Dispatched, running
	ciModuleCompleted                       // Completed successfully
	ciModuleFailed                          // Failed
	ciModuleSkipped                         // Skipped (cascade-failed)
)

// CIScheduleResult holds the result of a full CI scheduling run.
type CIScheduleResult struct {
	Dispatched    []string
	Completed     []string
	Failed        []string
	CascadeFailed []string
	Cached        []string
	TotalTime     time.Duration
}

// CIScheduler orchestrates CI workflow dispatch with concurrency limits.
type CIScheduler struct {
	cfg        CISchedulerConfig
	dispatcher CIWorkflowDispatcher

	// State built from filterCIDispatch
	dispatch []string            // Modules to dispatch
	deps     map[string][]string // CI dependency graph (module -> deps)
	cached   []string            // Modules skipped due to CI cache

	// Runtime state
	status       map[string]CIModuleStatus
	dispatchTime map[string]time.Time
}

// NewCIScheduler creates a scheduler with the given configuration and dispatcher.
func NewCIScheduler(cfg CISchedulerConfig, dispatcher CIWorkflowDispatcher) *CIScheduler {
	return &CIScheduler{
		cfg:          cfg,
		dispatcher:   dispatcher,
		status:       make(map[string]CIModuleStatus),
		dispatchTime: make(map[string]time.Time),
	}
}

// SetDispatchList sets the modules and dependencies to schedule.
// This is called by Schedule() after filtering, or directly in tests.
func (s *CIScheduler) SetDispatchList(dispatch []string, deps map[string][]string, cached []string) {
	s.dispatch = dispatch
	s.deps = deps
	s.cached = cached

	// Initialize all dispatch modules as pending.
	for _, m := range dispatch {
		s.status[m] = ciModulePending
	}
}

// Schedule runs the full dispatch lifecycle.
func (s *CIScheduler) Schedule(ctx context.Context) (*CIScheduleResult, error) {
	startTime := time.Now()

	// Nothing to dispatch.
	if len(s.dispatch) == 0 {
		return &CIScheduleResult{
			Cached:    s.cached,
			TotalTime: time.Since(startTime),
		}, nil
	}

	log.Infof("CI Scheduler: %d module(s) to dispatch, max-concurrent=%d", len(s.dispatch), s.cfg.MaxConcurrent)
	if len(s.cached) > 0 {
		log.Infof("CI Scheduler: %d module(s) skipped (cached): %s", len(s.cached), strings.Join(s.cached, " "))
	}
	log.Infof("CI Scheduler: dispatch order: %s", strings.Join(s.dispatch, " "))
	if len(s.deps) > 0 {
		log.Infof("CI Scheduler: dependencies: %v", s.deps)
	}

	for {
		// Check timeout.
		if time.Since(startTime) > s.cfg.Timeout {
			return s.buildResult(startTime), fmt.Errorf("timeout after %v", s.cfg.Timeout)
		}

		// Dispatch ready modules up to concurrency limit.
		dispatched := s.dispatchReady(ctx)
		if dispatched > 0 {
			log.Infof("CI Scheduler: dispatched %d module(s)", dispatched)
		}

		// Check if we're done.
		activeCount := s.countByStatus(ciModuleActive)
		pendingCount := s.countByStatus(ciModulePending)

		if activeCount == 0 && pendingCount == 0 {
			// All done.
			result := s.buildResult(startTime)
			if len(result.Failed) > 0 || len(result.CascadeFailed) > 0 {
				return result, fmt.Errorf("%d module(s) failed, %d cascade-skipped",
					len(result.Failed), len(result.CascadeFailed))
			}
			return result, nil
		}

		// Poll active dispatches.
		time.Sleep(s.cfg.PollInterval)

		if err := s.pollActive(ctx); err != nil {
			return s.buildResult(startTime), fmt.Errorf("poll error: %w", err)
		}
	}
}

// isReady returns true if all deps of module are satisfied (completed or cached).
func (s *CIScheduler) isReady(module string) bool {
	if s.status[module] != ciModulePending {
		return false
	}

	deps, ok := s.deps[module]
	if !ok {
		return true // No deps, always ready
	}

	// Deps are satisfied if they are completed OR not in the dispatch set (meaning cached).
	dispatchSet := make(map[string]bool, len(s.dispatch))
	for _, m := range s.dispatch {
		dispatchSet[m] = true
	}

	for _, dep := range deps {
		if !dispatchSet[dep] {
			continue // Not in dispatch set = cached/external, considered satisfied
		}
		if s.status[dep] != ciModuleCompleted {
			return false
		}
	}
	return true
}

// dispatchReady dispatches all ready modules up to the concurrency limit.
// Returns the number of modules dispatched.
func (s *CIScheduler) dispatchReady(ctx context.Context) int {
	activeCount := s.countByStatus(ciModuleActive)
	available := s.cfg.MaxConcurrent - activeCount
	if available <= 0 {
		return 0
	}

	dispatched := 0
	for _, module := range s.dispatch {
		if available <= 0 {
			break
		}
		if !s.isReady(module) {
			continue
		}

		log.Infof("CI Scheduler: dispatching %s", module)
		err := s.dispatcher.Dispatch(ctx, module, s.cfg.DispatchRef, s.cfg.HeadSHA, s.cfg.TriggerRunID)
		if err != nil {
			log.Errorf("CI Scheduler: failed to dispatch %s: %v", module, err)
			s.status[module] = ciModuleFailed
			s.cascadeFail(module)
			continue
		}

		s.status[module] = ciModuleActive
		s.dispatchTime[module] = time.Now()
		dispatched++
		available--
	}

	return dispatched
}

// pollActive checks status of all active dispatches.
func (s *CIScheduler) pollActive(ctx context.Context) error {
	var activeModules []string
	for _, m := range s.dispatch {
		if s.status[m] == ciModuleActive {
			activeModules = append(activeModules, m)
		}
	}

	for _, module := range activeModules {
		status, conclusion, err := s.dispatcher.GetStatus(ctx, module, s.cfg.HeadSHA)
		if err != nil {
			log.Warnf("CI Scheduler: error checking %s: %v", module, err)
			continue
		}

		switch status {
		case "completed":
			elapsed := time.Since(s.dispatchTime[module]).Round(time.Second)
			if conclusion == "success" || conclusion == "skipped" {
				log.Infof("CI Scheduler: %s completed successfully (%v)", module, elapsed)
				s.status[module] = ciModuleCompleted
			} else {
				log.Warnf("CI Scheduler: %s failed (conclusion=%s, %v)", module, conclusion, elapsed)
				s.status[module] = ciModuleFailed
				s.cascadeFail(module)
			}
		case "in_progress":
			// Still running, nothing to do.
		case "none":
			// Not yet visible in GitHub API. This can happen right after dispatch.
			// If it's been too long, we might have a problem, but for now just wait.
		}
	}

	// Log current state.
	active := s.countByStatus(ciModuleActive)
	pending := s.countByStatus(ciModulePending)
	completed := s.countByStatus(ciModuleCompleted)
	if active > 0 || pending > 0 {
		log.Infof("CI Scheduler: active=%d pending=%d completed=%d", active, pending, completed)
	}

	return nil
}

// cascadeFail marks all dependents of a failed module as skipped.
func (s *CIScheduler) cascadeFail(failedModule string) {
	// Build reverse dependency map: module -> modules that depend on it.
	reverseDeps := make(map[string][]string)
	for m, deps := range s.deps {
		for _, dep := range deps {
			reverseDeps[dep] = append(reverseDeps[dep], m)
		}
	}

	// BFS from failed module through reverse deps.
	queue := []string{failedModule}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dependent := range reverseDeps[current] {
			if s.status[dependent] == ciModulePending || s.status[dependent] == ciModuleActive {
				log.Warnf("CI Scheduler: cascade-skipping %s (depends on failed %s)", dependent, failedModule)
				s.status[dependent] = ciModuleSkipped
				queue = append(queue, dependent)
			}
		}
	}
}

// countByStatus counts modules with the given status.
func (s *CIScheduler) countByStatus(status CIModuleStatus) int {
	count := 0
	for _, m := range s.dispatch {
		if s.status[m] == status {
			count++
		}
	}
	return count
}

// buildResult constructs the final result from current state.
func (s *CIScheduler) buildResult(startTime time.Time) *CIScheduleResult {
	result := &CIScheduleResult{
		Cached:    s.cached,
		TotalTime: time.Since(startTime),
	}

	for _, m := range s.dispatch {
		switch s.status[m] {
		case ciModuleCompleted:
			result.Completed = append(result.Completed, m)
			result.Dispatched = append(result.Dispatched, m)
		case ciModuleFailed:
			result.Failed = append(result.Failed, m)
			result.Dispatched = append(result.Dispatched, m)
		case ciModuleSkipped:
			result.CascadeFailed = append(result.CascadeFailed, m)
		case ciModuleActive:
			result.Dispatched = append(result.Dispatched, m)
		}
	}

	return result
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
		PollInterval:  30 * time.Second,
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
	var dispatcher CIWorkflowDispatcher
	dispatcher = NewGHWorkflowDispatcher(workspaceRoot)

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
		api := github.NewGHClient(ghexec.New(cfg.WorkspaceRoot), cfg.WorkspaceRoot)
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
	log.Info("  --poll-interval <seconds>     Poll interval (default: 30)")
	log.Info("  --trigger-run-id <id>         Triggering workflow run ID")
	log.Info("  --mock <json>                 Mock CI cache status for testing")
	log.Info("  -h, --help                    Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  pipeline ci schedule --directly-changed \"core\" --invalidated \"eac docs\" --head-sha abc123")
	log.Info("  pipeline ci schedule --directly-changed \"core\" --max-concurrent 20 --timeout 3600")
}
