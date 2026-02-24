package ci

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDispatcher implements CIWorkflowDispatcher for testing.
type mockDispatcher struct {
	mu sync.Mutex

	// dispatched records the order of dispatched modules.
	dispatched []string

	// statusMap maps module -> (status, conclusion) to return from GetStatus.
	// Use setStatus() to change during a test.
	statusMap map[string]statusEntry

	// dispatchErr if set, Dispatch returns this error for any module.
	dispatchErr error

	// dispatchErrModules if set, Dispatch returns error only for these modules.
	dispatchErrModules map[string]error
}

type statusEntry struct {
	status     string
	conclusion string
}

func newMockDispatcher() *mockDispatcher {
	return &mockDispatcher{
		statusMap: make(map[string]statusEntry),
	}
}

func (m *mockDispatcher) Dispatch(_ context.Context, module, ref, sha, triggerRunID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dispatchErrModules != nil {
		if err, ok := m.dispatchErrModules[module]; ok {
			return err
		}
	}
	if m.dispatchErr != nil {
		return m.dispatchErr
	}
	m.dispatched = append(m.dispatched, module)

	// Default: after dispatch, module appears as in_progress.
	if _, ok := m.statusMap[module]; !ok {
		m.statusMap[module] = statusEntry{status: "in_progress"}
	}
	return nil
}

func (m *mockDispatcher) GetStatus(_ context.Context, module, sha string) (string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.statusMap[module]
	if !ok {
		return "none", "", nil
	}
	return entry.status, entry.conclusion, nil
}

func (m *mockDispatcher) BatchGetStatus(_ context.Context, modules []string, sha string) (map[string]ModuleRunStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]ModuleRunStatus, len(modules))
	for _, module := range modules {
		entry, ok := m.statusMap[module]
		if !ok {
			continue
		}
		result[module] = ModuleRunStatus{Status: entry.status, Conclusion: entry.conclusion}
	}
	return result, nil
}

func (m *mockDispatcher) setStatus(module, status, conclusion string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusMap[module] = statusEntry{status: status, conclusion: conclusion}
}

func (m *mockDispatcher) getDispatched() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.dispatched))
	copy(result, m.dispatched)
	return result
}

// helper to create a scheduler with fast polling for tests.
func testScheduler(maxConcurrent int, dispatcher CIWorkflowDispatcher) *CIScheduler {
	cfg := CISchedulerConfig{
		MaxConcurrent: maxConcurrent,
		HeadSHA:       "abc123",
		DispatchRef:   "main",
		Timeout:       5 * time.Second,
		PollInterval:  10 * time.Millisecond,
	}
	return NewCIScheduler(cfg, dispatcher)
}

func TestScheduler_EmptyDispatchList_ExitsImmediately(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	s.SetDispatchList(nil, nil, nil)

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result.Dispatched)
	assert.Empty(t, result.Failed)
	assert.Empty(t, result.CascadeFailed)
}

func TestScheduler_AllCached_ExitsImmediately(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	s.SetDispatchList(nil, nil, []string{"core", "docs"})

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result.Dispatched)
	assert.Equal(t, []string{"core", "docs"}, result.Cached)
}

func TestScheduler_NoDeps_AllDispatchedImmediately(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	s.SetDispatchList([]string{"core", "docs", "eac"}, nil, nil)

	// Complete all modules after a brief delay.
	go func() {
		time.Sleep(50 * time.Millisecond)
		d.setStatus("core", "completed", "success")
		d.setStatus("docs", "completed", "success")
		d.setStatus("eac", "completed", "success")
	}()

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"core", "docs", "eac"}, result.Dispatched)
	assert.ElementsMatch(t, []string{"core", "docs", "eac"}, result.Completed)
	assert.Empty(t, result.Failed)
}

func TestScheduler_WithDeps_RespectsOrder(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)

	// eac depends on core; docs depends on nothing.
	deps := map[string][]string{
		"eac": {"core"},
	}
	s.SetDispatchList([]string{"core", "docs", "eac"}, deps, nil)

	// Complete core and docs quickly, eac after that.
	go func() {
		time.Sleep(30 * time.Millisecond)
		d.setStatus("core", "completed", "success")
		d.setStatus("docs", "completed", "success")

		// Wait for eac to be dispatched then complete it.
		time.Sleep(100 * time.Millisecond)
		d.setStatus("eac", "completed", "success")
	}()

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"core", "docs", "eac"}, result.Completed)

	// Verify core was dispatched before eac.
	dispatched := d.getDispatched()
	coreIdx := indexOf(dispatched, "core")
	eacIdx := indexOf(dispatched, "eac")
	assert.True(t, coreIdx < eacIdx, "core should be dispatched before eac, got order: %v", dispatched)
}

