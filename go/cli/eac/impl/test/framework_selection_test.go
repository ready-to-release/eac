package test

import (
	"testing"

	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

func TestRemoveExistingModules(t *testing.T) {
	tests := []struct {
		name     string
		from     []string
		existing []string
		want     []string
	}{
		{
			name:     "removes existing items",
			from:     []string{"a", "b", "c"},
			existing: []string{"b"},
			want:     []string{"a", "c"},
		},
		{
			name:     "nothing to remove",
			from:     []string{"a", "b"},
			existing: []string{"x", "y"},
			want:     []string{"a", "b"},
		},
		{
			name:     "all removed",
			from:     []string{"a", "b"},
			existing: []string{"a", "b"},
			want:     nil,
		},
		{
			name:     "empty from",
			from:     nil,
			existing: []string{"a"},
			want:     nil,
		},
		{
			name:     "empty existing",
			from:     []string{"a", "b"},
			existing: nil,
			want:     []string{"a", "b"},
		},
		{
			name:     "both empty",
			from:     nil,
			existing: nil,
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeExistingModules(tt.from, tt.existing)
			if len(got) != len(tt.want) {
				t.Fatalf("removeExistingModules() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("removeExistingModules()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetSuitesIncluded(t *testing.T) {
	tests := []struct {
		name    string
		moniker string
		want    []string
	}{
		{
			name:    "composite suite",
			moniker: "unit+integration",
			want:    []string{"unit", "integration"},
		},
		{
			name:    "triple composite",
			moniker: "L0+L1+L2",
			want:    []string{"L0", "L1", "L2"},
		},
		{
			name:    "single suite returns nil",
			moniker: "unit",
			want:    nil,
		},
		{
			name:    "empty returns nil",
			moniker: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSuitesIncluded(tt.moniker)
			if len(got) != len(tt.want) {
				t.Fatalf("getSuitesIncluded(%q) = %v, want %v", tt.moniker, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getSuitesIncluded(%q)[%d] = %q, want %q", tt.moniker, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGroupTestsByType(t *testing.T) {
	tests := []struct {
		name  string
		input []coretesting.TestReference
		want  map[string]int // type -> count
	}{
		{
			name:  "empty slice",
			input: nil,
			want:  map[string]int{},
		},
		{
			name: "single type",
			input: []coretesting.TestReference{
				{Type: "gotest", TestName: "TestA"},
				{Type: "gotest", TestName: "TestB"},
			},
			want: map[string]int{"gotest": 2},
		},
		{
			name: "multiple types",
			input: []coretesting.TestReference{
				{Type: "gotest", TestName: "TestA"},
				{Type: "godog", TestName: "Scenario1"},
				{Type: "gotest", TestName: "TestB"},
				{Type: "godog", TestName: "Scenario2"},
				{Type: "mocha", TestName: "it1"},
			},
			want: map[string]int{"gotest": 2, "godog": 2, "mocha": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupTestsByType(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("groupTestsByType() returned %d types, want %d", len(got), len(tt.want))
			}
			for typ, wantCount := range tt.want {
				if len(got[typ]) != wantCount {
					t.Errorf("groupTestsByType()[%q] has %d tests, want %d", typ, len(got[typ]), wantCount)
				}
			}
		})
	}
}
