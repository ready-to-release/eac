package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ModuleTemplate defines a reusable module structure with placeholders.
// Templates are defined in defaults and referenced by modules via the "template" field.
type ModuleTemplate struct {
	// Versioning provides default versioning configuration
	Versioning *ModuleVersioning `yaml:"versioning,omitempty"`
	// Components provides default component definitions with placeholders
	Components ModuleComponents `yaml:"components,omitempty"`
	// DependsOn provides default dependencies
	DependsOn []string `yaml:"depends_on,omitempty"`
}

// ModuleTemplatesConfig holds all module templates loaded from defaults.
type ModuleTemplatesConfig struct {
	Templates map[string]ModuleTemplate `yaml:"module-templates"`
}

// DiscoveryConventions defines rules for auto-discovering components based on filesystem.
type DiscoveryConventions struct {
	Gherkin      *DiscoveryRule    `yaml:"gherkin,omitempty"`
	Structurizr  *DiscoveryRule    `yaml:"structurizr,omitempty"`
	GherkinSteps *GherkinStepsRule `yaml:"gherkin_steps,omitempty"`
	Markdown     *DeriveRule       `yaml:"markdown,omitempty"`
	YAML         *DeriveRule       `yaml:"yaml,omitempty"`
}

// GherkinStepsRule defines how to infer gherkin-steps component path.
// The steps path is derived from the module's primary code component.
// Supports multiple test runners: godog (Go), cucumber-js (TypeScript).
type GherkinStepsRule struct {
	// DeriveFromComponents maps component type to subdirectory suffix.
	// e.g., {"go": "specs", "typescript": "features"}
	// Path becomes: {component_root}/{subdirectory}
	DeriveFromComponents map[string]string `yaml:"derive_from_components,omitempty"`

	// FallbackPattern is used when no code component exists.
	// Uses {moniker} placeholder, e.g., "go/eac/specs/{moniker}"
	FallbackPattern string `yaml:"fallback_pattern,omitempty"`

	// RequiredFile must exist for the component to be discovered.
	// e.g., "godog_test.go" for Go executors
	RequiredFile string `yaml:"required_file"`
}

// DiscoveryRule defines how to find a component on the filesystem.
type DiscoveryRule struct {
	PathPattern  string `yaml:"path_pattern"`  // Pattern with {moniker} placeholder
	RequiredFile string `yaml:"required_file"` // File that must exist to enable component
}

// DeriveRule defines how to derive a component's root from another component.
type DeriveRule struct {
	DeriveFrom []string `yaml:"derive_from"` // Component names to try (first match wins)
}

// LoadModuleTemplates loads module templates from defaults.
// Returns nil, ErrNoDefaults if templates file doesn't exist.
func LoadModuleTemplates(repoRoot string) (*ModuleTemplatesConfig, error) {
	data, err := loadDefaultFile(repoRoot, "module-templates.yml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading module templates: %w", err)
	}

	var cfg ModuleTemplatesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing module templates: %w", err)
	}

	return &cfg, nil
}

// LoadDiscoveryConventions loads discovery conventions from defaults.
// Falls back to built-in defaults if file doesn't exist.
func LoadDiscoveryConventions(repoRoot string) *DiscoveryConventions {
	// Try to load from repository defaults
	data, err := loadDefaultFile(repoRoot, "repository.yml")
	if err == nil {
		var wrapper struct {
			Conventions struct {
				ComponentDiscovery *DiscoveryConventions `yaml:"component_discovery"`
			} `yaml:"conventions"`
		}
		if yaml.Unmarshal(data, &wrapper) == nil && wrapper.Conventions.ComponentDiscovery != nil {
			return wrapper.Conventions.ComponentDiscovery
		}
	}

	// Return built-in defaults
	return DefaultDiscoveryConventions()
}

