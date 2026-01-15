package config

// ModuleTypesConfig represents the module-types.yml configuration
type ModuleTypesConfig struct {
	Types []ModuleTypeDef `yaml:"types"`

	// Internal lookup map (built after load)
	typeMap map[string]*ModuleTypeDef
}

// ModuleTypeDef defines a module type
type ModuleTypeDef struct {
	Name          string             `yaml:"name"`
	Description   string             `yaml:"description"`
	BuildDeps     []string           `yaml:"build_deps"` // System dependencies required for building
	Capabilities  []string           `yaml:"capabilities"`
	TestFramework string             `yaml:"test_framework,omitempty"` // Unit test framework: mocha, jest, pytest, go (default)
	BDDFramework  string             `yaml:"bdd_framework,omitempty"`  // BDD test framework: godog, tscucumber (default: inferred from build_deps)
	DockerBuild   *DockerBuildConfig `yaml:"docker_build,omitempty"`   // Docker image build configuration (for docker-build dependency)
	Build         *BuildConfig       `yaml:"build,omitempty"`
	Defaults      *TypeDefaults      `yaml:"defaults,omitempty"`
}

// BuildConfig contains build output configuration for a module type
type BuildConfig struct {
	Artifacts []Artifact      `yaml:"artifacts"`
	PostBuild []PostBuildStep `yaml:"post_build,omitempty"`
}

// PostBuildStep defines a post-build action to execute after successful build
type PostBuildStep struct {
	Action  string   `yaml:"action"`            // copy, script
	Target  string   `yaml:"target,omitempty"`  // Target path for copy action
	Include []string `yaml:"include,omitempty"` // Glob patterns to include (for copy)
	Exclude []string `yaml:"exclude,omitempty"` // Glob patterns to exclude (for copy)
	Script  string   `yaml:"script,omitempty"`  // Script command to run (for script action)
}

// PostBuildAction constants
const (
	PostBuildActionCopy   = "copy"
	PostBuildActionScript = "script"
)

// Artifact defines an expected build artifact
type Artifact struct {
	Type        string   `yaml:"type"`                  // executable, file, directory, marker, image, glob
	Pattern     string   `yaml:"pattern"`               // Path pattern with variables: {moniker}, {os}, {arch}, {ext}
	ID          string   `yaml:"id,omitempty"`          // Optional explicit ID for metadata override key
	Platforms   []string `yaml:"platforms,omitempty"`   // For executables: linux, windows, darwin
	Verify      string   `yaml:"verify,omitempty"`      // Verification mode: current_platform (default), all, any
	Compression string   `yaml:"compression,omitempty"` // Compression type: none (default), strip, upx
	DeriveFrom  string   `yaml:"derive_from,omitempty"` // Source artifact pattern to derive from (for compressed variants)
}

// ArtifactType constants
const (
	ArtifactTypeExecutable = "executable"
	ArtifactTypeFile       = "file"
	ArtifactTypeDirectory  = "directory"
	ArtifactTypeImage      = "image"
	ArtifactTypeGlob       = "glob"
	ArtifactTypeTest       = "test"
)

// VerifyMode constants
const (
	VerifyCurrentPlatform = "current_platform"
	VerifyAll             = "all"
	VerifyAny             = "any"
)

// CompressionType constants
const (
	CompressionNone  = ""      // No compression (default)
	CompressionStrip = "strip" // Strip debug symbols only
	CompressionUPX   = "upx"   // UPX compression (implies strip)
)

// DockerBuildConfig contains Docker image build configuration
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

// DockerCacheConfig contains Docker build cache configuration
type DockerCacheConfig struct {
	Type  string `yaml:"type"`            // "gha" or "registry"
	Scope string `yaml:"scope,omitempty"` // Cache scope (for GHA cache)
	From  string `yaml:"from,omitempty"`  // Cache source image (for registry cache)
	To    string `yaml:"to,omitempty"`    // Cache destination image (for registry cache)
	Mode  string `yaml:"mode,omitempty"`  // Cache mode: "min" or "max"
}