func TestScheduler_ConcurrencyLimit_Respected(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(2, d) // Only allow 2 concurrent.

	s.SetDispatchList([]string{"a", "b", "c", "d"}, nil, nil)

	// Track dispatch timing to verify concurrency.
	go func() {
		// Wait until first two are dispatched.
		for {
			if len(d.getDispatched()) >= 2 {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}

		// At this point only 2 should be dispatched.
		dispatched := d.getDispatched()
		assert.LessOrEqual(t, len(dispatched), 2, "should have at most 2 dispatched initially")

		// Complete the first two.
		d.setStatus(dispatched[0], "completed", "success")
		d.setStatus(dispatched[1], "completed", "success")

		// Wait for the next batch.
		time.Sleep(50 * time.Millisecond)

		// Complete remaining.
		for _, m := range d.getDispatched() {
			d.setStatus(m, "completed", "success")
		}
	}()

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Completed, 4)
}

func TestScheduler_CachedDeps_PreMarkedComplete(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)

	// eac depends on core, but core is cached (not in dispatch set).
	deps := map[string][]string{
		"eac": {"core"},
	}
	// core is NOT in dispatch list (it's cached), eac IS.
	s.SetDispatchList([]string{"eac"}, deps, []string{"core"})

	// eac should be immediately ready since core is not in dispatch set.
	go func() {
		time.Sleep(30 * time.Millisecond)
		d.setStatus("eac", "completed", "success")
	}()

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"eac"}, result.Completed)
	assert.Equal(t, []string{"core"}, result.Cached)

	// Verify eac was dispatched (core should NOT have been dispatched).
	dispatched := d.getDispatched()
	assert.Equal(t, []string{"eac"}, dispatched)
}

func TestScheduler_FailedModule_CascadesFailure(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)

	// Chain: docs depends on eac, eac depends on core.
	deps := map[string][]string{
		"eac":  {"core"},
		"docs": {"eac"},
	}
	s.SetDispatchList([]string{"core", "eac", "docs"}, deps, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		// core fails.
		d.setStatus("core", "completed", "failure")
	}()

	result, err := s.Schedule(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")

	assert.Equal(t, []string{"core"}, result.Failed)
	// Both eac and docs should be cascade-failed.
	assert.ElementsMatch(t, []string{"eac", "docs"}, result.CascadeFailed)
}

func TestScheduler_Timeout_ReturnsError(t *testing.T) {
	d := newMockDispatcher()
	cfg := CISchedulerConfig{
		MaxConcurrent: 6,
		HeadSHA:       "abc123",
		DispatchRef:   "main",
		Timeout:       100 * time.Millisecond, // Very short timeout.
		PollInterval:  10 * time.Millisecond,
	}
	s := NewCIScheduler(cfg, d)
	s.SetDispatchList([]string{"slow"}, nil, nil)

	// Never complete the module — should hit timeout.
	result, err := s.Schedule(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
	assert.NotNil(t, result)
}

func TestScheduler_DispatchError_MarksFailedAndCascades(t *testing.T) {
	d := newMockDispatcher()
	d.dispatchErrModules = map[string]error{
		"core": fmt.Errorf("dispatch failed: rate limited"),
	}
	s := testScheduler(6, d)

	deps := map[string][]string{
		"eac": {"core"},
	}
	s.SetDispatchList([]string{"core", "eac", "docs"}, deps, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		d.setStatus("docs", "completed", "success")
	}()

	result, err := s.Schedule(context.Background())
	require.Error(t, err)

	assert.Equal(t, []string{"core"}, result.Failed)
	assert.Equal(t, []string{"eac"}, result.CascadeFailed)
	assert.Equal(t, []string{"docs"}, result.Completed)
}

func TestScheduler_DiamondDependency(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)

	// Diamond: A has no deps; B and C depend on A; D depends on B and C.
	deps := map[string][]string{
		"B": {"A"},
		"C": {"A"},
		"D": {"B", "C"},
	}
	s.SetDispatchList([]string{"A", "B", "C", "D"}, deps, nil)

	go func() {
		// A completes first.
		time.Sleep(30 * time.Millisecond)
		d.setStatus("A", "completed", "success")

		// B and C can now run; complete them.
		time.Sleep(50 * time.Millisecond)
		d.setStatus("B", "completed", "success")
		d.setStatus("C", "completed", "success")

		// D can now run; complete it.
		time.Sleep(50 * time.Millisecond)
		d.setStatus("D", "completed", "success")
	}()

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"A", "B", "C", "D"}, result.Completed)

	// Verify ordering: A before B and C, B and C before D.
	dispatched := d.getDispatched()
	aIdx := indexOf(dispatched, "A")
	bIdx := indexOf(dispatched, "B")
	cIdx := indexOf(dispatched, "C")
	dIdx := indexOf(dispatched, "D")

	assert.True(t, aIdx < bIdx, "A before B")
	assert.True(t, aIdx < cIdx, "A before C")
	assert.True(t, bIdx < dIdx, "B before D")
	assert.True(t, cIdx < dIdx, "C before D")
}

