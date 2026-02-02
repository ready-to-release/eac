package workunit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// UnitResult Type Tests
// =============================================================================

func TestUnitResult_FieldsExist(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: 0,
		Duration: 5 * time.Second,
		LogPath:  "/path/to/log",
		Metrics:  map[string]any{"tests_passed": 42},
	}

	assert.Equal(t, "mod", result.ID.Module)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, 5*time.Second, result.Duration)
	assert.Equal(t, "/path/to/log", result.LogPath)
	assert.Equal(t, 42, result.Metrics["tests_passed"])
}

// =============================================================================
// Success() Method Tests
// =============================================================================

func TestUnitResult_Success_ReturnsTrueWhenExitCodeIsZero(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: 0,
	}

	assert.True(t, result.Success())
}

func TestUnitResult_Success_ReturnsFalseWhenExitCodeIsNegative(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: -1,
	}

	assert.False(t, result.Success())
}

func TestUnitResult_Success_ReturnsFalseWhenExitCodeIsPositive(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: 1,
	}

	assert.False(t, result.Success())
}

func TestUnitResult_Success_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		expected bool
	}{
		{name: "exit code 0 is success", exitCode: 0, expected: true},
		{name: "exit code -1 is not success", exitCode: -1, expected: false},
		{name: "exit code 1 is not success", exitCode: 1, expected: false},
		{name: "exit code 127 is not success", exitCode: 127, expected: false},
		{name: "exit code -2 is not success", exitCode: -2, expected: false},
		{name: "exit code 255 is not success", exitCode: 255, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnitResult{ExitCode: tt.exitCode}
			assert.Equal(t, tt.expected, result.Success())
		})
	}
}

// =============================================================================
// Cached() Method Tests
// =============================================================================

func TestUnitResult_Cached_ReturnsTrueWhenExitCodeIsMinusOne(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: -1,
	}

	assert.True(t, result.Cached())
}

func TestUnitResult_Cached_ReturnsFalseWhenExitCodeIsZero(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: 0,
	}

	assert.False(t, result.Cached())
}

func TestUnitResult_Cached_ReturnsFalseWhenExitCodeIsPositive(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: 1,
	}

	assert.False(t, result.Cached())
}

func TestUnitResult_Cached_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		expected bool
	}{
		{name: "exit code -1 is cached", exitCode: -1, expected: true},
		{name: "exit code 0 is not cached", exitCode: 0, expected: false},
		{name: "exit code 1 is not cached", exitCode: 1, expected: false},
		{name: "exit code 127 is not cached", exitCode: 127, expected: false},
		{name: "exit code -2 is not cached", exitCode: -2, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnitResult{ExitCode: tt.exitCode}
			assert.Equal(t, tt.expected, result.Cached())
		})
	}
}

// =============================================================================
// Failed() Method Tests
// =============================================================================

func TestUnitResult_Failed_ReturnsTrueWhenExitCodeIsPositive(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: 1,
	}

	assert.True(t, result.Failed())
}

func TestUnitResult_Failed_ReturnsFalseWhenExitCodeIsZero(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: 0,
	}

	assert.False(t, result.Failed())
}

func TestUnitResult_Failed_ReturnsFalseWhenExitCodeIsNegative(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ExitCode: -1,
	}

	assert.False(t, result.Failed())
}

func TestUnitResult_Failed_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		expected bool
	}{
		{name: "exit code 1 is failed", exitCode: 1, expected: true},
		{name: "exit code 127 is failed", exitCode: 127, expected: true},
		{name: "exit code 255 is failed", exitCode: 255, expected: true},
		{name: "exit code 0 is not failed", exitCode: 0, expected: false},
		{name: "exit code -1 is not failed", exitCode: -1, expected: false},
		{name: "exit code -2 is not failed", exitCode: -2, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnitResult{ExitCode: tt.exitCode}
			assert.Equal(t, tt.expected, result.Failed())
		})
	}
}

// =============================================================================
// Exit Code Semantics - Comprehensive Tests
// =============================================================================

