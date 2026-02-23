package dependabot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile is a test helper that creates a file and any necessary parent directories.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// --- discoverGoModules tests ---

func TestDiscoverGoModules(t *testing.T) {
	t.Run("standard go.work with multiple modules", func(t *testing.T) {
		root := t.TempDir()
		goWork := `go 1.24.4

use (
	./go/core
	./go/cli/clie
	./go/adapters/ai
)
`
		writeFile(t, filepath.Join(root, "go.work"), goWork)

		entries, err := discoverGoModules(root)
		require.NoError(t, err)
		require.Len(t, entries, 3)

		assert.Equal(t, EcosystemEntry{Ecosystem: "gomod", Directory: "/go/core"}, entries[0])
		assert.Equal(t, EcosystemEntry{Ecosystem: "gomod", Directory: "/go/cli/clie"}, entries[1])
		assert.Equal(t, EcosystemEntry{Ecosystem: "gomod", Directory: "/go/adapters/ai"}, entries[2])
	})

	t.Run("go.work does not exist returns nil nil", func(t *testing.T) {
		root := t.TempDir()

		entries, err := discoverGoModules(root)
		assert.NoError(t, err)
		assert.Nil(t, entries)
	})

	t.Run("commented-out lines in go.work are skipped", func(t *testing.T) {
		root := t.TempDir()
		goWork := `go 1.24.4

use (
	./go/core
	// ./go/deprecated/old
	./go/cli/clie
	// ./go/experimental
)
`
		writeFile(t, filepath.Join(root, "go.work"), goWork)

		entries, err := discoverGoModules(root)
		require.NoError(t, err)
		require.Len(t, entries, 2)

		assert.Equal(t, "gomod", entries[0].Ecosystem)
		assert.Equal(t, "/go/core", entries[0].Directory)
		assert.Equal(t, "gomod", entries[1].Ecosystem)
		assert.Equal(t, "/go/cli/clie", entries[1].Directory)
	})

	t.Run("empty use block returns no entries", func(t *testing.T) {
		root := t.TempDir()
		goWork := `go 1.24.4

use (
)
`
		writeFile(t, filepath.Join(root, "go.work"), goWork)

		entries, err := discoverGoModules(root)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("modules without dot-slash prefix", func(t *testing.T) {
		root := t.TempDir()
		goWork := `go 1.24.4

use (
	go/core
)
`
		writeFile(t, filepath.Join(root, "go.work"), goWork)

		entries, err := discoverGoModules(root)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "/go/core", entries[0].Directory)
	})
}

// --- discoverByFile tests ---

func TestDiscoverByFile(t *testing.T) {
	t.Run("finds package.json in nested directories", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "web", "app", "package.json"), `{}`)
		writeFile(t, filepath.Join(root, "tools", "cli", "package.json"), `{}`)

		entries, err := discoverByFile(root, "package.json", "npm", nil)
		require.NoError(t, err)
		require.Len(t, entries, 2)

		dirs := make(map[string]bool)
		for _, e := range entries {
			assert.Equal(t, "npm", e.Ecosystem)
			dirs[e.Directory] = true
		}
		assert.True(t, dirs["/web/app"], "should find /web/app")
		assert.True(t, dirs["/tools/cli"], "should find /tools/cli")
	})

	t.Run("excludes node_modules directories", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "web", "package.json"), `{}`)
		writeFile(t, filepath.Join(root, "web", "node_modules", "dep", "package.json"), `{}`)

		entries, err := discoverByFile(root, "package.json", "npm", []string{"node_modules"})
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "/web", entries[0].Directory)
	})

	t.Run("finds Dockerfile in container directories", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "containers", "api", "Dockerfile"), "FROM golang:1.24")
		writeFile(t, filepath.Join(root, "containers", "web", "Dockerfile"), "FROM node:20")

		entries, err := discoverByFile(root, "Dockerfile", "docker", nil)
		require.NoError(t, err)
		require.Len(t, entries, 2)

		dirs := make(map[string]bool)
		for _, e := range entries {
			assert.Equal(t, "docker", e.Ecosystem)
			dirs[e.Directory] = true
		}
		assert.True(t, dirs["/containers/api"], "should find /containers/api")
		assert.True(t, dirs["/containers/web"], "should find /containers/web")
	})

	t.Run("file in repo root gets / directory", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "package.json"), `{}`)

		entries, err := discoverByFile(root, "package.json", "npm", nil)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "/", entries[0].Directory)
	})

	t.Run("excludes multiple directories", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "src", "package.json"), `{}`)
		writeFile(t, filepath.Join(root, "out", "package.json"), `{}`)
		writeFile(t, filepath.Join(root, "vendor", "package.json"), `{}`)
		writeFile(t, filepath.Join(root, ".cache", "package.json"), `{}`)

		entries, err := discoverByFile(root, "package.json", "npm", []string{"out", "vendor", ".cache"})
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "/src", entries[0].Directory)
	})

	t.Run("empty directory returns no entries", func(t *testing.T) {
		root := t.TempDir()

		entries, err := discoverByFile(root, "package.json", "npm", nil)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}

// --- discoverByGlob tests ---

func TestDiscoverByGlob(t *testing.T) {
	t.Run("deduplicates multiple requirements files in same directory", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "scripts", "requirements.txt"), "flask==2.0")
		writeFile(t, filepath.Join(root, "scripts", "requirements-dev.txt"), "pytest==7.0")

		entries, err := discoverByGlob(root, "requirements*.txt", "pip", nil)
		require.NoError(t, err)
		require.Len(t, entries, 1, "should deduplicate to one entry per directory")
		assert.Equal(t, "pip", entries[0].Ecosystem)
		assert.Equal(t, "/scripts", entries[0].Directory)
	})

	t.Run("finds requirements files in multiple directories", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "services", "api", "requirements.txt"), "flask==2.0")
		writeFile(t, filepath.Join(root, "tools", "requirements.txt"), "click==8.0")
		writeFile(t, filepath.Join(root, "tools", "requirements-dev.txt"), "pytest==7.0")

		entries, err := discoverByGlob(root, "requirements*.txt", "pip", nil)
		require.NoError(t, err)
		require.Len(t, entries, 2)

		dirs := make(map[string]bool)
		for _, e := range entries {
			assert.Equal(t, "pip", e.Ecosystem)
			dirs[e.Directory] = true
		}
		assert.True(t, dirs["/services/api"], "should find /services/api")
		assert.True(t, dirs["/tools"], "should find /tools")
	})

	t.Run("excludes directories", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "src", "requirements.txt"), "flask==2.0")
		writeFile(t, filepath.Join(root, "vendor", "requirements.txt"), "vendored==1.0")

		entries, err := discoverByGlob(root, "requirements*.txt", "pip", []string{"vendor"})
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "/src", entries[0].Directory)
	})

	t.Run("empty directory returns no entries", func(t *testing.T) {
		root := t.TempDir()

		entries, err := discoverByGlob(root, "requirements*.txt", "pip", nil)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("non-matching files are ignored", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "src", "readme.txt"), "not a requirements file")
		writeFile(t, filepath.Join(root, "src", "setup.py"), "from setuptools import setup")

		entries, err := discoverByGlob(root, "requirements*.txt", "pip", nil)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}

