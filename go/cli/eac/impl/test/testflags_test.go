package test

import (
	"testing"
)

func TestParseTestSpecificFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantSuite     string
		wantCoverage  bool
		wantListOnly  bool
		wantRemaining []string
		wantErr       bool
	}{
		{
			name:          "empty args",
			args:          nil,
			wantRemaining: nil,
		},
		{
			name:      "suite with separate value",
			args:      []string{"--suite", "L0"},
			wantSuite: "L0",
		},
		{
			name:      "suite with equals syntax",
			args:      []string{"--suite=L2"},
			wantSuite: "L2",
		},
		{
			name:         "coverage flag",
			args:         []string{"--coverage"},
			wantCoverage: true,
		},
		{
			name:         "list-only flag",
			args:         []string{"--list-only"},
			wantListOnly: true,
		},
		{
			name:         "all flags combined",
			args:         []string{"--suite", "L1", "--coverage", "--list-only"},
			wantSuite:    "L1",
			wantCoverage: true,
			wantListOnly: true,
		},
		{
			name:          "unknown flags passed through",
			args:          []string{"--verbose", "--suite", "L0", "--output", "json"},
			wantSuite:     "L0",
			wantRemaining: []string{"--verbose", "--output", "json"},
		},
		{
			name:    "suite missing value",
			args:    []string{"--suite"},
			wantErr: true,
		},
		{
			name:          "positional args passed through",
			args:          []string{"module-a", "--coverage"},
			wantCoverage:  true,
			wantRemaining: []string{"module-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, remaining, err := ParseTestSpecificFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if flags.SuiteName != tt.wantSuite {
				t.Errorf("SuiteName = %q, want %q", flags.SuiteName, tt.wantSuite)
			}
			if flags.Coverage != tt.wantCoverage {
				t.Errorf("Coverage = %v, want %v", flags.Coverage, tt.wantCoverage)
			}
			if flags.ListOnly != tt.wantListOnly {
				t.Errorf("ListOnly = %v, want %v", flags.ListOnly, tt.wantListOnly)
			}
			if len(remaining) != len(tt.wantRemaining) {
				t.Fatalf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
			for i := range remaining {
				if remaining[i] != tt.wantRemaining[i] {
					t.Errorf("remaining[%d] = %q, want %q", i, remaining[i], tt.wantRemaining[i])
				}
			}
		})
	}
}
