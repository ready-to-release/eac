package lint

import (
	"testing"
	"time"
)

func TestLintConfig_Defaults(t *testing.T) {
	cfg := &LintConfig{}

	if cfg.Fix {
		t.Error("Fix should default to false")
	}
	if cfg.Config != "" {
		t.Errorf("Config should default to empty, got %q", cfg.Config)
	}
	if cfg.ForceLint {
		t.Error("ForceLint should default to false")
	}
}

func TestLintConfig_AllFieldsSet(t *testing.T) {
	cfg := &LintConfig{
		Fix:       true,
		Config:    "/path/to/config.yml",
		ForceLint: true,
	}

	if !cfg.Fix {
		t.Error("Fix should be true")
	}
	if cfg.Config != "/path/to/config.yml" {
		t.Errorf("Config = %q, want %q", cfg.Config, "/path/to/config.yml")
	}
	if !cfg.ForceLint {
		t.Error("ForceLint should be true")
	}
}

func TestLintModuleResult_Defaults(t *testing.T) {
	result := &LintModuleResult{}

	if result.Moniker != "" {
		t.Errorf("Moniker should default to empty, got %q", result.Moniker)
	}
	if result.Success {
		t.Error("Success should default to false")
	}
	if result.IssueCount != 0 {
		t.Errorf("IssueCount should default to 0, got %d", result.IssueCount)
	}
	if result.FixedCount != 0 {
		t.Errorf("FixedCount should default to 0, got %d", result.FixedCount)
	}
	if result.Duration != 0 {
		t.Errorf("Duration should default to 0, got %v", result.Duration)
	}
	if len(result.Providers) != 0 {
		t.Errorf("Providers should default to empty, got %v", result.Providers)
	}
}

func TestLintModuleResult_FullyPopulated(t *testing.T) {
	result := &LintModuleResult{
		Moniker:      "mod-a:go",
		Success:      true,
		IssueCount:   5,
		FixedCount:   3,
		ErrorMessage: "some warning",
		Duration:     2 * time.Second,
		Providers:    []string{"golangci-lint", "staticcheck"},
	}

	if result.Moniker != "mod-a:go" {
		t.Errorf("Moniker = %q, want %q", result.Moniker, "mod-a:go")
	}
	if !result.Success {
		t.Error("Success should be true")
	}
	if result.IssueCount != 5 {
		t.Errorf("IssueCount = %d, want 5", result.IssueCount)
	}
	if result.FixedCount != 3 {
		t.Errorf("FixedCount = %d, want 3", result.FixedCount)
	}
	if result.ErrorMessage != "some warning" {
		t.Errorf("ErrorMessage = %q, want %q", result.ErrorMessage, "some warning")
	}
	if result.Duration != 2*time.Second {
		t.Errorf("Duration = %v, want %v", result.Duration, 2*time.Second)
	}
	if len(result.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(result.Providers))
	}
}

func TestLintContext_Initialization(t *testing.T) {
	lctx := &lintContext{
		cfg:         &LintConfig{Fix: true},
		results:     make(map[string]*LintModuleResult),
		moduleFiles: make(map[string][]string),
	}

	if lctx.cfg == nil {
		t.Fatal("cfg should not be nil")
	}
	if !lctx.cfg.Fix {
		t.Error("cfg.Fix should be true")
	}
	if lctx.results == nil {
		t.Error("results map should be initialized")
	}
	if lctx.moduleFiles == nil {
		t.Error("moduleFiles map should be initialized")
	}
	if lctx.cachedModules != nil {
		t.Error("cachedModules should be nil when not initialized")
	}
	if lctx.cachedUoWs != nil {
		t.Error("cachedUoWs should be nil when not initialized")
	}
}

func TestLintContext_CachedStateTracking(t *testing.T) {
	now := time.Now()
	lctx := &lintContext{
		cfg:           &LintConfig{},
		results:       make(map[string]*LintModuleResult),
		cachedModules: map[string]bool{"mod-a": true, "mod-b": false},
		cacheTimes:    map[string]time.Time{"mod-a": now},
		cachedUoWs:    map[string]bool{"lint:mod-a:go:go:golangci-lint": true},
		uowCacheTimes: map[string]time.Time{"lint:mod-a:go:go:golangci-lint": now},
	}

	if !lctx.cachedModules["mod-a"] {
		t.Error("mod-a should be cached")
	}
	if lctx.cachedModules["mod-b"] {
		t.Error("mod-b should not be cached")
	}
	if _, ok := lctx.cachedModules["mod-c"]; ok {
		t.Error("mod-c should not exist in cachedModules")
	}

	if !lctx.cacheTimes["mod-a"].Equal(now) {
		t.Errorf("cache time for mod-a = %v, want %v", lctx.cacheTimes["mod-a"], now)
	}

	if !lctx.cachedUoWs["lint:mod-a:go:go:golangci-lint"] {
		t.Error("UoW should be cached")
	}
}
