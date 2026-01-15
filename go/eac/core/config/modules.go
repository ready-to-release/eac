package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/defaults"
)

// Module represents a single module definition
type Module struct {
	Moniker       string                 `yaml:"moniker"`
	Name          string                 `yaml:"name"`
	Type          string                 `yaml:"type"`
	Description   string                 `yaml:"description"`
	DependsOn     []string               `yaml:"depends_on"`
	DependsOnCI   []string               `yaml:"depends_on_ci"`            // CI artifact dependencies (merged into DependsOn)
	CIDeps        []string               `yaml:"-"`                        // Computed: CI artifact deps for dispatch layering
	Build         *ModuleBuild           `yaml:"build,omitempty"`          // Per-module build configuration
	DockerBuild   map[string]interface{} `yaml:"docker_build,omitempty"`   // Per-module Docker build configuration
	Books         []string               `yaml:"books,omitempty"`          // Book names to build for this module (references books.yml)
	EvidenceBooks []string               `yaml:"evidence_books,omitempty"` // Evidence book names, built via 'update evidence' command
	ReleaseBundle *ReleaseBundle         `yaml:"release_bundle,omitempty"` // Release bundle configuration (for release modules)
	Files         Files                  `yaml:"files"`
	Flags         Flags                  `yaml:"flags"`
	Metadata      map[string]string      `yaml:"metadata,omitempty"` // Generic key-value store for module-specific data
	Versioning    *ModuleVersioning      `yaml:"versioning,omitempty"`
}

// ReleaseBundle configures how the release module creates GitHub releases
type ReleaseBundle struct {
	TitleFormat string                  `yaml:"title_format"` // Title template, e.g., "{r2r} ({r2r_version}) + {eac} ({eac_version})"
	Headline    map[string]string       `yaml:"headline"`     // Map of label -> moniker for title, e.g., {"r2r": "r2r-cli", "eac": "ext-eac"}
	Categories  []ReleaseBundleCategory `yaml:"categories"`   // Grouped modules for release notes
}

// ReleaseBundleCategory groups modules in release notes
type ReleaseBundleCategory struct {
	Name        string   `yaml:"name"`        // Category name, e.g., "Core Tools"
	Description string   `yaml:"description"` // Category description
	Modules     []string `yaml:"modules"`     // Module monikers in this category
}

// ModuleVersioning holds module versioning configuration
type ModuleVersioning struct {
	Scheme  string `yaml:"scheme"`  // SemVer, CalVer
	Current string `yaml:"current"` // Current version (optional)
}

// ModuleBuild contains per-module build configuration
type ModuleBuild struct {
	Handler   string           `yaml:"handler,omitempty"`   // Explicit build handler override
	Artifacts []ModuleArtifact `yaml:"artifacts,omitempty"` // Artifacts to produce
	Options   *BuildOptions    `yaml:"options,omitempty"`   // Build behavior options
}

// ModuleArtifact defines an artifact to be produced by a module build
type ModuleArtifact struct {
	ID          string `yaml:"id"`                    // Unique artifact identifier
	Type        string `yaml:"type"`                  // executable, file, directory, test
	Pattern     string `yaml:"pattern"`               // Output path pattern
	Compression string `yaml:"compression,omitempty"` // none, strip, upx
	DeriveFrom  string `yaml:"derive_from,omitempty"` // Source artifact to derive from
}

// BuildOptions contains optional build behavior flags
type BuildOptions struct {
	// Reserved for future build options
}

// GetBuildHandler returns the explicit build handler for this module
func (m *Module) GetBuildHandler() string {
	if m.Build == nil {
		return ""
	}
	return m.Build.Handler
}

