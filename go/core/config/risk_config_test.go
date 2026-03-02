package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRiskConfig_NoFiles(t *testing.T) {
	cfg, err := LoadRiskConfig(t.TempDir())
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

func TestLoadRiskConfig_EmbeddedDefaults(t *testing.T) {
	cfg, err := LoadRiskConfig(t.TempDir())
	require.NoError(t, err)

	// Embedded defaults provide scoring from risk-config.yml
	assert.Equal(t, 4, cfg.GetScoring().GetImpact("api"))
	assert.Equal(t, "high", cfg.GetScoring().GetCriticality("api"))
	assert.Equal(t, 1, cfg.GetScoring().GetImpact("docs"))

	// Embedded defaults provide NIST catalog URL
	assert.Contains(t, cfg.GetCatalogURL(), "nist.gov")

	// No profile loaded from embedded defaults (relative path can't resolve)
	_, err = cfg.GetProfile()
	assert.Error(t, err)
}

func TestLoadRiskConfig_UserProfileOverride(t *testing.T) {
	configDir := t.TempDir()

	profileJSON := `{
  "profile": {
    "uuid": "test-uuid",
    "metadata": {"title": "Test Profile", "version": "1.0.0", "oscal-version": "1.1.3", "last-modified": "2024-01-01T00:00:00Z"},
    "imports": [{"href": "https://example.com/catalog.json", "include-controls": [{"with-ids": ["ac-1", "ac-2"]}]}]
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-profile.json"), []byte(profileJSON), 0o644))

	userYAML := `
profile:
  path: risk-profile.json
  catalog_url: https://example.com/catalog.json
scoring:
  impact:
    api: 5
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(userYAML), 0o644))

	cfg, err := LoadRiskConfig(configDir)
	require.NoError(t, err)

	assert.Equal(t, 5, cfg.GetScoring().GetImpact("api"))
	assert.Equal(t, "https://example.com/catalog.json", cfg.GetCatalogURL())

	profile, err := cfg.GetProfile()
	require.NoError(t, err)
	assert.Equal(t, "Test Profile", profile.Title())
	assert.True(t, profile.HasControl("ac-1"))
}

func TestLoadRiskConfig_WithUserOverrides(t *testing.T) {
	configDir := t.TempDir()

	userYAML := `
profile:
  catalog_url: https://override.com/catalog.json
scoring:
  impact:
    api: 5
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(userYAML), 0o644))

	cfg, err := LoadRiskConfig(configDir)
	require.NoError(t, err)

	// Overridden values
	assert.Equal(t, 5, cfg.GetScoring().GetImpact("api"))
	assert.Equal(t, "https://override.com/catalog.json", cfg.GetCatalogURL())

	// Preserved embedded default values
	assert.Equal(t, 1, cfg.GetScoring().GetImpact("docs"))
}

func TestLoadRiskConfig_ModuleProfiles(t *testing.T) {
	configDir := t.TempDir()

	mainProfileJSON := `{
  "profile": {
    "uuid": "main-uuid",
    "metadata": {"title": "Main Profile", "version": "1.0.0", "oscal-version": "1.1.3", "last-modified": "2024-01-01T00:00:00Z"},
    "imports": [{"href": "https://example.com/catalog.json", "include-controls": [{"with-ids": ["ac-1"]}]}]
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "main-profile.json"), []byte(mainProfileJSON), 0o644))

	moduleProfileJSON := `{
  "profile": {
    "uuid": "billing-uuid",
    "metadata": {"title": "Billing Profile", "version": "1.0.0", "oscal-version": "1.1.3", "last-modified": "2024-01-01T00:00:00Z"},
    "imports": [{"href": "https://example.com/catalog.json", "include-controls": [{"with-ids": ["ac-1", "ac-2", "au-1"]}]}]
  }
}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "billing-profile.json"), []byte(moduleProfileJSON), 0o644))

	configYAML := `
profile:
  path: main-profile.json
module_profiles:
  billing-service:
    path: billing-profile.json
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(configYAML), 0o644))

	cfg, err := LoadRiskConfig(configDir)
	require.NoError(t, err)

	mainProfile, err := cfg.GetProfile()
	require.NoError(t, err)
	assert.Equal(t, "Main Profile", mainProfile.Title())
	assert.Len(t, mainProfile.ControlIDs(), 1)

	billingProfile, err := cfg.GetModuleProfile("billing-service")
	require.NoError(t, err)
	assert.Equal(t, "Billing Profile", billingProfile.Title())
	assert.Len(t, billingProfile.ControlIDs(), 3)

	unknownProfile, err := cfg.GetModuleProfile("unknown-service")
	require.NoError(t, err)
	assert.Equal(t, "Main Profile", unknownProfile.Title())

	modules := cfg.ListModuleProfiles()
	assert.Len(t, modules, 1)
	assert.Contains(t, modules, "billing-service")
}

func TestLoadRiskConfig_InvalidYAML(t *testing.T) {
	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte("invalid: yaml: here:"), 0o644))

	_, err := LoadRiskConfig(configDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing")
}

func TestLoadRiskConfig_InvalidProfileJSON(t *testing.T) {
	configDir := t.TempDir()

	configYAML := `
profile:
  path: invalid-profile.json
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(configYAML), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "invalid-profile.json"), []byte("not json"), 0o644))

	_, err := LoadRiskConfig(configDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading profile")
}

func TestLoadRiskConfig_MissingProfileIsOK(t *testing.T) {
	configDir := t.TempDir()

	configYAML := `
profile:
  path: nonexistent-profile.json
scoring:
  impact:
    api: 5
`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "risk-config.yml"), []byte(configYAML), 0o644))

	cfg, err := LoadRiskConfig(configDir)
	require.NoError(t, err)

	assert.Equal(t, 5, cfg.GetScoring().GetImpact("api"))

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

	t.Run("absolute path unchanged", func(t *testing.T) {
		absPath := filepath.Join(tmpDir, "other", "system.json")
		got := cfg.resolvePath(absPath)
		assert.Equal(t, absPath, got)
	})
}

func TestSecurityConfig_Risk_Integration(t *testing.T) {
	cfg, err := LoadSecurityConfig(t.TempDir())
	require.NoError(t, err)

	risk := cfg.Risk()
	require.NotNil(t, risk, "Risk() should return non-nil")

	assert.Equal(t, 4, risk.GetScoring().GetImpact("api"))
	assert.Equal(t, "high", risk.GetScoring().GetCriticality("api"))
}

func TestSecurityConfig_Risk_NilWhenMissing(t *testing.T) {
	cfg := &SecurityConfig{}
	assert.Nil(t, cfg.Risk(), "Risk() should return nil when not loaded")
}
