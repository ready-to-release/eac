package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToolIDConstants(t *testing.T) {
	// Test that constant values are correct
	t.Run("scanner tool IDs", func(t *testing.T) {
		tests := map[string]string{
			"ToolTrivySBOM":       ToolTrivySBOM,
			"ToolTrivyVuln":       ToolTrivyVuln,
			"ToolTrivySecrets":    ToolTrivySecrets,
			"ToolTrivyCompliance": ToolTrivyCompliance,
			"ToolTrivyIaC":        ToolTrivyIaC,
			"ToolSemgrep":         ToolSemgrep,
			"ToolZap":             ToolZap,
		}

		expected := map[string]string{
			"ToolTrivySBOM":       "trivy-sbom",
			"ToolTrivyVuln":       "trivy-vuln",
			"ToolTrivySecrets":    "trivy-secrets",
			"ToolTrivyCompliance": "trivy-compliance",
			"ToolTrivyIaC":        "trivy-iac",
			"ToolSemgrep":         "semgrep",
			"ToolZap":             "zap",
		}

		for name, got := range tests {
			want := expected[name]
			if got != want {
				t.Errorf("%s = %q, want %q", name, got, want)
			}
		}
	})

	t.Run("server tool IDs", func(t *testing.T) {
		if ToolStaticSite != "static-site" {
			t.Errorf("ToolStaticSite = %q, want %q", ToolStaticSite, "static-site")
		}
		if ToolMkDocsLive != "mkdocs-live" {
			t.Errorf("ToolMkDocsLive = %q, want %q", ToolMkDocsLive, "mkdocs-live")
		}
		if ToolStructurizrLite != "structurizr-lite" {
			t.Errorf("ToolStructurizrLite = %q, want %q", ToolStructurizrLite, "structurizr-lite")
		}
	})

	t.Run("linter tool IDs", func(t *testing.T) {
		if ToolGolangciLint != "golangci-lint" {
			t.Errorf("ToolGolangciLint = %q, want %q", ToolGolangciLint, "golangci-lint")
		}
		if ToolMarkdownlint != "markdownlint" {
			t.Errorf("ToolMarkdownlint = %q, want %q", ToolMarkdownlint, "markdownlint")
		}
	})

	t.Run("builder tool IDs", func(t *testing.T) {
		if ToolGo != "go" {
			t.Errorf("ToolGo = %q, want %q", ToolGo, "go")
		}
		if ToolNpmBuild != "npm-build" {
			t.Errorf("ToolNpmBuild = %q, want %q", ToolNpmBuild, "npm-build")
		}
		if ToolBuildx != "buildx" {
			t.Errorf("ToolBuildx = %q, want %q", ToolBuildx, "buildx")
		}
	})
}

func TestAllScannerToolIDs(t *testing.T) {
	ids := AllScannerToolIDs()

	if len(ids) != 7 {
		t.Errorf("AllScannerToolIDs() returned %d items, want 7", len(ids))
	}

	// Check that all expected scanners are present
	expected := []string{
		"trivy-sbom", "trivy-vuln", "trivy-secrets",
		"trivy-compliance", "trivy-iac", "semgrep", "zap",
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	for _, exp := range expected {
		if !idSet[exp] {
			t.Errorf("AllScannerToolIDs() missing %q", exp)
		}
	}
}

func TestAllServerToolIDs(t *testing.T) {
	ids := AllServerToolIDs()

	if len(ids) != 3 {
		t.Errorf("AllServerToolIDs() returned %d items, want 3", len(ids))
	}

	expected := []string{"static-site", "mkdocs-live", "structurizr-lite"}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	for _, exp := range expected {
		if !idSet[exp] {
			t.Errorf("AllServerToolIDs() missing %q", exp)
		}
	}
}

func TestAllLinterToolIDs(t *testing.T) {
	ids := AllLinterToolIDs()

	if len(ids) != 5 {
		t.Errorf("AllLinterToolIDs() returned %d items, want 5", len(ids))
	}

	expected := []string{"golangci-lint", "markdownlint", "eslint", "ruff", "clippy"}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	for _, exp := range expected {
		if !idSet[exp] {
			t.Errorf("AllLinterToolIDs() missing %q", exp)
		}
	}
}

func TestAllBuilderToolIDs(t *testing.T) {
	ids := AllBuilderToolIDs()

	if len(ids) != 9 {
		t.Errorf("AllBuilderToolIDs() returned %d items, want 9", len(ids))
	}

	// Spot check some builders
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	for _, exp := range []string{"go", "npm-build", "mkdocs-build", "buildx", "cargo"} {
		if !idSet[exp] {
			t.Errorf("AllBuilderToolIDs() missing %q", exp)
		}
	}
}

func TestToolIDsMatchToolConfig(t *testing.T) {
	// This test validates that tool ID constants exist in tool-config.yml.
	// Skip if not in repo root (e.g., running in CI without contracts).

	// Try to find the contracts directory
	repoRoot := findRepoRoot()
	if repoRoot == "" {
		t.Skip("Could not find repo root, skipping tool-config validation")
	}

	configDir := filepath.Join(repoRoot, ".eac")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}

	config, err := LoadToolConfig(repoRoot, configDir)
	if err != nil {
		t.Fatalf("LoadToolConfig failed: %v", err)
	}

	// Validate scanner tool IDs exist in config
	scannerIDs := AllScannerToolIDs()
	for _, id := range scannerIDs {
		if !config.hasToolCanonical(id) {
			t.Errorf("Scanner tool %q defined in ids.go but not in tool-config.yml", id)
		}
	}

	// Validate server tool IDs exist in config
	serverIDs := AllServerToolIDs()
	for _, id := range serverIDs {
		if !config.hasToolCanonical(id) {
			t.Errorf("Server tool %q defined in ids.go but not in tool-config.yml", id)
		}
	}

	// Validate linter tool IDs exist in config
	linterIDs := AllLinterToolIDs()
	for _, id := range linterIDs {
		if !config.hasToolCanonical(id) {
			t.Errorf("Linter tool %q defined in ids.go but not in tool-config.yml", id)
		}
	}

	// Validate builder tool IDs exist in config
	builderIDs := AllBuilderToolIDs()
	for _, id := range builderIDs {
		if !config.hasToolCanonical(id) {
			t.Errorf("Builder tool %q defined in ids.go but not in tool-config.yml", id)
		}
	}
}

// findRepoRoot walks up the directory tree to find the repo root.
func findRepoRoot() string {
	// Start from the test directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		// Check for contracts directory (repo marker)
		contractsDir := filepath.Join(dir, "contracts", "core")
		if _, err := os.Stat(contractsDir); err == nil {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, not found
			return ""
		}
		dir = parent
	}
}
