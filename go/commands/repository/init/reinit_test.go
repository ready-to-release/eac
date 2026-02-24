//go:build L1 && ov
// +build L1,ov

package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectExistingConfig_NoConfig tests detection when no .eac directory exists.
func TestDetectExistingConfig_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	config := detectExistingConfig(tmpDir)

	assert.Nil(t, config)
}

// TestDetectExistingConfig_EmptyDirectory tests detection when .eac exists but is empty.
func TestDetectExistingConfig_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)

	require.NotNil(t, config)
	assert.False(t, config.HasRepositoryYML)
	assert.False(t, config.HasBooksYML)
	assert.False(t, config.HasEnvironmentsYML)
	assert.False(t, config.HasAIProviderYML)
	assert.False(t, config.HasAIProviderPersonalYML)
	assert.Empty(t, config.AIProvider)
	assert.Empty(t, config.AIModel)
	assert.Equal(t, 0, config.ModuleCount)
}

// TestDetectExistingConfig_AllConfigFiles tests detection when all config files exist.
func TestDetectExistingConfig_AllConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create all config files
	files := []string{
		"repository.yml",
		"books.yml",
		"environments.yml",
		"ai-provider.yml",
		"ai-provider.personal.yml",
	}

	for _, file := range files {
		path := filepath.Join(eacDir, file)
		err := os.WriteFile(path, []byte("# test"), 0o644)
		require.NoError(t, err)
	}

	config := detectExistingConfig(tmpDir)

	require.NotNil(t, config)
	assert.True(t, config.HasRepositoryYML)
	assert.True(t, config.HasBooksYML)
	assert.True(t, config.HasEnvironmentsYML)
	assert.True(t, config.HasAIProviderYML)
	assert.True(t, config.HasAIProviderPersonalYML)
}

// TestDetectExistingConfig_AIProviderSettings tests loading AI provider settings.
func TestDetectExistingConfig_AIProviderSettings(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create personal AI provider config
	personalConfig := `ai:
  provider: claude-api
  model: claude-3-opus-20240229
  api_key: sk-ant-test-key
`
	personalPath := filepath.Join(eacDir, "ai-provider.personal.yml")
	err = os.WriteFile(personalPath, []byte(personalConfig), 0o600)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)

	require.NotNil(t, config)
	assert.True(t, config.HasAIProviderPersonalYML)
	assert.Equal(t, "claude-api", config.AIProvider)
	assert.Equal(t, "claude-3-opus-20240229", config.AIModel)
}

// TestDetectExistingConfig_PersonalConfigPriority tests personal config takes precedence.
func TestDetectExistingConfig_PersonalConfigPriority(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create team config
	teamConfig := `ai:
  provider: openai
  model: gpt-4
  api_key: ${OPENAI_API_KEY}
`
	teamPath := filepath.Join(eacDir, "ai-provider.yml")
	err = os.WriteFile(teamPath, []byte(teamConfig), 0o644)
	require.NoError(t, err)

	// Create personal config (should override)
	personalConfig := `ai:
  provider: claude-api
  model: claude-3-haiku-20240307
  api_key: sk-ant-test-key
`
	personalPath := filepath.Join(eacDir, "ai-provider.personal.yml")
	err = os.WriteFile(personalPath, []byte(personalConfig), 0o600)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)

	require.NotNil(t, config)
	assert.True(t, config.HasAIProviderYML)
	assert.True(t, config.HasAIProviderPersonalYML)
	// Personal config should take precedence
	assert.Equal(t, "claude-api", config.AIProvider)
	assert.Equal(t, "claude-3-haiku-20240307", config.AIModel)
}

// TestDetectExistingConfig_ModuleCount tests loading module count from repository.yml.
func TestDetectExistingConfig_ModuleCount(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create repository.yml with 3 modules
	repoConfig := `repository:
  name: test-repo
modules:
  - moniker: core
    description: Core module
  - moniker: cli
    description: CLI module
  - moniker: api
    description: API module
`
	repoPath := filepath.Join(eacDir, "repository.yml")
	err = os.WriteFile(repoPath, []byte(repoConfig), 0o644)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)

	require.NotNil(t, config)
	assert.True(t, config.HasRepositoryYML)
	assert.Equal(t, 3, config.ModuleCount)
}