// DefaultDiscoveryConventions returns the built-in default discovery rules.
func DefaultDiscoveryConventions() *DiscoveryConventions {
	return &DiscoveryConventions{
		Gherkin: &DiscoveryRule{
			PathPattern:  "specs/{moniker}",
			RequiredFile: "specification.feature",
		},
		Structurizr: &DiscoveryRule{
			PathPattern:  "specs/{moniker}/.design",
			RequiredFile: "workspace.dsl",
		},
		// GherkinSteps inference rules:
		// - Module with go component: {go_root}/specs (requires godog_test.go)
		// - Module with typescript component: {ts_root}/features (requires cucumber runner)
		// - Module without code: go/eac/specs/{moniker}
		GherkinSteps: &GherkinStepsRule{
			DeriveFromComponents: map[string]string{
				"go":         "specs",
				"typescript": "features",
			},
			FallbackPattern: "go/eac/specs/{moniker}",
			RequiredFile:    "godog_test.go",
		},
		Markdown: &DeriveRule{
			DeriveFrom: []string{"go", "typescript", "dockerfile"},
		},
	}
}

// ExpandModuleFromTemplate expands a module definition using its referenced template.
// This merges template defaults into the module, substitutes parameters, and
// optionally discovers components based on filesystem conventions.
func ExpandModuleFromTemplate(
	mod *Module,
	templates map[string]ModuleTemplate,
	conventions *DiscoveryConventions,
	repoRoot string,
	owner string,
) error {
	// 1. If module references a template, merge template as base
	if mod.Template != "" {
		tmpl, ok := templates[mod.Template]
		if !ok {
			return fmt.Errorf("module %s: unknown template %q", mod.Moniker, mod.Template)
		}
		mergeTemplateIntoModule(mod, &tmpl)
	}

	// 2. Build parameters and substitute in all fields
	params := buildModuleParams(mod, owner)
	substituteModuleParams(mod, params)

	// 3. Resolve null/empty components using conventions
	if conventions != nil {
		resolveConventionalComponents(mod, conventions, repoRoot)
	}

	// 4. Always discover core BDD components (gherkin, structurizr, gherkin-steps)
	// These are auto-inferred based on filesystem conventions without requiring explicit config
	if conventions != nil {
		discoverCoreComponents(mod, conventions, repoRoot)
	}

	// 5. If auto_discover is enabled, scan filesystem for ALL additional components
	if mod.AutoDiscover && conventions != nil {
		discoverComponents(mod, conventions, repoRoot)
	}

	return nil
}

// mergeTemplateIntoModule applies template defaults to module.
// Module values take precedence over template values.
// Deep merge is performed for components - module can override specific fields.
func mergeTemplateIntoModule(mod *Module, tmpl *ModuleTemplate) {
	// Versioning: use template if module doesn't have it
	if mod.Versioning == nil && tmpl.Versioning != nil {
		mod.Versioning = &ModuleVersioning{
			Scheme:      tmpl.Versioning.Scheme,
			Changelog:   tmpl.Versioning.Changelog,
			ReleaseType: tmpl.Versioning.ReleaseType,
		}
	}

	// Components: deep merge - template provides base, module can override specific fields
	if mod.Components == nil {
		mod.Components = make(ModuleComponents)
	}
	for name, tmplComp := range tmpl.Components {
		modComp, exists := mod.Components[name]
		if !exists {
			// Component doesn't exist in module - copy from template
			mod.Components[name] = cloneComponentEntry(tmplComp)
		} else {
			// Component exists in module - deep merge template as base
			mod.Components[name] = mergeComponentEntries(tmplComp, modComp)
		}
	}

	// DependsOn: merge unique
	if len(tmpl.DependsOn) > 0 {
		existing := make(map[string]bool)
		for _, d := range mod.DependsOn {
			existing[d] = true
		}
		for _, d := range tmpl.DependsOn {
			if !existing[d] {
				mod.DependsOn = append(mod.DependsOn, d)
			}
		}
	}
}