// --- DiscoverAll tests ---

func TestDiscoverAll(t *testing.T) {
	t.Run("fully populated repository", func(t *testing.T) {
		root := t.TempDir()

		// go.work with modules
		goWork := `go 1.24.4

use (
	./go/core
	./go/cli/clie
)
`
		writeFile(t, filepath.Join(root, "go.work"), goWork)

		// package.json
		writeFile(t, filepath.Join(root, "web", "package.json"), `{}`)

		// Dockerfile
		writeFile(t, filepath.Join(root, "containers", "api", "Dockerfile"), "FROM golang:1.24")

		// requirements.txt
		writeFile(t, filepath.Join(root, "scripts", "requirements.txt"), "flask==2.0")

		// .github/workflows for github-actions
		writeFile(t, filepath.Join(root, ".github", "workflows", "ci.yml"), "name: CI")

		opts := DefaultDiscoverOptions()
		entries, err := DiscoverAll(root, opts)
		require.NoError(t, err)

		// Build a map of ecosystem:directory for easy assertions
		keySet := make(map[string]bool)
		for _, e := range entries {
			keySet[e.Key()] = true
		}

		assert.True(t, keySet["github-actions:/"], "should discover github-actions")
		assert.True(t, keySet["gomod:/go/core"], "should discover gomod /go/core")
		assert.True(t, keySet["gomod:/go/cli/clie"], "should discover gomod /go/cli/clie")
		assert.True(t, keySet["npm:/web"], "should discover npm /web")
		assert.True(t, keySet["docker:/containers/api"], "should discover docker /containers/api")
		assert.True(t, keySet["pip:/scripts"], "should discover pip /scripts")
		assert.Len(t, entries, 6, "should discover exactly 6 entries")
	})

	t.Run("empty directory discovers nothing", func(t *testing.T) {
		root := t.TempDir()

		opts := DefaultDiscoverOptions()
		entries, err := DiscoverAll(root, opts)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("selective ecosystems via options", func(t *testing.T) {
		root := t.TempDir()

		goWork := `go 1.24.4

use (
	./go/core
)
`
		writeFile(t, filepath.Join(root, "go.work"), goWork)
		writeFile(t, filepath.Join(root, "web", "package.json"), `{}`)
		writeFile(t, filepath.Join(root, ".github", "workflows", "ci.yml"), "name: CI")

		// Only scan for gomod, disable everything else
		opts := DiscoverOptions{
			IncludeGomod:         true,
			IncludeNpm:           false,
			IncludePip:           false,
			IncludeDocker:        false,
			IncludeGithubActions: false,
		}

		entries, err := DiscoverAll(root, opts)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "gomod", entries[0].Ecosystem)
		assert.Equal(t, "/go/core", entries[0].Directory)
	})

	t.Run("github-actions not discovered without workflows directory", func(t *testing.T) {
		root := t.TempDir()

		// Create .github but not .github/workflows
		require.NoError(t, os.MkdirAll(filepath.Join(root, ".github"), 0o755))

		opts := DiscoverOptions{
			IncludeGithubActions: true,
		}

		entries, err := DiscoverAll(root, opts)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("node_modules excluded by default options", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "web", "package.json"), `{}`)
		writeFile(t, filepath.Join(root, "web", "node_modules", "dep", "package.json"), `{}`)

		opts := DefaultDiscoverOptions()
		// Disable other ecosystems for a focused test
		opts.IncludeGomod = false
		opts.IncludePip = false
		opts.IncludeDocker = false
		opts.IncludeGithubActions = false

		entries, err := DiscoverAll(root, opts)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "/web", entries[0].Directory)
	})
}

// --- DefaultDiscoverOptions tests ---

func TestDefaultDiscoverOptions(t *testing.T) {
	opts := DefaultDiscoverOptions()

	assert.True(t, opts.IncludeGomod)
	assert.True(t, opts.IncludeNpm)
	assert.True(t, opts.IncludePip)
	assert.True(t, opts.IncludeDocker)
	assert.True(t, opts.IncludeGithubActions)
	assert.Contains(t, opts.ExcludeDirs, "node_modules")
	assert.Contains(t, opts.ExcludeDirs, "out")
	assert.Contains(t, opts.ExcludeDirs, ".cache")
	assert.Contains(t, opts.ExcludeDirs, "vendor")
	assert.Contains(t, opts.ExcludeDirs, ".git")
	assert.Contains(t, opts.ExcludeDirs, "templates")
	assert.Contains(t, opts.ExcludeDirs, ".vscode")
}