// TypeDefaults contains default values for modules of this type.
// Supports variable substitution: {moniker}, {root}, {type}
type TypeDefaults struct {
	Files *FilesDefaults `yaml:"files,omitempty"`
	Repo  *RepoDefaults  `yaml:"repo,omitempty"`
}

// FilesDefaults contains default file patterns for a module type
type FilesDefaults struct {
	Source    []string          `yaml:"source,omitempty"`
	Config    []string          `yaml:"config,omitempty"`
	Assets    []string          `yaml:"assets,omitempty"`
	Tests     []string          `yaml:"tests,omitempty"`
	Changelog string            `yaml:"changelog,omitempty"`
	Workflows *WorkflowDefaults `yaml:"workflows,omitempty"`
}

// WorkflowDefaults contains default workflow file paths
type WorkflowDefaults struct {
	CI      string `yaml:"ci,omitempty"`
	Release string `yaml:"release,omitempty"`
}

// RepoDefaults contains default repo-level configurations
type RepoDefaults struct {
	Specs    []string `yaml:"specs,omitempty"`
	TestImpl string   `yaml:"test_impl,omitempty"`
	Design   string   `yaml:"design,omitempty"`
}

// buildTypeMap builds the internal lookup map
func (c *ModuleTypesConfig) buildTypeMap() {
	c.typeMap = make(map[string]*ModuleTypeDef, len(c.Types))
	for i := range c.Types {
		c.typeMap[c.Types[i].Name] = &c.Types[i]
	}
}

// Get returns a module type definition by name
func (c *ModuleTypesConfig) Get(typeName string) *ModuleTypeDef {
	if c.typeMap == nil {
		c.buildTypeMap()
	}
	return c.typeMap[typeName]
}

// HasCapability checks if a module type has a specific capability
func (c *ModuleTypesConfig) HasCapability(typeName, capability string) bool {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return false
	}
	for _, cap := range typeDef.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// GetCapabilities returns all capabilities for a module type
func (c *ModuleTypesConfig) GetCapabilities(typeName string) []string {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return nil
	}
	return typeDef.Capabilities
}

// GetBuildDeps returns the build dependencies for a module type.
// Deprecated: Use GetBuildDepsFromCapabilities with SystemDependenciesConfig for capability-driven resolution.
func (c *ModuleTypesConfig) GetBuildDeps(typeName string) []string {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return nil
	}
	return typeDef.BuildDeps
}

// GetBuildDepsFromCapabilities returns build dependencies by resolving through capabilities.
// This is the preferred method for capability-driven architecture.
func (c *ModuleTypesConfig) GetBuildDepsFromCapabilities(typeName string, sysDeps *SystemDependenciesConfig) []string {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return nil
	}

	// If explicit BuildDeps defined, use them (backward compatibility)
	if len(typeDef.BuildDeps) > 0 {
		return typeDef.BuildDeps
	}

	// Otherwise, resolve from capabilities
	if sysDeps != nil {
		return sysDeps.GetRequiredDeps(typeDef.Capabilities)
	}

	return nil
}

// GetPrimaryBuildDep returns the first build dependency (used for build dispatch)
func (c *ModuleTypesConfig) GetPrimaryBuildDep(typeName string) string {
	deps := c.GetBuildDeps(typeName)
	if len(deps) == 0 {
		return ""
	}
	return deps[0]
}

// GetDockerBuildConfig returns the docker_build configuration for a module type
func (c *ModuleTypesConfig) GetDockerBuildConfig(typeName string) *DockerBuildConfig {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return nil
	}
	return typeDef.DockerBuild
}

// GetDockerImageName returns the primary Docker image name for a module type.
// Returns empty string if not configured.
func (c *ModuleTypesConfig) GetDockerImageName(typeName string) string {
	dockerConfig := c.GetDockerBuildConfig(typeName)
	if dockerConfig == nil || len(dockerConfig.Tags) == 0 {
		return ""
	}
	return dockerConfig.Tags[0]
}

// GetDockerContainerDir returns the container context directory for a module type.
// Returns empty string if not configured.
func (c *ModuleTypesConfig) GetDockerContainerDir(typeName string) string {
	dockerConfig := c.GetDockerBuildConfig(typeName)
	if dockerConfig == nil {
		return ""
	}
	return dockerConfig.Context
}

