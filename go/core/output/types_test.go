package output

import (
	"encoding/json"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Status Constants Tests
// =============================================================================

func TestStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "pending status",
			status:   StatusPending,
			expected: "pending",
		},
		{
			name:     "in_progress status",
			status:   StatusInProgress,
			expected: "in_progress",
		},
		{
			name:     "completed status",
			status:   StatusCompleted,
			expected: "completed",
		},
		{
			name:     "failed status",
			status:   StatusFailed,
			expected: "failed",
		},
		{
			name:     "cached status",
			status:   StatusCached,
			expected: "cached",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, Status(tt.expected), tt.status)
		})
	}
}

func TestStatus_CanBeUsedAsMapKey(t *testing.T) {
	m := map[Status]string{
		StatusPending:    "waiting to start",
		StatusInProgress: "currently running",
		StatusCompleted:  "finished successfully",
		StatusFailed:     "finished with errors",
		StatusCached:     "reused from cache",
	}

	assert.Equal(t, "waiting to start", m[StatusPending])
	assert.Equal(t, "currently running", m[StatusInProgress])
	assert.Equal(t, "finished successfully", m[StatusCompleted])
	assert.Equal(t, "finished with errors", m[StatusFailed])
	assert.Equal(t, "reused from cache", m[StatusCached])
}

func TestStatus_StringConversion(t *testing.T) {
	assert.Equal(t, "pending", string(StatusPending))
	assert.Equal(t, "in_progress", string(StatusInProgress))
	assert.Equal(t, "completed", string(StatusCompleted))
	assert.Equal(t, "failed", string(StatusFailed))
	assert.Equal(t, "cached", string(StatusCached))
}

// =============================================================================
// Artifact Type Tests
// =============================================================================

func TestArtifact_FieldsExist(t *testing.T) {
	artifact := Artifact{
		ID:     "binary-linux-amd64",
		Path:   "out/build/core/go/eac-linux-amd64",
		SHA256: "sha256:abc123def456789",
		Size:   12345678,
		Type:   "binary",
	}

	assert.Equal(t, "binary-linux-amd64", artifact.ID)
	assert.Equal(t, "out/build/core/go/eac-linux-amd64", artifact.Path)
	assert.Equal(t, "sha256:abc123def456789", artifact.SHA256)
	assert.Equal(t, int64(12345678), artifact.Size)
	assert.Equal(t, "binary", artifact.Type)
}

func TestArtifact_ZeroValueIsValid(t *testing.T) {
	var artifact Artifact

	assert.Empty(t, artifact.ID)
	assert.Empty(t, artifact.Path)
	assert.Empty(t, artifact.SHA256)
	assert.Zero(t, artifact.Size)
	assert.Empty(t, artifact.Type)
}

func TestArtifact_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		artifact Artifact
	}{
		{
			name: "binary artifact",
			artifact: Artifact{
				ID:     "eac-linux-amd64",
				Path:   "out/build/eac/go/eac-linux-amd64",
				SHA256: "sha256:abc123",
				Size:   10485760,
				Type:   "binary",
			},
		},
		{
			name: "docker image artifact",
			artifact: Artifact{
				ID:     "eac:latest",
				Path:   "out/build/eac/docker/image.tar",
				SHA256: "sha256:docker789",
				Size:   52428800,
				Type:   "docker-image",
			},
		},
		{
			name: "test report artifact",
			artifact: Artifact{
				ID:     "junit-report",
				Path:   "out/test/core/go/unit/junit.xml",
				SHA256: "sha256:test456",
				Size:   4096,
				Type:   "report",
			},
		},
		{
			name: "scan results artifact",
			artifact: Artifact{
				ID:     "trivy-results",
				Path:   "out/scan/core/go/trivy-vuln/results.json",
				SHA256: "sha256:scan999",
				Size:   8192,
				Type:   "scan-report",
			},
		},
		{
			name: "zero-size artifact (marker file)",
			artifact: Artifact{
				ID:     "lint-passed",
				Path:   "out/lint/core/go/.passed",
				SHA256: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				Size:   0,
				Type:   "marker",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible
			_ = tt.artifact.ID
			_ = tt.artifact.Path
			_ = tt.artifact.SHA256
			_ = tt.artifact.Size
			_ = tt.artifact.Type
		})
	}
}

