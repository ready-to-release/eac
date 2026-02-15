package docsmanifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateDocsManifestCommand_Name(t *testing.T) {
	cmd := &updateDocsManifestCommand{}
	assert.Equal(t, "update docs-manifest", cmd.Name())
}

func TestUpdateDocsManifestCommand_Metadata(t *testing.T) {
	cmd := &updateDocsManifestCommand{}
	meta := cmd.Metadata()

	assert.Equal(t, "update-docs-manifest", meta.CanonicalName)
	assert.Equal(t, "Update the documentation assets manifest", meta.Short)
	assert.NotEmpty(t, meta.Long)
	assert.Len(t, meta.Flags, 3, "expected 3 flags: check, dry-run, verbose")

	flagNames := make(map[string]bool)
	for _, f := range meta.Flags {
		flagNames[f.Name] = true
	}
	assert.True(t, flagNames["check"], "missing check flag")
	assert.True(t, flagNames["dry-run"], "missing dry-run flag")
	assert.True(t, flagNames["verbose"], "missing verbose flag")
}

func TestCommands_ReturnsOneCommand(t *testing.T) {
	cmds := Commands()
	assert.Len(t, cmds, 1, "Commands() should return exactly one command")
}

func TestHandleCheckMode(t *testing.T) {
	tests := []struct {
		name     string
		result   *UpdateResult
		wantCode int
	}{
		{
			name: "up to date returns 0",
			result: &UpdateResult{
				NeedsCacheWrite: false,
			},
			wantCode: 0,
		},
		{
			name: "stale cache returns 1",
			result: &UpdateResult{
				NeedsCacheWrite: true,
			},
			wantCode: 1,
		},
		{
			name: "stale cache with added assets returns 1",
			result: &UpdateResult{
				NeedsCacheWrite: true,
				Added:           []string{"new/image.png"},
			},
			wantCode: 1,
		},
		{
			name: "stale cache with removed assets returns 1",
			result: &UpdateResult{
				NeedsCacheWrite: true,
				Removed:         []string{"old/image.png"},
			},
			wantCode: 1,
		},
		{
			name: "stale cache with both added and removed returns 1",
			result: &UpdateResult{
				NeedsCacheWrite: true,
				Added:           []string{"new/a.png", "new/b.png"},
				Removed:         []string{"old/c.png"},
			},
			wantCode: 1,
		},
		{
			name: "descriptions write needed but cache up to date returns 0",
			result: &UpdateResult{
				NeedsDescriptionsWrite: true,
				NeedsCacheWrite:        false,
			},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := handleCheckMode(tt.result)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestHandleDryRunMode(t *testing.T) {
	tests := []struct {
		name     string
		result   *UpdateResult
		wantCode int
	}{
		{
			name: "no changes returns 0",
			result: &UpdateResult{
				NeedsDescriptionsWrite: false,
				NeedsCacheWrite:        false,
			},
			wantCode: 0,
		},
		{
			name: "descriptions need write returns 0",
			result: &UpdateResult{
				NeedsDescriptionsWrite: true,
				NeedsCacheWrite:        false,
			},
			wantCode: 0,
		},
		{
			name: "cache needs write returns 0",
			result: &UpdateResult{
				NeedsDescriptionsWrite: false,
				NeedsCacheWrite:        true,
			},
			wantCode: 0,
		},
		{
			name: "both need write returns 0",
			result: &UpdateResult{
				NeedsDescriptionsWrite: true,
				NeedsCacheWrite:        true,
			},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := handleDryRunMode(tt.result, "/tmp/descriptions.yml", "/tmp/.manifest-cache.json")
			assert.Equal(t, tt.wantCode, code, "dry-run mode always returns 0")
		})
	}
}

func TestReportChanges_NoChanges(t *testing.T) {
	// Verify reportChanges does not panic with empty result
	result := &UpdateResult{}
	assert.NotPanics(t, func() {
		reportChanges(result)
	})
}

func TestReportChanges_WithChanges(t *testing.T) {
	result := &UpdateResult{
		Added:                  []string{"a.png", "b.png"},
		Removed:                []string{"c.png"},
		UpdatedUsedIn:          []string{"d.png"},
		NeedsDescriptionsWrite: true,
		NeedsCacheWrite:        true,
	}
	assert.NotPanics(t, func() {
		reportChanges(result)
	})
}

func TestReportSummaryStats_ZeroAssets(t *testing.T) {
	result := &UpdateResult{
		TotalAssets:  0,
		UsedAssets:   0,
		UnusedAssets: 0,
	}
	assert.NotPanics(t, func() {
		reportSummaryStats(result)
	})
}

func TestReportSummaryStats_WithAssets(t *testing.T) {
	result := &UpdateResult{
		TotalAssets:  10,
		UsedAssets:   7,
		UnusedAssets: 3,
	}
	assert.NotPanics(t, func() {
		reportSummaryStats(result)
	})
}

func TestReportMissingDescriptions_NoMissing(t *testing.T) {
	mergeResult := &MergeResult{
		Descriptions: map[string]Description{
			"a.png": {Active: true, Description: "has description"},
		},
		Result: &UpdateResult{},
	}
	assert.NotPanics(t, func() {
		reportMissingDescriptions(mergeResult, false)
	})
}

func TestReportMissingDescriptions_WithMissing(t *testing.T) {
	mergeResult := &MergeResult{
		Descriptions: map[string]Description{
			"a.png": {Active: true, Description: ""},
			"b.png": {Active: true, Description: "has one"},
			"c.png": {Active: true, Description: ""},
		},
		Result: &UpdateResult{
			Added: []string{"a.png"},
		},
	}
	assert.NotPanics(t, func() {
		reportMissingDescriptions(mergeResult, false)
	})
}

func TestReportMissingDescriptions_VerboseMode(t *testing.T) {
	mergeResult := &MergeResult{
		Descriptions: map[string]Description{
			"a.png": {Active: true, Description: ""},
			"b.png": {Active: true, Description: ""},
		},
		Result: &UpdateResult{},
	}
	assert.NotPanics(t, func() {
		reportMissingDescriptions(mergeResult, true)
	})
}
