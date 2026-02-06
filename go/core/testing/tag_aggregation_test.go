package testing

import (
	"reflect"
	"sort"
	"testing"
)

func TestAggregateTagsForTests(t *testing.T) {
	t.Run("nil input returns empty TagSummary", func(t *testing.T) {
		result := AggregateTagsForTests(nil)
		if !result.IsEmpty() {
			t.Errorf("expected empty TagSummary for nil input, got %+v", result)
		}
	})

	t.Run("empty slice returns empty TagSummary", func(t *testing.T) {
		result := AggregateTagsForTests([]TestReference{})
		if !result.IsEmpty() {
			t.Errorf("expected empty TagSummary for empty slice input, got %+v", result)
		}
	})

	t.Run("single test with level tags", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L0", "@L1"}},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "Levels", result.Levels, []string{"@L0", "@L1"})
		assertSlice(t, "All", result.All, []string{"@L0", "@L1"})
	})

	t.Run("single test with verification tags", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@ov", "@iv"}},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "Verification", result.Verification, []string{"@iv", "@ov"})
		assertSlice(t, "All", result.All, []string{"@iv", "@ov"})
	})

	t.Run("all verification tag types recognized", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@ov", "@iv", "@pv", "@piv", "@ppv"}},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "Verification", result.Verification, []string{"@iv", "@ov", "@piv", "@ppv", "@pv"})
	})

	t.Run("all level tags recognized", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L0", "@L1", "@L2", "@L3", "@L4"}},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "Levels", result.Levels, []string{"@L0", "@L1", "@L2", "@L3", "@L4"})
	})

	t.Run("risk control tags stored without prefix", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@risk-control:RC-001", "@risk-control:RC-002"}},
		}
		result := AggregateTagsForTests(tests)

		// Risk control tags are stored without the @risk-control: prefix (like deps)
		assertSlice(t, "RiskControls", result.RiskControls, []string{"RC-001", "RC-002"})
		// Full tag still present in All
		assertSlice(t, "All", result.All, []string{"@risk-control:RC-001", "@risk-control:RC-002"})
	})

	t.Run("system dependency tags stored without prefix", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@deps:docker", "@deps:git"}},
		}
		result := AggregateTagsForTests(tests)

		// System deps stored WITHOUT the @deps: prefix
		assertSlice(t, "SystemDeps", result.SystemDeps, []string{"docker", "git"})
		// Raw tags in All
		assertSlice(t, "All", result.All, []string{"@deps:docker", "@deps:git"})
	})

	t.Run("module dependency tags stored without prefix", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@depm:core", "@depm:tui"}},
		}
		result := AggregateTagsForTests(tests)

		// Module deps stored WITHOUT the @depm: prefix
		assertSlice(t, "ModuleDeps", result.ModuleDeps, []string{"core", "tui"})
		// Raw tags in All
		assertSlice(t, "All", result.All, []string{"@depm:core", "@depm:tui"})
	})

	t.Run("gxp tag sets IsGxP", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@gxp"}},
		}
		result := AggregateTagsForTests(tests)

		if !result.IsGxP {
			t.Error("expected IsGxP to be true for @gxp tag")
		}
	})

	t.Run("Manual tag sets HasManual", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@Manual"}},
		}
		result := AggregateTagsForTests(tests)

		if !result.HasManual {
			t.Error("expected HasManual to be true for @Manual tag")
		}
	})

	t.Run("single test with mixed tags classifies all categories", func(t *testing.T) {
		tests := []TestReference{
			{
				Tags: []string{
					"@L0", "@L2",
					"@ov", "@pv",
					"@risk-control:RC-100",
					"@deps:docker",
					"@depm:core",
					"@gxp",
					"@Manual",
				},
			},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "Levels", result.Levels, []string{"@L0", "@L2"})
		assertSlice(t, "Verification", result.Verification, []string{"@ov", "@pv"})
		assertSlice(t, "RiskControls", result.RiskControls, []string{"RC-100"})
		assertSlice(t, "SystemDeps", result.SystemDeps, []string{"docker"})
		assertSlice(t, "ModuleDeps", result.ModuleDeps, []string{"core"})
		if !result.IsGxP {
			t.Error("expected IsGxP true")
		}
		if !result.HasManual {
			t.Error("expected HasManual true")
		}

		// All should contain every raw tag
		if len(result.All) != 9 {
			t.Errorf("expected 9 items in All, got %d: %v", len(result.All), result.All)
		}
	})

	t.Run("multiple tests union tags with deduplication", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L0", "@ov", "@risk-control:RC-001"}},
			{Tags: []string{"@L0", "@L1", "@ov", "@iv", "@risk-control:RC-001", "@risk-control:RC-002"}},
			{Tags: []string{"@L2", "@pv"}},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "Levels", result.Levels, []string{"@L0", "@L1", "@L2"})
		assertSlice(t, "Verification", result.Verification, []string{"@iv", "@ov", "@pv"})
		assertSlice(t, "RiskControls", result.RiskControls, []string{"RC-001", "RC-002"})
	})

	t.Run("all slices are sorted", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L3", "@L1", "@ppv"}},
			{Tags: []string{"@L0", "@L2", "@ov", "@iv"}},
		}
		result := AggregateTagsForTests(tests)

		if !sort.StringsAreSorted(result.Levels) {
			t.Errorf("Levels should be sorted, got %v", result.Levels)
		}
		if !sort.StringsAreSorted(result.Verification) {
			t.Errorf("Verification should be sorted, got %v", result.Verification)
		}
		if !sort.StringsAreSorted(result.All) {
			t.Errorf("All should be sorted, got %v", result.All)
		}
	})

	t.Run("boolean flags ORd: one gxp one not", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L0"}, IsGxP: false},
			{Tags: []string{"@L1", "@gxp"}, IsGxP: true},
		}
		result := AggregateTagsForTests(tests)

		if !result.IsGxP {
			t.Error("expected IsGxP true when at least one test has gxp")
		}
	})

	t.Run("boolean flags ORd: one manual one not", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L0"}, IsManual: false},
			{Tags: []string{"@Manual"}, IsManual: true},
		}
		result := AggregateTagsForTests(tests)

		if !result.HasManual {
			t.Error("expected HasManual true when at least one test is manual")
		}
	})

	t.Run("boolean flags ORd: both false stays false", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L0"}, IsGxP: false, IsManual: false},
			{Tags: []string{"@L1"}, IsGxP: false, IsManual: false},
		}
		result := AggregateTagsForTests(tests)

		if result.IsGxP {
			t.Error("expected IsGxP false when no test has gxp")
		}
		if result.HasManual {
			t.Error("expected HasManual false when no test is manual")
		}
	})

	t.Run("structured SystemDependencies included in aggregation", func(t *testing.T) {
		tests := []TestReference{
			{
				Tags:               []string{"@L0"},
				SystemDependencies: []string{"docker", "git"},
			},
			{
				Tags:               []string{"@L1"},
				SystemDependencies: []string{"git", "npm"},
			},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "SystemDeps", result.SystemDeps, []string{"docker", "git", "npm"})
	})

	t.Run("structured ModuleDependencies included in aggregation", func(t *testing.T) {
		tests := []TestReference{
			{
				Tags:               []string{"@L0"},
				ModuleDependencies: []string{"core", "cli"},
			},
			{
				Tags:               []string{"@L1"},
				ModuleDependencies: []string{"cli", "tui"},
			},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "ModuleDeps", result.ModuleDeps, []string{"cli", "core", "tui"})
	})

	t.Run("structured RiskControls included in aggregation", func(t *testing.T) {
		tests := []TestReference{
			{
				Tags:         []string{"@L0"},
				RiskControls: []string{"RC-001", "RC-003"},
			},
			{
				Tags:         []string{"@L1"},
				RiskControls: []string{"RC-002", "RC-003"},
			},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "RiskControls", result.RiskControls, []string{"RC-001", "RC-002", "RC-003"})
	})

	t.Run("structured fields merged with tag-derived fields", func(t *testing.T) {
		tests := []TestReference{
			{
				Tags:               []string{"@deps:docker", "@depm:core", "@risk-control:RC-001"},
				SystemDependencies: []string{"git"},
				ModuleDependencies: []string{"tui"},
				RiskControls:       []string{"RC-002"},
			},
		}
		result := AggregateTagsForTests(tests)

		assertSlice(t, "SystemDeps", result.SystemDeps, []string{"docker", "git"})
		assertSlice(t, "ModuleDeps", result.ModuleDeps, []string{"core", "tui"})
		assertSlice(t, "RiskControls", result.RiskControls, []string{"RC-001", "RC-002"})
	})

	t.Run("structured IsGxP ORd from field", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L0"}, IsGxP: true},
			{Tags: []string{"@L1"}, IsGxP: false},
		}
		result := AggregateTagsForTests(tests)

		if !result.IsGxP {
			t.Error("expected IsGxP true from structured field")
		}
	})

	t.Run("structured IsManual ORd from field", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@L0"}, IsManual: true},
			{Tags: []string{"@L1"}, IsManual: false},
		}
		result := AggregateTagsForTests(tests)

		if !result.HasManual {
			t.Error("expected HasManual true from structured field")
		}
	})

	t.Run("deterministic output regardless of input order", func(t *testing.T) {
		testsOrderA := []TestReference{
			{Tags: []string{"@L2", "@iv", "@risk-control:RC-003", "@deps:npm"}},
			{Tags: []string{"@L0", "@ov", "@risk-control:RC-001", "@deps:docker"}},
			{Tags: []string{"@L1", "@pv", "@risk-control:RC-002", "@deps:git"}},
		}
		testsOrderB := []TestReference{
			{Tags: []string{"@L1", "@pv", "@risk-control:RC-002", "@deps:git"}},
			{Tags: []string{"@L2", "@iv", "@risk-control:RC-003", "@deps:npm"}},
			{Tags: []string{"@L0", "@ov", "@risk-control:RC-001", "@deps:docker"}},
		}
		testsOrderC := []TestReference{
			{Tags: []string{"@L0", "@ov", "@risk-control:RC-001", "@deps:docker"}},
			{Tags: []string{"@L1", "@pv", "@risk-control:RC-002", "@deps:git"}},
			{Tags: []string{"@L2", "@iv", "@risk-control:RC-003", "@deps:npm"}},
		}

		resultA := AggregateTagsForTests(testsOrderA)
		resultB := AggregateTagsForTests(testsOrderB)
		resultC := AggregateTagsForTests(testsOrderC)

		if !reflect.DeepEqual(resultA, resultB) {
			t.Errorf("results differ for order A vs B:\nA: %+v\nB: %+v", resultA, resultB)
		}
		if !reflect.DeepEqual(resultA, resultC) {
			t.Errorf("results differ for order A vs C:\nA: %+v\nC: %+v", resultA, resultC)
		}
	})

	t.Run("test with no tags and no structured fields is empty", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{}},
			{Tags: nil},
		}
		result := AggregateTagsForTests(tests)

		if !result.IsEmpty() {
			t.Errorf("expected empty result for tests with no tags, got %+v", result)
		}
	})

	t.Run("gxp from tag even when IsGxP field is false", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@gxp"}, IsGxP: false},
		}
		result := AggregateTagsForTests(tests)

		if !result.IsGxP {
			t.Error("expected IsGxP true from @gxp tag even when structured field is false")
		}
	})

	t.Run("Manual from tag even when IsManual field is false", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@Manual"}, IsManual: false},
		}
		result := AggregateTagsForTests(tests)

		if !result.HasManual {
			t.Error("expected HasManual true from @Manual tag even when structured field is false")
		}
	})

	t.Run("unrecognized tags only appear in All", func(t *testing.T) {
		tests := []TestReference{
			{Tags: []string{"@custom-tag", "@L0"}},
		}
		result := AggregateTagsForTests(tests)

		if !sliceHas(result.All, "@custom-tag") {
			t.Error("expected @custom-tag in All")
		}
		if sliceHas(result.Levels, "@custom-tag") {
			t.Error("did not expect @custom-tag in Levels")
		}
		if sliceHas(result.Verification, "@custom-tag") {
			t.Error("did not expect @custom-tag in Verification")
		}
	})
}

// assertSlice compares two string slices for exact equality.
func assertSlice(t *testing.T, field string, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: got %v, want %v", field, got, want)
	}
}

// sliceHas checks if a string slice contains a given value.
func sliceHas(s []string, val string) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}
