package cache

import "testing"

func TestDefaultSkipSpecs(t *testing.T) {
	// Verify the default specs are what we expect
	if len(DefaultSkipSpecs) != 2 {
		t.Errorf("expected 2 default specs, got %d", len(DefaultSkipSpecs))
	}

	// First spec should be local:state
	if DefaultSkipSpecs[0].Level != LevelLocal || DefaultSkipSpecs[0].Type != TypeState {
		t.Errorf("expected first spec to be local:state, got %s", DefaultSkipSpecs[0].String())
	}

	// Second spec should be local:work
	if DefaultSkipSpecs[1].Level != LevelLocal || DefaultSkipSpecs[1].Type != TypeWork {
		t.Errorf("expected second spec to be local:work, got %s", DefaultSkipSpecs[1].String())
	}
}

func TestDefaultSkipSpecsMatching(t *testing.T) {
	// Create a config with default specs
	cfg := NewConfig()
	for _, spec := range DefaultSkipSpecs {
		cfg.Skip(spec)
	}

	// Should skip state
	if !cfg.ShouldSkipState() {
		t.Error("expected ShouldSkipState() to return true with default specs")
	}

	// Should skip work
	if !cfg.ShouldSkipWork() {
		t.Error("expected ShouldSkipWork() to return true with default specs")
	}

	// Should NOT skip assets
	if cfg.ShouldSkipAsset() {
		t.Error("expected ShouldSkipAsset() to return false with default specs")
	}

	// Should NOT skip registry
	if cfg.ShouldSkipLocalRegistry() {
		t.Error("expected ShouldSkipLocalRegistry() to return false with default specs")
	}

	// Should NOT skip layer
	if cfg.ShouldSkipLocalLayer() {
		t.Error("expected ShouldSkipLocalLayer() to return false with default specs")
	}
}
