package config

import (
	"fmt"
	"path/filepath"
)

// ResolvedArtifact represents an artifact with metadata overrides applied and existence checked.
type ResolvedArtifact struct {
	// From artifact definition
	Type       string   `json:"type" yaml:"type"`
	ID         string   `json:"id" yaml:"id"`
	Pattern    string   `json:"pattern" yaml:"pattern"`
	Platforms  []string `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	VerifyMode string   `json:"verify_mode,omitempty" yaml:"verify_mode,omitempty"`

	// Resolved information
	ResolvedName     string `json:"resolved_name" yaml:"resolved_name"`
	ResolvedPath     string `json:"resolved_path" yaml:"resolved_path"`
	MetadataOverride string `json:"metadata_override,omitempty" yaml:"metadata_override,omitempty"`
	Exists           bool   `json:"exists" yaml:"exists"`
	IsDirectory      bool   `json:"is_directory,omitempty" yaml:"is_directory,omitempty"`
	Error            string `json:"error,omitempty" yaml:"error,omitempty"`
}

// ArtifactResolutionSummary provides summary statistics for artifact resolution.
type ArtifactResolutionSummary struct {
	Total     int `json:"total" yaml:"total"`
	Exists    int `json:"exists" yaml:"exists"`
	Missing   int `json:"missing" yaml:"missing"`
	Overrides int `json:"overrides" yaml:"overrides"`
}

// ResolveArtifactsForModule resolves all artifacts for a module with metadata applied.
func ResolveArtifactsForModule(
	module *Module,
	buildDir string,
	targetOS, targetArch string,
) ([]ResolvedArtifact, *ArtifactResolutionSummary, error) {
	return ResolveArtifactsForModuleWithConfig(module, buildDir, targetOS, targetArch, nil)
}

// ResolveArtifactsForModuleWithConfig resolves all artifacts for a module with optional books config.
// Artifact resolution uses cfg.GetBuildArtifacts() which properly merges module-level and
// type-level artifacts (module-level takes priority).
func ResolveArtifactsForModuleWithConfig(
	module *Module,
	buildDir string,
	targetOS, targetArch string,
	cfg *EACConfig,
) ([]ResolvedArtifact, *ArtifactResolutionSummary, error) {
	if module == nil {
		return nil, nil, fmt.Errorf("module cannot be nil")
	}

	// Get merged artifacts from config (module-level takes priority over type-level)
	// Pass buildAll=true to get all artifacts including UPX variants for resolution
	var artifacts []Artifact
	if cfg != nil {
		artifacts = cfg.GetBuildArtifacts(module.Moniker, true)
	}

	if len(artifacts) == 0 {
		// No artifacts defined - not a buildable module
		return []ResolvedArtifact{}, &ArtifactResolutionSummary{}, nil
	}

	// Create artifact resolver
	resolver := NewArtifactResolverFull(
		module.Moniker,
		buildDir,
		targetOS,
		targetArch,
		module.Metadata,
	)

	// If no artifacts defined, module type doesn't produce artifacts.
	// The per-module manifest is the contract for validation (not markers).

	// Verify all artifacts
	results := resolver.VerifyArtifacts(artifacts)

	// Convert to ResolvedArtifact format
	resolved := make([]ResolvedArtifact, len(results))
	summary := &ArtifactResolutionSummary{}

	for i, result := range results {
		artifact := ResolvedArtifact{
			Type:         result.Artifact.Type,
			ID:           deriveArtifactID(result.Artifact, targetOS, targetArch),
			Pattern:      result.Artifact.Pattern,
			Platforms:    result.Artifact.Platforms,
			VerifyMode:   result.Artifact.Verify,
			ResolvedName: result.Pattern,
			ResolvedPath: result.Path,
			Exists:       result.Exists,
			IsDirectory:  result.IsDirectory,
		}

		if result.Error != nil {
			artifact.Error = result.Error.Error()
		}

		// Check if this artifact used a metadata override
		if module.Metadata != nil {
			metadataKey := fmt.Sprintf("%s-%s", result.Artifact.Type, artifact.ID)
			if _, ok := module.Metadata[metadataKey]; ok {
				artifact.MetadataOverride = metadataKey
				summary.Overrides++
			}
		}

		resolved[i] = artifact
		summary.Total++
		if artifact.Exists {
			summary.Exists++
		} else {
			summary.Missing++
		}
	}

	return resolved, summary, nil
}

// deriveArtifactID derives an ID for an artifact based on its type and context.
func deriveArtifactID(artifact Artifact, targetOS, targetArch string) string {
	// If artifact has explicit ID, use it
	if artifact.ID != "" {
		return artifact.ID
	}

	// For executables, use {os}-{arch} format with compression suffix if applicable
	if artifact.Type == ArtifactTypeExecutable {
		baseID := fmt.Sprintf("%s-%s", targetOS, targetArch)

		// Append compression suffix for compressed variants
		switch artifact.Compression {
		case CompressionUPX:
			return baseID + "-upx"
		case CompressionStrip:
			return baseID + "-strip"
		}

		return baseID
	}

	// For other types, extract from pattern
	// Remove common variables to get a reasonable ID
	id := artifact.Pattern
	id = filepath.Base(id)

	// Remove extension for files
	if artifact.Type == ArtifactTypeFile {
		ext := filepath.Ext(id)
		if ext != "" {
			id = id[:len(id)-len(ext)]
		}
	}

	return id
}
