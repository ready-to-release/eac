package systemdeps

import (
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

func TestVerify_Claude(t *testing.T) {
	result := Verify("@deps:claude")

	assert.Equal(t, "@deps:claude", result.Dependency)
	// Claude API key may or may not be configured
	// Just check it doesn't panic
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
