//go:build L0 && ov

package config

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSecurityConfig(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	// Use a temp dir as config root (no user overrides)
	configRoot := t.TempDir()

	cfg, err := LoadSecurityConfig(repoRoot, configRoot)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	t.Run("loads scanner definitions", func(t *testing.T) {
		scanners := cfg.ListScanners()
		assert.NotEmpty(t, scanners, "should have scanners loaded")

		// Check for expected scanners
		scanner, ok := cfg.GetScanner("trivy-sbom")
		assert.True(t, ok, "should have trivy-sbom scanner")
		assert.Equal(t, "ghcr.io/aquasecurity/trivy", scanner.Image())
		assert.NotEmpty(t, scanner.Tag())
		assert.NotEmpty(t, scanner.Command())
	})

	t.Run("loads policies", func(t *testing.T) {
		// Go modules should have scanners
		goScanners := cfg.GetDefaultScanners("go")
		assert.NotEmpty(t, goScanners, "go should have default scanners")
		assert.Contains(t, goScanners, "trivy-sbom")

		// Assets should have secrets scanner only
		assetsScanners := cfg.GetDefaultScanners("assets")
		assert.Contains(t, assetsScanners, "trivy-secret", "assets should have secrets scanner")
	})

	t.Run("skip modules", func(t *testing.T) {
		assert.True(t, cfg.ShouldSkipModule("repository"), "repository should be skipped")
		assert.False(t, cfg.ShouldSkipModule("eac"), "eac should not be skipped")
	})
}

func TestSecurityConfig_GetScanner(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadSecurityConfig(repoRoot, t.TempDir())
	require.NoError(t, err)

	tests := []struct {
		id       string
		wantOk   bool
		category string
	}{
		{"trivy-sbom", true, "sbom"},
		{"trivy-vuln", true, "vuln"},
		{"trivy-secrets", true, "secrets"},
		{"semgrep", true, "sast"},
		{"zap", true, "dast"},
		{"nonexistent", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			scanner, ok := cfg.GetScanner(tt.id)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.category, scanner.Category())
				assert.NotEmpty(t, scanner.FullImage())
			}
		})
	}
}

func TestSecurityConfig_GetDefaultScanners_Fallback(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadSecurityConfig(repoRoot, t.TempDir())
	require.NoError(t, err)

	// Unknown component type should fall back to default
	scanners := cfg.GetDefaultScanners("unknown-type")
	assert.NotEmpty(t, scanners, "unknown type should get default scanners")
}