// mergeComponentEntries merges two ComponentEntry structs.
// Module values override template values for non-zero fields.
func mergeComponentEntries(tmpl, mod *ComponentEntry) *ComponentEntry {
	if tmpl == nil {
		return mod
	}
	if mod == nil {
		return cloneComponentEntry(tmpl)
	}

	// Start with template as base
	result := cloneComponentEntry(tmpl)

	// Override with module values
	if mod.Type != "" {
		result.Type = mod.Type
	}
	if mod.Root != "" {
		result.Root = mod.Root
	}
	if mod.Patterns != nil {
		if result.Patterns == nil {
			result.Patterns = &ComponentPatterns{}
		}
		if mod.Patterns.Source != nil {
			result.Patterns.Source = make([]string, len(mod.Patterns.Source))
			copy(result.Patterns.Source, mod.Patterns.Source)
		}
		if mod.Patterns.Tests != nil {
			result.Patterns.Tests = make([]string, len(mod.Patterns.Tests))
			copy(result.Patterns.Tests, mod.Patterns.Tests)
		}
		if mod.Patterns.Config != nil {
			result.Patterns.Config = make([]string, len(mod.Patterns.Config))
			copy(result.Patterns.Config, mod.Patterns.Config)
		}
	}
	if mod.Build != nil {
		result.Build = mod.Build.Clone()
	}
	if mod.Amp != nil {
		result.Amp = &AmpConfig{
			Build: mod.Amp.Build,
			Lint:  mod.Amp.Lint,
			Test:  mod.Amp.Test,
			Scan:  mod.Amp.Scan,
		}
	}

	// Deep merge DockerBuild config
	if mod.DockerBuild != nil {
		result.DockerBuild = mergeDockerBuildConfigs(result.DockerBuild, mod.DockerBuild)
	}

	return result
}

// mergeDockerBuildConfigs merges DockerBuildConfig with module overrides.
// Module values override template values for non-zero/non-nil fields.
func mergeDockerBuildConfigs(tmpl, mod *DockerBuildConfig) *DockerBuildConfig {
	if tmpl == nil {
		if mod != nil {
			return mod.Clone()
		}
		return nil
	}
	if mod == nil {
		return tmpl.Clone()
	}

	// Start with template as base
	result := tmpl.Clone()

	// Override with module values (only non-zero values)
	if mod.Container != "" {
		result.Container = mod.Container
	}
	if mod.Context != "" {
		result.Context = mod.Context
	}
	if mod.Dockerfile != "" {
		result.Dockerfile = mod.Dockerfile
	}
	if len(mod.Platforms) > 0 {
		result.Platforms = make([]string, len(mod.Platforms))
		copy(result.Platforms, mod.Platforms)
	}
	if len(mod.Tags) > 0 {
		result.Tags = make([]string, len(mod.Tags))
		copy(result.Tags, mod.Tags)
	}
	// Push: can't differentiate between explicit false and unset, so only override if true
	if mod.Push {
		result.Push = mod.Push
	}
	if mod.Registry != "" {
		result.Registry = mod.Registry
	}
	if mod.Cache != nil {
		result.Cache = &DockerCacheConfig{
			Type: mod.Cache.Type,
			From: mod.Cache.From,
			To:   mod.Cache.To,
			Mode: mod.Cache.Mode,
		}
	}
	if mod.SBOM {
		result.SBOM = mod.SBOM
	}
	if mod.Provenance {
		result.Provenance = mod.Provenance
	}

	return result
}

