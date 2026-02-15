package get

import (
	"fmt"
	"testing"

	"github.com/ready-to-release/eac/go/core/cache"
)

// helper: build a CICacheChecker from the legacy mock format (map[string]bool).
func checkerFromMock(mockStatus map[string]bool, headSHA string) *cache.CICacheChecker {
	return mockCheckerFromStatus(mockStatus, headSHA)
}

// helper: build a CICacheChecker from a MockCIRunQuerier with optional cache config.
func checkerFromQuerier(runs map[string]string, cfg *cache.Config) *cache.CICacheChecker {
	querier := cache.NewMockCIRunQuerier(runs)
	return cache.NewCICacheChecker(querier, cfg)
}

func TestFilterCIDispatch_DirectlyChangedAlwaysDispatched(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"moduleA": true}, "abc123")

	result, err := filterCIDispatchWithDeps("moduleA", "", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 1 || result.Dispatch[0] != "moduleA" {
		t.Errorf("expected moduleA in dispatch, got %v", result.Dispatch)
	}
	if result.Reasons["moduleA"] != "directly_changed" {
		t.Errorf("expected reason 'directly_changed', got %s", result.Reasons["moduleA"])
	}
}

func TestFilterCIDispatch_InvalidatedWithValidCI_Skipped(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"moduleB": true}, "abc123")

	result, err := filterCIDispatchWithDeps("", "moduleB", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Skipped) != 1 || result.Skipped[0] != "moduleB" {
		t.Errorf("expected moduleB in skipped, got %v", result.Skipped)
	}
	if len(result.Dispatch) != 0 {
		t.Errorf("expected empty dispatch, got %v", result.Dispatch)
	}
}

func TestFilterCIDispatch_InvalidatedWithoutValidCI_Dispatched(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"moduleC": false}, "abc123")

	result, err := filterCIDispatchWithDeps("", "moduleC", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 1 || result.Dispatch[0] != "moduleC" {
		t.Errorf("expected moduleC in dispatch, got %v", result.Dispatch)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected empty skipped, got %v", result.Skipped)
	}
}

func TestFilterCIDispatch_MixedModules(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"dep1": true, "dep2": false}, "abc123")

	result, err := filterCIDispatchWithDeps("core", "dep1 dep2", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 2 {
		t.Errorf("expected 2 in dispatch, got %v", result.Dispatch)
	}
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 in skipped, got %v", result.Skipped)
	}

	dispatchMap := make(map[string]bool)
	for _, m := range result.Dispatch {
		dispatchMap[m] = true
	}
	if !dispatchMap["core"] || !dispatchMap["dep2"] {
		t.Errorf("expected core and dep2 in dispatch, got %v", result.Dispatch)
	}
	if result.Skipped[0] != "dep1" {
		t.Errorf("expected dep1 in skipped, got %v", result.Skipped)
	}
}

func TestFilterCIDispatch_ModuleNotInMock_Dispatched(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"known": true}, "abc123")

	result, err := filterCIDispatchWithDeps("", "unknown", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 1 || result.Dispatch[0] != "unknown" {
		t.Errorf("expected unknown in dispatch, got %v", result.Dispatch)
	}
	if result.Reasons["unknown"] != "no_ci_run" {
		t.Errorf("expected reason 'no_ci_run', got %s", result.Reasons["unknown"])
	}
}

func TestFilterCIDispatch_EmptyInputs(t *testing.T) {
	checker := checkerFromQuerier(nil, nil)

	result, err := filterCIDispatchWithDeps("", "", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 0 {
		t.Errorf("expected empty dispatch, got %v", result.Dispatch)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected empty skipped, got %v", result.Skipped)
	}
}

func TestFilterCIDispatch_HeadSHA(t *testing.T) {
	checker := checkerFromQuerier(nil, nil)

	result, err := filterCIDispatchWithDeps("", "", "my-head-sha", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HeadSHA != "my-head-sha" {
		t.Errorf("expected HeadSHA 'my-head-sha', got %s", result.HeadSHA)
	}
}

func TestFilterCIDispatch_MultipleDirectlyChanged(t *testing.T) {
	checker := checkerFromQuerier(nil, nil)

	result, err := filterCIDispatchWithDeps("mod1 mod2 mod3", "", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 3 {
		t.Errorf("expected 3 modules in dispatch, got %d", len(result.Dispatch))
	}
	for _, mod := range result.Dispatch {
		if result.Reasons[mod] != "directly_changed" {
			t.Errorf("expected 'directly_changed' for %s, got %s", mod, result.Reasons[mod])
		}
	}
}

func TestParseModuleList(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"mod1", []string{"mod1"}},
		{"mod1 mod2", []string{"mod1", "mod2"}},
		{"mod1  mod2   mod3", []string{"mod1", "mod2", "mod3"}},
		{"  mod1  ", []string{"mod1"}},
	}

	for _, tt := range tests {
		result := parseModuleList(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("parseModuleList(%q): expected %v, got %v", tt.input, tt.expected, result)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("parseModuleList(%q)[%d]: expected %s, got %s", tt.input, i, tt.expected[i], v)
			}
		}
	}
}

