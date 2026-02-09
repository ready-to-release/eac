package hash

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModuleInputHashProvider(t *testing.T) {
	t.Run("returns hash for known module", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create test files
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "util.go"), []byte("package main"), 0644))

		moduleFiles := map[string][]string{
			"core": {"main.go", "util.go"},
		}

		provider := NewModuleInputHashProvider(tmpDir, moduleFiles)

		hash, err := provider(workunit.UnitID{Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"})
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA256 hex
	})

	t.Run("returns empty string for unknown module", func(t *testing.T) {
		tmpDir := t.TempDir()
		moduleFiles := map[string][]string{
			"core": {"main.go"},
		}

		provider := NewModuleInputHashProvider(tmpDir, moduleFiles)

		hash, err := provider(workunit.UnitID{Module: "unknown", ComponentType: "go", ComponentName: "go", Tool: "go"})
		require.NoError(t, err)
		assert.Empty(t, hash, "unknown module should return empty string")
	})

	t.Run("returns same hash for same module different components", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644))

		moduleFiles := map[string][]string{
			"core": {"main.go"},
		}

		provider := NewModuleInputHashProvider(tmpDir, moduleFiles)

		// Different components in same module should get same hash (based on module files)
		hash1, err := provider(workunit.UnitID{Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"})
		require.NoError(t, err)

		hash2, err := provider(workunit.UnitID{Module: "core", ComponentType: "ts", ComponentName: "ts", Tool: "tsc"})
		require.NoError(t, err)

		assert.Equal(t, hash1, hash2, "same module should have same input hash regardless of component")
	})

	t.Run("returns different hash for different modules", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "core.go"), []byte("package core"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "cli.go"), []byte("package cli"), 0644))

		moduleFiles := map[string][]string{
			"core": {"core.go"},
			"cli":  {"cli.go"},
		}

		provider := NewModuleInputHashProvider(tmpDir, moduleFiles)

		hash1, err := provider(workunit.UnitID{Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"})
		require.NoError(t, err)

		hash2, err := provider(workunit.UnitID{Module: "cli", ComponentType: "go", ComponentName: "go", Tool: "go"})
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2, "different modules should have different hashes")
	})

	t.Run("handles empty file list", func(t *testing.T) {
		tmpDir := t.TempDir()
		moduleFiles := map[string][]string{
			"core": {}, // Empty file list
		}

		provider := NewModuleInputHashProvider(tmpDir, moduleFiles)

		hash, err := provider(workunit.UnitID{Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"})
		require.NoError(t, err)
		assert.Empty(t, hash, "empty file list should return empty hash")
	})

	t.Run("returns error for missing files", func(t *testing.T) {
		tmpDir := t.TempDir()
		moduleFiles := map[string][]string{
			"core": {"nonexistent.go"},
		}

		provider := NewModuleInputHashProvider(tmpDir, moduleFiles)

		_, err := provider(workunit.UnitID{Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"})
		assert.Error(t, err, "missing files should return error")
	})
}
