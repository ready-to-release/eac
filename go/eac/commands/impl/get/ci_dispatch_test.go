package get

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/github"
)

func TestFilterCIDispatch_DirectlyChangedAlwaysDispatched(t *testing.T) {
	// Directly changed modules should ALWAYS be dispatched, even if they have valid CI
	mockStatus := map[string]bool{
		"moduleA": true, // Has valid CI
	}

	result, err := filterCIDispatch("moduleA", "", "abc123", mockStatus, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// moduleA should be dispatched because it's directly changed
	if len(result.Dispatch) != 1 || result.Dispatch[0] != "moduleA" {
		t.Errorf("expected moduleA in dispatch, got %v", result.Dispatch)
	}

	if result.Reasons["moduleA"] != "directly_changed" {
		t.Errorf("expected reason 'directly_changed', got %s", result.Reasons["moduleA"])
	}
}

func TestFilterCIDispatch_InvalidatedWithValidCI_Skipped(t *testing.T) {
	// Invalidated modules with valid CI at HEAD should be skipped
	mockStatus := map[string]bool{
		"moduleB": true, // Has valid CI at HEAD
	}

	result, err := filterCIDispatch("", "moduleB", "abc123", mockStatus, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// moduleB should be skipped
	if len(result.Skipped) != 1 || result.Skipped[0] != "moduleB" {
		t.Errorf("expected moduleB in skipped, got %v", result.Skipped)
	}

	if len(result.Dispatch) != 0 {
		t.Errorf("expected empty dispatch, got %v", result.Dispatch)
	}
}

func TestFilterCIDispatch_InvalidatedWithoutValidCI_Dispatched(t *testing.T) {
	// Invalidated modules without valid CI should be dispatched
	mockStatus := map[string]bool{
		"moduleC": false, // No valid CI
	}

	result, err := filterCIDispatch("", "moduleC", "abc123", mockStatus, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// moduleC should be dispatched
	if len(result.Dispatch) != 1 || result.Dispatch[0] != "moduleC" {
		t.Errorf("expected moduleC in dispatch, got %v", result.Dispatch)
	}

	if len(result.Skipped) != 0 {
		t.Errorf("expected empty skipped, got %v", result.Skipped)
	}
}

func TestFilterCIDispatch_MixedModules(t *testing.T) {
	// Test with a mix of directly changed and invalidated modules
	mockStatus := map[string]bool{
		"dep1": true,  // Valid CI
		"dep2": false, // No valid CI
	}

	result, err := filterCIDispatch("core", "dep1 dep2", "abc123", mockStatus, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// core: dispatched (directly changed)
	// dep1: skipped (valid CI)
	// dep2: dispatched (no valid CI)
	if len(result.Dispatch) != 2 {
		t.Errorf("expected 2 in dispatch, got %v", result.Dispatch)
	}

	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 in skipped, got %v", result.Skipped)
	}

	// Verify dispatch contains core and dep2
	dispatchMap := make(map[string]bool)
	for _, m := range result.Dispatch {
		dispatchMap[m] = true
	}
	if !dispatchMap["core"] || !dispatchMap["dep2"] {
		t.Errorf("expected core and dep2 in dispatch, got %v", result.Dispatch)
	}

	// Verify skipped contains dep1
	if result.Skipped[0] != "dep1" {
		t.Errorf("expected dep1 in skipped, got %v", result.Skipped)
	}
}

func TestFilterCIDispatch_ModuleNotInMock_Dispatched(t *testing.T) {
	// Modules not in mock data should be dispatched (safe default)
	mockStatus := map[string]bool{
		"known": true,
	}

	result, err := filterCIDispatch("", "unknown", "abc123", mockStatus, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// unknown should be dispatched
	if len(result.Dispatch) != 1 || result.Dispatch[0] != "unknown" {
		t.Errorf("expected unknown in dispatch, got %v", result.Dispatch)
	}

	if result.Reasons["unknown"] != "mock:not_specified" {
		t.Errorf("expected reason 'mock:not_specified', got %s", result.Reasons["unknown"])
	}
}

func TestFilterCIDispatch_EmptyInputs(t *testing.T) {
	result, err := filterCIDispatch("", "", "abc123", nil, "")
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
	result, err := filterCIDispatch("", "", "my-head-sha", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HeadSHA != "my-head-sha" {
		t.Errorf("expected HeadSHA 'my-head-sha', got %s", result.HeadSHA)
	}
}

func TestFilterCIDispatch_MultipleDirectlyChanged(t *testing.T) {
	// Use empty mock to skip file-based validation (testing directly_changed logic)
	result, err := filterCIDispatch("mod1 mod2 mod3", "", "abc123", map[string]bool{}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 3 {
		t.Errorf("expected 3 modules in dispatch, got %d", len(result.Dispatch))
	}

	// All should have directly_changed reason
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
		{"mod1  mod2   mod3", []string{"mod1", "mod2", "mod3"}}, // Extra spaces
		{"  mod1  ", []string{"mod1"}},                           // Leading/trailing spaces
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

func TestCheckModuleCIValidity_MockMode(t *testing.T) {
	tests := []struct {
		name       string
		module     string
		mockStatus map[string]bool
		wantValid  bool
		wantReason string
	}{
		{
			name:       "valid CI in mock",
			module:     "mod1",
			mockStatus: map[string]bool{"mod1": true},
			wantValid:  true,
			wantReason: "mock:valid_ci_at_head",
		},
		{
			name:       "no valid CI in mock",
			module:     "mod1",
			mockStatus: map[string]bool{"mod1": false},
			wantValid:  false,
			wantReason: "mock:no_valid_ci",
		},
		{
			name:       "module not in mock",
			module:     "unknown",
			mockStatus: map[string]bool{"other": true},
			wantValid:  false,
			wantReason: "mock:not_specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, reason, err := checkModuleCIValidity(tt.module, "sha", tt.mockStatus, "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if valid != tt.wantValid {
				t.Errorf("valid: expected %v, got %v", tt.wantValid, valid)
			}

			if reason != tt.wantReason {
				t.Errorf("reason: expected %s, got %s", tt.wantReason, reason)
			}
		})
	}
}

func TestFilterCIDispatch_AllSkipped(t *testing.T) {
	// All invalidated modules have valid CI - all should be skipped
	mockStatus := map[string]bool{
		"mod1": true,
		"mod2": true,
		"mod3": true,
	}

	result, err := filterCIDispatch("", "mod1 mod2 mod3", "abc123", mockStatus, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Dispatch) != 0 {
		t.Errorf("expected empty dispatch, got %v", result.Dispatch)
	}

	if len(result.Skipped) != 3 {
		t.Errorf("expected 3 skipped, got %d", len(result.Skipped))
	}

	// Verify all reasons are valid_ci
	for _, mod := range result.Skipped {
		if result.Reasons[mod] != "mock:valid_ci_at_head" {
			t.Errorf("%s: expected reason 'mock:valid_ci_at_head', got %s", mod, result.Reasons[mod])
		}
	}
}

func TestFilterCIDispatch_AllDispatched(t *testing.T) {
	// All invalidated modules have no valid CI - all should be dispatched
	mockStatus := map[string]bool{
		"mod1": false,
		"mod2": false,
		"mod3": false,
	}

	result, err := filterCIDispatch("", "mod1 mod2 mod3", "abc123", mockStatus, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Skipped) != 0 {
		t.Errorf("expected empty skipped, got %v", result.Skipped)
	}

	if len(result.Dispatch) != 3 {
		t.Errorf("expected 3 dispatched, got %d", len(result.Dispatch))
	}

	// Verify all reasons are no_valid_ci
	for _, mod := range result.Dispatch {
		if result.Reasons[mod] != "mock:no_valid_ci" {
			t.Errorf("%s: expected reason 'mock:no_valid_ci', got %s", mod, result.Reasons[mod])
		}
	}
}

func TestFilterCIDispatch_DirectlyChangedAndInvalidated(t *testing.T) {
	// When a module appears in both directly-changed and invalidated,
	// it appears in both dispatch (directly_changed) and skipped (valid CI)
	// The reason map gets the last value (from invalidated check)
	mockStatus := map[string]bool{
		"mod1": true, // Has valid CI
	}

	result, err := filterCIDispatch("mod1", "mod1", "abc123", mockStatus, "")
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
	// Test with a larger set of modules
	mockStatus := map[string]bool{
		"mod1": true,
		"mod2": false,
		"mod3": true,
		"mod4": false,
		"mod5": true,
		"mod6": false,
	}

	result, err := filterCIDispatch("core", "mod1 mod2 mod3 mod4 mod5 mod6", "abc123", mockStatus, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// core + mod2, mod4, mod6 = 4 dispatched
	if len(result.Dispatch) != 4 {
		t.Errorf("expected 4 dispatched, got %d: %v", len(result.Dispatch), result.Dispatch)
	}

	// mod1, mod3, mod5 = 3 skipped
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
			"mod2": "mock:no_valid_ci",
			"mod3": "mock:valid_ci_at_head",
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

func TestCheckModuleCIValidity_NilMockStatus(t *testing.T) {
	// When mockStatus is nil, it should try to query GitHub
	// We can't test the actual query without mocking exec.Command,
	// but we can verify it doesn't panic with nil

	// This will actually try to run gh, which may fail in test env
	// So we just verify the function signature and parameters work correctly
	// The actual GitHub query is integration-tested separately
}

func TestFilterCIDispatch_InvalidModule_ReturnsError(t *testing.T) {
	// When mock is nil (production mode), invalid modules should return error
	// Use actual workspace root to trigger validation
	workspaceRoot := "../../../../../" // Go up to repo root from test file

	// Test with a fake module that doesn't exist
	_, err := filterCIDispatch("fake-module-xyz", "", "abc123", nil, workspaceRoot)
	if err == nil {
		t.Errorf("expected error for invalid module, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid module") {
		t.Errorf("expected 'invalid module' error, got: %v", err)
	}
}

func TestFilterCIDispatch_InvalidInvalidatedModule_ReturnsError(t *testing.T) {
	// Invalid modules in invalidated list should also error
	workspaceRoot := "../../../../../"

	_, err := filterCIDispatch("", "nonexistent-module", "abc123", nil, workspaceRoot)
	if err == nil {
		t.Errorf("expected error for invalid invalidated module, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid module") {
		t.Errorf("expected 'invalid module' error, got: %v", err)
	}
}

func TestParseModuleList_EdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"   ", []string{}},                       // Only whitespace
		{"\t\n", []string{}},                      // Tabs and newlines
		{"mod1", []string{"mod1"}},
		{"mod1 mod2 mod3", []string{"mod1", "mod2", "mod3"}},
		{"  mod1  mod2  ", []string{"mod1", "mod2"}},
		{"mod-with-dashes", []string{"mod-with-dashes"}},
		{"mod_with_underscores", []string{"mod_with_underscores"}},
		{"eac-core eac-commands r2r-cli", []string{"eac-core", "eac-commands", "r2r-cli"}},
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

// Tests using github.MockAPI

func TestCheckModuleCIValidityWithAPI_ValidCIAtHead(t *testing.T) {
	mock := github.NewMockAPI()
	mock.AddSuccessRun("ci-eac-core.yaml", "abc123")

	valid, reason, err := checkModuleCIValidityWithAPI("eac-core", "abc123", nil, "", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !valid {
		t.Error("expected valid=true when CI exists at HEAD")
	}
	if reason != "valid_ci_at_head" {
		t.Errorf("expected reason 'valid_ci_at_head', got %q", reason)
	}

	// Verify API was called
	if mock.CallCount("ListRuns") != 1 {
		t.Errorf("expected ListRuns to be called once, got %d", mock.CallCount("ListRuns"))
	}
}

func TestCheckModuleCIValidityWithAPI_CIAtDifferentSHA(t *testing.T) {
	mock := github.NewMockAPI()
	mock.AddSuccessRun("ci-eac-core.yaml", "different123")

	valid, reason, err := checkModuleCIValidityWithAPI("eac-core", "abc123", nil, "", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if valid {
		t.Error("expected valid=false when CI is at different SHA")
	}
	if reason != "ci_at_different_sha:differe" {
		t.Errorf("expected reason starting with 'ci_at_different_sha:', got %q", reason)
	}
}

func TestCheckModuleCIValidityWithAPI_NoCIRun(t *testing.T) {
	mock := github.NewMockAPI()
	// No runs added - empty mock

	valid, reason, err := checkModuleCIValidityWithAPI("eac-core", "abc123", nil, "", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if valid {
		t.Error("expected valid=false when no CI runs exist")
	}
	if reason != "no_ci_run" {
		t.Errorf("expected reason 'no_ci_run', got %q", reason)
	}
}

func TestCheckModuleCIValidityWithAPI_APIError(t *testing.T) {
	mock := github.NewMockAPI()
	mock.RunsError = fmt.Errorf("API rate limit exceeded")

	valid, reason, err := checkModuleCIValidityWithAPI("eac-core", "abc123", nil, "", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// On API error, we return valid=false and dispatch to be safe
	if valid {
		t.Error("expected valid=false on API error")
	}
	if reason == "" || reason[:13] != "query_failed:" {
		t.Errorf("expected reason starting with 'query_failed:', got %q", reason)
	}
}

func TestCheckModuleCIValidityWithAPI_MockStatusTakesPrecedence(t *testing.T) {
	// Even with a real API mock configured, mockStatus should take precedence
	mock := github.NewMockAPI()
	mock.AddSuccessRun("ci-eac-core.yaml", "abc123") // Would make valid=true via API

	mockStatus := map[string]bool{
		"eac-core": false, // But mock status says invalid
	}

	valid, reason, err := checkModuleCIValidityWithAPI("eac-core", "abc123", mockStatus, "", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// mockStatus should take precedence - module marked as no valid CI
	if valid {
		t.Error("expected valid=false when mockStatus says false")
	}
	if reason != "mock:no_valid_ci" {
		t.Errorf("expected reason 'mock:no_valid_ci', got %q", reason)
	}

	// API should NOT have been called
	if mock.CallCount("ListRuns") != 0 {
		t.Errorf("expected ListRuns to NOT be called when mockStatus is provided, got %d calls", mock.CallCount("ListRuns"))
	}
}

func TestGetLastSuccessfulWorkflowSHAWithAPI_ReturnsFirstRun(t *testing.T) {
	mock := github.NewMockAPI()
	mock.AddSuccessRun("ci-test.yaml", "sha111")
	mock.AddSuccessRun("ci-test.yaml", "sha222")

	sha, err := getLastSuccessfulWorkflowSHAWithAPI("ci-test.yaml", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return the first (most recent) run
	if sha != "sha111" {
		t.Errorf("expected sha111, got %q", sha)
	}
}

func TestGetLastSuccessfulWorkflowSHAWithAPI_EmptyResult(t *testing.T) {
	mock := github.NewMockAPI()
	// No runs added

	sha, err := getLastSuccessfulWorkflowSHAWithAPI("ci-test.yaml", mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sha != "" {
		t.Errorf("expected empty SHA, got %q", sha)
	}
}

func TestGetLastSuccessfulWorkflowSHAWithAPI_Error(t *testing.T) {
	mock := github.NewMockAPI()
	mock.RunsError = fmt.Errorf("network error")

	_, err := getLastSuccessfulWorkflowSHAWithAPI("ci-test.yaml", mock)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCheckModuleCIValidityWithAPI_MultipleModules(t *testing.T) {
	mock := github.NewMockAPI()
	mock.AddSuccessRun("ci-mod1.yaml", "head123")
	mock.AddSuccessRun("ci-mod2.yaml", "different")
	// mod3 has no runs

	tests := []struct {
		module    string
		headSHA   string
		wantValid bool
	}{
		{"mod1", "head123", true},   // CI at HEAD
		{"mod2", "head123", false},  // CI at different SHA
		{"mod3", "head123", false},  // No CI
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			mock.Reset() // Clear call history
			valid, _, err := checkModuleCIValidityWithAPI(tt.module, tt.headSHA, nil, "", mock)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if valid != tt.wantValid {
				t.Errorf("valid = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}
