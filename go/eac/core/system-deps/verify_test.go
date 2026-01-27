//go:build L1 && ov
// +build L1,ov

package systemdeps

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerify_Docker(t *testing.T) {
	result := Verify("@deps:docker")

	assert.Equal(t, "@deps:docker", result.Dependency)
	// Result depends on whether Docker is installed
	// Just check it doesn't panic
}

func TestVerify_Git(t *testing.T) {
	result := Verify("@deps:git")

	assert.Equal(t, "@deps:git", result.Dependency)
	// Git is likely available in development
	if result.Available {
		assert.NotEmpty(t, result.Version)
	}
}

func TestVerify_Go(t *testing.T) {
	result := Verify("@deps:go")

	assert.Equal(t, "@deps:go", result.Dependency)
	// Go must be available (we're running tests with it!)
	assert.True(t, result.Available)
	assert.NotEmpty(t, result.Version)
}

func TestVerify_AzureCLI(t *testing.T) {
	result := Verify("@deps:az-cli")

	assert.Equal(t, "@deps:az-cli", result.Dependency)
	// Azure CLI may or may not be installed
	// Just check it doesn't panic
}

func TestVerify_UnknownDependency(t *testing.T) {
	result := Verify("@deps:unknown")

	assert.Equal(t, "@deps:unknown", result.Dependency)
	assert.False(t, result.Available)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "unknown dependency")
}

func TestVerifyAll(t *testing.T) {
	deps := []string{"@deps:go", "@deps:git", "@deps:docker"}

	results := VerifyAll(deps)

	assert.Len(t, results, 3)

	// Find Go result (must be available)
	var goResult *Result
	for _, r := range results {
		if r.Dependency == "@deps:go" {
			goResult = &r
			break
		}
	}

	require.NotNil(t, goResult)
	assert.True(t, goResult.Available)
}

func TestIsAvailable(t *testing.T) {
	// Go must be available
	assert.True(t, IsAvailable("@deps:go"))

	// Unknown dependency is not available
	assert.False(t, IsAvailable("@deps:unknown"))
}

func TestGetMissingDependencies(t *testing.T) {
	deps := []string{"@deps:go", "@deps:unknown-xyz"}

	missing := GetMissingDependencies(deps)

	// @deps:go should be available, @deps:unknown-xyz should not
	assert.Contains(t, missing, "@deps:unknown-xyz")
	assert.NotContains(t, missing, "@deps:go")
}

// TestVerifyAllForPhase_ExistingBehavior tests existing phase filtering.
// This should pass before platform filtering is implemented.
func TestVerifyAllForPhase_ExistingBehavior(t *testing.T) {
	t.Run("phase-agnostic dependency is always verified", func(t *testing.T) {
		// Go has phases=["build","test"] so should be verified in build phase
		results := VerifyAllForPhase([]string{"go"}, "build")

		require.Len(t, results, 1)
		assert.Equal(t, "go", results[0].Dependency)
		assert.True(t, results[0].Available)
		assert.NotEmpty(t, results[0].Version)
		assert.NotContains(t, results[0].Version, "n/a")
	})

	t.Run("dependency with wrong phase is skipped", func(t *testing.T) {
		// Bash has phases=["command","test"] - NOT "build"
		// So in build phase it should be marked as available with "n/a (build phase)"
		results := VerifyAllForPhase([]string{"bash"}, "build")

		require.Len(t, results, 1)
		assert.Equal(t, "bash", results[0].Dependency)
		assert.True(t, results[0].Available, "should be marked available (not needed)")
		assert.Contains(t, results[0].Version, "n/a")
		assert.Contains(t, results[0].Version, "build phase")
	})

	t.Run("dependency with matching phase is verified", func(t *testing.T) {
		// Bash has phases=["command","test"] - includes "test"
		results := VerifyAllForPhase([]string{"bash"}, "test")

		require.Len(t, results, 1)
		assert.Equal(t, "bash", results[0].Dependency)
		// Result depends on whether bash is installed
		// Just verify it was actually checked (not skipped)
		if !results[0].Available {
			assert.Error(t, results[0].Error)
			assert.NotContains(t, results[0].Version, "n/a")
		}
	})
}

