package environments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvironmentContract(t *testing.T) {
	contract, err := LoadEnvironmentContract()
	require.NoError(t, err)
	assert.NotNil(t, contract)
	assert.Equal(t, "0.1.0", contract.Metadata.Version)
	assert.NotEmpty(t, contract.Environments)
}

func TestGetEnvironment(t *testing.T) {
	contract, err := LoadEnvironmentContract()
	require.NoError(t, err)

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
	contract, err := LoadEnvironmentContract()
	require.NoError(t, err)

	l0Envs := contract.GetEnvironmentsByLevel("L0")
	assert.NotEmpty(t, l0Envs)
	for _, env := range l0Envs {
		assert.Equal(t, "L0", env.Level)
	}

	l1Envs := contract.GetEnvironmentsByLevel("L1")
	assert.NotEmpty(t, l1Envs)
	for _, env := range l1Envs {
		assert.Equal(t, "L1", env.Level)
	}

	l2Envs := contract.GetEnvironmentsByLevel("L2")
	assert.NotEmpty(t, l2Envs)
	for _, env := range l2Envs {
		assert.Equal(t, "L2", env.Level)
	}

	l3Envs := contract.GetEnvironmentsByLevel("L3")
	assert.NotEmpty(t, l3Envs)
	for _, env := range l3Envs {
		assert.Equal(t, "L3", env.Level)
	}

	l4Envs := contract.GetEnvironmentsByLevel("L4")
	assert.NotEmpty(t, l4Envs)
	for _, env := range l4Envs {
		assert.Equal(t, "L4", env.Level)
	}
}

func TestGetEnvironmentsByType(t *testing.T) {
	contract, err := LoadEnvironmentContract()
	require.NoError(t, err)

	unitEnvs := contract.GetEnvironmentsByType("unit")
	assert.NotEmpty(t, unitEnvs)
	for _, env := range unitEnvs {
		assert.Equal(t, "unit", env.Type)
	}

	dockerEnvs := contract.GetEnvironmentsByType("docker")
	assert.NotEmpty(t, dockerEnvs)
	for _, env := range dockerEnvs {
		assert.Equal(t, "docker", env.Type)
	}

	plteEnvs := contract.GetEnvironmentsByType("plte")
	assert.NotEmpty(t, plteEnvs)
	for _, env := range plteEnvs {
		assert.Equal(t, "plte", env.Type)
	}

	productionEnvs := contract.GetEnvironmentsByType("production")
	assert.NotEmpty(t, productionEnvs)
	for _, env := range productionEnvs {
		assert.Equal(t, "production", env.Type)
	}
}

func TestValidateContract(t *testing.T) {
	contract, err := LoadEnvironmentContract()
	require.NoError(t, err)

	// Valid contract should pass validation
	err = contract.ValidateContract()
	assert.NoError(t, err)
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
	contract, err := LoadEnvironmentContract()
	require.NoError(t, err)

	monikers := contract.GetAllMonikers()
	assert.NotEmpty(t, monikers)
	assert.Contains(t, monikers, "l00-01")
	assert.Contains(t, monikers, "l00-02")
	assert.Contains(t, monikers, "l01-01")
	assert.Contains(t, monikers, "l01-02")
	assert.Contains(t, monikers, "local01")
	assert.Contains(t, monikers, "local02")
	assert.Contains(t, monikers, "plte01")
	assert.Contains(t, monikers, "plte02")
	assert.Contains(t, monikers, "production")
	assert.Contains(t, monikers, "production-inactive")
}

func TestEnvironmentSystemDeps(t *testing.T) {
	contract, err := LoadEnvironmentContract()
	require.NoError(t, err)

	l00_01, err := contract.GetEnvironment("l00-01")
	require.NoError(t, err)
	assert.Empty(t, l00_01.SystemDeps)

	l01_01, err := contract.GetEnvironment("l01-01")
	require.NoError(t, err)
	assert.Contains(t, l01_01.SystemDeps, "@deps:go")

	local01, err := contract.GetEnvironment("local01")
	require.NoError(t, err)
	assert.Contains(t, local01.SystemDeps, "@deps:docker")

	plte01, err := contract.GetEnvironment("plte01")
	require.NoError(t, err)
	assert.Contains(t, plte01.SystemDeps, "@deps:kubectl")
	assert.Contains(t, plte01.SystemDeps, "@deps:helm")

	production, err := contract.GetEnvironment("production")
	require.NoError(t, err)
	assert.Contains(t, production.SystemDeps, "@deps:kubectl")
	assert.Contains(t, production.SystemDeps, "@deps:helm")
}

func TestGetTestTag(t *testing.T) {
	contract, err := LoadEnvironmentContract()
	require.NoError(t, err)

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
