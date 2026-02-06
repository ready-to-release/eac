package testview

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTestUoWManifest creates a UoW manifest in the expected directory structure.
func writeTestUoWManifest(t *testing.T, workspaceRoot, module string, manifest *coreoutput.UoWManifest) string {
	t.Helper()
	dir := filepath.Join(workspaceRoot, "out", "test", module, manifest.DirName())
	require.NoError(t, os.MkdirAll(dir, 0755))

	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "uow.manifest.json"), data, 0644))

	return dir
}

func TestLoadModuleTestView_NoManifests(t *testing.T) {
	ws := t.TempDir()
	view, err := LoadModuleTestView(ws, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, view)
}

func TestLoadModuleTestView_SingleUoW(t *testing.T) {
	ws := t.TempDir()

	manifest := &coreoutput.UoWManifest{
		Context:    workunit.ContextTest,
		Module:     "core",
		Component:  "go",
		Tool:       "gotest",
		Extra:      map[string]string{"testname": "unit"},
		ExitCode:   0,
		InputHash:  "sha256:abc",
		ExecutedAt: time.Now().Add(-5 * time.Minute),
		Duration:   30 * time.Second,
		Artifacts:  []coreoutput.Artifact{},
	}

	writeTestUoWManifest(t, ws, "core", manifest)

	view, err := LoadModuleTestView(ws, "core")
	require.NoError(t, err)
	require.NotNil(t, view)

	assert.Equal(t, "core", view.Module)
	assert.Equal(t, 1, view.UoWCount)
	assert.Equal(t, 0, view.ExitCode)
}

func TestLoadModuleTestView_WithCTRFArtifact(t *testing.T) {
	ws := t.TempDir()

	// Create CTRF report file first
	uowDir := filepath.Join(ws, "out", "test", "core", "go-gotest-unit")
	require.NoError(t, os.MkdirAll(uowDir, 0755))

	ctrfContent := `{
		"results": {
			"tests": [
				{"name": "TestFoo", "status": "passed", "duration": 100},
				{"name": "TestBar", "status": "failed", "duration": 200}
			]
		}
	}`
	ctrfPath := filepath.Join(uowDir, "unit.json")
	require.NoError(t, os.WriteFile(ctrfPath, []byte(ctrfContent), 0644))

	// Hash the file for the artifact
	size, hash, err := coreoutput.HashFile(ctrfPath)
	require.NoError(t, err)

	manifest := &coreoutput.UoWManifest{
		Context:    workunit.ContextTest,
		Module:     "core",
		Component:  "go",
		Tool:       "gotest",
		Extra:      map[string]string{"testname": "unit"},
		ExitCode:   0,
		ExecutedAt: time.Now(),
		Duration:   10 * time.Second,
		Artifacts: []coreoutput.Artifact{
			{ID: "ctrf-report", Path: "unit.json", SHA256: hash, Size: size, Type: "ctrf-report"},
		},
	}

	// Write manifest (which goes into the same dir)
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(uowDir, "uow.manifest.json"), data, 0644))

	view, err := LoadModuleTestView(ws, "core")
	require.NoError(t, err)
	require.NotNil(t, view)

	assert.Equal(t, 2, view.Summary.Total)
	assert.Equal(t, 1, view.Summary.Passed)
	assert.Equal(t, 1, view.Summary.Failed)
	assert.Len(t, view.CTRFReports, 1)
}

func TestLoadModuleTestView_WithCucumberArtifact(t *testing.T) {
	ws := t.TempDir()

	uowDir := filepath.Join(ws, "out", "test", "mymod", "go-godog-impl")
	require.NoError(t, os.MkdirAll(uowDir, 0755))

	cucumberContent := `[
		{
			"uri": "specs/mymod/login/specification.feature",
			"elements": [
				{
					"name": "User logs in",
					"type": "scenario",
					"tags": [{"name": "@L0"}, {"name": "@control:auth-1"}],
					"steps": [
						{"result": {"status": "passed", "duration": 1000000}}
					]
				}
			]
		}
	]`
	cucPath := filepath.Join(uowDir, "cucumber.json")
	require.NoError(t, os.WriteFile(cucPath, []byte(cucumberContent), 0644))

	size, hash, err := coreoutput.HashFile(cucPath)
	require.NoError(t, err)

	manifest := &coreoutput.UoWManifest{
		Context:    workunit.ContextTest,
		Module:     "mymod",
		Component:  "go",
		Tool:       "godog",
		Extra:      map[string]string{"testname": "impl"},
		ExitCode:   0,
		ExecutedAt: time.Now(),
		Duration:   5 * time.Second,
		Artifacts: []coreoutput.Artifact{
			{ID: "cucumber-report", Path: "cucumber.json", SHA256: hash, Size: size, Type: "cucumber-report"},
		},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(uowDir, "uow.manifest.json"), data, 0644))

	view, err := LoadModuleTestView(ws, "mymod")
	require.NoError(t, err)
	require.NotNil(t, view)

	assert.Equal(t, 1, view.Summary.Total)
	assert.Equal(t, 1, view.Summary.Passed)
	assert.Len(t, view.CucumberReports, 1)

	// Check test entry details
	require.Len(t, view.Tests, 1)
	assert.Equal(t, "User logs in", view.Tests[0].Name)
	assert.Equal(t, "godog", view.Tests[0].Type)
	assert.Equal(t, "unit", view.Tests[0].Suite)
	assert.Contains(t, view.Tests[0].Tags, "@control:auth-1")
}

func TestLoadAllTestViews_MultipleModules(t *testing.T) {
	ws := t.TempDir()

	for _, mod := range []string{"alpha", "beta"} {
		manifest := &coreoutput.UoWManifest{
			Context:    workunit.ContextTest,
			Module:     mod,
			Component:  "go",
			Tool:       "gotest",
			Extra:      map[string]string{"testname": "unit"},
			ExitCode:   0,
			ExecutedAt: time.Now(),
			Duration:   5 * time.Second,
		}
		writeTestUoWManifest(t, ws, mod, manifest)
	}

	views, err := LoadAllTestViews(ws)
	require.NoError(t, err)
	assert.Len(t, views, 2)
	// Should be sorted by module name
	assert.Equal(t, "alpha", views[0].Module)
	assert.Equal(t, "beta", views[1].Module)
}

func TestLoadAllTestViews_EmptyTestDir(t *testing.T) {
	ws := t.TempDir()
	views, err := LoadAllTestViews(ws)
	assert.NoError(t, err)
	assert.Nil(t, views)
}

func TestLoadModuleTestView_FailedExitCode(t *testing.T) {
	ws := t.TempDir()

	manifest := &coreoutput.UoWManifest{
		Context:    workunit.ContextTest,
		Module:     "core",
		Component:  "go",
		Tool:       "gotest",
		Extra:      map[string]string{"testname": "unit"},
		ExitCode:   1,
		ExecutedAt: time.Now(),
		Duration:   10 * time.Second,
	}
	writeTestUoWManifest(t, ws, "core", manifest)

	view, err := LoadModuleTestView(ws, "core")
	require.NoError(t, err)
	require.NotNil(t, view)
	assert.Equal(t, 1, view.ExitCode)
}