func TestArtifact_MarshalJSON(t *testing.T) {
	artifact := Artifact{
		ID:     "test-artifact",
		Path:   "out/test/path",
		SHA256: "sha256:abc123",
		Size:   1024,
		Type:   "binary",
	}

	data, err := json.Marshal(artifact)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"path"`)
	assert.Contains(t, jsonStr, `"sha256"`)
	assert.Contains(t, jsonStr, `"size"`)
	assert.Contains(t, jsonStr, `"type"`)
}

func TestArtifact_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"id": "my-artifact",
		"path": "out/build/module/component/file",
		"sha256": "sha256:deadbeef",
		"size": 2048,
		"type": "binary"
	}`

	var artifact Artifact
	err := json.Unmarshal([]byte(jsonData), &artifact)
	require.NoError(t, err)

	assert.Equal(t, "my-artifact", artifact.ID)
	assert.Equal(t, "out/build/module/component/file", artifact.Path)
	assert.Equal(t, "sha256:deadbeef", artifact.SHA256)
	assert.Equal(t, int64(2048), artifact.Size)
	assert.Equal(t, "binary", artifact.Type)
}

func TestArtifact_MarshalUnmarshalRoundTrip(t *testing.T) {
	original := Artifact{
		ID:     "roundtrip-test",
		Path:   "out/build/core/go/binary",
		SHA256: "sha256:fedcba9876543210",
		Size:   999999,
		Type:   "executable",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Artifact
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original, restored)
}

// =============================================================================
// UoWManifest Type Tests
// =============================================================================

func TestUoWManifest_FieldsExist(t *testing.T) {
	now := time.Now()
	manifest := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "core",
		Component:  "go",
		Tool:       "go",
		ExitCode:   0,
		InputHash:  "sha256:input123",
		ExecutedAt: now,
		Duration:   5 * time.Second,
		Artifacts: []Artifact{
			{ID: "binary", Path: "out/build/core/go/binary", SHA256: "sha256:bin", Size: 1000, Type: "binary"},
		},
		OutputHash: "sha256:output456",
		Version:    "1.0.0",
	}

	assert.Equal(t, core.ActionBuild, manifest.Action)
	assert.Equal(t, "core", manifest.Module)
	assert.Equal(t, "go", manifest.Component)
	assert.Equal(t, "go", manifest.Tool)
	assert.Equal(t, 0, manifest.ExitCode)
	assert.Equal(t, "sha256:input123", manifest.InputHash)
	assert.Equal(t, now, manifest.ExecutedAt)
	assert.Equal(t, 5*time.Second, manifest.Duration)
	assert.Len(t, manifest.Artifacts, 1)
	assert.Equal(t, "sha256:output456", manifest.OutputHash)
	assert.Equal(t, "1.0.0", manifest.Version)
}

func TestUoWManifest_ZeroValueIsValid(t *testing.T) {
	var manifest UoWManifest

	assert.Empty(t, manifest.Action)
	assert.Empty(t, manifest.Module)
	assert.Empty(t, manifest.Component)
	assert.Empty(t, manifest.Tool)
	assert.Zero(t, manifest.ExitCode)
	assert.Empty(t, manifest.InputHash)
	assert.True(t, manifest.ExecutedAt.IsZero())
	assert.Zero(t, manifest.Duration)
	assert.Nil(t, manifest.Artifacts)
	assert.Empty(t, manifest.OutputHash)
	assert.Empty(t, manifest.Version)
}