func TestUnitResult_ExitCodeSemantics(t *testing.T) {
	tests := []struct {
		name      string
		exitCode  int
		isSuccess bool
		isCached  bool
		isFailed  bool
		desc      string
	}{
		{
			name:      "cached/skipped execution",
			exitCode:  -1,
			isSuccess: false,
			isCached:  true,
			isFailed:  false,
			desc:      "Exit code -1 means cached/skipped",
		},
		{
			name:      "successful execution",
			exitCode:  0,
			isSuccess: true,
			isCached:  false,
			isFailed:  false,
			desc:      "Exit code 0 means success",
		},
		{
			name:      "general failure",
			exitCode:  1,
			isSuccess: false,
			isCached:  false,
			isFailed:  true,
			desc:      "Exit code 1 means general failure",
		},
		{
			name:      "command not found",
			exitCode:  127,
			isSuccess: false,
			isCached:  false,
			isFailed:  true,
			desc:      "Exit code 127 typically means command not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnitResult{ExitCode: tt.exitCode}

			assert.Equal(t, tt.isSuccess, result.Success(), "Success() for %s", tt.desc)
			assert.Equal(t, tt.isCached, result.Cached(), "Cached() for %s", tt.desc)
			assert.Equal(t, tt.isFailed, result.Failed(), "Failed() for %s", tt.desc)
		})
	}
}

// =============================================================================
// Mutually Exclusive States Tests
// =============================================================================

func TestUnitResult_MutuallyExclusiveStates(t *testing.T) {
	exitCodes := []int{-1, 0, 1, 127, 255}

	for _, code := range exitCodes {
		t.Run("exit code "+string(rune('0'+code)), func(t *testing.T) {
			result := UnitResult{ExitCode: code}

			// Count how many states are true
			trueCount := 0
			if result.Success() {
				trueCount++
			}
			if result.Cached() {
				trueCount++
			}
			if result.Failed() {
				trueCount++
			}

			// Exactly one state should be true (except for edge cases like -2)
			if code == -1 || code == 0 || code > 0 {
				assert.Equal(t, 1, trueCount,
					"Exactly one of Success/Cached/Failed should be true for exit code %d", code)
			}
		})
	}
}

func TestUnitResult_CommonExitCodes_ExactlyOneStateTrue(t *testing.T) {
	// These are the three canonical exit codes
	tests := []struct {
		exitCode int
		state    string
	}{
		{-1, "Cached"},
		{0, "Success"},
		{1, "Failed"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			result := UnitResult{ExitCode: tt.exitCode}

			success := result.Success()
			cached := result.Cached()
			failed := result.Failed()

			switch tt.state {
			case "Cached":
				assert.True(t, cached)
				assert.False(t, success)
				assert.False(t, failed)
			case "Success":
				assert.True(t, success)
				assert.False(t, cached)
				assert.False(t, failed)
			case "Failed":
				assert.True(t, failed)
				assert.False(t, success)
				assert.False(t, cached)
			}
		})
	}
}

// =============================================================================
// UnitResult Metrics Tests
// =============================================================================

func TestUnitResult_MetricsCanStoreTestCounts(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextTest, Module: "mod", Component: "go", Tool: "gotest"},
		ExitCode: 0,
		Metrics: map[string]any{
			"tests_passed":  42,
			"tests_failed":  0,
			"tests_skipped": 3,
		},
	}

	assert.Equal(t, 42, result.Metrics["tests_passed"])
	assert.Equal(t, 0, result.Metrics["tests_failed"])
	assert.Equal(t, 3, result.Metrics["tests_skipped"])
}

func TestUnitResult_MetricsCanStoreScanFindings(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextScan, Module: "mod", Component: "go", Tool: "trivy-vuln"},
		ExitCode: 0,
		Metrics: map[string]any{
			"issues_found": 5,
			"critical":     0,
			"high":         2,
			"medium":       3,
			"low":          0,
		},
	}

	assert.Equal(t, 5, result.Metrics["issues_found"])
	assert.Equal(t, 2, result.Metrics["high"])
}

func TestUnitResult_MetricsCanStoreLintIssues(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextLint, Module: "mod", Component: "go", Tool: "golangci-lint"},
		ExitCode: 1,
		Metrics: map[string]any{
			"issues_found": 12,
			"auto_fixed":   3,
		},
	}

	assert.Equal(t, 12, result.Metrics["issues_found"])
	assert.Equal(t, 3, result.Metrics["auto_fixed"])
}

func TestUnitResult_MetricsCanBeNil(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "go", Tool: "go"},
		ExitCode: 0,
		Metrics:  nil,
	}

	assert.Nil(t, result.Metrics)
}

func TestUnitResult_MetricsCanBeEmpty(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "go", Tool: "go"},
		ExitCode: 0,
		Metrics:  map[string]any{},
	}

	assert.NotNil(t, result.Metrics)
	assert.Empty(t, result.Metrics)
}

