package testrunners

import (
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

	// No descriptors registered - should fall back to "godog"
	got := ResolveFeatureTestType(FeatureModuleInfo{HasGo: true})
	if got != "godog" {
		t.Errorf("ResolveFeatureTestType with no descriptors = %q, want %q", got, "godog")
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

	// No BDD types registered
	RegisterDescriptor(&TestTypeDescriptor{TestType: "gotest", IsBDD: false})

	got := BDDComponentNames()
	if len(got) != 1 || got[0] != "godog" {
		t.Errorf("BDDComponentNames with no BDD types = %v, want [godog]", got)
	}
}

func TestBDDComponentNamesEmpty(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	// No descriptors at all
	got := BDDComponentNames()
	if len(got) != 1 || got[0] != "godog" {
		t.Errorf("BDDComponentNames with no descriptors = %v, want [godog]", got)
	}
}
