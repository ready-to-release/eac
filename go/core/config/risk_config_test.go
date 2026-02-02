package config

import (
	"os"
	"path/filepath"
	"testing"

	security "github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRiskConfig_ImplementsPort(t *testing.T) {
	var _ security.RiskConfigPort = (*RiskConfig)(nil)
}

func TestLoadRiskConfig_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	cfg, err := LoadRiskConfig(tmpDir, configDir)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Should have default scoring
	assert.Equal(t, 4, cfg.GetScoring().GetImpact("api"))
	assert.Equal(t, "high", cfg.GetScoring().GetCriticality("api"))

	// No profile should be loaded
	_, err = cfg.GetProfile()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no risk profile configured")
}

func TestLoadRiskConfig_WithDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create contract defaults structure
	defaultsDir := filepath.Join(tmpDir, "contracts", "security", "0.1.0", "defaults")
	require.NoError(t, os.MkdirAll(defaultsDir, 0o755))

	// Create risk-config.yml
	riskConfigYAML := `
profile:
  path: risk-profile.json
  catalog_url: https://example.com/catalog.json
scoring:
  impact:
    api: 5
  criticality:
    api: high
`
	require.NoError(t, os.WriteFile(filepath.Join(defaultsDir, "risk-config.yml"), []byte(riskConfigYAML), 0o644))

	// Create profile file
	profileJSON := `{
  "profile": {
    "uuid": "test-uuid",
    "metadata": {"title": "Test Profile", "version": "1.0.0", "oscal-version": "1.1.3", "last-modified": "2024-01-01T00:00:00Z"},
    "imports": [{"href": "https://example.com/catalog.json", "include-controls": [{"with-ids": ["ac-1", "ac-2"]}]}]
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(defaultsDir, "risk-profile.json"), []byte(profileJSON), 0o644))

	configDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	cfg, err := LoadRiskConfig(tmpDir, configDir)
	require.NoError(t, err)

	// Check scoring
	assert.Equal(t, 5, cfg.GetScoring().GetImpact("api"))

	// Check catalog URL
	assert.Equal(t, "https://example.com/catalog.json", cfg.GetCatalogURL())

	// Check profile
	profile, err := cfg.GetProfile()
	require.NoError(t, err)
	assert.Equal(t, "Test Profile", profile.Title())
	assert.True(t, profile.HasControl("ac-1"))
}

func TestLoadRiskConfig_WithUserOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	// Create contract defaults
	defaultsDir := filepath.Join(tmpDir, "contracts", "security", "0.1.0", "defaults")
	require.NoError(t, os.MkdirAll(defaultsDir, 0o755))

	defaultsYAML := `
profile:
  catalog_url: https://default.com/catalog.json
scoring:
  impact:
    api: 4
    docs: 1
`
	require.NoError(t, os.WriteFile(filepath.Join(defaultsDir, "risk-config.yml"), []byte(defaultsYAML), 0o644))

	// Create user overrides
	configDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	userYAML := `
profile:
  catalog_url: https://override.com/catalog.json
scoring:
  impact:
    api: 5
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(userYAML), 0o644))

	cfg, err := LoadRiskConfig(tmpDir, configDir)
	require.NoError(t, err)

	// Overridden values
	assert.Equal(t, 5, cfg.GetScoring().GetImpact("api"))
	assert.Equal(t, "https://override.com/catalog.json", cfg.GetCatalogURL())

	// Preserved default values
	assert.Equal(t, 1, cfg.GetScoring().GetImpact("docs"))
}