func TestFilterCIDispatch_AllSkipped(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"mod1": true, "mod2": true, "mod3": true}, "abc123")

	result, err := filterCIDispatchWithDeps("", "mod1 mod2 mod3", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 0 {
		t.Errorf("expected empty dispatch, got %v", result.Dispatch)
	}
	if len(result.Skipped) != 3 {
		t.Errorf("expected 3 skipped, got %d", len(result.Skipped))
	}
}

func TestFilterCIDispatch_AllDispatched(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"mod1": false, "mod2": false, "mod3": false}, "abc123")

	result, err := filterCIDispatchWithDeps("", "mod1 mod2 mod3", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Skipped) != 0 {
		t.Errorf("expected empty skipped, got %v", result.Skipped)
	}
	if len(result.Dispatch) != 3 {
		t.Errorf("expected 3 dispatched, got %d", len(result.Dispatch))
	}
}

func TestFilterCIDispatch_DirectlyChangedAndInvalidated(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"mod1": true}, "abc123")

	result, err := filterCIDispatchWithDeps("mod1", "mod1", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// mod1 appears in dispatch (from directly-changed)
	found := false
	for _, mod := range result.Dispatch {
		if mod == "mod1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected mod1 in dispatch from directly-changed")
	}

	// mod1 also appears in skipped (from invalidated with valid CI)
	foundSkipped := false
	for _, mod := range result.Skipped {
		if mod == "mod1" {
			foundSkipped = true
			break
		}
	}
	if !foundSkipped {
		t.Error("expected mod1 in skipped from invalidated with valid CI")
	}
}

func TestFilterCIDispatch_LargeModuleSet(t *testing.T) {
	checker := checkerFromMock(map[string]bool{
		"mod1": true, "mod2": false, "mod3": true,
		"mod4": false, "mod5": true, "mod6": false,
	}, "abc123")

	result, err := filterCIDispatchWithDeps("core", "mod1 mod2 mod3 mod4 mod5 mod6", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 4 {
		t.Errorf("expected 4 dispatched, got %d: %v", len(result.Dispatch), result.Dispatch)
	}
	if len(result.Skipped) != 3 {
		t.Errorf("expected 3 skipped, got %d: %v", len(result.Skipped), result.Skipped)
	}
}

func TestCIDispatchResult_Struct(t *testing.T) {
	result := CIDispatchResult{
		Dispatch: []string{"mod1", "mod2"},
		Skipped:  []string{"mod3"},
		Reasons: map[string]string{
			"mod1": "directly_changed",
			"mod2": "no_ci_run",
			"mod3": "valid_ci_at_head",
		},
		HeadSHA: "abc123def456",
	}

	if len(result.Dispatch) != 2 {
		t.Errorf("Dispatch length = %d, want 2", len(result.Dispatch))
	}
	if len(result.Skipped) != 1 {
		t.Errorf("Skipped length = %d, want 1", len(result.Skipped))
	}
	if len(result.Reasons) != 3 {
		t.Errorf("Reasons length = %d, want 3", len(result.Reasons))
	}
	if result.HeadSHA != "abc123def456" {
		t.Errorf("HeadSHA = %q, want %q", result.HeadSHA, "abc123def456")
	}
}

func TestFilterCIDispatch_InvalidModule_ReturnsError(t *testing.T) {
	checker := checkerFromQuerier(nil, nil)
	workspaceRoot := "../../../../../"

	_, err := filterCIDispatchWithDeps("fake-module-xyz", "", "abc123", checker, workspaceRoot, nil)
	if err == nil {
		t.Errorf("expected error for invalid module, got nil")
	}
}

func TestFilterCIDispatch_InvalidInvalidatedModule_ReturnsError(t *testing.T) {
	checker := checkerFromQuerier(nil, nil)
	workspaceRoot := "../../../../../"

	_, err := filterCIDispatchWithDeps("", "nonexistent-module", "abc123", checker, workspaceRoot, nil)
	if err == nil {
		t.Errorf("expected error for invalid invalidated module, got nil")
	}
}

func TestParseModuleList_EdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"   ", []string{}},
		{"\t\n", []string{}},
		{"mod1", []string{"mod1"}},
		{"mod1 mod2 mod3", []string{"mod1", "mod2", "mod3"}},
		{"  mod1  mod2  ", []string{"mod1", "mod2"}},
		{"mod-with-dashes", []string{"mod-with-dashes"}},
		{"mod_with_underscores", []string{"mod_with_underscores"}},
		{"core eac-cli clie", []string{"core", "eac-cli", "clie"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseModuleList(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseModuleList(%q): expected %d items, got %d", tt.input, len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseModuleList(%q)[%d]: expected %s, got %s", tt.input, i, tt.expected[i], v)
				}
			}
		})
	}
}

