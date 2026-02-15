package ci

import (
	"context"
	"fmt"
	"strings"
	"time"
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