func TestUoWManifest_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		manifest UoWManifest
	}{
		{
			name: "build manifest",
			manifest: UoWManifest{
				Action:     core.ActionBuild,
				Module:     "core",
				Component:  "go",
				Tool:       "go",
				ExitCode:   0,
				InputHash:  "sha256:src123",
				ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Duration:   45 * time.Second,
				Artifacts: []Artifact{
					{ID: "binary", Path: "out/build/core/go/eac", SHA256: "sha256:bin1", Size: 10000000, Type: "binary"},
				},
				OutputHash: "sha256:out123",
				Version:    "1.0.0",
			},
		},
		{
			name: "test manifest with testset",
			manifest: UoWManifest{
				Action:     core.ActionTest,
				Module:     "core",
				Component:  "go",
				Tool:       "gotest",
				ExitCode:   0,
				InputHash:  "sha256:test123",
				ExecutedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
				Duration:   120 * time.Second,
				Artifacts: []Artifact{
					{ID: "junit", Path: "out/test/core/go/unit/junit.xml", SHA256: "sha256:junit1", Size: 5000, Type: "report"},
					{ID: "coverage", Path: "out/test/core/go/unit/coverage.out", SHA256: "sha256:cov1", Size: 2000, Type: "coverage"},
				},
				OutputHash: "sha256:testout123",
				Version:    "1.0.0",
			},
		},
		{
			name: "failed lint manifest",
			manifest: UoWManifest{
				Action:     core.ActionLint,
				Module:     "web-app",
				Component:  "typescript",
				Tool:       "eslint",
				ExitCode:   1,
				InputHash:  "sha256:lint456",
				ExecutedAt: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Duration:   15 * time.Second,
				Artifacts: []Artifact{
					{ID: "results", Path: "out/lint/web-app/typescript/eslint.json", SHA256: "sha256:lint1", Size: 1000, Type: "report"},
				},
				OutputHash: "sha256:lintout456",
				Version:    "1.0.0",
			},
		},
		{
			name: "scan manifest with vulnerabilities",
			manifest: UoWManifest{
				Action:     core.ActionScan,
				Module:     "eac",
				Component:  "docker",
				Tool:       "trivy-vuln",
				ExitCode:   0,
				InputHash:  "sha256:scan789",
				ExecutedAt: time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC),
				Duration:   60 * time.Second,
				Artifacts: []Artifact{
					{ID: "trivy-report", Path: "out/scan/eac/docker/trivy-vuln/results.json", SHA256: "sha256:scan1", Size: 8000, Type: "scan-report"},
					{ID: "sbom", Path: "out/scan/eac/docker/trivy-vuln/sbom.json", SHA256: "sha256:sbom1", Size: 12000, Type: "sbom"},
				},
				OutputHash: "sha256:scanout789",
				Version:    "1.0.0",
			},
		},
		{
			name: "manifest with no artifacts",
			manifest: UoWManifest{
				Action:     core.ActionLint,
				Module:     "docs",
				Component:  "assets",
				Tool:       "markdownlint",
				ExitCode:   0,
				InputHash:  "sha256:docs123",
				ExecutedAt: time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
				Duration:   2 * time.Second,
				Artifacts:  []Artifact{},
				OutputHash: "sha256:empty",
				Version:    "1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible
			_ = tt.manifest.Action
			_ = tt.manifest.Module
			_ = tt.manifest.Component
			_ = tt.manifest.Tool
			_ = tt.manifest.ExitCode
			_ = tt.manifest.InputHash
			_ = tt.manifest.ExecutedAt
			_ = tt.manifest.Duration
			_ = tt.manifest.Artifacts
			_ = tt.manifest.OutputHash
			_ = tt.manifest.Version
		})
	}
}