// cloneComponentEntry creates a deep copy of a ComponentEntry.
func cloneComponentEntry(entry *ComponentEntry) *ComponentEntry {
	if entry == nil {
		return nil
	}
	clone := &ComponentEntry{
		Type:     entry.Type,
		Root:     entry.Root,
		Resolved: entry.Resolved,
	}
	if entry.Patterns != nil {
		clone.Patterns = &ComponentPatterns{}
		if entry.Patterns.Source != nil {
			clone.Patterns.Source = make([]string, len(entry.Patterns.Source))
			copy(clone.Patterns.Source, entry.Patterns.Source)
		}
		if entry.Patterns.Tests != nil {
			clone.Patterns.Tests = make([]string, len(entry.Patterns.Tests))
			copy(clone.Patterns.Tests, entry.Patterns.Tests)
		}
		if entry.Patterns.Config != nil {
			clone.Patterns.Config = make([]string, len(entry.Patterns.Config))
			copy(clone.Patterns.Config, entry.Patterns.Config)
		}
	}
	if entry.Build != nil {
		clone.Build = entry.Build.Clone()
	}
	if entry.DockerBuild != nil {
		clone.DockerBuild = entry.DockerBuild.Clone()
	}
	if entry.Amp != nil {
		clone.Amp = &AmpConfig{
			Build: entry.Amp.Build,
			Lint:  entry.Amp.Lint,
			Test:  entry.Amp.Test,
			Scan:  entry.Amp.Scan,
		}
	}
	return clone
}

// buildModuleParams creates parameter map for substitution.
func buildModuleParams(mod *Module, owner string) map[string]string {
	params := map[string]string{
		"moniker": mod.Moniker,
		"owner":   owner,
	}

	// Add explicit parameters from module
	for k, v := range mod.Parameters {
		params[k] = v
	}

	// Infer go_root from go component if present and not already set
	if _, hasGoRoot := params["go_root"]; !hasGoRoot {
		if goComp, ok := mod.Components["go"]; ok && goComp != nil && goComp.Root != "" {
			params["go_root"] = goComp.Root
		}
	}

	return params
}

// substituteModuleParams replaces {placeholder} in all string fields.
func substituteModuleParams(mod *Module, params map[string]string) {
	// Substitute in versioning
	if mod.Versioning != nil {
		mod.Versioning.Changelog = substituteParams(mod.Versioning.Changelog, params)
	}

	// Substitute in components
	for name, comp := range mod.Components {
		if comp == nil {
			continue
		}
		comp.Root = substituteParams(comp.Root, params)
		if comp.Patterns != nil {
			comp.Patterns.Source = substituteParamsSlice(comp.Patterns.Source, params)
			comp.Patterns.Tests = substituteParamsSlice(comp.Patterns.Tests, params)
			comp.Patterns.Config = substituteParamsSlice(comp.Patterns.Config, params)
		}
		if comp.DockerBuild != nil {
			substituteDockerBuildParams(comp.DockerBuild, params)
		}
		mod.Components[name] = comp
	}
}

// substituteParams replaces {placeholder} patterns in a string.
func substituteParams(s string, params map[string]string) string {
	if s == "" {
		return s
	}
	result := s
	for key, value := range params {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}
	return result
}

// substituteParamsSlice replaces {placeholder} patterns in a slice of strings.
func substituteParamsSlice(slice []string, params map[string]string) []string {
	if slice == nil {
		return nil
	}
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = substituteParams(s, params)
	}
	return result
}

// substituteDockerBuildParams substitutes parameters in DockerBuildConfig fields.
func substituteDockerBuildParams(dbc *DockerBuildConfig, params map[string]string) {
	if dbc == nil {
		return
	}
	dbc.Container = substituteParams(dbc.Container, params)
	dbc.Context = substituteParams(dbc.Context, params)
	dbc.Dockerfile = substituteParams(dbc.Dockerfile, params)
	dbc.Tags = substituteParamsSlice(dbc.Tags, params)
	if dbc.Cache != nil {
		dbc.Cache.From = substituteParams(dbc.Cache.From, params)
		dbc.Cache.To = substituteParams(dbc.Cache.To, params)
	}
}

