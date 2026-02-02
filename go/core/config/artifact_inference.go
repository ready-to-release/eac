package config

import (
	"os"
	"path/filepath"
	"sort"
)

// DefaultArtifactPattern is the fallback pattern when none is configured.
const DefaultArtifactPattern = "{moniker}{ext}"

// mainLocation represents a detected main.go location.
type mainLocation struct {
	name string // artifact name (e.g., "r2r" for single binary, "server" for cmd/server/main.go)
	path string // relative path to main.go from component root
}

// detectMainPackages scans absRoot for main.go files and returns their locations.
// Priority: root/main.go > cmd/main.go > cmd/*/main.go
// Results for cmd/*/main.go are sorted alphabetically by name.
func detectMainPackages(absRoot, moniker string) []mainLocation {
	// Priority 1: root/main.go
	if _, err := os.Stat(filepath.Join(absRoot, "main.go")); err == nil {
		return []mainLocation{{name: moniker, path: "main.go"}}
	}

	// Priority 2: cmd/main.go
	cmdDir := filepath.Join(absRoot, "cmd")
	if _, err := os.Stat(filepath.Join(cmdDir, "main.go")); err == nil {
		return []mainLocation{{name: moniker, path: "cmd/main.go"}}
	}

	// Priority 3: cmd/*/main.go (multi-binary)
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil
	}

	var results []mainLocation
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		mainPath := filepath.Join(cmdDir, entry.Name(), "main.go")
		if _, err := os.Stat(mainPath); err == nil {
			// Use forward slashes for consistency across platforms
			// Name includes moniker prefix for multi-binary (e.g., "myapp-server")
			results = append(results, mainLocation{
				name: moniker + "-" + entry.Name(),
				path: "cmd/" + entry.Name() + "/main.go",
			})
		}
	}

	// Sort alphabetically by name for deterministic output
	sort.Slice(results, func(i, j int) bool {
		return results[i].name < results[j].name
	})

	return results
}

// InferGoArtifacts infers artifact configuration for Go modules based on main.go detection.
// Returns nil if no main.go is found (library module).
// Uses artifact_pattern from ComponentTypesConfig if available, otherwise falls back to DefaultArtifactPattern.
func InferGoArtifacts(moniker, goRoot, repoRoot string, compTypes *ComponentTypesConfig) []ModuleArtifact {
	// Resolve absolute path
	absRoot := filepath.Join(repoRoot, goRoot)

	// Detect main packages
	mainLocs := detectMainPackages(absRoot, moniker)
	if len(mainLocs) == 0 {
		return nil
	}

	// Get artifact pattern from component types or use default
	pattern := getArtifactPattern(compTypes)

	// Create artifacts for each detected main package
	artifacts := make([]ModuleArtifact, 0, len(mainLocs))
	for _, loc := range mainLocs {
		artifacts = append(artifacts, ModuleArtifact{
			ID:      loc.name,
			Type:    ArtifactTypeExecutable,
			Pattern: pattern,
		})
	}

	return artifacts
}

// InferGoArtifactsForModule checks if a module's Go component has explicit artifacts.
// Returns nil if explicit artifacts exist (no inference needed).
// Otherwise delegates to InferGoArtifacts.
func InferGoArtifactsForModule(m *Module, repoRoot string, compTypes *ComponentTypesConfig) []ModuleArtifact {
	// Check if module has a Go component
	goComp, ok := m.Components["go"]
	if !ok {
		return nil
	}

	// Skip if explicit artifacts are already defined
	if goComp.Build != nil && len(goComp.Build.Artifacts) > 0 {
		return nil
	}

	return InferGoArtifacts(m.Moniker, goComp.Root, repoRoot, compTypes)
}

// getArtifactPattern returns the artifact pattern from component types config.
// Falls back to DefaultArtifactPattern if not configured.
func getArtifactPattern(compTypes *ComponentTypesConfig) string {
	if compTypes == nil {
		return DefaultArtifactPattern
	}

	goType, ok := compTypes.ComponentTypes["go"]
	if !ok {
		return DefaultArtifactPattern
	}

	if goType.Defaults == nil {
		return DefaultArtifactPattern
	}

	if goType.Defaults.ArtifactPattern == "" {
		return DefaultArtifactPattern
	}

	return goType.Defaults.ArtifactPattern
}