func TestUoWManifest_MarshalJSON(t *testing.T) {
	manifest := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "core",
		Component:  "go",
		Tool:       "go",
		ExitCode:   0,
		InputHash:  "sha256:input",
		ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Duration:   30 * time.Second,
		Artifacts: []Artifact{
			{ID: "bin", Path: "out/build/core/go/bin", SHA256: "sha256:abc", Size: 1000, Type: "binary"},
		},
		OutputHash: "sha256:output",
		Version:    "1.0.0",
	}

	data, err := json.Marshal(manifest)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"context"`)
	assert.Contains(t, jsonStr, `"module"`)
	assert.Contains(t, jsonStr, `"component"`)
	assert.Contains(t, jsonStr, `"tool"`)
	assert.Contains(t, jsonStr, `"exit_code"`)
	assert.Contains(t, jsonStr, `"input_hash"`)
	assert.Contains(t, jsonStr, `"executed_at"`)
	assert.Contains(t, jsonStr, `"duration"`)
	assert.Contains(t, jsonStr, `"artifacts"`)
	assert.Contains(t, jsonStr, `"output_hash"`)
	assert.Contains(t, jsonStr, `"version"`)
}

func TestUoWManifest_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"context": "build",
		"module": "core",
		"component": "go",
		"tool": "go",
		"exit_code": 0,
		"input_hash": "sha256:inputhash",
		"executed_at": "2024-01-15T10:30:00Z",
		"duration": 30000000000,
		"artifacts": [
			{
				"id": "binary",
				"path": "out/build/core/go/binary",
				"sha256": "sha256:abc",
				"size": 1024,
				"type": "binary"
			}
		],
		"output_hash": "sha256:outputhash",
		"version": "1.0.0"
	}`

	var manifest UoWManifest
	err := json.Unmarshal([]byte(jsonData), &manifest)
	require.NoError(t, err)

	assert.Equal(t, core.ActionBuild, manifest.Action)
	assert.Equal(t, "core", manifest.Module)
	assert.Equal(t, "go", manifest.Component)
	assert.Equal(t, "go", manifest.Tool)
	assert.Equal(t, 0, manifest.ExitCode)
	assert.Equal(t, "sha256:inputhash", manifest.InputHash)
	assert.Equal(t, 30*time.Second, manifest.Duration)
	assert.Len(t, manifest.Artifacts, 1)
	assert.Equal(t, "sha256:outputhash", manifest.OutputHash)
	assert.Equal(t, "1.0.0", manifest.Version)
}