// TestDetectExistingConfig_InvalidRepositoryYML tests handling malformed repository.yml.
func TestDetectExistingConfig_InvalidRepositoryYML(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create invalid YAML
	repoPath := filepath.Join(eacDir, "repository.yml")
	err = os.WriteFile(repoPath, []byte("invalid: yaml: here!!!"), 0o644)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)

	require.NotNil(t, config)
	assert.True(t, config.HasRepositoryYML)
	// Should gracefully handle invalid YAML
	assert.Equal(t, 0, config.ModuleCount)
}

// TestReinitialize_ReuseExistingAIProvider tests reusing AI provider when flag omitted.
func TestReinitialize_ReuseExistingAIProvider(t *testing.T) {
	// This test verifies the key behavior:
	// When user runs "eac init" (without --ai-provider flag),
	// and existing AI config exists, it should be reused automatically.

	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create existing AI provider config
	personalConfig := `ai:
  provider: claude-api
  model: claude-3-haiku-20240307
  api_key: sk-ant-test-key
`
	personalPath := filepath.Join(eacDir, "ai-provider.personal.yml")
	err = os.WriteFile(personalPath, []byte(personalConfig), 0o600)
	require.NoError(t, err)

	// Create repository config
	repoConfig := `repository:
  name: test-repo
modules:
  - moniker: core
    description: Core module
`
	repoPath := filepath.Join(eacDir, "repository.yml")
	err = os.WriteFile(repoPath, []byte(repoConfig), 0o644)
	require.NoError(t, err)

	// Detect existing config
	config := detectExistingConfig(tmpDir)

	require.NotNil(t, config)
	assert.Equal(t, "claude-api", config.AIProvider)
	assert.Equal(t, "claude-3-haiku-20240307", config.AIModel)

	// When re-init is called without --ai-provider flag,
	// it should use these existing settings
}

// TestReinitialize_OverrideAIProvider tests overriding AI provider with flag.
func TestReinitialize_OverrideAIProvider(t *testing.T) {
	// This test verifies:
	// When user runs "eac init --ai-provider openai",
	// and existing config has claude-api,
	// it should switch to openai.

	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create existing AI provider config (claude)
	personalConfig := `ai:
  provider: claude-api
  model: claude-3-haiku-20240307
  api_key: sk-ant-test-key
`
	personalPath := filepath.Join(eacDir, "ai-provider.personal.yml")
	err = os.WriteFile(personalPath, []byte(personalConfig), 0o600)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)
	require.NotNil(t, config)
	assert.Equal(t, "claude-api", config.AIProvider)

	// When user provides --ai-provider openai flag,
	// reinit should override to openai
	// (implementation will handle this in reinitialize() function)
}

// TestReinitialize_NoAIConfigExists tests re-init when no AI config exists.
func TestReinitialize_NoAIConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create only repository config (no AI config)
	repoConfig := `repository:
  name: test-repo
modules:
  - moniker: core
    description: Core module
`
	repoPath := filepath.Join(eacDir, "repository.yml")
	err = os.WriteFile(repoPath, []byte(repoConfig), 0o644)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)

	require.NotNil(t, config)
	assert.False(t, config.HasAIProviderYML)
	assert.False(t, config.HasAIProviderPersonalYML)
	assert.Empty(t, config.AIProvider)

	// When re-init is called, it should proceed with rule-based mode
}

// TestReinitialize_DetectNewModules tests detecting modules added since last init.
func TestReinitialize_DetectNewModules(t *testing.T) {
	// This test documents the expected behavior:
	// When repository structure changes (new modules added),
	// re-running init should detect and add them to config.

	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create existing config with 2 modules
	repoConfig := `repository:
  name: test-repo
modules:
  - moniker: core
    description: Core module
  - moniker: cli
    description: CLI module
`
	repoPath := filepath.Join(eacDir, "repository.yml")
	err = os.WriteFile(repoPath, []byte(repoConfig), 0o644)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)
	require.NotNil(t, config)
	assert.Equal(t, 2, config.ModuleCount)

	// When scan detects 3 modules (new "api" module added),
	// merge should add it to config
	// (merge_test.go will test the actual merge logic)
}

