package testrunners

import (
	"io"
	"sort"
	"testing"
)

func TestRegisterDescriptor(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	desc := &TestTypeDescriptor{
		TestType:      "gotest",
		IsBDD:         false,
		ComponentType: "go",
		MonikerStyle:  "file",
	}
	RegisterDescriptor(desc)

	got := GetDescriptor("gotest")
	if got == nil {
		t.Fatal("expected descriptor for gotest, got nil")
	}
	if got.ComponentType != "go" {
		t.Errorf("ComponentType = %q, want %q", got.ComponentType, "go")
	}
}

func TestRegisterDescriptorNil(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	// Should not panic
	RegisterDescriptor(nil)

	if got := GetDescriptor("anything"); got != nil {
		t.Errorf("expected nil for unregistered type, got %v", got)
	}
}

func TestGetComponentType(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	RegisterDescriptor(&TestTypeDescriptor{
		TestType:      "godog",
		ComponentType: "gherkin",
	})
	RegisterDescriptor(&TestTypeDescriptor{
		TestType:      "mocha",
		ComponentType: "typescript",
	})

	tests := []struct {
		testType string
		want     string
	}{
		{"godog", "gherkin"},
		{"mocha", "typescript"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		got := GetComponentType(tt.testType)
		if got != tt.want {
			t.Errorf("GetComponentType(%q) = %q, want %q", tt.testType, got, tt.want)
		}
	}
}

func TestGetRunnerFileConventions(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	RegisterDescriptor(&TestTypeDescriptor{
		TestType:             "godog",
		RunnerFileConvention: "godog_test.go",
	})
	RegisterDescriptor(&TestTypeDescriptor{
		TestType:             "mocha",
		RunnerFileConvention: "", // no convention
	})

	got := GetRunnerFileConventions()
	if len(got) != 1 {
		t.Fatalf("expected 1 convention, got %d", len(got))
	}
	if !got["godog_test.go"] {
		t.Error("expected godog_test.go in conventions")
	}
}

func TestResolveFeatureTestType(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	RegisterDescriptor(&TestTypeDescriptor{
		TestType: "godog",
		IsBDD:    true,
		FeatureTestTypeResolver: func(info FeatureModuleInfo) bool {
			return !info.HasTypeScript
		},
	})
	RegisterDescriptor(&TestTypeDescriptor{
		TestType: "tscucumber",
		IsBDD:    true,
		FeatureTestTypeResolver: func(info FeatureModuleInfo) bool {
			return info.HasTypeScript
		},
	})

	tests := []struct {
		name string
		info FeatureModuleInfo
		want string
	}{
		{"go module", FeatureModuleInfo{HasGo: true}, "godog"},
		{"ts module", FeatureModuleInfo{HasTypeScript: true}, "tscucumber"},
		{"neither", FeatureModuleInfo{}, "godog"},
	}

	for _, tt := range tests {
		got := ResolveFeatureTestType(tt.info)
		if got != tt.want {
			t.Errorf("ResolveFeatureTestType(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestResolveFeatureTestTypeFallback(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	// No descriptors registered - returns empty (no hardcoded fallback)
	got := ResolveFeatureTestType(FeatureModuleInfo{HasGo: true})
	if got != "" {
		t.Errorf("ResolveFeatureTestType with no descriptors = %q, want %q", got, "")
	}
}

func TestGetMonikerStyle(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	RegisterDescriptor(&TestTypeDescriptor{
		TestType:     "godog",
		MonikerStyle: "feature",
	})
	RegisterDescriptor(&TestTypeDescriptor{
		TestType:     "gotest",
		MonikerStyle: "file",
	})

	tests := []struct {
		testType string
		want     string
	}{
		{"godog", "feature"},
		{"gotest", "file"},
		{"unknown", "file"}, // default
	}

	for _, tt := range tests {
		got := GetMonikerStyle(tt.testType)
		if got != tt.want {
			t.Errorf("GetMonikerStyle(%q) = %q, want %q", tt.testType, got, tt.want)
		}
	}
}

func TestCollectInferences(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	RegisterDescriptor(&TestTypeDescriptor{
		TestType: "gotest",
		DefaultInferences: []Inference{
			{TestTypes: []string{"gotest"}, ThenAddTags: []string{"@deps:go"}, Description: "Go tests need Go"},
			{TestTypes: []string{"gotest"}, ThenAddTags: []string{"@L1"}, Description: "Go tests default L1"},
		},
	})
	RegisterDescriptor(&TestTypeDescriptor{
		TestType: "mocha",
		// No inferences
	})

	got := CollectInferences()
	if len(got) != 2 {
		t.Fatalf("expected 2 inferences, got %d", len(got))
	}
	if got[0].Description != "Go tests need Go" {
		t.Errorf("inference[0].Description = %q, want %q", got[0].Description, "Go tests need Go")
	}
}

func TestCollectInferencesEmpty(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	got := CollectInferences()
	if len(got) != 0 {
		t.Errorf("expected 0 inferences from empty registry, got %d", len(got))
	}
}

func TestAllDescriptors(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	RegisterDescriptor(&TestTypeDescriptor{TestType: "gotest"})
	RegisterDescriptor(&TestTypeDescriptor{TestType: "godog"})

	got := AllDescriptors()
	if len(got) != 2 {
		t.Fatalf("expected 2 descriptors, got %d", len(got))
	}
}

func TestBDDComponentNames(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	RegisterDescriptor(&TestTypeDescriptor{TestType: "godog", IsBDD: true})
	RegisterDescriptor(&TestTypeDescriptor{TestType: "tscucumber", IsBDD: true})
	RegisterDescriptor(&TestTypeDescriptor{TestType: "gotest", IsBDD: false})

	got := BDDComponentNames()
	sort.Strings(got)
	if len(got) != 2 {
		t.Fatalf("expected 2 BDD names, got %d: %v", len(got), got)
	}
	if got[0] != "godog" || got[1] != "tscucumber" {
		t.Errorf("BDDComponentNames() = %v, want [godog tscucumber]", got)
	}
}

func TestBDDComponentNamesFallback(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	// No BDD types registered - returns empty (no hardcoded fallback)
	RegisterDescriptor(&TestTypeDescriptor{TestType: "gotest", IsBDD: false})

	got := BDDComponentNames()
	if len(got) != 0 {
		t.Errorf("BDDComponentNames with no BDD types = %v, want []", got)
	}
}

func TestBDDComponentNamesEmpty(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	// No descriptors at all - returns empty (no hardcoded fallback)
	got := BDDComponentNames()
	if len(got) != 0 {
		t.Errorf("BDDComponentNames with no descriptors = %v, want []", got)
	}
}

// --- Mock runner for isolated registry tests ---

// mockRunner is a minimal TestTypeRunner for testing.
type mockRunner struct {
	types []string
	bdd   bool
}

func (m *mockRunner) TestTypes() []string                                     { return m.types }
func (m *mockRunner) IsBDD() bool                                             { return m.bdd }
func (m *mockRunner) GetTestInfo(_ TestReference, _ string, _ any) *TestInfo  { return nil }
func (m *mockRunner) FindTestRoot(_ string, _ any) string                     { return "" }
func (m *mockRunner) BuildPackagePath(_, _ string) string                     { return "" }
func (m *mockRunner) Execute(_ string, _ []TestReference, _ io.Writer, _ RunConfig) RunResult {
	return RunResult{}
}

// --- Isolated Registry tests ---

func TestNewRegistryIsolation(t *testing.T) {
	// Verify that two NewRegistry() instances are fully isolated from
	// each other and from the default (package-level) registry.

	ResetForTesting()
	defer ResetForTesting()

	regA := NewRegistry()
	regB := NewRegistry()

	runnerA := &mockRunner{types: []string{"typeA"}}
	runnerB := &mockRunner{types: []string{"typeB"}}

	regA.Register(runnerA)
	regB.Register(runnerB)

	t.Run("registry A has only typeA", func(t *testing.T) {
		if got := regA.Get("typeA"); got == nil {
			t.Fatal("expected runner for typeA in regA, got nil")
		}
		if got := regA.Get("typeB"); got != nil {
			t.Error("regA should not have typeB")
		}
	})

	t.Run("registry B has only typeB", func(t *testing.T) {
		if got := regB.Get("typeB"); got == nil {
			t.Fatal("expected runner for typeB in regB, got nil")
		}
		if got := regB.Get("typeA"); got != nil {
			t.Error("regB should not have typeA")
		}
	})

	t.Run("default registry is unaffected", func(t *testing.T) {
		if got := Get("typeA"); got != nil {
			t.Error("default registry should not have typeA")
		}
		if got := Get("typeB"); got != nil {
			t.Error("default registry should not have typeB")
		}
	})
}

func TestNewRegistryDescriptors(t *testing.T) {
	reg := NewRegistry()

	reg.RegisterDescriptor(&TestTypeDescriptor{
		TestType:      "custom",
		ComponentType: "custom-lang",
		MonikerStyle:  "feature",
		IsBDD:         true,
	})

	t.Run("GetDescriptor", func(t *testing.T) {
		got := reg.GetDescriptor("custom")
		if got == nil {
			t.Fatal("expected descriptor for custom, got nil")
		}
		if got.ComponentType != "custom-lang" {
			t.Errorf("ComponentType = %q, want %q", got.ComponentType, "custom-lang")
		}
	})

	t.Run("GetComponentType", func(t *testing.T) {
		if got := reg.GetComponentType("custom"); got != "custom-lang" {
			t.Errorf("GetComponentType = %q, want %q", got, "custom-lang")
		}
		if got := reg.GetComponentType("missing"); got != "" {
			t.Errorf("GetComponentType for missing = %q, want empty", got)
		}
	})

	t.Run("GetMonikerStyle", func(t *testing.T) {
		if got := reg.GetMonikerStyle("custom"); got != "feature" {
			t.Errorf("GetMonikerStyle = %q, want %q", got, "feature")
		}
		if got := reg.GetMonikerStyle("missing"); got != "file" {
			t.Errorf("GetMonikerStyle for missing = %q, want %q", got, "file")
		}
	})

	t.Run("AllDescriptors", func(t *testing.T) {
		all := reg.AllDescriptors()
		if len(all) != 1 {
			t.Fatalf("expected 1 descriptor, got %d", len(all))
		}
	})

	t.Run("BDDComponentNames", func(t *testing.T) {
		names := reg.BDDComponentNames()
		if len(names) != 1 || names[0] != "custom" {
			t.Errorf("BDDComponentNames = %v, want [custom]", names)
		}
	})

	t.Run("SupportedTypes", func(t *testing.T) {
		// No runners registered, only descriptors
		types := reg.SupportedTypes()
		if len(types) != 0 {
			t.Errorf("SupportedTypes = %v, want empty", types)
		}
	})
}

func TestNewRegistryFallback(t *testing.T) {
	reg := NewRegistry()

	fb := &mockRunner{types: []string{"fallback-type"}}
	reg.RegisterFallback(fb)

	t.Run("fallback returned for unknown type", func(t *testing.T) {
		got := reg.Get("unknown")
		if got != fb {
			t.Error("expected fallback runner for unknown type")
		}
	})

	t.Run("registered runner takes precedence", func(t *testing.T) {
		specific := &mockRunner{types: []string{"specific"}}
		reg.Register(specific)

		got := reg.Get("specific")
		if got != specific {
			t.Error("expected specific runner, not fallback")
		}
	})
}

func TestNewRegistryGetAll(t *testing.T) {
	reg := NewRegistry()

	r1 := &mockRunner{types: []string{"a", "b"}}
	r2 := &mockRunner{types: []string{"c"}}
	reg.Register(r1)
	reg.Register(r2)

	all := reg.GetAll()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries in GetAll, got %d", len(all))
	}
	if all["a"] != r1 || all["b"] != r1 {
		t.Error("runner r1 should be registered for both a and b")
	}
	if all["c"] != r2 {
		t.Error("runner r2 should be registered for c")
	}
}

func TestNewRegistryResetForTesting(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&mockRunner{types: []string{"x"}})
	reg.RegisterDescriptor(&TestTypeDescriptor{TestType: "x"})
	reg.RegisterFallback(&mockRunner{types: []string{"fb"}})

	// Verify populated
	if len(reg.GetAll()) == 0 {
		t.Fatal("expected non-empty registry before reset")
	}

	reg.ResetForTesting()

	if len(reg.GetAll()) != 0 {
		t.Error("expected empty runners after reset")
	}
	if len(reg.AllDescriptors()) != 0 {
		t.Error("expected empty descriptors after reset")
	}
	if reg.Get("x") != nil {
		t.Error("expected nil for Get after reset (fallback should be cleared)")
	}
}