// GetTestFramework returns the test framework for a module type.
// Returns empty string if not specified (caller should use default based on file type).
func (c *ModuleTypesConfig) GetTestFramework(typeName string) string {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return ""
	}
	return typeDef.TestFramework
}

// GetBDDFramework returns the BDD test framework for a module type.
// Returns empty string if not specified (caller should infer from build_deps).
func (c *ModuleTypesConfig) GetBDDFramework(typeName string) string {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return ""
	}
	return typeDef.BDDFramework
}

// GetTypesWithCapability returns all type names that have the given capability
func (c *ModuleTypesConfig) GetTypesWithCapability(capability string) []string {
	var result []string
	for _, t := range c.Types {
		for _, cap := range t.Capabilities {
			if cap == capability {
				result = append(result, t.Name)
				break
			}
		}
	}
	return result
}

// GetTypesWithBuildDep returns all type names that require the given build dependency
func (c *ModuleTypesConfig) GetTypesWithBuildDep(dep string) []string {
	var result []string
	for _, t := range c.Types {
		for _, d := range t.BuildDeps {
			if d == dep {
				result = append(result, t.Name)
				break
			}
		}
	}
	return result
}

// GetPostBuildSteps returns the post-build steps for a module type
func (c *ModuleTypesConfig) GetPostBuildSteps(typeName string) []PostBuildStep {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return nil
	}
	return typeDef.GetPostBuildSteps()
}

// HasCapability checks if this type definition has a specific capability
func (t *ModuleTypeDef) HasCapability(capability string) bool {
	for _, cap := range t.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// GetArtifacts returns the build artifacts for this module type
func (t *ModuleTypeDef) GetArtifacts() []Artifact {
	if t.Build == nil {
		return nil
	}
	return t.Build.Artifacts
}

// HasArtifacts returns true if this module type defines build artifacts
func (t *ModuleTypeDef) HasArtifacts() bool {
	return t.Build != nil && len(t.Build.Artifacts) > 0
}

// GetArtifactsByType returns artifacts of a specific type
func (t *ModuleTypeDef) GetArtifactsByType(artifactType string) []Artifact {
	if t.Build == nil {
		return nil
	}
	var result []Artifact
	for _, a := range t.Build.Artifacts {
		if a.Type == artifactType {
			result = append(result, a)
		}
	}
	return result
}

// GetVerifyMode returns the verification mode, defaulting to current_platform
func (a *Artifact) GetVerifyMode() string {
	if a.Verify == "" {
		return VerifyCurrentPlatform
	}
	return a.Verify
}

// IsExecutable returns true if this is an executable artifact
func (a *Artifact) IsExecutable() bool {
	return a.Type == ArtifactTypeExecutable
}

// GetCompression returns the compression type, defaulting to none
func (a *Artifact) GetCompression() string {
	if a.Compression == "" {
		return CompressionNone
	}
	return a.Compression
}

// IsDerived returns true if this artifact is derived from another
func (a *Artifact) IsDerived() bool {
	return a.DeriveFrom != ""
}

// RequiresCompression returns true if this artifact needs compression
func (a *Artifact) RequiresCompression() bool {
	return a.Compression == CompressionStrip || a.Compression == CompressionUPX
}

// GetPostBuildSteps returns the post-build steps for this module type
func (t *ModuleTypeDef) GetPostBuildSteps() []PostBuildStep {
	if t.Build == nil {
		return nil
	}
	return t.Build.PostBuild
}

// HasPostBuild returns true if this module type defines post-build steps
func (t *ModuleTypeDef) HasPostBuild() bool {
	return t.Build != nil && len(t.Build.PostBuild) > 0
}

// IsCopyAction returns true if this is a copy action
func (p *PostBuildStep) IsCopyAction() bool {
	return p.Action == PostBuildActionCopy
}

// IsScriptAction returns true if this is a script action
func (p *PostBuildStep) IsScriptAction() bool {
	return p.Action == PostBuildActionScript
}