// TestVerifyAllForPhase_PlatformFiltering tests new platform-aware filtering.
// These tests will FAIL until platform filtering is implemented in verify.go.
func TestVerifyAllForPhase_PlatformFiltering(t *testing.T) {
	currentPlatform := runtime.GOOS

	t.Run("dependency not matching platform is skipped with n/a reason", func(t *testing.T) {
		// Determine which dependency to test based on current platform
		var testDep string
		var expectedPlatform string

		switch currentPlatform {
		case "windows":
			// On Windows, bash (linux/darwin only) should be skipped
			testDep = "bash"
			expectedPlatform = "windows"
		case "linux", "darwin":
			// On Unix, pwsh (windows only) should be skipped
			testDep = "pwsh"
			expectedPlatform = currentPlatform
		default:
			t.Skip("Unsupported platform for this test")
		}

		// Test in "test" phase where the dependency would normally apply
		results := VerifyAllForPhase([]string{testDep}, "test")

		require.Len(t, results, 1)
		result := results[0]

		assert.Equal(t, testDep, result.Dependency)
		assert.True(t, result.Available, "should be marked available (not needed on this platform)")
		assert.Contains(t, result.Version, "n/a", "should indicate not applicable")
		assert.Contains(t, result.Version, expectedPlatform, "should mention current platform")
		assert.Contains(t, result.Version, "platform", "should indicate platform reason")
	})

	t.Run("dependency matching platform is verified normally", func(t *testing.T) {
		// Test platform-specific dependency that matches current platform
		var testDep string

		switch currentPlatform {
		case "windows":
			testDep = "pwsh" // PowerShell is for Windows
		case "linux", "darwin":
			testDep = "bash" // Bash is for Unix
		default:
			t.Skip("Unsupported platform for this test")
		}

		// Test in "test" phase where the dependency applies
		results := VerifyAllForPhase([]string{testDep}, "test")

		require.Len(t, results, 1)
		result := results[0]

		assert.Equal(t, testDep, result.Dependency)
		// Should be actually verified, not skipped
		// Availability depends on whether it's installed
		if !result.Available {
			assert.Error(t, result.Error)
			assert.NotContains(t, result.Version, "n/a")
		} else {
			// If available, should have real version
			assert.NotContains(t, result.Version, "n/a")
		}
	})

	t.Run("platform-agnostic dependency is verified on all platforms", func(t *testing.T) {
		// Go has no platform restriction (empty platforms array)
		results := VerifyAllForPhase([]string{"go"}, "build")

		require.Len(t, results, 1)
		result := results[0]

		assert.Equal(t, "go", result.Dependency)
		assert.True(t, result.Available, "Go must be available (we're running tests)")
		assert.NotEmpty(t, result.Version)
		assert.NotContains(t, result.Version, "n/a")
	})
}

