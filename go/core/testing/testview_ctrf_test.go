package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCTRFFile_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "unit.json")

	content := `{
		"results": {
			"tests": [
				{
					"name": "TestFoo",
					"status": "passed",
					"duration": 150,
					"suite": "unit",
					"filePath": "pkg/foo/foo_test.go"
				},
				{
					"name": "TestBar",
					"status": "failed",
					"duration": 200,
					"suite": "unit",
					"filePath": "pkg/bar/bar_test.go"
				},
				{
					"name": "TestBaz",
					"status": "skipped",
					"duration": 0
				}
			]
		}
	}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	entries := parseCTRFFile(path, "mymodule")

	assert.Len(t, entries, 3)

	assert.Equal(t, "TestFoo", entries[0].Name)
	assert.Equal(t, "mymodule", entries[0].Module)
	assert.Equal(t, "gotest", entries[0].Type)
	assert.Equal(t, StatusPassed, entries[0].Status)
	assert.Equal(t, int64(150), entries[0].DurationMs)
	assert.Equal(t, "unit", entries[0].Suite)
	assert.Equal(t, "pkg/foo/foo_test.go", entries[0].FilePath)

	assert.Equal(t, "TestBar", entries[1].Name)
	assert.Equal(t, StatusFailed, entries[1].Status)

	assert.Equal(t, "TestBaz", entries[2].Name)
	assert.Equal(t, StatusSkipped, entries[2].Status)
	assert.Equal(t, "unit", entries[2].Suite) // default when empty
}

func TestParseCTRFFile_NonexistentFile(t *testing.T) {
	entries := parseCTRFFile("/nonexistent/unit.json", "mod")
	assert.Nil(t, entries)
}

func TestParseCTRFFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "unit.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

	entries := parseCTRFFile(path, "mod")
	assert.Nil(t, entries)
}

func TestParseCTRFFile_EmptyResults(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "unit.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"results":{"tests":[]}}`), 0644))

	entries := parseCTRFFile(path, "mod")
	assert.Empty(t, entries)
}

func TestNormalizeCTRFStatus(t *testing.T) {
	assert.Equal(t, StatusPassed, normalizeCTRFStatus("passed"))
	assert.Equal(t, StatusFailed, normalizeCTRFStatus("failed"))
	assert.Equal(t, StatusSkipped, normalizeCTRFStatus("skipped"))
	assert.Equal(t, StatusSkipped, normalizeCTRFStatus("pending"))
	assert.Equal(t, "other", normalizeCTRFStatus("other"))
}