// =============================================================================
// UnitResult Duration Tests
// =============================================================================

func TestUnitResult_DurationIsAccessible(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "go", Tool: "go"},
		ExitCode: 0,
		Duration: 2*time.Minute + 30*time.Second,
	}

	assert.Equal(t, 2*time.Minute+30*time.Second, result.Duration)
}

func TestUnitResult_DurationCanBeZero(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "go", Tool: "go"},
		ExitCode: -1, // Cached, so duration might be 0
		Duration: 0,
	}

	assert.Equal(t, time.Duration(0), result.Duration)
}

// =============================================================================
// UnitResult LogPath Tests
// =============================================================================

func TestUnitResult_LogPathIsAccessible(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "core", Component: "go", Tool: "go"},
		ExitCode: 0,
		LogPath:  "out/build/core/go/execution.log",
	}

	assert.Equal(t, "out/build/core/go/execution.log", result.LogPath)
}

func TestUnitResult_LogPathCanBeEmpty(t *testing.T) {
	result := UnitResult{
		ID:       UnitID{Context: ContextBuild, Module: "mod", Component: "go", Tool: "go"},
		ExitCode: -1, // Cached, might not have log
		LogPath:  "",
	}

	assert.Empty(t, result.LogPath)
}

// =============================================================================
// UnitResult Real-World Scenarios
// =============================================================================

func TestUnitResult_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name     string
		result   UnitResult
		expected struct {
			success bool
			cached  bool
			failed  bool
		}
	}{
		{
			name: "successful build",
			result: UnitResult{
				ID:       UnitID{Context: ContextBuild, Module: "core", Component: "go", Tool: "go"},
				ExitCode: 0,
				Duration: 15 * time.Second,
				LogPath:  "out/build/core/go/execution.log",
			},
			expected: struct{ success, cached, failed bool }{true, false, false},
		},
		{
			name: "cached build (up-to-date)",
			result: UnitResult{
				ID:       UnitID{Context: ContextBuild, Module: "core", Component: "go", Tool: "go"},
				ExitCode: -1,
				Duration: 0,
				LogPath:  "",
			},
			expected: struct{ success, cached, failed bool }{false, true, false},
		},
		{
			name: "failed test run",
			result: UnitResult{
				ID:       UnitID{Context: ContextTest, Module: "core", Component: "go", Tool: "gotest", Extra: map[string]string{"testset": "unit"}},
				ExitCode: 1,
				Duration: 45 * time.Second,
				LogPath:  "out/test/core/go/unit/execution.log",
				Metrics:  map[string]any{"tests_passed": 95, "tests_failed": 3},
			},
			expected: struct{ success, cached, failed bool }{false, false, true},
		},
		{
			name: "successful lint with no issues",
			result: UnitResult{
				ID:       UnitID{Context: ContextLint, Module: "core", Component: "go", Tool: "golangci-lint"},
				ExitCode: 0,
				Duration: 8 * time.Second,
				LogPath:  "out/lint/core/go/execution.log",
				Metrics:  map[string]any{"issues_found": 0},
			},
			expected: struct{ success, cached, failed bool }{true, false, false},
		},
		{
			name: "scan found vulnerabilities",
			result: UnitResult{
				ID:       UnitID{Context: ContextScan, Module: "core", Component: "go", Tool: "trivy-vuln"},
				ExitCode: 1,
				Duration: 30 * time.Second,
				LogPath:  "out/scan/core/go/execution.log",
				Metrics:  map[string]any{"vulnerabilities": 5, "critical": 1},
			},
			expected: struct{ success, cached, failed bool }{false, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.success, tt.result.Success(), "Success()")
			assert.Equal(t, tt.expected.cached, tt.result.Cached(), "Cached()")
			assert.Equal(t, tt.expected.failed, tt.result.Failed(), "Failed()")
		})
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestUnitResult_EdgeCase_ExitCodeMinusTwo(t *testing.T) {
	// Exit code -2 is unusual but should be handled
	result := UnitResult{ExitCode: -2}

	// -2 is not success (not 0)
	assert.False(t, result.Success())
	// -2 is not cached (not -1)
	assert.False(t, result.Cached())
	// -2 is not failed (not > 0)
	assert.False(t, result.Failed())
}

func TestUnitResult_ZeroValue(t *testing.T) {
	var result UnitResult

	// Zero value has ExitCode 0, which means success
	assert.True(t, result.Success())
	assert.False(t, result.Cached())
	assert.False(t, result.Failed())
}