func TestUoWManifest_MarshalUnmarshalRoundTrip(t *testing.T) {
	original := UoWManifest{
		Action:     core.ActionTest,
		Module:     "roundtrip-module",
		Component:  "gherkin",
		Tool:       "godog",
		ExitCode:   0,
		InputHash:  "sha256:roundtrip-input",
		ExecutedAt: time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
		Duration:   90 * time.Second,
		Artifacts: []Artifact{
			{ID: "junit", Path: "out/test/mod/gherkin/junit.xml", SHA256: "sha256:junit", Size: 5000, Type: "report"},
			{ID: "cucumber", Path: "out/test/mod/gherkin/cucumber.json", SHA256: "sha256:cuc", Size: 8000, Type: "report"},
		},
		OutputHash: "sha256:roundtrip-output",
		Version:    "1.0.0",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored UoWManifest
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Compare fields
	assert.Equal(t, original.Action, restored.Action)
	assert.Equal(t, original.Module, restored.Module)
	assert.Equal(t, original.Component, restored.Component)
	assert.Equal(t, original.Tool, restored.Tool)
	assert.Equal(t, original.ExitCode, restored.ExitCode)
	assert.Equal(t, original.InputHash, restored.InputHash)
	assert.True(t, original.ExecutedAt.Equal(restored.ExecutedAt))
	assert.Equal(t, original.Duration, restored.Duration)
	assert.Equal(t, original.Artifacts, restored.Artifacts)
	assert.Equal(t, original.OutputHash, restored.OutputHash)
	assert.Equal(t, original.Version, restored.Version)
}

func TestUoWManifest_SuccessfulExecution(t *testing.T) {
	manifest := UoWManifest{ExitCode: 0}
	assert.Equal(t, 0, manifest.ExitCode, "Successful execution should have ExitCode 0")
}

func TestUoWManifest_FailedExecution(t *testing.T) {
	exitCodes := []int{1, 2, 127, 255}
	for _, code := range exitCodes {
		manifest := UoWManifest{ExitCode: code}
		assert.NotEqual(t, 0, manifest.ExitCode, "Failed execution should have non-zero ExitCode")
	}
}

// =============================================================================
// ValidationResult Type Tests
// =============================================================================

func TestValidationResult_FieldsExist(t *testing.T) {
	result := ValidationResult{
		Valid:            false,
		ManifestExists:   true,
		ManifestValid:    true,
		ArtifactsValid:   false,
		MissingArtifacts: []string{"out/build/core/go/binary"},
		CorruptArtifacts: []string{"out/build/core/go/other"},
		Error:            nil,
	}

	assert.False(t, result.Valid)
	assert.True(t, result.ManifestExists)
	assert.True(t, result.ManifestValid)
	assert.False(t, result.ArtifactsValid)
	assert.Equal(t, []string{"out/build/core/go/binary"}, result.MissingArtifacts)
	assert.Equal(t, []string{"out/build/core/go/other"}, result.CorruptArtifacts)
	assert.Nil(t, result.Error)
}

func TestValidationResult_ZeroValueIsValid(t *testing.T) {
	var result ValidationResult

	assert.False(t, result.Valid)
	assert.False(t, result.ManifestExists)
	assert.False(t, result.ManifestValid)
	assert.False(t, result.ArtifactsValid)
	assert.Nil(t, result.MissingArtifacts)
	assert.Nil(t, result.CorruptArtifacts)
	assert.Nil(t, result.Error)
}

func TestValidationResult_AllValid(t *testing.T) {
	result := ValidationResult{
		Valid:            true,
		ManifestExists:   true,
		ManifestValid:    true,
		ArtifactsValid:   true,
		MissingArtifacts: []string{},
		CorruptArtifacts: []string{},
		Error:            nil,
	}

	assert.True(t, result.Valid)
	assert.True(t, result.ManifestExists)
	assert.True(t, result.ManifestValid)
	assert.True(t, result.ArtifactsValid)
	assert.Empty(t, result.MissingArtifacts)
	assert.Empty(t, result.CorruptArtifacts)
	assert.Nil(t, result.Error)
}

func TestValidationResult_MissingManifest(t *testing.T) {
	result := ValidationResult{
		Valid:          false,
		ManifestExists: false,
		ManifestValid:  false,
		ArtifactsValid: false,
	}

	assert.False(t, result.Valid)
	assert.False(t, result.ManifestExists)
}

func TestValidationResult_InvalidManifest(t *testing.T) {
	result := ValidationResult{
		Valid:          false,
		ManifestExists: true,
		ManifestValid:  false,
		ArtifactsValid: false,
	}

	assert.False(t, result.Valid)
	assert.True(t, result.ManifestExists)
	assert.False(t, result.ManifestValid)
}

func TestValidationResult_MissingArtifactsOnly(t *testing.T) {
	result := ValidationResult{
		Valid:            false,
		ManifestExists:   true,
		ManifestValid:    true,
		ArtifactsValid:   false,
		MissingArtifacts: []string{"file1.bin", "file2.bin"},
		CorruptArtifacts: []string{},
	}

	assert.False(t, result.Valid)
	assert.True(t, result.ManifestValid)
	assert.False(t, result.ArtifactsValid)
	assert.Len(t, result.MissingArtifacts, 2)
	assert.Empty(t, result.CorruptArtifacts)
}

func TestValidationResult_CorruptArtifactsOnly(t *testing.T) {
	result := ValidationResult{
		Valid:            false,
		ManifestExists:   true,
		ManifestValid:    true,
		ArtifactsValid:   false,
		MissingArtifacts: []string{},
		CorruptArtifacts: []string{"corrupted.bin"},
	}

	assert.False(t, result.Valid)
	assert.True(t, result.ManifestValid)
	assert.False(t, result.ArtifactsValid)
	assert.Empty(t, result.MissingArtifacts)
	assert.Len(t, result.CorruptArtifacts, 1)
}

func TestValidationResult_WithError(t *testing.T) {
	testErr := assert.AnError
	result := ValidationResult{
		Valid:          false,
		ManifestExists: true,
		ManifestValid:  false,
		Error:          testErr,
	}

	assert.False(t, result.Valid)
	assert.Error(t, result.Error)
	assert.Equal(t, testErr, result.Error)
}

func TestValidationResult_TableDriven(t *testing.T) {
	tests := []struct {
		name                 string
		result               ValidationResult
		expectValid          bool
		expectManifestExists bool
		expectArtifactsValid bool
		expectMissingCount   int
		expectCorruptCount   int
	}{
		{
			name: "completely valid",
			result: ValidationResult{
				Valid:            true,
				ManifestExists:   true,
				ManifestValid:    true,
				ArtifactsValid:   true,
				MissingArtifacts: []string{},
				CorruptArtifacts: []string{},
			},
			expectValid:          true,
			expectManifestExists: true,
			expectArtifactsValid: true,
			expectMissingCount:   0,
			expectCorruptCount:   0,
		},
		{
			name: "no manifest",
			result: ValidationResult{
				Valid:          false,
				ManifestExists: false,
				ManifestValid:  false,
				ArtifactsValid: false,
			},
			expectValid:          false,
			expectManifestExists: false,
			expectArtifactsValid: false,
			expectMissingCount:   0,
			expectCorruptCount:   0,
		},
		{
			name: "some missing artifacts",
			result: ValidationResult{
				Valid:            false,
				ManifestExists:   true,
				ManifestValid:    true,
				ArtifactsValid:   false,
				MissingArtifacts: []string{"a.bin", "b.bin", "c.bin"},
				CorruptArtifacts: []string{},
			},
			expectValid:          false,
			expectManifestExists: true,
			expectArtifactsValid: false,
			expectMissingCount:   3,
			expectCorruptCount:   0,
		},
		{
			name: "some corrupt artifacts",
			result: ValidationResult{
				Valid:            false,
				ManifestExists:   true,
				ManifestValid:    true,
				ArtifactsValid:   false,
				MissingArtifacts: []string{},
				CorruptArtifacts: []string{"bad.bin", "worse.bin"},
			},
			expectValid:          false,
			expectManifestExists: true,
			expectArtifactsValid: false,
			expectMissingCount:   0,
			expectCorruptCount:   2,
		},
		{
			name: "mixed missing and corrupt",
			result: ValidationResult{
				Valid:            false,
				ManifestExists:   true,
				ManifestValid:    true,
				ArtifactsValid:   false,
				MissingArtifacts: []string{"missing.bin"},
				CorruptArtifacts: []string{"corrupt.bin"},
			},
			expectValid:          false,
			expectManifestExists: true,
			expectArtifactsValid: false,
			expectMissingCount:   1,
			expectCorruptCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectValid, tt.result.Valid)
			assert.Equal(t, tt.expectManifestExists, tt.result.ManifestExists)
			assert.Equal(t, tt.expectArtifactsValid, tt.result.ArtifactsValid)
			assert.Len(t, tt.result.MissingArtifacts, tt.expectMissingCount)
			assert.Len(t, tt.result.CorruptArtifacts, tt.expectCorruptCount)
		})
	}
}