// ============================================================================
// Tests using cache.CICacheChecker directly
// ============================================================================

func TestCICacheChecker_ValidCIAtHead(t *testing.T) {
	checker := checkerFromQuerier(map[string]string{
		"ci-core.yaml": "abc123",
	}, nil)

	result := checker.Check("core", "abc123")

	if !result.Cached {
		t.Error("expected Cached=true when CI exists at HEAD")
	}
	if result.Reason != "valid_ci_at_head" {
		t.Errorf("expected reason 'valid_ci_at_head', got %q", result.Reason)
	}
}

func TestCICacheChecker_CIAtDifferentSHA(t *testing.T) {
	checker := checkerFromQuerier(map[string]string{
		"ci-core.yaml": "different123",
	}, nil)

	result := checker.Check("core", "abc123")

	if result.Cached {
		t.Error("expected Cached=false when CI is at different SHA")
	}
	if result.Reason[:20] != "ci_at_different_sha:" {
		t.Errorf("expected reason starting with 'ci_at_different_sha:', got %q", result.Reason)
	}
}

func TestCICacheChecker_NoCIRun(t *testing.T) {
	checker := checkerFromQuerier(nil, nil)

	result := checker.Check("core", "abc123")

	if result.Cached {
		t.Error("expected Cached=false when no CI runs exist")
	}
	if result.Reason != "no_ci_run" {
		t.Errorf("expected reason 'no_ci_run', got %q", result.Reason)
	}
}

func TestCICacheChecker_QuerierError(t *testing.T) {
	querier := &cache.MockCIRunQuerier{Err: fmt.Errorf("API rate limit exceeded")}
	checker := cache.NewCICacheChecker(querier, nil)

	result := checker.Check("core", "abc123")

	if result.Cached {
		t.Error("expected Cached=false on querier error")
	}
	if result.Reason == "" || result.Reason[:13] != "query_failed:" {
		t.Errorf("expected reason starting with 'query_failed:', got %q", result.Reason)
	}
}

func TestCICacheChecker_MultipleModules(t *testing.T) {
	checker := checkerFromQuerier(map[string]string{
		"ci-mod1.yaml": "head123",
		"ci-mod2.yaml": "different",
	}, nil)

	tests := []struct {
		module     string
		headSHA    string
		wantCached bool
	}{
		{"mod1", "head123", true},
		{"mod2", "head123", false},
		{"mod3", "head123", false},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			result := checker.Check(tt.module, tt.headSHA)
			if result.Cached != tt.wantCached {
				t.Errorf("Cached = %v, want %v", result.Cached, tt.wantCached)
			}
		})
	}
}

// ============================================================================
// Tests for --skip-cache=remote:ci bypass
// ============================================================================

