package execution

import (
	"context"
	"time"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// LayerPolicy is the port interface for execution ordering.
// Adapters (orchestrator) query this to know what/when to execute.
//
// This interface defines the domain's contract for execution scheduling:
//   - ComputePlan: builds the execution plan from work units
//   - IsReady: checks if a specific unit can execute now
//   - GetReadyUnits: returns all units that can execute now
//
// Implementations should be thread-safe if used concurrently.
type LayerPolicy interface {
	// ComputePlan creates a LayerPlan from work units.
	// This is the domain's single source of truth for execution order.
	//
	// The plan organizes units into nested layers:
	// - Module layers: inter-module dependency ordering
	// - Component layers: intra-module build_after ordering
	//
	// Returns error if:
	// - Circular dependencies are detected
	// - Required modules/components are not found
	ComputePlan(units []workunit.UnitSpec) (*LayerPlan, error)

	// IsReady checks if a unit can execute given current completion state.
	// Encapsulates all layer constraint logic in the domain.
	//
	// The check depends on the LayerMode:
	// - Strict: module layers + component layers + DependsOn
	// - ModuleOnly: module layers + DependsOn
	// - None: only DependsOn
	//
	// Parameters:
	//   - plan: the execution plan from ComputePlan
	//   - unitID: the unit to check
	//   - completed: map of unit longnames that have completed (success or failure)
	//
	// Returns true if the unit can execute now.
	IsReady(plan *LayerPlan, unitID workunit.UnitID, completed map[string]bool) bool

	// GetReadyUnits returns all units that can execute now.
	// This is a convenience method for batch scheduling.
	//
	// Returns units in priority order (implementation-defined).
	// Typically returns heavier units first (LPT scheduling).
	GetReadyUnits(plan *LayerPlan, completed map[string]bool) []workunit.UnitSpec
}

// ReadyChecker is a simpler interface for checking unit readiness.
// Used when the full LayerPolicy is not needed.
type ReadyChecker interface {
	// IsReady returns true if the unit can execute given completion state.
	IsReady(unitID workunit.UnitID, completed map[string]bool) bool
}

// CompletionTracker tracks which units have completed.
// This is a simple interface for tracking completion state.
type CompletionTracker interface {
	// MarkComplete marks a unit as completed successfully.
	MarkComplete(id workunit.UnitID)

	// MarkFailed marks a unit as failed.
	// Failed units are considered "complete" for dependency resolution
	// but their dependents should also fail.
	MarkFailed(id workunit.UnitID)

	// IsComplete returns true if the unit has completed (success or failure).
	IsComplete(id workunit.UnitID) bool

	// IsFailed returns true if the unit failed.
	IsFailed(id workunit.UnitID) bool

	// Completed returns a snapshot of all completed units.
	// The map key is UnitID.Longname().
	Completed() map[string]bool
}

// PlanExecutor combines planning and execution tracking.
// This is the full interface for managing execution state.
type PlanExecutor interface {
	LayerPolicy
	CompletionTracker

	// Mode returns the current LayerMode.
	Mode() LayerMode
}

// CacheResult represents the cache status of a work unit.
type CacheResult struct {
	// Cached indicates whether the unit's output is cached and valid.
	Cached bool
	// CacheTime is when the cache was created/validated.
	// Zero value if not cached.
	CacheTime time.Time
}

// CacheVerifier is the port interface for verifying work unit cache status.
// Commands implement this interface to provide their own cache verification logic.
// The orchestrator uses this interface for background cache detection.
//
// Implementations should be safe for concurrent use - multiple goroutines
// may call Verify simultaneously for different units.
type CacheVerifier interface {
	// Verify checks if a work unit's output is cached and valid.
	// Returns the cache status and any error encountered during verification.
	//
	// The context should be used for cancellation - long-running verification
	// should respect ctx.Done().
	//
	// Implementations should not modify shared state during verification.
	Verify(ctx context.Context, unit workunit.UnitSpec) (CacheResult, error)
}

// CacheVerifierFunc is an adapter that allows ordinary functions to be used
// as CacheVerifier implementations.
type CacheVerifierFunc func(ctx context.Context, unit workunit.UnitSpec) (CacheResult, error)

// Verify implements CacheVerifier by calling the underlying function.
func (f CacheVerifierFunc) Verify(ctx context.Context, unit workunit.UnitSpec) (CacheResult, error) {
	return f(ctx, unit)
}
