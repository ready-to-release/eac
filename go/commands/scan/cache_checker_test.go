package scan

import (
	"context"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// newScanUnit creates a workunit.UnitSpec for scan testing with the given module, component, and scanner.
func newScanUnit(module, component, scanner string) workunit.UnitSpec {
	return workunit.UnitSpec{
		ID: workunit.UnitID{
			Action:        core.ActionScan,
			Module:        module,
			ComponentType: component,
			ComponentName: component,
			Tool:          scanner,
		},
	}
}

func TestScanCacheVerifier_UoWCacheHit(t *testing.T) {
	cachedTime := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	unit := newScanUnit("mod-a", "go", "trivy")
	longname := unit.ID.Longname() // "scan:mod-a:go:trivy"

	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{longname: true},
		uowCacheTimes: map[string]time.Time{longname: cachedTime},
	}

	result, err := verifier.Verify(context.Background(), unit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if !result.Cached {
		t.Error("expected Cached=true for UoW cache hit, got false")
	}
	if !result.CacheTime.Equal(cachedTime) {
		t.Errorf("CacheTime = %v, want %v", result.CacheTime, cachedTime)
	}
}

func TestScanCacheVerifier_UoWCacheMiss(t *testing.T) {
	cachedTime := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	cachedUnit := newScanUnit("mod-a", "go", "trivy")
	uncachedUnit := newScanUnit("mod-a", "go", "grype")

	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{cachedUnit.ID.Longname(): true},
		uowCacheTimes: map[string]time.Time{cachedUnit.ID.Longname(): cachedTime},
	}

	result, err := verifier.Verify(context.Background(), uncachedUnit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if result.Cached {
		t.Error("expected Cached=false for UoW cache miss, got true")
	}
	if !result.CacheTime.IsZero() {
		t.Errorf("CacheTime should be zero for cache miss, got %v", result.CacheTime)
	}
}

func TestScanCacheVerifier_ModuleFallbackHit(t *testing.T) {
	unit := newScanUnit("mod-b", "go", "trivy")

	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{},
		uowCacheTimes: map[string]time.Time{},
		cachedModules: map[string]bool{"mod-b": true},
	}

	result, err := verifier.Verify(context.Background(), unit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if !result.Cached {
		t.Error("expected Cached=true for module-level fallback, got false")
	}
	if !result.CacheTime.IsZero() {
		t.Errorf("CacheTime should be zero for module-level hit, got %v", result.CacheTime)
	}
}

func TestScanCacheVerifier_CompletelyUncached(t *testing.T) {
	unit := newScanUnit("mod-c", "go", "trivy")

	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{},
		uowCacheTimes: map[string]time.Time{},
		cachedModules: map[string]bool{},
	}

	result, err := verifier.Verify(context.Background(), unit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if result.Cached {
		t.Error("expected Cached=false for completely uncached unit, got true")
	}
	if !result.CacheTime.IsZero() {
		t.Errorf("CacheTime should be zero for uncached unit, got %v", result.CacheTime)
	}
}

func TestScanCacheVerifier_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	unit := newScanUnit("mod-a", "go", "trivy")

	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{unit.ID.Longname(): true},
		uowCacheTimes: map[string]time.Time{unit.ID.Longname(): time.Now()},
	}

	result, err := verifier.Verify(ctx, unit)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if result.Cached {
		t.Error("expected Cached=false when context is cancelled")
	}
}

func TestScanCacheVerifier_NilMaps(t *testing.T) {
	unit := newScanUnit("mod-a", "go", "trivy")

	verifier := &ScanCacheVerifier{} // all maps nil

	result, err := verifier.Verify(context.Background(), unit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error with nil maps: %v", err)
	}
	if result.Cached {
		t.Error("expected Cached=false for verifier with nil maps, got true")
	}
	if !result.CacheTime.IsZero() {
		t.Errorf("CacheTime should be zero for nil maps, got %v", result.CacheTime)
	}
}

func TestScanCacheVerifier_CacheTimeReturned(t *testing.T) {
	times := []time.Time{
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		time.Now().UTC().Truncate(time.Second),
	}

	for _, cachedTime := range times {
		unit := newScanUnit("mod-a", "go", "trivy")
		longname := unit.ID.Longname()

		verifier := &ScanCacheVerifier{
			cachedUoWs:    map[string]bool{longname: true},
			uowCacheTimes: map[string]time.Time{longname: cachedTime},
		}

		result, err := verifier.Verify(context.Background(), unit)
		if err != nil {
			t.Fatalf("Verify() returned unexpected error: %v", err)
		}
		if !result.CacheTime.Equal(cachedTime) {
			t.Errorf("CacheTime = %v, want %v", result.CacheTime, cachedTime)
		}
	}
}