// resolveConventionalComponents resolves null/marker components using conventions.
// A component with ConventionMarker=true or nil entry is resolved from conventions.
func resolveConventionalComponents(mod *Module, conv *DiscoveryConventions, repoRoot string) {
	for name, comp := range mod.Components {
		// Skip components that already have a root or are explicitly configured
		if comp != nil && comp.Root != "" {
			continue
		}

		var path string

		// Handle gherkin-steps specially - it derives from code components
		if name == "gherkin-steps" {
			path = getGherkinStepsPath(mod.Components, mod.Moniker, conv, repoRoot)
		} else {
			// Check if this is a convention-based component
			path = getConventionalPath(name, mod.Moniker, conv)
			if path == "" {
				// Try deriving from other components
				path = derivePathFrom(mod.Components, name, conv)
			}
		}

		if path == "" {
			// No convention found, remove the nil marker
			if comp == nil {
				delete(mod.Components, name)
			}
			continue
		}

		// Check if path exists (for gherkin-steps, this was already checked)
		if name != "gherkin-steps" && !pathExists(repoRoot, path) {
			// Path doesn't exist, remove the nil marker
			if comp == nil {
				delete(mod.Components, name)
			}
			continue
		}

		// Path exists, set it
		if comp == nil {
			mod.Components[name] = &ComponentEntry{Root: path}
		} else {
			comp.Root = path
		}
	}
}

// getConventionalPath returns the conventional path for a component type.
func getConventionalPath(compName, moniker string, conv *DiscoveryConventions) string {
	if conv == nil {
		return ""
	}

	var rule *DiscoveryRule
	switch compName {
	case "gherkin":
		rule = conv.Gherkin
	case "structurizr":
		rule = conv.Structurizr
	}

	if rule == nil || rule.PathPattern == "" {
		return ""
	}

	return strings.ReplaceAll(rule.PathPattern, "{moniker}", moniker)
}

// getGherkinStepsPath derives the gherkin-steps path based on module components.
// Inference rules:
// - Module with go component: {go_root}/specs
// - Module with typescript component: {ts_root}/features
// - Module without code: go/eac/specs/{moniker}
// Returns empty string if path doesn't exist or no rule applies.
func getGherkinStepsPath(components ModuleComponents, moniker string, conv *DiscoveryConventions, repoRoot string) string {
	if conv == nil || conv.GherkinSteps == nil {
		return ""
	}

	rule := conv.GherkinSteps

	// Try deriving from code components (in priority order: go, then typescript)
	for _, compType := range []string{"go", "typescript"} {
		if subdir, ok := rule.DeriveFromComponents[compType]; ok {
			if entry, exists := components[compType]; exists && entry != nil && entry.Root != "" {
				path := filepath.ToSlash(filepath.Join(entry.Root, subdir))
				if pathHasRequiredFile(repoRoot, path, rule.RequiredFile) {
					return path
				}
			}
		}
	}

	// Fallback for modules without code components
	if rule.FallbackPattern != "" {
		path := strings.ReplaceAll(rule.FallbackPattern, "{moniker}", moniker)
		if pathHasRequiredFile(repoRoot, path, rule.RequiredFile) {
			return path
		}
	}

	return ""
}

// derivePathFrom attempts to derive a component's path from another component.
func derivePathFrom(components ModuleComponents, targetName string, conv *DiscoveryConventions) string {
	if conv == nil {
		return ""
	}

	var rule *DeriveRule
	switch targetName {
	case "markdown":
		rule = conv.Markdown
	case "yaml":
		rule = conv.YAML
	}

	if rule == nil || len(rule.DeriveFrom) == 0 {
		return ""
	}

	// Try each source component in order
	for _, sourceName := range rule.DeriveFrom {
		if entry, ok := components[sourceName]; ok && entry != nil && entry.Root != "" {
			return entry.Root
		}
	}

	return ""
}

