package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectTestArtifacts_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	artifacts := collectTestArtifacts(tmpDir)
	assert.Empty(t, artifacts)
}

func TestCollectTestArtifacts_NonexistentDir(t *testing.T) {
	artifacts := collectTestArtifacts(filepath.Join(t.TempDir(), "nonexistent"))
	assert.Empty(t, artifacts)
}

func TestCollectTestArtifacts_KnownFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create known test output files (test.log excluded — owned by orchestrator)
	files := map[string]string{
		"cucumber.json": `[{"uri":"test.feature"}]`,
		"unit.json":     `{"results":{"tests":[]}}`,
		"coverage.out":  "mode: set",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644))
	}

	artifacts := collectTestArtifacts(tmpDir)
	assert.Len(t, artifacts, 3)

	// Verify artifact types
	typeMap := make(map[string]string)
	for _, a := range artifacts {
		typeMap[a.Path] = a.Type
	}
	assert.Equal(t, "cucumber-report", typeMap["cucumber.json"])
	assert.Equal(t, "ctrf-report", typeMap["unit.json"])
	assert.Equal(t, "coverage", typeMap["coverage.out"])

	// All artifacts should have hashes and sizes
	for _, a := range artifacts {
		assert.NotEmpty(t, a.SHA256, "artifact %s should have hash", a.Path)
		assert.True(t, a.Size > 0, "artifact %s should have size", a.Path)
		assert.NotEmpty(t, a.ID, "artifact %s should have ID", a.Path)
	}
}

func TestCollectTestArtifacts_IgnoresTestLog(t *testing.T) {
	tmpDir := t.TempDir()

	// test.log should NOT be collected as a tracked artifact
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.log"), []byte("log content"), 0644))

	artifacts := collectTestArtifacts(tmpDir)
	assert.Empty(t, artifacts)
}

func TestCollectTestArtifacts_SkipsManifest(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "uow.manifest.json"), []byte("{}"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "unit.json"), []byte(`{"results":{}}`), 0644))

	artifacts := collectTestArtifacts(tmpDir)
	assert.Len(t, artifacts, 1)
	assert.Equal(t, "unit.json", artifacts[0].Path)
}

func TestCollectTestArtifacts_SkipsUnknownFiles(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "random.txt"), []byte("data"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "unit.json"), []byte(`{"results":{}}`), 0644))

	artifacts := collectTestArtifacts(tmpDir)
	assert.Len(t, artifacts, 1)
	assert.Equal(t, "unit.json", artifacts[0].Path)
}

func TestCollectTestArtifacts_CucumberJsonSuffix(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "godog.cucumber.json"), []byte("[]"), 0644))

	artifacts := collectTestArtifacts(tmpDir)
	assert.Len(t, artifacts, 1)
	assert.Equal(t, "cucumber-report", artifacts[0].Type)
	assert.Equal(t, "cucumber-report", artifacts[0].ID)
}

func TestCollectTestArtifacts_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "coverage.out"), []byte("mode: set"), 0644))

	artifacts := collectTestArtifacts(tmpDir)
	assert.Len(t, artifacts, 1)
}
