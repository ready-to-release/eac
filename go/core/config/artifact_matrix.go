package config

import "gopkg.in/yaml.v3"

// ArtifactMatrixEntry represents a single entry in an artifact matrix.
type ArtifactMatrixEntry struct {
	ID          string `yaml:"id"`
	Type        string `yaml:"type"`
	Pattern     string `yaml:"pattern"`
	Compression string `yaml:"compression,omitempty"`
	DeriveFrom  string `yaml:"derive_from,omitempty"`
}

// ArtifactMatrix represents a named artifact matrix that can be extended.
type ArtifactMatrix struct {
	Extends    string                `yaml:"extends,omitempty"`
	Entries    []ArtifactMatrixEntry `yaml:"-"`          // Inline entries (for simple matrices)
	Additional []ArtifactMatrixEntry `yaml:"additional"` // Additional entries when extending
}

// UnmarshalYAML implements custom unmarshaling for ArtifactMatrix.
// Handles both simple array format and object format with extends/additional.
func (am *ArtifactMatrix) UnmarshalYAML(node *yaml.Node) error {
	// Handle array format (simple matrix)
	if node.Kind == yaml.SequenceNode {
		return node.Decode(&am.Entries)
	}

	// Handle object format (with extends/additional)
	if node.Kind == yaml.MappingNode {
		type rawMatrix struct {
			Extends    string                `yaml:"extends,omitempty"`
			Additional []ArtifactMatrixEntry `yaml:"additional,omitempty"`
		}
		var raw rawMatrix
		if err := node.Decode(&raw); err != nil {
			return err
		}
		am.Extends = raw.Extends
		am.Additional = raw.Additional
		return nil
	}

	return nil
}

// ExpandArtifactMatrix expands a matrix reference into concrete ComponentArtifact entries.
// Returns nil if the matrix name is not found.
// Parameters:
//   - matrixName: the name of the matrix to expand
//   - params: parameter substitution map (e.g., {"moniker": "myapp", "ext": ".exe"})
//
// Returns expanded artifacts with all parameters substituted.
func (cfg *BlueprintsConfig) ExpandArtifactMatrix(matrixName string, params map[string]string) []ModuleArtifact {
	if cfg == nil || cfg.ArtifactMatrices == nil {
		return nil
	}

	// Track visited matrices to detect circular inheritance
	visited := make(map[string]bool)
	return cfg.expandMatrixWithVisited(matrixName, params, visited)
}

// expandMatrixWithVisited expands a matrix while tracking visited matrices for cycle detection.
func (cfg *BlueprintsConfig) expandMatrixWithVisited(matrixName string, params map[string]string, visited map[string]bool) []ModuleArtifact {
	// Check for circular reference
	if visited[matrixName] {
		return nil // Circular reference detected
	}

	matrix, ok := cfg.ArtifactMatrices[matrixName]
	if !ok || matrix == nil {
		return nil
	}

	visited[matrixName] = true

	var artifacts []ModuleArtifact

	// If extending another matrix, get its entries first
	if matrix.Extends != "" {
		parentArtifacts := cfg.expandMatrixWithVisited(matrix.Extends, params, visited)
		artifacts = append(artifacts, parentArtifacts...)
	}

	// Add this matrix's entries
	for _, entry := range matrix.Entries {
		artifact := ModuleArtifact{
			ID:          substituteMatrixParams(entry.ID, params),
			Type:        entry.Type,
			Pattern:     substituteMatrixParams(entry.Pattern, params),
			Compression: entry.Compression,
			DeriveFrom:  substituteMatrixParams(entry.DeriveFrom, params),
		}
		artifacts = append(artifacts, artifact)
	}

	// Add additional entries (for extended matrices)
	for _, entry := range matrix.Additional {
		artifact := ModuleArtifact{
			ID:          substituteMatrixParams(entry.ID, params),
			Type:        entry.Type,
			Pattern:     substituteMatrixParams(entry.Pattern, params),
			Compression: entry.Compression,
			DeriveFrom:  substituteMatrixParams(entry.DeriveFrom, params),
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts
}

// expandArtifactMatrixForModule expands a module's artifact_matrix reference into concrete
// Go component artifacts. Only applies when:
// 1. Module has ArtifactMatrixRef set
// 2. Module has a Go component
// 3. Go component doesn't already have explicit artifacts
func expandArtifactMatrixForModule(mod *Module, matrices *BlueprintsConfig) {
	if mod.ArtifactMatrixRef == "" || matrices == nil {
		return
	}

	goComp, ok := mod.Components["go"]
	if !ok || goComp == nil {
		return
	}

	// Skip if explicit artifacts already defined
	if goComp.Build != nil && len(goComp.Build.Artifacts) > 0 {
		return
	}

	// Expand the matrix with the module's moniker and binary name from component
	params := map[string]string{"moniker": mod.Moniker}
	if goComp.Build != nil && goComp.Build.BinaryName != "" {
		params["binary_name"] = goComp.Build.BinaryName
	} else {
		params["binary_name"] = mod.Moniker
	}
	artifacts := matrices.ExpandArtifactMatrix(mod.ArtifactMatrixRef, params)
	if len(artifacts) == 0 {
		return
	}

	// Set the expanded artifacts on the Go component
	if goComp.Build == nil {
		goComp.Build = &ModuleBuild{}
	}
	goComp.Build.Artifacts = artifacts
}

// substituteMatrixParams replaces {placeholder} patterns in a string.
func substituteMatrixParams(s string, params map[string]string) string {
	if s == "" {
		return s
	}
	result := s
	for key, value := range params {
		result = replaceMatrixPlaceholder(result, key, value)
	}
	return result
}

// replaceMatrixPlaceholder replaces a single {key} placeholder with value.
func replaceMatrixPlaceholder(s, key, value string) string {
	placeholder := "{" + key + "}"
	result := s
	for {
		idx := indexOfSubstring(result, placeholder)
		if idx == -1 {
			break
		}
		result = result[:idx] + value + result[idx+len(placeholder):]
	}
	return result
}

// indexOfSubstring returns the index of substr in s, or -1 if not found.
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
