package workunit

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTagSummary_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		ts   TagSummary
		want bool
	}{
		{
			name: "zero value is empty",
			ts:   TagSummary{},
			want: true,
		},
		{
			name: "nil All is empty",
			ts: TagSummary{
				All: nil,
			},
			want: true,
		},
		{
			name: "empty All slice is empty",
			ts: TagSummary{
				All: []string{},
			},
			want: true,
		},
		{
			name: "non-empty All means not empty",
			ts: TagSummary{
				All: []string{"@L0"},
			},
			want: false,
		},
		{
			name: "single tag in All means not empty",
			ts: TagSummary{
				All: []string{"@gxp"},
			},
			want: false,
		},
		{
			name: "IsEmpty only checks All field: other fields populated without All is still empty",
			ts: TagSummary{
				Levels:       []string{"L0"},
				Verification: []string{"ov"},
				RiskControls: []string{"RC-001"},
				SystemDeps:   []string{"docker"},
				ModuleDeps:   []string{"core"},
				IsGxP:        true,
				HasManual:    true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ts.IsEmpty()
			if got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagSummary_Summary(t *testing.T) {
	tests := []struct {
		name string
		ts   TagSummary
		want string
	}{
		{
			name: "empty returns empty string",
			ts:   TagSummary{},
			want: "",
		},
		{
			name: "empty All returns empty string even with other fields",
			ts: TagSummary{
				Levels: []string{"L0"},
				IsGxP:  true,
			},
			want: "",
		},
		{
			name: "single level",
			ts: TagSummary{
				All:    []string{"@L0"},
				Levels: []string{"L0"},
			},
			want: "L0",
		},
		{
			name: "multiple levels joined with comma",
			ts: TagSummary{
				All:    []string{"@L0", "@L1"},
				Levels: []string{"L0", "L1"},
			},
			want: "L0,L1",
		},
		{
			name: "levels with at-prefix are stripped",
			ts: TagSummary{
				All:    []string{"@L0"},
				Levels: []string{"@L0"},
			},
			want: "L0",
		},
		{
			name: "verification only",
			ts: TagSummary{
				All:          []string{"@ov"},
				Verification: []string{"ov"},
			},
			want: "ov",
		},
		{
			name: "multiple verification tags",
			ts: TagSummary{
				All:          []string{"@iv", "@ov"},
				Verification: []string{"iv", "ov"},
			},
			want: "iv,ov",
		},
		{
			name: "verification with at-prefix stripped",
			ts: TagSummary{
				All:          []string{"@ov"},
				Verification: []string{"@ov"},
			},
			want: "ov",
		},
		{
			name: "gxp flag",
			ts: TagSummary{
				All:   []string{"@gxp"},
				IsGxP: true,
			},
			want: "gxp",
		},
		{
			name: "manual flag",
			ts: TagSummary{
				All:       []string{"@Manual"},
				HasManual: true,
			},
			want: "manual",
		},
		{
			name: "risk controls count plural",
			ts: TagSummary{
				All:          []string{"@risk-control:RC-001", "@risk-control:RC-002"},
				RiskControls: []string{"RC-001", "RC-002"},
			},
			want: "2 controls",
		},
		{
			name: "single risk control",
			ts: TagSummary{
				All:          []string{"@risk-control:RC-001"},
				RiskControls: []string{"RC-001"},
			},
			want: "1 control",
		},
		{
			name: "combined levels and verification",
			ts: TagSummary{
				All:          []string{"@L0", "@L1", "@ov"},
				Levels:       []string{"L0", "L1"},
				Verification: []string{"ov"},
			},
			want: "L0,L1 ov",
		},
		{
			name: "full combined output",
			ts: TagSummary{
				All:          []string{"@L0", "@L1", "@Manual", "@gxp", "@ov", "@risk-control:RC-001", "@risk-control:RC-002"},
				Levels:       []string{"L0", "L1"},
				Verification: []string{"ov"},
				IsGxP:        true,
				HasManual:    true,
				RiskControls: []string{"RC-001", "RC-002"},
			},
			want: "L0,L1 ov gxp manual 2 controls",
		},
		{
			name: "gxp and manual without levels",
			ts: TagSummary{
				All:       []string{"@Manual", "@gxp"},
				IsGxP:     true,
				HasManual: true,
			},
			want: "gxp manual",
		},
		{
			name: "levels and gxp and controls",
			ts: TagSummary{
				All:          []string{"@L2", "@gxp", "@risk-control:RC-100"},
				Levels:       []string{"L2"},
				IsGxP:        true,
				RiskControls: []string{"RC-100"},
			},
			want: "L2 gxp 1 control",
		},
		{
			name: "only levels no verification no flags",
			ts: TagSummary{
				All:    []string{"@L0"},
				Levels: []string{"L0"},
			},
			want: "L0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ts.Summary()
			if got != tt.want {
				t.Errorf("Summary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTagSummary_Summary_OrderOfParts(t *testing.T) {
	// Verify the order: levels, verification, gxp, manual, controls
	ts := TagSummary{
		All:          []string{"@L0", "@Manual", "@gxp", "@ov", "@risk-control:RC-001"},
		Levels:       []string{"L0"},
		Verification: []string{"ov"},
		IsGxP:        true,
		HasManual:    true,
		RiskControls: []string{"RC-001"},
	}
	got := ts.Summary()
	want := "L0 ov gxp manual 1 control"
	if got != want {
		t.Errorf("Summary() order = %q, want %q", got, want)
	}
}

func TestTagSummary_Merge(t *testing.T) {
	tests := []struct {
		name  string
		base  TagSummary
		other TagSummary
		check func(t *testing.T, result TagSummary)
	}{
		{
			name:  "merge two empty summaries",
			base:  TagSummary{},
			other: TagSummary{},
			check: func(t *testing.T, result TagSummary) {
				if !result.IsEmpty() {
					t.Error("merge of two empty should be empty")
				}
			},
		},
		{
			name: "merge empty into populated preserves original",
			base: TagSummary{
				All:    []string{"@L0", "@ov"},
				Levels: []string{"L0"},
				IsGxP:  true,
			},
			other: TagSummary{},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "All", result.All, []string{"@L0", "@ov"})
				assertStringSlice(t, "Levels", result.Levels, []string{"L0"})
				if !result.IsGxP {
					t.Error("expected IsGxP to be true")
				}
			},
		},
		{
			name: "merge populated into empty",
			base: TagSummary{},
			other: TagSummary{
				All:          []string{"@L1"},
				Levels:       []string{"L1"},
				Verification: []string{"iv"},
				HasManual:    true,
			},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "All", result.All, []string{"@L1"})
				assertStringSlice(t, "Levels", result.Levels, []string{"L1"})
				assertStringSlice(t, "Verification", result.Verification, []string{"iv"})
				if !result.HasManual {
					t.Error("expected HasManual to be true")
				}
			},
		},
		{
			name: "unions All slices with deduplication and sorting",
			base: TagSummary{
				All: []string{"@L1", "@gxp", "@ov"},
			},
			other: TagSummary{
				All: []string{"@L0", "@iv", "@ov"},
			},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "All", result.All, []string{"@L0", "@L1", "@gxp", "@iv", "@ov"})
			},
		},
		{
			name: "unions Levels with deduplication and sorting",
			base: TagSummary{
				Levels: []string{"L0", "L2"},
			},
			other: TagSummary{
				Levels: []string{"L0", "L1"},
			},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "Levels", result.Levels, []string{"L0", "L1", "L2"})
			},
		},
		{
			name: "unions Verification with deduplication and sorting",
			base: TagSummary{
				Verification: []string{"ov", "pv"},
			},
			other: TagSummary{
				Verification: []string{"iv", "ov"},
			},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "Verification", result.Verification, []string{"iv", "ov", "pv"})
			},
		},
		{
			name: "unions RiskControls with deduplication and sorting",
			base: TagSummary{
				RiskControls: []string{"RC-001", "RC-002"},
			},
			other: TagSummary{
				RiskControls: []string{"RC-001", "RC-003"},
			},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "RiskControls", result.RiskControls, []string{"RC-001", "RC-002", "RC-003"})
			},
		},
		{
			name: "unions SystemDeps with deduplication and sorting",
			base: TagSummary{
				SystemDeps: []string{"docker", "git"},
			},
			other: TagSummary{
				SystemDeps: []string{"git", "npm"},
			},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "SystemDeps", result.SystemDeps, []string{"docker", "git", "npm"})
			},
		},
		{
			name: "unions ModuleDeps with deduplication and sorting",
			base: TagSummary{
				ModuleDeps: []string{"cli", "core"},
			},
			other: TagSummary{
				ModuleDeps: []string{"cli", "tui"},
			},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "ModuleDeps", result.ModuleDeps, []string{"cli", "core", "tui"})
			},
		},
		{
			name: "ORs IsGxP: false + true = true",
			base:  TagSummary{IsGxP: false},
			other: TagSummary{IsGxP: true},
			check: func(t *testing.T, result TagSummary) {
				if !result.IsGxP {
					t.Error("expected IsGxP true (false OR true)")
				}
			},
		},
		{
			name: "ORs IsGxP: true + false = true",
			base:  TagSummary{IsGxP: true},
			other: TagSummary{IsGxP: false},
			check: func(t *testing.T, result TagSummary) {
				if !result.IsGxP {
					t.Error("expected IsGxP true (true OR false)")
				}
			},
		},
		{
			name: "ORs IsGxP: false + false = false",
			base:  TagSummary{IsGxP: false},
			other: TagSummary{IsGxP: false},
			check: func(t *testing.T, result TagSummary) {
				if result.IsGxP {
					t.Error("expected IsGxP false (false OR false)")
				}
			},
		},
		{
			name: "ORs HasManual: false + true = true",
			base:  TagSummary{HasManual: false},
			other: TagSummary{HasManual: true},
			check: func(t *testing.T, result TagSummary) {
				if !result.HasManual {
					t.Error("expected HasManual true (false OR true)")
				}
			},
		},
		{
			name: "ORs HasManual: true + true = true",
			base:  TagSummary{HasManual: true},
			other: TagSummary{HasManual: true},
			check: func(t *testing.T, result TagSummary) {
				if !result.HasManual {
					t.Error("expected HasManual true (true OR true)")
				}
			},
		},
		{
			name: "full merge of two rich summaries",
			base: TagSummary{
				All:          []string{"@L0", "@gxp", "@ov"},
				Levels:       []string{"L0"},
				Verification: []string{"ov"},
				RiskControls: []string{"RC-001"},
				SystemDeps:   []string{"docker"},
				ModuleDeps:   []string{"core"},
				IsGxP:        true,
				HasManual:    false,
			},
			other: TagSummary{
				All:          []string{"@L1", "@Manual", "@iv"},
				Levels:       []string{"L1"},
				Verification: []string{"iv"},
				RiskControls: []string{"RC-002"},
				SystemDeps:   []string{"git"},
				ModuleDeps:   []string{"tui"},
				IsGxP:        false,
				HasManual:    true,
			},
			check: func(t *testing.T, result TagSummary) {
				assertStringSlice(t, "All", result.All, []string{"@L0", "@L1", "@Manual", "@gxp", "@iv", "@ov"})
				assertStringSlice(t, "Levels", result.Levels, []string{"L0", "L1"})
				assertStringSlice(t, "Verification", result.Verification, []string{"iv", "ov"})
				assertStringSlice(t, "RiskControls", result.RiskControls, []string{"RC-001", "RC-002"})
				assertStringSlice(t, "SystemDeps", result.SystemDeps, []string{"docker", "git"})
				assertStringSlice(t, "ModuleDeps", result.ModuleDeps, []string{"core", "tui"})
				if !result.IsGxP {
					t.Error("expected IsGxP true")
				}
				if !result.HasManual {
					t.Error("expected HasManual true")
				}
			},
		},
		{
			name: "merge returns nil slices when both empty",
			base:  TagSummary{},
			other: TagSummary{},
			check: func(t *testing.T, result TagSummary) {
				if result.All != nil {
					t.Errorf("expected nil All, got %v", result.All)
				}
				if result.Levels != nil {
					t.Errorf("expected nil Levels, got %v", result.Levels)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.base.Merge(tt.other)
			tt.check(t, result)
		})
	}
}

