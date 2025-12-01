package changelog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name    string
		version string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{"valid version", "1.2.3", 1, 2, 3, false},
		{"zero version", "0.0.0", 0, 0, 0, false},
		{"large numbers", "100.200.300", 100, 200, 300, false},
		{"invalid - has v prefix", "v1.2.3", 0, 0, 0, true},
		{"invalid - two parts", "1.2", 0, 0, 0, true},
		{"invalid - four parts", "1.2.3.4", 0, 0, 0, true},
		{"invalid - non-numeric", "a.b.c", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, err := ParseSemver(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.major, major)
				assert.Equal(t, tt.minor, minor)
				assert.Equal(t, tt.patch, patch)
			}
		})
	}
}

func TestBumpSemver(t *testing.T) {
	tests := []struct {
		name    string
		version string
		bump    BumpType
		want    string
	}{
		{"patch bump", "1.2.3", BumpPatch, "1.2.4"},
		{"minor bump", "1.2.3", BumpMinor, "1.3.0"},
		{"major bump", "1.2.3", BumpMajor, "2.0.0"},
		{"no bump", "1.2.3", BumpNone, "1.2.3"},
		{"patch from zero", "0.0.0", BumpPatch, "0.0.1"},
		{"minor from zero", "0.0.0", BumpMinor, "0.1.0"},
		{"major from zero", "0.0.0", BumpMajor, "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BumpSemver(tt.version, tt.bump)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestNextCalver(t *testing.T) {
	date := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		existing []string
		want     string
	}{
		{
			name:     "no existing versions",
			existing: []string{},
			want:     "2025.12.01",
		},
		{
			name:     "existing different date",
			existing: []string{"2025.11.30"},
			want:     "2025.12.01",
		},
		{
			name:     "existing same date",
			existing: []string{"2025.12.01"},
			want:     "2025.12.01.1",
		},
		{
			name:     "existing same date with suffix",
			existing: []string{"2025.12.01", "2025.12.01.1"},
			want:     "2025.12.01.2",
		},
		{
			name:     "existing with gaps",
			existing: []string{"2025.12.01", "2025.12.01.1", "2025.12.01.5"},
			want:     "2025.12.01.6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NextCalver(date, tt.existing)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCommitTypeToBump(t *testing.T) {
	tests := []struct {
		commitType string
		breaking   bool
		want       BumpType
	}{
		{"feat", false, BumpMinor},
		{"feat", true, BumpMajor},
		{"fix", false, BumpPatch},
		{"fix", true, BumpMajor},
		{"refactor", false, BumpPatch},
		{"refactor", true, BumpMajor},
		{"docs", false, BumpNone},
		{"docs", true, BumpMajor},
		{"chore", false, BumpNone},
		{"test", false, BumpNone},
		{"unknown", false, BumpNone},
	}

	for _, tt := range tests {
		t.Run(tt.commitType, func(t *testing.T) {
			result := CommitTypeToBump(tt.commitType, tt.breaking)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCalculateBump(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		want    BumpType
	}{
		{
			name:    "empty entries",
			entries: []Entry{},
			want:    BumpNone,
		},
		{
			name: "only docs",
			entries: []Entry{
				{CommitType: "docs"},
			},
			want: BumpNone,
		},
		{
			name: "fix only",
			entries: []Entry{
				{CommitType: "fix"},
			},
			want: BumpPatch,
		},
		{
			name: "feat only",
			entries: []Entry{
				{CommitType: "feat"},
			},
			want: BumpMinor,
		},
		{
			name: "breaking change",
			entries: []Entry{
				{CommitType: "fix", Breaking: true},
			},
			want: BumpMajor,
		},
		{
			name: "mixed - highest wins",
			entries: []Entry{
				{CommitType: "docs"},
				{CommitType: "fix"},
				{CommitType: "feat"},
			},
			want: BumpMinor,
		},
		{
			name: "breaking overrides all",
			entries: []Entry{
				{CommitType: "docs"},
				{CommitType: "fix"},
				{CommitType: "feat"},
				{CommitType: "chore", Breaking: true},
			},
			want: BumpMajor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBump(tt.entries)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMaxBump(t *testing.T) {
	assert.Equal(t, BumpMajor, MaxBump(BumpMajor, BumpMinor))
	assert.Equal(t, BumpMajor, MaxBump(BumpMinor, BumpMajor))
	assert.Equal(t, BumpMinor, MaxBump(BumpMinor, BumpPatch))
	assert.Equal(t, BumpPatch, MaxBump(BumpNone, BumpPatch))
	assert.Equal(t, BumpNone, MaxBump(BumpNone, BumpNone))
}

func TestBumpType_String(t *testing.T) {
	assert.Equal(t, "none", BumpNone.String())
	assert.Equal(t, "patch", BumpPatch.String())
	assert.Equal(t, "minor", BumpMinor.String())
	assert.Equal(t, "major", BumpMajor.String())
}

func TestCommitTypeToChangeType(t *testing.T) {
	tests := []struct {
		commitType string
		want       ChangeType
	}{
		{"feat", Added},
		{"fix", Fixed},
		{"refactor", Changed},
		{"perf", Changed},
		{"style", Changed},
		{"docs", Changed},
		{"chore", Changed},
		{"test", Changed},
		{"unknown", Changed},
	}

	for _, tt := range tests {
		t.Run(tt.commitType, func(t *testing.T) {
			result := CommitTypeToChangeType(tt.commitType)
			assert.Equal(t, tt.want, result)
		})
	}
}