// discoverCoreComponents auto-discovers gherkin, structurizr, and gherkin-steps components.
// These core BDD components are ALWAYS inferred based on filesystem conventions,
// unlike other components which require auto_discover: true.
func discoverCoreComponents(mod *Module, conv *DiscoveryConventions, repoRoot string) {
	if conv == nil {
		return
	}

	// Ensure Components map exists
	if mod.Components == nil {
		mod.Components = make(ModuleComponents)
	}

	// Discover gherkin specs from specs/{moniker}/
	// Features are in subdirectories (e.g., specs/core/cache-invalidation/specification.feature)
	// So we check for directory existence and any .feature files
	if !mod.Components.HasComponent("gherkin") && conv.Gherkin != nil {
		path := strings.ReplaceAll(conv.Gherkin.PathPattern, "{moniker}", mod.Moniker)
		if pathHasFeatureFiles(repoRoot, path) {
			mod.Components["gherkin"] = &ComponentEntry{Root: path}
		}
	}

	// Discover structurizr designs from specs/{moniker}/.design/
	if !mod.Components.HasComponent("structurizr") && conv.Structurizr != nil {
		path := strings.ReplaceAll(conv.Structurizr.PathPattern, "{moniker}", mod.Moniker)
		if pathHasRequiredFile(repoRoot, path, conv.Structurizr.RequiredFile) {
			mod.Components["structurizr"] = &ComponentEntry{Root: path}
		}
	}

	// Discover gherkin-steps based on code component location
	if !mod.Components.HasComponent("gherkin-steps") && conv.GherkinSteps != nil {
		path := getGherkinStepsPath(mod.Components, mod.Moniker, conv, repoRoot)
		if path != "" {
			mod.Components["gherkin-steps"] = &ComponentEntry{Root: path}
		}
	}
}

// discoverComponents scans filesystem for conventional components not already defined.
func discoverComponents(mod *Module, conv *DiscoveryConventions, repoRoot string) {
	if conv == nil {
		return
	}

	// Try to discover gherkin
	if !mod.Components.HasComponent("gherkin") && conv.Gherkin != nil {
		path := strings.ReplaceAll(conv.Gherkin.PathPattern, "{moniker}", mod.Moniker)
		if pathHasRequiredFile(repoRoot, path, conv.Gherkin.RequiredFile) {
			mod.Components["gherkin"] = &ComponentEntry{Root: path}
		}
	}

	// Try to discover structurizr
	if !mod.Components.HasComponent("structurizr") && conv.Structurizr != nil {
		path := strings.ReplaceAll(conv.Structurizr.PathPattern, "{moniker}", mod.Moniker)
		if pathHasRequiredFile(repoRoot, path, conv.Structurizr.RequiredFile) {
			mod.Components["structurizr"] = &ComponentEntry{Root: path}
		}
	}

	// Try to discover gherkin-steps
	if !mod.Components.HasComponent("gherkin-steps") && conv.GherkinSteps != nil {
		path := getGherkinStepsPath(mod.Components, mod.Moniker, conv, repoRoot)
		if path != "" {
			mod.Components["gherkin-steps"] = &ComponentEntry{Root: path}
		}
	}

	// Try to derive markdown from primary component
	if !mod.Components.HasComponent("markdown") && conv.Markdown != nil {
		path := derivePathFrom(mod.Components, "markdown", conv)
		if path != "" && pathExists(repoRoot, path) {
			mod.Components["markdown"] = &ComponentEntry{Root: path}
		}
	}
}

// pathExists checks if a path exists in the repo.
func pathExists(repoRoot, path string) bool {
	fullPath := filepath.Join(repoRoot, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// pathHasFeatureFiles checks if a path contains any .feature files (including subdirectories).
// This is used for gherkin component discovery where features are in subdirectories.
func pathHasFeatureFiles(repoRoot, path string) bool {
	fullPath := filepath.Join(repoRoot, path)

	// Check if directory exists
	info, err := os.Stat(fullPath)
	if err != nil || !info.IsDir() {
		return false
	}

	// Check for .feature files in immediate subdirectories
	// (features are organized as specs/{moniker}/{feature-name}/specification.feature)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			featurePath := filepath.Join(fullPath, entry.Name(), "specification.feature")
			if _, err := os.Stat(featurePath); err == nil {
				return true
			}
		}
	}

	return false
}

// pathHasRequiredFile checks if a path contains the required file.
func pathHasRequiredFile(repoRoot, path, requiredFile string) bool {
	if requiredFile == "" {
		return pathExists(repoRoot, path)
	}
	fullPath := filepath.Join(repoRoot, path, requiredFile)
	_, err := os.Stat(fullPath)
	return err == nil
}