func TestFilterCIDispatch_SkipCacheRemoteCI_BypassesAllChecking(t *testing.T) {
	cfg := cache.NewConfig()
	cfg.Skip(cache.Spec{Level: cache.LevelRemote, Type: cache.TypeCI})

	// Even though runs exist at HEAD, they should be bypassed
	checker := checkerFromQuerier(map[string]string{
		"ci-mod1.yaml": "abc123",
		"ci-mod2.yaml": "abc123",
	}, cfg)

	result, err := filterCIDispatchWithDeps("", "mod1 mod2", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All modules should be dispatched (bypass)
	if len(result.Dispatch) != 2 {
		t.Errorf("expected 2 dispatched with bypass, got %d: %v", len(result.Dispatch), result.Dispatch)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected 0 skipped with bypass, got %d: %v", len(result.Skipped), result.Skipped)
	}

	// Reasons should be cache_bypassed
	for _, mod := range result.Dispatch {
		if result.Reasons[mod] != "cache_bypassed" {
			t.Errorf("expected reason 'cache_bypassed' for %s, got %s", mod, result.Reasons[mod])
		}
	}
}

func TestFilterCIDispatch_SkipCacheAll_BypassesCIChecking(t *testing.T) {
	cfg := cache.NewConfig()
	cfg.SkipAll()

	checker := checkerFromQuerier(map[string]string{
		"ci-mod1.yaml": "abc123",
	}, cfg)

	result, err := filterCIDispatchWithDeps("", "mod1", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 1 {
		t.Errorf("expected 1 dispatched, got %d", len(result.Dispatch))
	}
	if result.Reasons["mod1"] != "cache_bypassed" {
		t.Errorf("expected 'cache_bypassed', got %s", result.Reasons["mod1"])
	}
}

// ============================================================================
// Tests for topological sort ordering (CI dependencies)
// ============================================================================

func TestComputeCIDispatchOrder_NoDeps(t *testing.T) {
	modules := []string{"core", "docs", "contracts-bundle"}
	ciDeps := map[string][]string{}

	result := computeCIDispatchOrder(modules, ciDeps)

	expected := []string{"contracts-bundle", "core", "docs"}
	if !slicesEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestComputeCIDispatchOrder_SimpleDep(t *testing.T) {
	modules := []string{"eac-ext", "clie", "core"}
	ciDeps := map[string][]string{
		"eac-ext": {"clie"},
	}

	result := computeCIDispatchOrder(modules, ciDeps)

	clieIdx := indexOf(result, "clie")
	extIdx := indexOf(result, "eac-ext")
	if clieIdx == -1 || extIdx == -1 {
		t.Fatalf("missing modules in result: %v", result)
	}
	if clieIdx >= extIdx {
		t.Errorf("clie (idx %d) should come before eac-ext (idx %d), got %v", clieIdx, extIdx, result)
	}
}

func TestComputeCIDispatchOrder_MultipleDepsOnSameModule(t *testing.T) {
	modules := []string{"eac-ext", "clie-installer", "implicit-cli", "clie", "core"}
	ciDeps := map[string][]string{
		"eac-ext":        {"clie"},
		"clie-installer": {"clie"},
		"implicit-cli":   {"clie"},
	}

	result := computeCIDispatchOrder(modules, ciDeps)

	clieIdx := indexOf(result, "clie")
	for _, dep := range []string{"eac-ext", "clie-installer", "implicit-cli"} {
		depIdx := indexOf(result, dep)
		if depIdx == -1 {
			t.Fatalf("missing %s in result: %v", dep, result)
		}
		if clieIdx >= depIdx {
			t.Errorf("clie (idx %d) should come before %s (idx %d), got %v", clieIdx, dep, depIdx, result)
		}
	}
}

func TestComputeCIDispatchOrder_DepNotInDispatchSet(t *testing.T) {
	modules := []string{"eac-ext", "core"}
	ciDeps := map[string][]string{
		"eac-ext": {"clie"},
	}

	result := computeCIDispatchOrder(modules, ciDeps)

	if len(result) != 2 {
		t.Errorf("expected 2 modules, got %d: %v", len(result), result)
	}
	if indexOf(result, "eac-ext") == -1 {
		t.Errorf("eac-ext should be in result: %v", result)
	}
}

func TestComputeCIDispatchOrder_EmptyList(t *testing.T) {
	result := computeCIDispatchOrder([]string{}, map[string][]string{})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestComputeCIDispatchOrder_SingleModule(t *testing.T) {
	result := computeCIDispatchOrder([]string{"core"}, map[string][]string{})
	if len(result) != 1 || result[0] != "core" {
		t.Errorf("expected [core], got %v", result)
	}
}

func TestComputeCIDispatchOrder_Deterministic(t *testing.T) {
	modules := []string{"eac-ext", "clie", "core", "docs", "clie-installer"}
	ciDeps := map[string][]string{
		"eac-ext":        {"clie"},
		"clie-installer": {"clie"},
	}

	first := computeCIDispatchOrder(modules, ciDeps)
	for i := 0; i < 10; i++ {
		result := computeCIDispatchOrder(modules, ciDeps)
		if !slicesEqual(result, first) {
			t.Errorf("non-deterministic: run 1 got %v, run %d got %v", first, i+2, result)
		}
	}
}

func TestComputeCIDispatchOrder_ChainedDeps(t *testing.T) {
	modules := []string{"A", "B", "C"}
	ciDeps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
	}

	result := computeCIDispatchOrder(modules, ciDeps)

	cIdx := indexOf(result, "C")
	bIdx := indexOf(result, "B")
	aIdx := indexOf(result, "A")

	if cIdx >= bIdx || bIdx >= aIdx {
		t.Errorf("expected C < B < A, got %v (C=%d, B=%d, A=%d)", result, cIdx, bIdx, aIdx)
	}
}

func TestFilterCIDispatch_OrderedByDeps(t *testing.T) {
	checker := checkerFromMock(map[string]bool{
		"clie": false, "eac-ext": false, "core": false,
	}, "abc123")

	ciDeps := map[string][]string{
		"eac-ext": {"clie"},
	}

	result, err := filterCIDispatchWithDeps("", "clie eac-ext core", "abc123", checker, "", ciDeps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	clieIdx := indexOf(result.Dispatch, "clie")
	extIdx := indexOf(result.Dispatch, "eac-ext")

	if clieIdx == -1 || extIdx == -1 {
		t.Fatalf("missing modules in dispatch: %v", result.Dispatch)
	}
	if clieIdx >= extIdx {
		t.Errorf("clie should come before eac-ext, got %v", result.Dispatch)
	}
}

func TestFilterCIDispatch_CIDependenciesField(t *testing.T) {
	checker := checkerFromMock(map[string]bool{"clie": false, "eac-ext": false}, "abc123")

	ciDeps := map[string][]string{
		"eac-ext": {"clie"},
	}

	result, err := filterCIDispatchWithDeps("", "clie eac-ext", "abc123", checker, "", ciDeps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CIDependencies == nil {
		t.Fatal("CIDependencies should not be nil")
	}
	deps, ok := result.CIDependencies["eac-ext"]
	if !ok {
		t.Errorf("expected eac-ext in CIDependencies, got %v", result.CIDependencies)
	}
	if len(deps) != 1 || deps[0] != "clie" {
		t.Errorf("expected eac-ext deps [clie], got %v", deps)
	}
}

func TestFilterCIDispatch_TotalModulesField(t *testing.T) {
	checker := checkerFromMock(map[string]bool{
		"mod1": false, "mod2": true, "mod3": false,
	}, "abc123")

	result, err := filterCIDispatchWithDeps("", "mod1 mod2 mod3", "abc123", checker, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalModules != 3 {
		t.Errorf("expected TotalModules=3, got %d", result.TotalModules)
	}
}

func TestCIDispatchResult_NewFields(t *testing.T) {
	result := CIDispatchResult{
		Dispatch: []string{"core", "clie", "eac-ext"},
		Skipped:  []string{"docs"},
		Reasons: map[string]string{
			"core":    "directly_changed",
			"clie":    "no_ci_run",
			"eac-ext": "no_ci_run",
			"docs":    "valid_ci_at_head",
		},
		CIDependencies: map[string][]string{
			"eac-ext": {"clie"},
		},
		HeadSHA:      "abc123def456",
		TotalModules: 4,
	}

	if len(result.Dispatch) != 3 {
		t.Errorf("Dispatch length = %d, want 3", len(result.Dispatch))
	}
	if len(result.Skipped) != 1 {
		t.Errorf("Skipped length = %d, want 1", len(result.Skipped))
	}
	if result.TotalModules != 4 {
		t.Errorf("TotalModules = %d, want 4", result.TotalModules)
	}
	if len(result.CIDependencies) != 1 {
		t.Errorf("CIDependencies length = %d, want 1", len(result.CIDependencies))
	}
	deps, ok := result.CIDependencies["eac-ext"]
	if !ok || len(deps) != 1 || deps[0] != "clie" {
		t.Errorf("CIDependencies[eac-ext] = %v, want [clie]", deps)
	}
}

// Helper functions for tests

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}