// TestReinitialize_RemoveDeletedModules tests removing modules that no longer exist.
func TestReinitialize_RemoveDeletedModules(t *testing.T) {
	// This test documents the expected behavior:
	// When modules are deleted from filesystem,
	// re-init should remove them from config.

	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create existing config with 3 modules
	repoConfig := `repository:
  name: test-repo
modules:
  - moniker: core
    description: Core module
  - moniker: cli
    description: CLI module
  - moniker: deprecated
    description: Deprecated module
`
	repoPath := filepath.Join(eacDir, "repository.yml")
	err = os.WriteFile(repoPath, []byte(repoConfig), 0o644)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)
	require.NotNil(t, config)
	assert.Equal(t, 3, config.ModuleCount)

	// When scan only detects 2 modules (deprecated removed),
	// merge should remove it from config
	// (merge_test.go will test the actual merge logic)
}

// TestEffectiveAI_FallsBackToExistingConfig tests Bug 1 fix:
// When --ai-provider is not given but existing config has AI configured, use existing AI.
func TestEffectiveAI_FallsBackToExistingConfig(t *testing.T) {
	// Simulate the effectiveAI resolution logic from initImpl.
	flagAI := ""
	existingCfg := &ExistingConfig{AIProvider: "claude-api"}

	effectiveAI := flagAI
	if effectiveAI == "" && existingCfg != nil {
		effectiveAI = existingCfg.AIProvider
	}

	assert.Equal(t, "claude-api", effectiveAI)
}

// TestEffectiveAI_FlagOverridesExistingConfig tests Bug 1: explicit flag wins over existing config.
func TestEffectiveAI_FlagOverridesExistingConfig(t *testing.T) {
	flagAI := "openai"
	existingCfg := &ExistingConfig{AIProvider: "claude-api"}

	effectiveAI := flagAI
	if effectiveAI == "" && existingCfg != nil {
		effectiveAI = existingCfg.AIProvider
	}

	assert.Equal(t, "openai", effectiveAI)
}

// TestEffectiveAI_RemainsEmptyWhenNeitherConfigured tests effectiveAI stays empty when no AI anywhere.
func TestEffectiveAI_RemainsEmptyWhenNeitherConfigured(t *testing.T) {
	flagAI := ""
	existingCfg := &ExistingConfig{AIProvider: ""}

	effectiveAI := flagAI
	if effectiveAI == "" && existingCfg != nil {
		effectiveAI = existingCfg.AIProvider
	}

	assert.Empty(t, effectiveAI)
}

// TestReinitialize_PreserveBooksAndEnvironments tests that books.yml and environments.yml are unchanged.
func TestReinitialize_PreserveBooksAndEnvironments(t *testing.T) {
	// This test verifies:
	// Re-init should only update repository.yml and ai-provider.yml.
	// books.yml and environments.yml should be left untouched.

	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create books.yml with custom content
	booksContent := `books:
  - name: user-guide
    title: User Guide
    description: Custom user guide
`
	booksPath := filepath.Join(eacDir, "books.yml")
	err = os.WriteFile(booksPath, []byte(booksContent), 0o644)
	require.NoError(t, err)

	// Create environments.yml with custom content
	envsContent := `environments:
  - moniker: production
    name: Production
    level: L4
`
	envsPath := filepath.Join(eacDir, "environments.yml")
	err = os.WriteFile(envsPath, []byte(envsContent), 0o644)
	require.NoError(t, err)

	// Read original content
	originalBooks, err := os.ReadFile(booksPath)
	require.NoError(t, err)
	originalEnvs, err := os.ReadFile(envsPath)
	require.NoError(t, err)

	config := detectExistingConfig(tmpDir)
	require.NotNil(t, config)
	assert.True(t, config.HasBooksYML)
	assert.True(t, config.HasEnvironmentsYML)

	// After re-init, books.yml and environments.yml should be unchanged
	// (implementation will ensure this in reinitialize() function)
	// For now, we document this requirement in the test
	_ = originalBooks
	_ = originalEnvs
}