func TestLoadRiskConfig_ModuleProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	// Create main profile
	mainProfileJSON := `{
  "profile": {
    "uuid": "main-uuid",
    "metadata": {"title": "Main Profile", "version": "1.0.0", "oscal-version": "1.1.3", "last-modified": "2024-01-01T00:00:00Z"},
    "imports": [{"href": "https://example.com/catalog.json", "include-controls": [{"with-ids": ["ac-1"]}]}]
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "main-profile.json"), []byte(mainProfileJSON), 0o644))

	// Create module profile
	moduleProfileJSON := `{
  "profile": {
    "uuid": "billing-uuid",
    "metadata": {"title": "Billing Profile", "version": "1.0.0", "oscal-version": "1.1.3", "last-modified": "2024-01-01T00:00:00Z"},
    "imports": [{"href": "https://example.com/catalog.json", "include-controls": [{"with-ids": ["ac-1", "ac-2", "au-1"]}]}]
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "billing-profile.json"), []byte(moduleProfileJSON), 0o644))

	// Create config
	configYAML := `
profile:
  path: main-profile.json
module_profiles:
  billing-service:
    path: billing-profile.json
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(configYAML), 0o644))

	cfg, err := LoadRiskConfig(tmpDir, configDir)
	require.NoError(t, err)

	// Main profile
	mainProfile, err := cfg.GetProfile()
	require.NoError(t, err)
	assert.Equal(t, "Main Profile", mainProfile.Title())
	assert.Len(t, mainProfile.ControlIDs(), 1)

	// Module profile
	billingProfile, err := cfg.GetModuleProfile("billing-service")
	require.NoError(t, err)
	assert.Equal(t, "Billing Profile", billingProfile.Title())
	assert.Len(t, billingProfile.ControlIDs(), 3)

	// Unknown module falls back to main
	unknownProfile, err := cfg.GetModuleProfile("unknown-service")
	require.NoError(t, err)
	assert.Equal(t, "Main Profile", unknownProfile.Title())

	// List module profiles
	modules := cfg.ListModuleProfiles()
	assert.Len(t, modules, 1)
	assert.Contains(t, modules, "billing-service")
}

func TestLoadRiskConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte("invalid: yaml: here:"), 0o644))

	_, err := LoadRiskConfig(tmpDir, configDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing")
}

func TestLoadRiskConfig_InvalidProfileJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	configYAML := `
profile:
  path: invalid-profile.json
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(configYAML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "invalid-profile.json"), []byte("not json"), 0o644))

	_, err := LoadRiskConfig(tmpDir, configDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading profile")
}

func TestLoadRiskConfig_MissingProfileIsOK(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	// Profile path points to non-existent file - this is OK
	configYAML := `
profile:
  path: nonexistent-profile.json
scoring:
  impact:
    api: 5
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(configYAML), 0o644))

	cfg, err := LoadRiskConfig(tmpDir, configDir)
	require.NoError(t, err)

	// Scoring should still work
	assert.Equal(t, 5, cfg.GetScoring().GetImpact("api"))

	// Profile should be nil
	_, err = cfg.GetProfile()
	assert.Error(t, err)
}

func TestRiskConfig_resolvePath(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &RiskConfig{configDir: tmpDir}

	t.Run("relative path", func(t *testing.T) {
		got := cfg.resolvePath("profile.json")
		expected := filepath.Join(tmpDir, "profile.json")
		assert.Equal(t, expected, got)
	})

	t.Run("relative with subdirectory", func(t *testing.T) {
		got := cfg.resolvePath(filepath.Join("profiles", "main.json"))
		expected := filepath.Join(tmpDir, "profiles", "main.json")
		assert.Equal(t, expected, got)
	})

	t.Run("absolute path unchanged", func(t *testing.T) {
		absPath := filepath.Join(tmpDir, "other", "system.json")
		got := cfg.resolvePath(absPath)
		assert.Equal(t, absPath, got)
	})
}

func TestSecurityConfig_Risk_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".r2r", "eac")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	// Create contract defaults structure
	defaultsDir := filepath.Join(tmpDir, "contracts", "security", "0.1.0", "defaults")
	require.NoError(t, os.MkdirAll(defaultsDir, 0o755))

	// Create minimal scanners.yml
	scannersYAML := `scanners: {}`
	require.NoError(t, os.WriteFile(filepath.Join(defaultsDir, "scanners.yml"), []byte(scannersYAML), 0o644))

	// Create minimal policies.yml
	policiesYAML := `default: []`
	require.NoError(t, os.WriteFile(filepath.Join(defaultsDir, "policies.yml"), []byte(policiesYAML), 0o644))

	// Create risk-config.yml with scoring
	riskConfigYAML := `
scoring:
  impact:
    api: 5
  criticality:
    api: high
`
	require.NoError(t, os.WriteFile(filepath.Join(defaultsDir, "risk-config.yml"), []byte(riskConfigYAML), 0o644))

	cfg, err := LoadSecurityConfig(tmpDir, configDir)
	require.NoError(t, err)

	// Access risk config through SecurityConfig
	risk := cfg.Risk()
	require.NotNil(t, risk, "Risk() should return non-nil")

	// Verify scoring works
	assert.Equal(t, 5, risk.GetScoring().GetImpact("api"))
	assert.Equal(t, "high", risk.GetScoring().GetCriticality("api"))
}

func TestSecurityConfig_Risk_NilWhenMissing(t *testing.T) {
	cfg := &SecurityConfig{}
	assert.Nil(t, cfg.Risk(), "Risk() should return nil when not loaded")
}