// GetDockerBuildConfig parses and returns the per-module docker_build configuration.
// Returns nil if no docker_build is defined at module level.
// This allows modules to override or extend type-level docker_build config.
func (m *Module) GetDockerBuildConfig() *DockerBuildConfig {
	if m.DockerBuild == nil || len(m.DockerBuild) == 0 {
		return nil
	}

	// Parse the raw map[string]interface{} into DockerBuildConfig
	cfg := &DockerBuildConfig{}

	if v, ok := m.DockerBuild["container"].(string); ok {
		cfg.Container = v
	}
	if v, ok := m.DockerBuild["context"].(string); ok {
		cfg.Context = v
	}
	if v, ok := m.DockerBuild["dockerfile"].(string); ok {
		cfg.Dockerfile = v
	}
	if v, ok := m.DockerBuild["platforms"].([]interface{}); ok {
		for _, p := range v {
			if ps, ok := p.(string); ok {
				cfg.Platforms = append(cfg.Platforms, ps)
			}
		}
	}
	if v, ok := m.DockerBuild["tags"].([]interface{}); ok {
		for _, t := range v {
			if ts, ok := t.(string); ok {
				cfg.Tags = append(cfg.Tags, ts)
			}
		}
	}
	if v, ok := m.DockerBuild["load"].(bool); ok {
		cfg.Load = v
	}
	if v, ok := m.DockerBuild["push"].(bool); ok {
		cfg.Push = v
	}
	if v, ok := m.DockerBuild["registry"].(string); ok {
		cfg.Registry = v
	}
	if v, ok := m.DockerBuild["sbom"].(bool); ok {
		cfg.SBOM = v
	}
	if v, ok := m.DockerBuild["provenance"].(bool); ok {
		cfg.Provenance = v
	}

	// Parse cache config if present
	if cacheMap, ok := m.DockerBuild["cache"].(map[string]interface{}); ok {
		cfg.Cache = &DockerCacheConfig{}
		if v, ok := cacheMap["type"].(string); ok {
			cfg.Cache.Type = v
		}
		if v, ok := cacheMap["scope"].(string); ok {
			cfg.Cache.Scope = v
		}
		if v, ok := cacheMap["from"].(string); ok {
			cfg.Cache.From = v
		}
		if v, ok := cacheMap["to"].(string); ok {
			cfg.Cache.To = v
		}
		if v, ok := cacheMap["mode"].(string); ok {
			cfg.Cache.Mode = v
		}
	}

	return cfg
}

// Files defines file ownership patterns for a module
type Files struct {
	Root      string    `yaml:"root"`
	Source    []string  `yaml:"source"`
	Config    []string  `yaml:"config"`
	Assets    []string  `yaml:"assets"`
	Tests     []string  `yaml:"tests"`
	Exclude   []string  `yaml:"exclude"`
	Changelog string    `yaml:"changelog"`
	Workflows Workflows `yaml:"workflows"`
	Repo      RepoFiles `yaml:"repo"`
}

// Workflows defines GitHub Actions workflow file ownership
type Workflows struct {
	CI      string `yaml:"ci"`      // CI workflow file path
	Release string `yaml:"release"` // Release workflow file path
}

// RepoFiles defines repository-level file ownership
type RepoFiles struct {
	Specs    []string `yaml:"specs"`
	TestImpl string   `yaml:"test_impl"` // Test implementation directory path
	Design   string   `yaml:"design"`    // Design workspace directory path
	Other    []string `yaml:"other"`
	Exclude  []string `yaml:"exclude"`
}

// Flags defines module behavior flags
type Flags struct {
	// ExplicitOwnership disables the default "all files under root" ownership.
	// When true, the module only owns files that explicitly match its patterns.
	ExplicitOwnership bool `yaml:"explicit_ownership,omitempty"`
}

// applyModuleDefaults applies default values to all modules (generic defaults only).
// Call ApplyTypeDefaults after loading ModuleTypes for type-specific defaults.
func (c *RepositoryConfig) applyModuleDefaults() {
	for i := range c.Modules {
		m := &c.Modules[i]

		if m.Type == "" {
			m.Type = defaults.ModuleType
		}
		if m.Description == "" {
			m.Description = m.Name
		}
		if m.DependsOn == nil {
			m.DependsOn = []string{}
		}

		// Merge depends_on_ci into depends_on and track separately as CIDeps
		if len(m.DependsOnCI) > 0 {
			m.CIDeps = make([]string, len(m.DependsOnCI))
			copy(m.CIDeps, m.DependsOnCI)

			// Merge into DependsOn (avoiding duplicates)
			existing := make(map[string]bool)
			for _, dep := range m.DependsOn {
				existing[dep] = true
			}
			for _, ciDep := range m.DependsOnCI {
				if !existing[ciDep] {
					m.DependsOn = append(m.DependsOn, ciDep)
					existing[ciDep] = true
				}
			}
		}

		// Note: Other defaults (Source, Config, Changelog, Specs, etc.) are now
		// applied by ApplyTypeDefaults using type-specific defaults with fallback.
	}
}

