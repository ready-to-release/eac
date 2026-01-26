package config

// This file contains build-related types used in component-types.yml configuration.
// These define component-type-level defaults for artifacts and Docker builds.

// BuildConfig contains build output configuration for a component type.
type BuildConfig struct {
	Artifacts []Artifact `yaml:"artifacts"`
}

// Artifact defines an expected build artifact.
type Artifact struct {
	Type        string   `yaml:"type"`                  // executable, file, directory, marker, image, glob
	Pattern     string   `yaml:"pattern"`               // Path pattern with variables: {moniker}, {os}, {arch}, {ext}
	ID          string   `yaml:"id,omitempty"`          // Optional explicit ID for metadata override key
	Platforms   []string `yaml:"platforms,omitempty"`   // For executables: linux, windows, darwin
	Verify      string   `yaml:"verify,omitempty"`      // Verification mode: current_platform (default), all, any
	Compression string   `yaml:"compression,omitempty"` // Compression type: none (default), strip, upx
	DeriveFrom  string   `yaml:"derive_from,omitempty"` // Source artifact pattern to derive from (for compressed variants)
}

// ArtifactType constants.
const (
	ArtifactTypeExecutable = "executable"
	ArtifactTypeFile       = "file"
	ArtifactTypeDirectory  = "directory"
	ArtifactTypeImage      = "image"
	ArtifactTypeGlob       = "glob"
	ArtifactTypeTest       = "test"
)

// VerifyMode constants.
const (
	VerifyCurrentPlatform = "current_platform"
	VerifyAll             = "all"
	VerifyAny             = "any"
)

// CompressionType constants.
const (
	CompressionNone  = ""      // No compression (default)
	CompressionStrip = "strip" // Strip debug symbols only
	CompressionUPX   = "upx"   // UPX compression (implies strip)
)

// DockerBuildConfig contains Docker image build configuration.
type DockerBuildConfig struct {
	Container  string             `yaml:"container"`            // Container name (references containers/{container}/)
	Context    string             `yaml:"context"`              // Build context path
	Dockerfile string             `yaml:"dockerfile,omitempty"` // Path to Dockerfile (default: {context}/Dockerfile)
	Platforms  []string           `yaml:"platforms,omitempty"`  // Target platforms (e.g., linux/amd64, linux/arm64)
	Tags       []string           `yaml:"tags"`                 // Image tags
	Load       bool               `yaml:"load,omitempty"`       // Load image to local Docker daemon
	Push       bool               `yaml:"push,omitempty"`       // Push image to registry
	Registry   string             `yaml:"registry,omitempty"`   // Registry to push to (if push=true)
	Cache      *DockerCacheConfig `yaml:"cache,omitempty"`      // Cache configuration
	SBOM       bool               `yaml:"sbom,omitempty"`       // Generate SBOM
	Provenance bool               `yaml:"provenance,omitempty"` // Generate provenance attestation
}

// DockerCacheConfig contains Docker build cache configuration.
type DockerCacheConfig struct {
	Type  string `yaml:"type"`            // "gha" or "registry"
	Scope string `yaml:"scope,omitempty"` // Cache scope (for GHA cache)
	From  string `yaml:"from,omitempty"`  // Cache source image (for registry cache)
	To    string `yaml:"to,omitempty"`    // Cache destination image (for registry cache)
	Mode  string `yaml:"mode,omitempty"`  // Cache mode: "min" or "max"
}

// Clone creates a deep copy of DockerBuildConfig.
func (d *DockerBuildConfig) Clone() *DockerBuildConfig {
	if d == nil {
		return nil
	}
	clone := &DockerBuildConfig{
		Container:  d.Container,
		Context:    d.Context,
		Dockerfile: d.Dockerfile,
		Load:       d.Load,
		Push:       d.Push,
		Registry:   d.Registry,
		SBOM:       d.SBOM,
		Provenance: d.Provenance,
	}
	if d.Platforms != nil {
		clone.Platforms = make([]string, len(d.Platforms))
		copy(clone.Platforms, d.Platforms)
	}
	if d.Tags != nil {
		clone.Tags = make([]string, len(d.Tags))
		copy(clone.Tags, d.Tags)
	}
	if d.Cache != nil {
		clone.Cache = &DockerCacheConfig{
			Type:  d.Cache.Type,
			Scope: d.Cache.Scope,
			From:  d.Cache.From,
			To:    d.Cache.To,
			Mode:  d.Cache.Mode,
		}
	}
	return clone
}

// DockerBuildConfigToMap converts DockerBuildConfig to map[string]interface{}.
// This is used when converting between typed config and contract types.
func DockerBuildConfigToMap(d *DockerBuildConfig) map[string]interface{} {
	if d == nil {
		return nil
	}
	result := make(map[string]interface{})
	if d.Container != "" {
		result["container"] = d.Container
	}
	if d.Context != "" {
		result["context"] = d.Context
	}
	if d.Dockerfile != "" {
		result["dockerfile"] = d.Dockerfile
	}
	if len(d.Platforms) > 0 {
		result["platforms"] = d.Platforms
	}
	if len(d.Tags) > 0 {
		result["tags"] = d.Tags
	}
	if d.Load {
		result["load"] = d.Load
	}
	if d.Push {
		result["push"] = d.Push
	}
	if d.Registry != "" {
		result["registry"] = d.Registry
	}
	if d.SBOM {
		result["sbom"] = d.SBOM
	}
	if d.Provenance {
		result["provenance"] = d.Provenance
	}
	if d.Cache != nil {
		cache := make(map[string]interface{})
		if d.Cache.Type != "" {
			cache["type"] = d.Cache.Type
		}
		if d.Cache.Scope != "" {
			cache["scope"] = d.Cache.Scope
		}
		if d.Cache.From != "" {
			cache["from"] = d.Cache.From
		}
		if d.Cache.To != "" {
			cache["to"] = d.Cache.To
		}
		if d.Cache.Mode != "" {
			cache["mode"] = d.Cache.Mode
		}
		result["cache"] = cache
	}
	return result
}

// GetVerifyMode returns the verification mode, defaulting to current_platform.
func (a *Artifact) GetVerifyMode() string {
	if a.Verify == "" {
		return VerifyCurrentPlatform
	}
	return a.Verify
}

// IsExecutable returns true if this is an executable artifact.
func (a *Artifact) IsExecutable() bool {
	return a.Type == ArtifactTypeExecutable
}

// GetCompression returns the compression type, defaulting to none.
func (a *Artifact) GetCompression() string {
	if a.Compression == "" {
		return CompressionNone
	}
	return a.Compression
}

// IsDerived returns true if this artifact is derived from another.
func (a *Artifact) IsDerived() bool {
	return a.DeriveFrom != ""
}

// RequiresCompression returns true if this artifact needs compression.
func (a *Artifact) RequiresCompression() bool {
	return a.Compression == CompressionStrip || a.Compression == CompressionUPX
}