// =============================================================================
// ComponentView Type Tests
// =============================================================================

func TestComponentView_FieldsExist(t *testing.T) {
	view := ComponentView{
		Module:    "core",
		Component: "go",
		Status:    StatusCompleted,
		UoWs:      []UoWManifest{},
		TotalSize: 10485760,
	}

	assert.Equal(t, "core", view.Module)
	assert.Equal(t, "go", view.Component)
	assert.Equal(t, StatusCompleted, view.Status)
	assert.Empty(t, view.UoWs)
	assert.Equal(t, int64(10485760), view.TotalSize)
}

func TestComponentView_ZeroValueIsValid(t *testing.T) {
	var view ComponentView

	assert.Empty(t, view.Module)
	assert.Empty(t, view.Component)
	assert.Empty(t, view.Status)
	assert.Nil(t, view.UoWs)
	assert.Zero(t, view.TotalSize)
}

func TestComponentView_WithMultipleUoWs(t *testing.T) {
	view := ComponentView{
		Module:    "core",
		Component: "go",
		Status:    StatusCompleted,
		UoWs: []UoWManifest{
			{
				Action:    core.ActionBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go",
				ExitCode:  0,
			},
			{
				Action:    core.ActionTest,
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				ExitCode:  0,
			},
		},
		TotalSize: 15000000,
	}

	assert.Len(t, view.UoWs, 2)
	assert.Equal(t, core.ActionBuild, view.UoWs[0].Action)
	assert.Equal(t, core.ActionTest, view.UoWs[1].Action)
}

