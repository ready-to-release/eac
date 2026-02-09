package cache

import "testing"

func TestDefaultSkipSpecs(t *testing.T) {
	specs := DefaultSkipSpecs()

	// Verify the default specs are what we expect
	if len(specs) != 2 {
		t.Errorf("expected 2 default specs, got %d", len(specs))
	}

	// First spec should be local:state
	if specs[0].Level != LevelLocal || specs[0].Type != TypeState {
		t.Errorf("expected first spec to be local:state, got %s", specs[0].String())
	}

	// Second spec should be local:work
	if specs[1].Level != LevelLocal || specs[1].Type != TypeWork {
		t.Errorf("expected second spec to be local:work, got %s", specs[1].String())
	}
}

func TestDefaultSkipSpecsMatching(t *testing.T) {
	// Create a config with default specs
	cfg := NewConfig()
	for _, spec := range DefaultSkipSpecs() {
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

func TestDefaultSkipSpecs_CICacheNotSkippedByDefault(t *testing.T) {
	specs := DefaultSkipSpecs()

	// CI cache should NOT be in default skip specs
	for _, spec := range specs {
		if spec.Level == LevelRemote && spec.Type == TypeCI {
			t.Error("CI cache should NOT be in default skip specs")
		}
	}
}

func TestShouldSkipCI_FalseByDefault(t *testing.T) {
	// With default specs, ShouldSkipCI should return false
	cfg := NewConfig()
	for _, spec := range DefaultSkipSpecs() {
		cfg.Skip(spec)
	}

	if cfg.ShouldSkipCI() {
		t.Error("expected ShouldSkipCI() to return false with default specs")
	}
}

func TestShouldSkipCI_NilConfig(t *testing.T) {
	var cfg *Config
	if cfg.ShouldSkipCI() {
		t.Error("expected ShouldSkipCI() to return false with nil config")
	}
}

func TestDefaultSkipSpecs_ReturnsCopy(t *testing.T) {
	a := DefaultSkipSpecs()
	b := DefaultSkipSpecs()

	// Mutating one should not affect the other
	a[0].Type = TypeAll
	if b[0].Type == TypeAll {
		t.Error("DefaultSkipSpecs must return independent copies")
	}
}