// TestVerifyAllForPhase_CombinedFiltering tests both phase AND platform filtering.
// These tests will FAIL until both filters work together correctly.
func TestVerifyAllForPhase_CombinedFiltering(t *testing.T) {
	currentPlatform := runtime.GOOS

	t.Run("wrong phase AND wrong platform shows both in reason", func(t *testing.T) {
		var testDep string
		var expectedPhase string

		switch currentPlatform {
		case "windows":
			// Bash: platforms=[linux,darwin], phases=[command,test]
			// Test in "build" phase on Windows = both wrong
			testDep = "bash"
			expectedPhase = "build"
		case "linux", "darwin":
			// PowerShell: platforms=[windows], phases=[command,test]
			// Test in "build" phase on Unix = both wrong
			testDep = "pwsh"
			expectedPhase = "build"
		default:
			t.Skip("Unsupported platform")
		}

		results := VerifyAllForPhase([]string{testDep}, expectedPhase)

		require.Len(t, results, 1)
		result := results[0]

		assert.True(t, result.Available, "should be marked available (not needed)")
		assert.Contains(t, result.Version, "n/a")

		// Should mention EITHER phase OR platform (implementation can choose which to report)
		hasPhaseReason := strings.Contains(result.Version, "phase")
		hasPlatformReason := strings.Contains(result.Version, "platform")
		assert.True(t, hasPhaseReason || hasPlatformReason,
			"should mention why dependency is not applicable")
	})

	t.Run("right phase BUT wrong platform shows platform reason", func(t *testing.T) {
		var testDep string

		switch currentPlatform {
		case "windows":
			// Bash: platforms=[linux,darwin], phases=[command,test]
			// Test in "test" phase (right) on Windows (wrong)
			testDep = "bash"
		case "linux", "darwin":
			// PowerShell: platforms=[windows], phases=[command,test]
			// Test in "test" phase (right) on Unix (wrong)
			testDep = "pwsh"
		default:
			t.Skip("Unsupported platform")
		}

		results := VerifyAllForPhase([]string{testDep}, "test")

		require.Len(t, results, 1)
		result := results[0]

		assert.True(t, result.Available, "should be marked available (not needed)")
		assert.Contains(t, result.Version, "n/a")
		assert.Contains(t, result.Version, "platform", "should indicate platform mismatch")
		assert.Contains(t, result.Version, currentPlatform, "should mention current platform")
	})

	t.Run("wrong phase BUT right platform shows phase reason", func(t *testing.T) {
		var testDep string

		switch currentPlatform {
		case "windows":
			// PowerShell: platforms=[windows], phases=[command,test]
			// Test in "build" phase (wrong) on Windows (right)
			testDep = "pwsh"
		case "linux", "darwin":
			// Bash: platforms=[linux,darwin], phases=[command,test]
			// Test in "build" phase (wrong) on Unix (right)
			testDep = "bash"
		default:
			t.Skip("Unsupported platform")
		}

		results := VerifyAllForPhase([]string{testDep}, "build")

		require.Len(t, results, 1)
		result := results[0]

		assert.True(t, result.Available, "should be marked available (not needed)")
		assert.Contains(t, result.Version, "n/a")
		assert.Contains(t, result.Version, "build phase", "should indicate phase mismatch")
	})

	t.Run("right phase AND right platform verifies normally", func(t *testing.T) {
		var testDep string

		switch currentPlatform {
		case "windows":
			testDep = "pwsh" // PowerShell for Windows
		case "linux", "darwin":
			testDep = "bash" // Bash for Unix
		default:
			t.Skip("Unsupported platform")
		}

		// Test in "test" phase where it applies
		results := VerifyAllForPhase([]string{testDep}, "test")

		require.Len(t, results, 1)
		result := results[0]

		// Should be actually verified, not skipped
		if !result.Available {
			assert.Error(t, result.Error, "if not available, should have error")
			assert.NotContains(t, result.Version, "n/a", "real check should not say n/a")
		}
	})
}

// TestVerifyAllForPhase_MultipleDependencies tests filtering with mixed dependency lists.
// These tests will FAIL until platform filtering is fully implemented.
func TestVerifyAllForPhase_MultipleDependencies(t *testing.T) {
	currentPlatform := runtime.GOOS

	t.Run("mix of platform-specific and universal dependencies", func(t *testing.T) {
		// Test with: go (universal), platform-specific shell
		var platformDep string
		switch currentPlatform {
		case "windows":
			platformDep = "bash" // Should be skipped on Windows
		case "linux", "darwin":
			platformDep = "pwsh" // Should be skipped on Unix
		default:
			t.Skip("Unsupported platform")
		}

		deps := []string{"go", platformDep}
		results := VerifyAllForPhase(deps, "build")

		require.Len(t, results, 2)

		// Find results
		var goResult, shellResult *Result
		for i := range results {
			if results[i].Dependency == "go" {
				goResult = &results[i]
			} else if results[i].Dependency == platformDep {
				shellResult = &results[i]
			}
		}

		require.NotNil(t, goResult, "should have Go result")
		require.NotNil(t, shellResult, "should have shell result")

		// Go should be verified (universal + right phase)
		assert.True(t, goResult.Available)
		assert.NotContains(t, goResult.Version, "n/a")

		// Platform-specific shell should be skipped (wrong phase or platform)
		assert.True(t, shellResult.Available, "should be marked available (not needed)")
		assert.Contains(t, shellResult.Version, "n/a")
	})

	t.Run("all dependencies skipped due to platform", func(t *testing.T) {
		var deps []string

		switch currentPlatform {
		case "windows":
			// Both Unix-only tools
			deps = []string{"bash"}
		case "linux", "darwin":
			// Windows-only tool
			deps = []string{"pwsh"}
		default:
			t.Skip("Unsupported platform")
		}

		results := VerifyAllForPhase(deps, "test")

		require.Len(t, results, len(deps))

		// All should be skipped
		for _, result := range results {
			assert.True(t, result.Available, "should be marked available (not needed)")
			assert.Contains(t, result.Version, "n/a")
		}
	})
}

