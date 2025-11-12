package systemdeps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerify_Docker(t *testing.T) {
	result := Verify("@dep:docker")

	assert.Equal(t, "@dep:docker", result.Dependency)
	// Result depends on whether Docker is installed
	// Just check it doesn't panic
}

func TestVerify_Git(t *testing.T) {
	result := Verify("@dep:git")

	assert.Equal(t, "@dep:git", result.Dependency)
	// Git is likely available in development
	if result.Available {
		assert.NotEmpty(t, result.Version)
	}
}

func TestVerify_Go(t *testing.T) {
	result := Verify("@dep:go")

	assert.Equal(t, "@dep:go", result.Dependency)
	// Go must be available (we're running tests with it!)
	assert.True(t, result.Available)
	assert.NotEmpty(t, result.Version)
}

func TestVerify_Claude(t *testing.T) {
	result := Verify("@dep:claude")

	assert.Equal(t, "@dep:claude", result.Dependency)
	// Claude API key may or may not be configured
	// Just check it doesn't panic
}

func TestVerify_AzureCLI(t *testing.T) {
	result := Verify("@dep:az-cli")

	assert.Equal(t, "@dep:az-cli", result.Dependency)
	// Azure CLI may or may not be installed
	// Just check it doesn't panic
}

func TestVerify_UnknownDependency(t *testing.T) {
	result := Verify("@dep:unknown")

	assert.Equal(t, "@dep:unknown", result.Dependency)
	assert.False(t, result.Available)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "unknown dependency")
}

func TestVerifyAll(t *testing.T) {
	deps := []string{"@dep:go", "@dep:git", "@dep:docker"}

	results := VerifyAll(deps)

	assert.Len(t, results, 3)

	// Find Go result (must be available)
	var goResult *Result
	for _, r := range results {
		if r.Dependency == "@dep:go" {
			goResult = &r
			break
		}
	}

	require.NotNil(t, goResult)
	assert.True(t, goResult.Available)
}

func TestIsAvailable(t *testing.T) {
	// Go must be available
	assert.True(t, IsAvailable("@dep:go"))

	// Unknown dependency is not available
	assert.False(t, IsAvailable("@dep:unknown"))
}

func TestGetMissingDependencies(t *testing.T) {
	deps := []string{"@dep:go", "@dep:unknown-xyz"}

	missing := GetMissingDependencies(deps)

	// @dep:go should be available, @dep:unknown-xyz should not
	assert.Contains(t, missing, "@dep:unknown-xyz")
	assert.NotContains(t, missing, "@dep:go")
}
