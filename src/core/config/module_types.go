package config

// ModuleTypesConfig represents the module-types.yml configuration
type ModuleTypesConfig struct {
	Types []ModuleTypeDef `yaml:"types"`

	// Internal lookup map (built after load)
	typeMap map[string]*ModuleTypeDef
}

// ModuleTypeDef defines a module type
type ModuleTypeDef struct {
	Name          string        `yaml:"name"`
	Description   string        `yaml:"description"`
	BuildDeps     []string      `yaml:"build_deps"`     // System dependencies required for building
	Capabilities  []string      `yaml:"capabilities"`
	TestFramework string        `yaml:"test_framework,omitempty"` // Test framework: mocha, jest, pytest, go (default)
	Build         *BuildConfig  `yaml:"build,omitempty"`
	Defaults      *TypeDefaults `yaml:"defaults,omitempty"`
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
	Type      string   `yaml:"type"`               // executable, file, directory, marker, image, glob
	Pattern   string   `yaml:"pattern"`            // Path pattern with variables: {moniker}, {os}, {arch}, {ext}
	Platforms []string `yaml:"platforms,omitempty"` // For executables: linux, windows, darwin
	Verify    string   `yaml:"verify,omitempty"`   // Verification mode: current_platform (default), all, any
}

// ArtifactType constants
const (
	ArtifactTypeExecutable = "executable"
	ArtifactTypeFile       = "file"
	ArtifactTypeDirectory  = "directory"
	ArtifactTypeMarker     = "marker"
	ArtifactTypeImage      = "image"
	ArtifactTypeGlob       = "glob"
)

// VerifyMode constants
const (
	VerifyCurrentPlatform = "current_platform"
	VerifyAll             = "all"
	VerifyAny             = "any"
)

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

// GetBuildDeps returns the build dependencies for a module type
func (c *ModuleTypesConfig) GetBuildDeps(typeName string) []string {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return nil
	}
	return typeDef.BuildDeps
}

// GetPrimaryBuildDep returns the first build dependency (used for build dispatch)
func (c *ModuleTypesConfig) GetPrimaryBuildDep(typeName string) string {
	deps := c.GetBuildDeps(typeName)
	if len(deps) == 0 {
		return ""
	}
	return deps[0]
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

// IsMarker returns true if this is a marker artifact
func (a *Artifact) IsMarker() bool {
	return a.Type == ArtifactTypeMarker
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