func TestTagSummary_Merge_DoesNotMutateOriginal(t *testing.T) {
	base := TagSummary{
		All:          []string{"@L0"},
		Levels:       []string{"L0"},
		RiskControls: []string{"RC-001"},
		IsGxP:        false,
	}
	other := TagSummary{
		All:          []string{"@L1"},
		Levels:       []string{"L1"},
		RiskControls: []string{"RC-002"},
		IsGxP:        true,
	}

	// Save copies of original slices.
	origAll := make([]string, len(base.All))
	copy(origAll, base.All)
	origLevels := make([]string, len(base.Levels))
	copy(origLevels, base.Levels)

	_ = base.Merge(other)

	// Verify base was not mutated.
	if !reflect.DeepEqual(base.All, origAll) {
		t.Errorf("base.All was mutated: got %v, originally %v", base.All, origAll)
	}
	if !reflect.DeepEqual(base.Levels, origLevels) {
		t.Errorf("base.Levels was mutated: got %v, originally %v", base.Levels, origLevels)
	}
	if base.IsGxP {
		t.Error("base.IsGxP was mutated to true")
	}
}

func TestTagSummary_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		ts   TagSummary
	}{
		{
			name: "zero value round-trips",
			ts:   TagSummary{},
		},
		{
			name: "fully populated round-trips",
			ts: TagSummary{
				All:          []string{"@L0", "@L1", "@Manual", "@gxp", "@ov"},
				Levels:       []string{"L0", "L1"},
				Verification: []string{"ov"},
				RiskControls: []string{"RC-001", "RC-002"},
				SystemDeps:   []string{"docker", "git"},
				ModuleDeps:   []string{"core", "tui"},
				IsGxP:        true,
				HasManual:    true,
			},
		},
		{
			name: "partial fields round-trip",
			ts: TagSummary{
				All:    []string{"@L2", "@iv"},
				Levels: []string{"L2"},
				IsGxP:  false,
			},
		},
		{
			name: "only boolean flags set with All present",
			ts: TagSummary{
				All:       []string{"@gxp", "@Manual"},
				IsGxP:     true,
				HasManual: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.ts)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}

			var got TagSummary
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}

			// Compare slices (nil vs empty may differ, normalize).
			assertStringSliceNullSafe(t, "All", got.All, tt.ts.All)
			assertStringSliceNullSafe(t, "Levels", got.Levels, tt.ts.Levels)
			assertStringSliceNullSafe(t, "Verification", got.Verification, tt.ts.Verification)
			assertStringSliceNullSafe(t, "RiskControls", got.RiskControls, tt.ts.RiskControls)
			assertStringSliceNullSafe(t, "SystemDeps", got.SystemDeps, tt.ts.SystemDeps)
			assertStringSliceNullSafe(t, "ModuleDeps", got.ModuleDeps, tt.ts.ModuleDeps)

			if got.IsGxP != tt.ts.IsGxP {
				t.Errorf("IsGxP: got %v, want %v", got.IsGxP, tt.ts.IsGxP)
			}
			if got.HasManual != tt.ts.HasManual {
				t.Errorf("HasManual: got %v, want %v", got.HasManual, tt.ts.HasManual)
			}
		})
	}
}

func TestTagSummary_JSONFieldNames(t *testing.T) {
	ts := TagSummary{
		All:          []string{"@L0"},
		Levels:       []string{"L0"},
		Verification: []string{"ov"},
		RiskControls: []string{"RC-001"},
		SystemDeps:   []string{"docker"},
		ModuleDeps:   []string{"core"},
		IsGxP:        true,
		HasManual:    true,
	}

	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	expectedKeys := []string{
		"all", "levels", "verification", "risk_controls",
		"system_deps", "module_deps", "is_gxp", "has_manual",
	}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q not found in marshaled output", key)
		}
	}

	// Verify no unexpected keys.
	if len(raw) != len(expectedKeys) {
		t.Errorf("expected %d JSON keys, got %d", len(expectedKeys), len(raw))
	}
}

// assertStringSlice compares two string slices for exact equality.
func assertStringSlice(t *testing.T, field string, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: got %v, want %v", field, got, want)
	}
}

// assertStringSliceNullSafe compares two string slices, treating nil and empty as equivalent.
func assertStringSliceNullSafe(t *testing.T, field string, got, want []string) {
	t.Helper()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: got %v, want %v", field, got, want)
	}
}