func TestScheduler_SkippedConclusionTreatedAsSuccess(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)

	s.SetDispatchList([]string{"core"}, nil, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		d.setStatus("core", "completed", "skipped")
	}()

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"core"}, result.Completed)
}

func TestScheduler_SingleModule_NoTimeout(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(1, d)
	s.SetDispatchList([]string{"core"}, nil, nil)

	go func() {
		time.Sleep(20 * time.Millisecond)
		d.setStatus("core", "completed", "success")
	}()

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"core"}, result.Completed)
	assert.Equal(t, []string{"core"}, result.Dispatched)
}

func TestScheduler_isReady_NoDeps(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	s.SetDispatchList([]string{"core"}, nil, nil)

	assert.True(t, s.isReady("core"))
}

func TestScheduler_isReady_DepNotComplete(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	deps := map[string][]string{"eac": {"core"}}
	s.SetDispatchList([]string{"core", "eac"}, deps, nil)

	assert.True(t, s.isReady("core"))
	assert.False(t, s.isReady("eac"))
}

func TestScheduler_isReady_DepComplete(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	deps := map[string][]string{"eac": {"core"}}
	s.SetDispatchList([]string{"core", "eac"}, deps, nil)

	s.status["core"] = ciModuleCompleted
	assert.True(t, s.isReady("eac"))
}

func TestScheduler_isReady_DepNotInDispatchSet(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	// eac depends on core, but core is cached (not in dispatch list).
	deps := map[string][]string{"eac": {"core"}}
	s.SetDispatchList([]string{"eac"}, deps, []string{"core"})

	// core is not in dispatch set → considered satisfied.
	assert.True(t, s.isReady("eac"))
}

func TestScheduler_cascadeFail_Chain(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	deps := map[string][]string{
		"B": {"A"},
		"C": {"B"},
	}
	s.SetDispatchList([]string{"A", "B", "C"}, deps, nil)

	s.status["A"] = ciModuleFailed
	s.cascadeFail("A")

	assert.Equal(t, ciModuleSkipped, s.status["B"])
	assert.Equal(t, ciModuleSkipped, s.status["C"])
}

func TestScheduler_cascadeFail_Diamond(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	deps := map[string][]string{
		"B": {"A"},
		"C": {"A"},
		"D": {"B", "C"},
	}
	s.SetDispatchList([]string{"A", "B", "C", "D"}, deps, nil)

	s.status["A"] = ciModuleFailed
	s.cascadeFail("A")

	assert.Equal(t, ciModuleSkipped, s.status["B"])
	assert.Equal(t, ciModuleSkipped, s.status["C"])
	assert.Equal(t, ciModuleSkipped, s.status["D"])
}

func TestScheduler_buildResult_AllStates(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	s.SetDispatchList([]string{"a", "b", "c", "d"}, nil, []string{"cached1"})

	s.status["a"] = ciModuleCompleted
	s.status["b"] = ciModuleFailed
	s.status["c"] = ciModuleSkipped
	s.status["d"] = ciModuleActive

	result := s.buildResult(time.Now())

	assert.Equal(t, []string{"a"}, result.Completed)
	assert.Equal(t, []string{"b"}, result.Failed)
	assert.Equal(t, []string{"c"}, result.CascadeFailed)
	assert.Contains(t, result.Dispatched, "a")
	assert.Contains(t, result.Dispatched, "b")
	assert.Contains(t, result.Dispatched, "d")
	assert.Equal(t, []string{"cached1"}, result.Cached)
}

func TestScheduler_BatchPollActive_MultipleModules(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)
	s.SetDispatchList([]string{"core", "docs", "eac"}, nil, nil)

	// Simulate: all dispatched and active
	go func() {
		time.Sleep(30 * time.Millisecond)
		// Complete core and eac, leave docs in_progress
		d.setStatus("core", "completed", "success")
		d.setStatus("eac", "completed", "success")

		// After another delay, complete docs
		time.Sleep(50 * time.Millisecond)
		d.setStatus("docs", "completed", "success")
	}()

	result, err := s.Schedule(context.Background())
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"core", "docs", "eac"}, result.Completed)
}

func TestScheduler_BatchPollActive_MixedResults(t *testing.T) {
	d := newMockDispatcher()
	s := testScheduler(6, d)

	deps := map[string][]string{
		"eac": {"core"},
	}
	s.SetDispatchList([]string{"core", "docs", "eac"}, deps, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		// core fails, docs succeeds
		d.setStatus("core", "completed", "failure")
		d.setStatus("docs", "completed", "success")
	}()

	result, err := s.Schedule(context.Background())
	require.Error(t, err)
	assert.Equal(t, []string{"core"}, result.Failed)
	assert.Equal(t, []string{"docs"}, result.Completed)
	// eac should be cascade-failed because core failed
	assert.Equal(t, []string{"eac"}, result.CascadeFailed)
}

// indexOf returns the index of s in slice, or -1 if not found.
func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}