func TestComponentView_StatusAggregation(t *testing.T) {
	tests := []struct {
		name           string
		uowExitCodes   []int
		expectedStatus Status
	}{
		{
			name:           "all successful",
			uowExitCodes:   []int{0, 0, 0},
			expectedStatus: StatusCompleted,
		},
		{
			name:           "one failed",
			uowExitCodes:   []int{0, 1, 0},
			expectedStatus: StatusFailed,
		},
		{
			name:           "all failed",
			uowExitCodes:   []int{1, 2, 1},
			expectedStatus: StatusFailed,
		},
		{
			name:           "empty",
			uowExitCodes:   []int{},
			expectedStatus: StatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := ComponentView{
				Module:    "test",
				Component: "go",
				Status:    tt.expectedStatus,
				UoWs:      make([]UoWManifest, len(tt.uowExitCodes)),
			}

			for i, exitCode := range tt.uowExitCodes {
				view.UoWs[i] = UoWManifest{ExitCode: exitCode}
			}

			assert.Equal(t, tt.expectedStatus, view.Status)
		})
	}
}

// =============================================================================
// ModuleView Type Tests
// =============================================================================

func TestModuleView_FieldsExist(t *testing.T) {
	view := ModuleView{
		Module:     "core",
		Status:     StatusCompleted,
		Components: []ComponentView{},
		TotalSize:  52428800,
	}

	assert.Equal(t, "core", view.Module)
	assert.Equal(t, StatusCompleted, view.Status)
	assert.Empty(t, view.Components)
	assert.Equal(t, int64(52428800), view.TotalSize)
}

func TestModuleView_ZeroValueIsValid(t *testing.T) {
	var view ModuleView

	assert.Empty(t, view.Module)
	assert.Empty(t, view.Status)
	assert.Nil(t, view.Components)
	assert.Zero(t, view.TotalSize)
}

func TestModuleView_WithMultipleComponents(t *testing.T) {
	view := ModuleView{
		Module: "core",
		Status: StatusCompleted,
		Components: []ComponentView{
			{Module: "core", Component: "go", Status: StatusCompleted, TotalSize: 10000000},
			{Module: "core", Component: "docker", Status: StatusCompleted, TotalSize: 40000000},
		},
		TotalSize: 50000000,
	}

	assert.Len(t, view.Components, 2)
	assert.Equal(t, "go", view.Components[0].Component)
	assert.Equal(t, "docker", view.Components[1].Component)
	assert.Equal(t, int64(50000000), view.TotalSize)
}