// TestVerifyAllForPhase_ReasonStringFormats tests the exact format of skip reasons.
// This documents the expected user-facing messages.
func TestVerifyAllForPhase_ReasonStringFormats(t *testing.T) {
	currentPlatform := runtime.GOOS

	t.Run("phase-only skip reason format", func(t *testing.T) {
		// Use dependency that matches platform but not phase
		var testDep string

		switch currentPlatform {
		case "windows":
			testDep = "pwsh" // Windows only, command/test phases
		case "linux", "darwin":
			testDep = "bash" // Unix only, command/test phases
		default:
			t.Skip("Unsupported platform")
		}

		results := VerifyAllForPhase([]string{testDep}, "build")

		require.Len(t, results, 1)
		result := results[0]

		// Expected format: "n/a (build phase)"
		expected := "n/a (build phase)"
		assert.Equal(t, expected, result.Version,
			"phase-only skip reason should match format: %s", expected)
	})

	t.Run("platform-only skip reason format", func(t *testing.T) {
		// Use dependency that matches phase but not platform
		var testDep string

		switch currentPlatform {
		case "windows":
			testDep = "bash" // Unix only, command/test phases
		case "linux", "darwin":
			testDep = "pwsh" // Windows only, command/test phases
		default:
			t.Skip("Unsupported platform")
		}

		results := VerifyAllForPhase([]string{testDep}, "test")

		require.Len(t, results, 1)
		result := results[0]

		// Expected format: "n/a (windows platform)" or "n/a (linux platform)"
		expected := fmt.Sprintf("n/a (%s platform)", currentPlatform)
		assert.Equal(t, expected, result.Version,
			"platform-only skip reason should match format: %s", expected)
	})

	t.Run("combined skip reason format", func(t *testing.T) {
		// Use dependency that matches neither phase nor platform
		var testDep string

		switch currentPlatform {
		case "windows":
			testDep = "bash" // Unix only, command/test phases
		case "linux", "darwin":
			testDep = "pwsh" // Windows only, command/test phases
		default:
			t.Skip("Unsupported platform")
		}

		results := VerifyAllForPhase([]string{testDep}, "build")

		require.Len(t, results, 1)
		result := results[0]

		// Expected format: "n/a (build phase, windows platform)" or similar
		// Implementation can choose which reason takes precedence
		// or show both - we just verify it contains "n/a"
		assert.Contains(t, result.Version, "n/a")
		assert.True(t,
			strings.Contains(result.Version, "phase") ||
				strings.Contains(result.Version, "platform"),
			"should mention either phase or platform reason")
	})
}

// TestVerifyAllForPhase_EdgeCases tests edge cases and error conditions.
func TestVerifyAllForPhase_EdgeCases(t *testing.T) {
	t.Run("empty dependency list", func(t *testing.T) {
		results := VerifyAllForPhase([]string{}, "build")
		assert.Empty(t, results)
	})

	t.Run("unknown dependency still returns error", func(t *testing.T) {
		results := VerifyAllForPhase([]string{"unknown-dep-xyz"}, "build")

		require.Len(t, results, 1)
		result := results[0]

		assert.False(t, result.Available)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "unknown")
	})

	t.Run("empty phase string", func(t *testing.T) {
		// Dependencies with explicit phases should not match empty phase
		results := VerifyAllForPhase([]string{"bash"}, "")

		require.Len(t, results, 1)
		// Implementation-dependent: could verify or skip
		// Just verify it doesn't panic
	})

	t.Run("case sensitivity of phase names", func(t *testing.T) {
		// Phase names should be case-sensitive
		results := VerifyAllForPhase([]string{"bash"}, "BUILD")

		require.Len(t, results, 1)
		result := results[0]

		// "BUILD" should not match "build" phase
		assert.True(t, result.Available, "should be marked available (wrong phase)")
		assert.Contains(t, result.Version, "n/a")
	})
}
