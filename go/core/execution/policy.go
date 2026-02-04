package execution

import (
	"context"
	"time"

	"github.com/ready-to-release/eac/go/core/workunit"
)

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
