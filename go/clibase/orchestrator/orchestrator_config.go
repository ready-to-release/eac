package orchestrator

import "time"

// UpdateConfig updates the orchestrator's configuration with loaded values.
// This is used when the orchestrator is created early (before config is loaded)
// and needs to be updated with actual values once available.
// Only non-zero/non-nil fields are applied.
func (o *Orchestrator) UpdateConfig(update ConfigUpdate) {
	if update.WorkspaceRoot != "" {
		o.config.WorkspaceRoot = update.WorkspaceRoot
	}
	if update.OutputBaseDir != "" {
		o.config.OutputBaseDir = update.OutputBaseDir
	}
	if update.MaxConcurrency > 0 {
		o.config.MaxConcurrency = update.MaxConcurrency
	}
	if update.Turbo > 0 {
		o.config.Turbo = update.Turbo
	}
	if update.ComponentTypesDisplay != nil {
		o.config.ComponentTypesDisplay = update.ComponentTypesDisplay
	}
	if update.ShowTimings {
		o.config.ShowTimings = true
	}
	if update.DryRun {
		o.config.DryRun = true
	}
}

// SetComponentTypesDisplay updates the component types display map in the config.
// Useful when component types are determined after orchestrator creation.
func (o *Orchestrator) SetComponentTypesDisplay(componentTypes map[string]string) {
	o.config.ComponentTypesDisplay = componentTypes
}

// SetMaxConcurrency updates the maximum concurrency for subsequent Run calls.
// Useful for running sequential tests after parallel tests.
func (o *Orchestrator) SetMaxConcurrency(maxConcurrency int) {
	o.config.MaxConcurrency = maxConcurrency
}

// SetCacheTimes sets the cache times for modules that are up-to-date (cached).
// These times are passed to the TUI to display when cached artifacts were built.
// If the scheduler hasn't been created yet, times are stored and applied when it starts.
func (o *Orchestrator) SetCacheTimes(times map[string]time.Time) {
	if o.currentScheduler != nil {
		o.currentScheduler.SetCacheTimes(times)
	} else {
		// Store for later application when scheduler is created
		o.pendingCacheTimes = times
	}
}

// SetCacheDetection configures background cache detection for early TUI feedback.
// When set, cached tabs will progressively "light up" blue as detection completes,
// and workers will short-circuit for already-detected cached items.
//
// verifier: function to check if a component is cached
// cachedModules: pre-computed set of modules known to be cached
//
// If the scheduler hasn't been created yet, config is stored and applied when it starts.
func (o *Orchestrator) SetCacheDetection(verifier CacheVerifier, cachedModules map[string]bool) {
	if o.currentScheduler != nil {
		o.currentScheduler.SetCacheDetection(verifier, cachedModules)
	} else {
		// Store for later application when scheduler is created
		o.pendingCacheVerifier = verifier
		o.pendingCachedModules = cachedModules
	}
}

// SetSummaryBuilder sets the summary builder for incremental summary computation.
// The builder receives component results as they complete, enabling parallel
// summary computation during execution.
func (o *Orchestrator) SetSummaryBuilder(builder SummaryBuilder) {
	o.summaryBuilder = builder
}