func TestScanCacheVerifier_UoWTakesPrecedenceOverModule(t *testing.T) {
	uowTime := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	unit := newScanUnit("mod-a", "go", "trivy")
	longname := unit.ID.Longname()

	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{longname: true},
		uowCacheTimes: map[string]time.Time{longname: uowTime},
		cachedModules: map[string]bool{"mod-a": true},
	}

	result, err := verifier.Verify(context.Background(), unit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if !result.Cached {
		t.Error("expected Cached=true, got false")
	}
	// UoW-level cache should return the specific cache time, not zero.
	if !result.CacheTime.Equal(uowTime) {
		t.Errorf("CacheTime = %v, want %v (UoW should take precedence over module)", result.CacheTime, uowTime)
	}
}

func TestScanCacheVerifier_UoWFalseDoesNotBlockModuleFallback(t *testing.T) {
	unit := newScanUnit("mod-a", "go", "trivy")
	longname := unit.ID.Longname()

	// UoW entry exists but is false; module-level says cached.
	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{longname: false},
		uowCacheTimes: map[string]time.Time{},
		cachedModules: map[string]bool{"mod-a": true},
	}

	result, err := verifier.Verify(context.Background(), unit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	// cachedUoWs[longname] is false, so UoW check fails; falls through to module check.
	if !result.Cached {
		t.Error("expected Cached=true from module fallback when UoW entry is false")
	}
}

func TestScanCacheVerifier_ModuleFalseIsNotCached(t *testing.T) {
	unit := newScanUnit("mod-a", "go", "trivy")

	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{},
		uowCacheTimes: map[string]time.Time{},
		cachedModules: map[string]bool{"mod-a": false},
	}

	result, err := verifier.Verify(context.Background(), unit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if result.Cached {
		t.Error("expected Cached=false when module entry is false")
	}
}

func TestScanCacheVerifier_ContextDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	unit := newScanUnit("mod-a", "go", "trivy")

	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{unit.ID.Longname(): true},
		uowCacheTimes: map[string]time.Time{unit.ID.Longname(): time.Now()},
	}

	_, err := verifier.Verify(ctx, unit)
	if err == nil {
		t.Fatal("expected error from expired deadline, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestScanCacheVerifier_MultipleUnits(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	earlier := now.Add(-1 * time.Hour)

	unitA := newScanUnit("mod-a", "go", "trivy")
	unitB := newScanUnit("mod-a", "go", "grype")
	unitC := newScanUnit("mod-b", "docker", "trivy")
	unitD := newScanUnit("mod-c", "npm", "npm-audit")

	verifier := &ScanCacheVerifier{
		cachedUoWs: map[string]bool{
			unitA.ID.Longname(): true,
			unitB.ID.Longname(): true,
		},
		uowCacheTimes: map[string]time.Time{
			unitA.ID.Longname(): now,
			unitB.ID.Longname(): earlier,
		},
		cachedModules: map[string]bool{
			"mod-b": true,
		},
	}

	tests := []struct {
		name      string
		unit      workunit.UnitSpec
		wantCache bool
		wantTime  time.Time
	}{
		{
			name:      "unitA UoW cached",
			unit:      unitA,
			wantCache: true,
			wantTime:  now,
		},
		{
			name:      "unitB UoW cached with earlier time",
			unit:      unitB,
			wantCache: true,
			wantTime:  earlier,
		},
		{
			name:      "unitC module fallback cached",
			unit:      unitC,
			wantCache: true,
			wantTime:  time.Time{},
		},
		{
			name:      "unitD completely uncached",
			unit:      unitD,
			wantCache: false,
			wantTime:  time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := verifier.Verify(context.Background(), tt.unit)
			if err != nil {
				t.Fatalf("Verify() returned unexpected error: %v", err)
			}
			if result.Cached != tt.wantCache {
				t.Errorf("Cached = %v, want %v", result.Cached, tt.wantCache)
			}
			if !result.CacheTime.Equal(tt.wantTime) {
				t.Errorf("CacheTime = %v, want %v", result.CacheTime, tt.wantTime)
			}
		})
	}
}

func TestScanCacheVerifier_MissingCacheTimeForCachedUoW(t *testing.T) {
	unit := newScanUnit("mod-a", "go", "trivy")
	longname := unit.ID.Longname()

	// UoW is marked cached but no entry in uowCacheTimes.
	verifier := &ScanCacheVerifier{
		cachedUoWs:    map[string]bool{longname: true},
		uowCacheTimes: map[string]time.Time{},
	}

	result, err := verifier.Verify(context.Background(), unit)
	if err != nil {
		t.Fatalf("Verify() returned unexpected error: %v", err)
	}
	if !result.Cached {
		t.Error("expected Cached=true, got false")
	}
	// Missing time entry yields zero time from map lookup.
	if !result.CacheTime.IsZero() {
		t.Errorf("expected zero CacheTime when uowCacheTimes has no entry, got %v", result.CacheTime)
	}
}