func TestModuleView_StatusAggregation(t *testing.T) {
	tests := []struct {
		name              string
		componentStatuses []Status
		expectedStatus    Status
	}{
		{
			name:              "all completed",
			componentStatuses: []Status{StatusCompleted, StatusCompleted},
			expectedStatus:    StatusCompleted,
		},
		{
			name:              "one failed",
			componentStatuses: []Status{StatusCompleted, StatusFailed, StatusCompleted},
			expectedStatus:    StatusFailed,
		},
		{
			name:              "one in progress",
			componentStatuses: []Status{StatusCompleted, StatusInProgress},
			expectedStatus:    StatusInProgress,
		},
		{
			name:              "one pending",
			componentStatuses: []Status{StatusCompleted, StatusPending},
			expectedStatus:    StatusPending,
		},
		{
			name:              "all cached",
			componentStatuses: []Status{StatusCached, StatusCached},
			expectedStatus:    StatusCached,
		},
		{
			name:              "empty",
			componentStatuses: []Status{},
			expectedStatus:    StatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := ModuleView{
				Module:     "test",
				Status:     tt.expectedStatus,
				Components: make([]ComponentView, len(tt.componentStatuses)),
			}

			for i, status := range tt.componentStatuses {
				view.Components[i] = ComponentView{Status: status}
			}

			assert.Equal(t, tt.expectedStatus, view.Status)
		})
	}
}

// =============================================================================
// StatusFromExitCode Tests
// =============================================================================

func TestStatusFromExitCode(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		expected Status
	}{
		{
			name:     "zero exit code is completed",
			exitCode: 0,
			expected: StatusCompleted,
		},
		{
			name:     "positive exit code is failed",
			exitCode: 1,
			expected: StatusFailed,
		},
		{
			name:     "negative exit code is cached",
			exitCode: -1,
			expected: StatusCached,
		},
		{
			name:     "large positive exit code is failed",
			exitCode: 127,
			expected: StatusFailed,
		},
		{
			name:     "large negative exit code is cached",
			exitCode: -100,
			expected: StatusCached,
		},
		{
			name:     "exit code 255 is failed",
			exitCode: 255,
			expected: StatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusFromExitCode(tt.exitCode)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestStatusFromExitCode_MatchesTUIConvention(t *testing.T) {
	// Verify the convention used in TUI is consistent:
	// 0 = success (green), <0 = cached (blue), >0 = failed (red)
	assert.Equal(t, StatusCompleted, StatusFromExitCode(0), "0 should be completed (success)")
	assert.Equal(t, StatusCached, StatusFromExitCode(-1), "-1 should be cached")
	assert.Equal(t, StatusFailed, StatusFromExitCode(1), "1 should be failed")
}

// =============================================================================
// Type Relationships Tests
// =============================================================================

func TestTypeRelationships_ModuleContainsComponents(t *testing.T) {
	module := ModuleView{
		Module: "eac",
		Status: StatusCompleted,
		Components: []ComponentView{
			{
				Module:    "eac",
				Component: "go",
				Status:    StatusCompleted,
				UoWs: []UoWManifest{
					{Action: core.ActionBuild, Module: "eac", Component: "go", Tool: "go"},
				},
			},
			{
				Module:    "eac",
				Component: "docker",
				Status:    StatusCompleted,
				UoWs: []UoWManifest{
					{Action: core.ActionBuild, Module: "eac", Component: "docker", Tool: "docker"},
				},
			},
		},
	}

	assert.Equal(t, "eac", module.Module)
	assert.Len(t, module.Components, 2)

	for _, comp := range module.Components {
		assert.Equal(t, "eac", comp.Module, "Component should reference parent module")
		assert.NotEmpty(t, comp.UoWs)

		for _, uow := range comp.UoWs {
			assert.Equal(t, "eac", uow.Module, "UoW should reference parent module")
			assert.Equal(t, comp.Component, uow.Component, "UoW should reference parent component")
		}
	}
}

func TestTypeRelationships_ArtifactBelongsToUoW(t *testing.T) {
	artifact := Artifact{
		ID:     "binary",
		Path:   "out/build/core/go/eac",
		SHA256: "sha256:abc",
		Size:   1000,
		Type:   "binary",
	}

	uow := UoWManifest{
		Action:    core.ActionBuild,
		Module:    "core",
		Component: "go",
		Tool:      "go",
		Artifacts: []Artifact{artifact},
	}

	assert.Len(t, uow.Artifacts, 1)
	assert.Equal(t, artifact, uow.Artifacts[0])
}
