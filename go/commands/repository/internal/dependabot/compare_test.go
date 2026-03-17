package dependabot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name           string
		declared       []UpdateEntry
		discovered     []EcosystemEntry
		wantMissing    int
		wantExtra      int
		wantMatched    int
		wantHasIssues  bool
	}{
		{
			name: "perfect match with zero missing and zero extra",
			declared: []UpdateEntry{
				{PackageEcosystem: "gomod", Directory: "/go/core"},
				{PackageEcosystem: "npm", Directory: "/web"},
			},
			discovered: []EcosystemEntry{
				{Ecosystem: "gomod", Directory: "/go/core"},
				{Ecosystem: "npm", Directory: "/web"},
			},
			wantMissing:   0,
			wantExtra:     0,
			wantMatched:   2,
			wantHasIssues: false,
		},
		{
			name:     "only missing entries",
			declared: []UpdateEntry{},
			discovered: []EcosystemEntry{
				{Ecosystem: "gomod", Directory: "/go/core"},
				{Ecosystem: "npm", Directory: "/web"},
			},
			wantMissing:   2,
			wantExtra:     0,
			wantMatched:   0,
			wantHasIssues: true,
		},
		{
			name: "only extra entries",
			declared: []UpdateEntry{
				{PackageEcosystem: "gomod", Directory: "/go/core"},
				{PackageEcosystem: "docker", Directory: "/containers/app"},
			},
			discovered:    []EcosystemEntry{},
			wantMissing:   0,
			wantExtra:     2,
			wantMatched:   0,
			wantHasIssues: true,
		},
		{
			name: "both missing and extra",
			declared: []UpdateEntry{
				{PackageEcosystem: "gomod", Directory: "/go/core"},
				{PackageEcosystem: "docker", Directory: "/containers/old"},
			},
			discovered: []EcosystemEntry{
				{Ecosystem: "gomod", Directory: "/go/core"},
				{Ecosystem: "npm", Directory: "/web"},
			},
			wantMissing:   1,
			wantExtra:     1,
			wantMatched:   1,
			wantHasIssues: true,
		},
		{
			name:     "empty declared list",
			declared: []UpdateEntry{},
			discovered: []EcosystemEntry{
				{Ecosystem: "pip", Directory: "/scripts"},
			},
			wantMissing:   1,
			wantExtra:     0,
			wantMatched:   0,
			wantHasIssues: true,
		},
		{
			name: "empty discovered list",
			declared: []UpdateEntry{
				{PackageEcosystem: "pip", Directory: "/scripts"},
			},
			discovered:    []EcosystemEntry{},
			wantMissing:   0,
			wantExtra:     1,
			wantMatched:   0,
			wantHasIssues: true,
		},
		{
			name:           "both empty",
			declared:       []UpdateEntry{},
			discovered:     []EcosystemEntry{},
			wantMissing:    0,
			wantExtra:      0,
			wantMatched:    0,
			wantHasIssues:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := Compare(tt.declared, tt.discovered)

			assert.Equal(t, tt.wantMissing, len(report.Missing), "missing count")
			assert.Equal(t, tt.wantExtra, len(report.Extra), "extra count")
			assert.Equal(t, tt.wantMatched, len(report.Matched), "matched count")
			assert.Equal(t, tt.wantHasIssues, report.HasIssues(), "has issues")

			// Verify the report preserves the original inputs
			assert.Equal(t, tt.declared, report.Declared, "declared preserved")
			assert.Equal(t, tt.discovered, report.Discovered, "discovered preserved")
		})
	}
}

func TestCompare_ConsolidatedEntryCoversAllDirectories(t *testing.T) {
	declared := []UpdateEntry{
		{
			PackageEcosystem: "gomod",
			Directories:      []string{"/go/core", "/go/cli/eac", "/go/clibase"},
		},
		{PackageEcosystem: "npm", Directory: "/web"},
	}
	discovered := []EcosystemEntry{
		{Ecosystem: "gomod", Directory: "/go/core"},
		{Ecosystem: "gomod", Directory: "/go/cli/eac"},
		{Ecosystem: "gomod", Directory: "/go/clibase"},
		{Ecosystem: "npm", Directory: "/web"},
	}

	report := Compare(declared, discovered)

	assert.Empty(t, report.Missing, "no missing: all gomod dirs covered by consolidated entry")
	assert.Empty(t, report.Extra, "no extra")
	assert.Len(t, report.Matched, 1, "npm matched")
	assert.Len(t, report.Consolidated, 1, "one consolidated entry")
	assert.False(t, report.HasIssues())
}

func TestCompare_ConsolidatedEntryDoesNotCoverNewModule(t *testing.T) {
	declared := []UpdateEntry{
		{
			PackageEcosystem: "gomod",
			Directories:      []string{"/go/core", "/go/cli/eac"},
		},
	}
	discovered := []EcosystemEntry{
		{Ecosystem: "gomod", Directory: "/go/core"},
		{Ecosystem: "gomod", Directory: "/go/cli/eac"},
		{Ecosystem: "gomod", Directory: "/go/new-module"},
	}

	report := Compare(declared, discovered)

	assert.Len(t, report.Missing, 1, "new module not covered")
	assert.Equal(t, "/go/new-module", report.Missing[0].Directory)
	assert.True(t, report.HasIssues())
}

func TestCompare_MissingContainsCorrectEntries(t *testing.T) {
	declared := []UpdateEntry{
		{PackageEcosystem: "gomod", Directory: "/go/core"},
	}
	discovered := []EcosystemEntry{
		{Ecosystem: "gomod", Directory: "/go/core"},
		{Ecosystem: "npm", Directory: "/web"},
		{Ecosystem: "docker", Directory: "/containers/app"},
	}

	report := Compare(declared, discovered)

	assert.Len(t, report.Missing, 2)

	missingKeys := make(map[string]bool)
	for _, m := range report.Missing {
		missingKeys[m.Key()] = true
	}
	assert.True(t, missingKeys["npm:/web"], "npm:/web should be missing")
	assert.True(t, missingKeys["docker:/containers/app"], "docker:/containers/app should be missing")
}

func TestCompare_ExtraContainsCorrectEntries(t *testing.T) {
	declared := []UpdateEntry{
		{PackageEcosystem: "gomod", Directory: "/go/core"},
		{PackageEcosystem: "docker", Directory: "/containers/removed"},
		{PackageEcosystem: "pip", Directory: "/scripts/old"},
	}
	discovered := []EcosystemEntry{
		{Ecosystem: "gomod", Directory: "/go/core"},
	}

	report := Compare(declared, discovered)

	assert.Len(t, report.Extra, 2)

	extraKeys := make(map[string]bool)
	for _, e := range report.Extra {
		extraKeys[e.Key()] = true
	}
	assert.True(t, extraKeys["docker:/containers/removed"], "docker entry should be extra")
	assert.True(t, extraKeys["pip:/scripts/old"], "pip entry should be extra")
}
