//go:build L1 && ov
// +build L1,ov

package environments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ready-to-release/eac/go/core/workspace"
)

// loadTestContract is a helper that loads the environment contract using workspace detection.
func loadTestContract(t *testing.T) *EnvironmentContract {
	t.Helper()
	ws, err := workspace.Detect()
	require.NoError(t, err, "workspace detection should succeed")
	contract, err := LoadEnvironmentContract(ws.Root)
	require.NoError(t, err, "contract loading should succeed")
	return contract
}

func TestLoadEnvironmentContract(t *testing.T) {
	contract := loadTestContract(t)
	assert.NotNil(t, contract)
	assert.Equal(t, "0.1.0", contract.Metadata.Version)
	assert.NotEmpty(t, contract.Environments)
}

func TestGetEnvironment(t *testing.T) {
	contract := loadTestContract(t)

	tests := []struct {
		moniker     string
		expectFound bool
		expectLevel string
		expectType  string
	}{
		{"l00-01", true, "L0", "unit"},
		{"l00-02", true, "L0", "unit"},
		{"l01-01", true, "L1", "unit"},
		{"l01-02", true, "L1", "unit"},
		{"local01", true, "L2", "docker"},
		{"local02", true, "L2", "docker-compose"},
		{"plte01", true, "L3", "plte"},
		{"plte02", true, "L3", "plte"},
		{"production", true, "L4", "production"},
		{"production-inactive", true, "L4", "production"},
		{"nonexistent", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			env, err := contract.GetEnvironment(tt.moniker)
			if tt.expectFound {
				require.NoError(t, err)
				assert.Equal(t, tt.moniker, env.Moniker)
				assert.Equal(t, tt.expectLevel, env.Level)
				assert.Equal(t, tt.expectType, env.Type)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGetEnvironmentsByLevel(t *testing.T) {
	contract := loadTestContract(t)

	levels := []string{"L0", "L1", "L2", "L3", "L4"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			envs := contract.GetEnvironmentsByLevel(level)
			assert.NotEmpty(t, envs, "should have environments for level %s", level)
			for _, env := range envs {
				assert.Equal(t, level, env.Level)
			}
		})
	}
}

func TestGetEnvironmentsByType(t *testing.T) {
	contract := loadTestContract(t)

	types := []string{"unit", "docker", "plte", "production"}
	for _, envType := range types {
		t.Run(envType, func(t *testing.T) {
			envs := contract.GetEnvironmentsByType(envType)
			assert.NotEmpty(t, envs, "should have environments for type %s", envType)
			for _, env := range envs {
				assert.Equal(t, envType, env.Type)
			}
		})
	}
}

func TestValidateContract(t *testing.T) {
	contract := loadTestContract(t)
	err := contract.ValidateContract()
	assert.NoError(t, err, "valid contract should pass validation")
}

func TestValidateContract_MissingVersion(t *testing.T) {
	contract := &EnvironmentContract{
		Metadata: Metadata{Version: ""},
		Environments: []Environment{
			{Moniker: "test", Name: "Test", Level: "L2", Type: "docker"},
		},
	}

	err := contract.ValidateContract()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing version")
}

func TestValidateContract_NoEnvironments(t *testing.T) {
	contract := &EnvironmentContract{
		Metadata:     Metadata{Version: "0.1.0"},
		Environments: []Environment{},
	}

	err := contract.ValidateContract()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no environments")
}

func TestValidateContract_DuplicateMonikers(t *testing.T) {
	contract := &EnvironmentContract{
		Metadata: Metadata{Version: "0.1.0"},
		Environments: []Environment{
			{Moniker: "test", Name: "Test 1", Level: "L2", Type: "docker"},
			{Moniker: "test", Name: "Test 2", Level: "L3", Type: "plte"},
		},
	}

	err := contract.ValidateContract()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestValidateContract_InvalidLevel(t *testing.T) {
	contract := &EnvironmentContract{
		Metadata: Metadata{Version: "0.1.0"},
		Environments: []Environment{
			{Moniker: "test", Name: "Test", Level: "L5", Type: "docker"},
		},
	}

	err := contract.ValidateContract()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid level")
}

func TestGetAllMonikers(t *testing.T) {
	contract := loadTestContract(t)

	monikers := contract.GetAllMonikers()
	assert.NotEmpty(t, monikers)

	expectedMonikers := []string{
		"l00-01", "l00-02", "l01-01", "l01-02",
		"local01", "local02", "plte01", "plte02",
		"production", "production-inactive",
	}
	for _, expected := range expectedMonikers {
		assert.Contains(t, monikers, expected)
	}
}

func TestEnvironmentSystemDeps(t *testing.T) {
	contract := loadTestContract(t)

	tests := []struct {
		moniker      string
		expectedDeps []string
		expectEmpty  bool
	}{
		{"l00-01", nil, true},
		{"l01-01", []string{"@deps:go"}, false},
		{"local01", []string{"@deps:docker"}, false},
		{"plte01", []string{"@deps:kubectl", "@deps:helm"}, false},
		{"production", []string{"@deps:kubectl", "@deps:helm"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			env, err := contract.GetEnvironment(tt.moniker)
			require.NoError(t, err)

			if tt.expectEmpty {
				assert.Empty(t, env.SystemDeps)
			} else {
				for _, dep := range tt.expectedDeps {
					assert.Contains(t, env.SystemDeps, dep)
				}
			}
		})
	}
}

func TestGetTestTag(t *testing.T) {
	contract := loadTestContract(t)

	tests := []struct {
		moniker     string
		expectedTag string
	}{
		{"l00-01", "@env:l00-01"},
		{"l00-02", "@env:l00-02"},
		{"l01-01", "@env:l01-01"},
		{"l01-02", "@env:l01-02"},
		{"local01", "@env:local01"},
		{"local02", "@env:local02"},
		{"plte01", "@env:plte01"},
		{"plte02", "@env:plte02"},
		{"production", "@env:production"},
		{"production-inactive", "@env:production-inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			env, err := contract.GetEnvironment(tt.moniker)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTag, env.GetTestTag())
		})
	}
}