// ApplyTypeDefaults applies type-specific defaults to all modules.
// This should be called after both Modules and ModuleTypes are loaded.
func (c *RepositoryConfig) ApplyTypeDefaults(types *ModuleTypesConfig) {
	// Build path variables from repository config
	pathVars := c.GetPathVariables()

	for i := range c.Modules {
		m := &c.Modules[i]

		// Get type definition
		var typeDef *defaults.TypeDefaults
		if types != nil {
			if td := types.Get(m.Type); td != nil && td.Defaults != nil {
				typeDef = convertTypeDefaults(td.Defaults)
			}
		}

		// Resolve defaults using type-specific + generic fallback
		resolved := defaults.ResolveDefaults(
			typeDef,
			m.Moniker, m.Files.Root, m.Type,
			pathVars,
			m.Files.Source, m.Files.Config, m.Files.Assets, m.Files.Tests,
			m.Files.Changelog,
			m.Files.Workflows.CI, m.Files.Workflows.Release,
			m.Files.Repo.Specs,
			m.Files.Repo.TestImpl, m.Files.Repo.Design,
		)

		// Apply resolved values (only if not already set)
		if m.Files.Source == nil {
			m.Files.Source = resolved.Source
		}
		if m.Files.Config == nil {
			m.Files.Config = resolved.Config
		}
		if m.Files.Assets == nil {
			m.Files.Assets = resolved.Assets
		}
		if m.Files.Tests == nil {
			m.Files.Tests = resolved.Tests
		}
		if m.Files.Changelog == "" {
			m.Files.Changelog = resolved.Changelog
		}
		if m.Files.Workflows.CI == "" {
			m.Files.Workflows.CI = resolved.WorkflowCI
		}
		if m.Files.Workflows.Release == "" {
			m.Files.Workflows.Release = resolved.WorkflowRelease
		}
		if m.Files.Repo.Specs == nil {
			m.Files.Repo.Specs = resolved.Specs
		}
		if m.Files.Repo.TestImpl == "" {
			m.Files.Repo.TestImpl = resolved.TestImpl
		}
		if m.Files.Repo.Design == "" {
			m.Files.Repo.Design = resolved.Design
		}
	}
}

// convertTypeDefaults converts config.TypeDefaults to defaults.TypeDefaults
func convertTypeDefaults(td *TypeDefaults) *defaults.TypeDefaults {
	if td == nil {
		return nil
	}

	result := &defaults.TypeDefaults{}

	if td.Files != nil {
		result.Files = &defaults.FilesDefaults{
			Source:    td.Files.Source,
			Config:    td.Files.Config,
			Assets:    td.Files.Assets,
			Tests:     td.Files.Tests,
			Changelog: td.Files.Changelog,
		}
		if td.Files.Workflows != nil {
			result.Files.WorkflowCI = td.Files.Workflows.CI
			result.Files.WorkflowRelease = td.Files.Workflows.Release
		}
	}

	if td.Repo != nil {
		result.Repo = &defaults.RepoDefaults{
			Specs:    td.Repo.Specs,
			TestImpl: td.Repo.TestImpl,
			Design:   td.Repo.Design,
		}
	}

	return result
}

// ValidateAndDiscoverWorkflows validates workflow paths and auto-discovers missing ones.
// For each module:
//   - If workflow is defined but file doesn't exist → returns error
//   - If workflow is empty but conventional file exists → sets the path (auto-discovery)
//   - If workflow is empty and file doesn't exist → remains empty (no workflow)
func (c *RepositoryConfig) ValidateAndDiscoverWorkflows(repoRoot string) error {
	var errs []error

	for i := range c.Modules {
		m := &c.Modules[i]

		// Validate/discover CI workflow
		ciPath := defaults.WorkflowCIPath(m.Moniker)
		ciFullPath := filepath.Join(repoRoot, ciPath)

		if m.Files.Workflows.CI != "" {
			// Defined - validate it exists
			definedPath := filepath.Join(repoRoot, m.Files.Workflows.CI)
			if _, err := os.Stat(definedPath); os.IsNotExist(err) {
				errs = append(errs, fmt.Errorf("module %s: CI workflow file not found: %s", m.Moniker, m.Files.Workflows.CI))
			}
		} else {
			// Not defined - check for auto-discovery
			if _, err := os.Stat(ciFullPath); err == nil {
				m.Files.Workflows.CI = ciPath
			}
		}

		// Validate/discover Release workflow
		releasePath := defaults.WorkflowReleasePath(m.Moniker)
		releaseFullPath := filepath.Join(repoRoot, releasePath)

		if m.Files.Workflows.Release != "" {
			// Defined - validate it exists
			definedPath := filepath.Join(repoRoot, m.Files.Workflows.Release)
			if _, err := os.Stat(definedPath); os.IsNotExist(err) {
				errs = append(errs, fmt.Errorf("module %s: release workflow file not found: %s", m.Moniker, m.Files.Workflows.Release))
			}
		} else {
			// Not defined - check for auto-discovery
			if _, err := os.Stat(releaseFullPath); err == nil {
				m.Files.Workflows.Release = releasePath
			}
		}
	}

	if len(errs) > 0 {
		return &MultiWorkflowError{Errors: errs}
	}
	return nil
}

// MultiWorkflowError holds multiple workflow validation errors
type MultiWorkflowError struct {
	Errors []error
}

func (e *MultiWorkflowError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	msg := fmt.Sprintf("%d workflow errors:", len(e.Errors))
	for _, err := range e.Errors {
		msg += "\n  - " + err.Error()
	}
	return msg
}
