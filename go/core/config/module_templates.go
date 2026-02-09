package config

import (
	"fmt"
	"strings"
)

// ModuleTemplate defines a reusable module structure with placeholders.
// Templates are defined in defaults and referenced by modules via the "template" field.
// Templates provide versioning defaults and a discover list of named discovery rules
// for auto-discovering auxiliary components (markdown, yaml, gherkin, etc.).
type ModuleTemplate struct {
	// Versioning provides default versioning configuration
	Versioning *ModuleVersioning `yaml:"versioning,omitempty"`
	// Discover lists named discovery rules to run for this template.
	// Rules are resolved from the discovery-rules section of blueprints.yml.
	// Each rule auto-discovers an auxiliary component if its filesystem check passes.
	Discover []string `yaml:"discover,omitempty"`
	// DependsOn provides default dependencies
	DependsOn []string `yaml:"depends_on,omitempty"`
	// Components provides default component entries (e.g., docker_build defaults).
	// Template components are merged into module components as defaults —
	// module values take precedence over template values.
	Components map[string]*ComponentEntry `yaml:"components,omitempty"`
}

// BlueprintsConfig holds all blueprints loaded from defaults.
// This is the unified config: component kinds (tool bindings), module templates,
// discovery rules, and artifact matrices — all in one file.
type BlueprintsConfig struct {
	// ComponentKinds maps component kind name to its definition (tool bindings, default patterns, resources).
	// Component kinds define tool bindings, default file patterns, and resources.
	ComponentKinds map[string]*ComponentType `yaml:"component-kinds"`
	// DiscoveryRules maps rule name to its discovery specification.
	// Templates reference these by name via their "discover" field.
	DiscoveryRules map[string]ComponentDiscoveryRule `yaml:"discovery-rules"`
	// Templates maps template name to its specification
	Templates map[string]ModuleTemplate `yaml:"module-templates"`
	// ArtifactMatrices maps matrix name to its definition
	ArtifactMatrices map[string]*ArtifactMatrix `yaml:"artifact-matrices"`
}

// BlueprintsFileName is the config file for blueprints.
const BlueprintsFileName = "blueprints.yml"

// LoadBlueprintsDefaults loads blueprints from contract defaults.
func LoadBlueprintsDefaults(repoRoot string) (*BlueprintsConfig, error) {
	cfg, err := cloneBlueprintsDefaults()
	if err != nil {
		return nil, fmt.Errorf("loading blueprints defaults: %w", err)
	}
	return cfg, nil
}

// MergeBlueprintsConfig merges override into base. Override keys replace base keys.
// Either argument may be nil.
func MergeBlueprintsConfig(base, override *BlueprintsConfig) *BlueprintsConfig {
	if base == nil && override == nil {
		return &BlueprintsConfig{}
	}
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := &BlueprintsConfig{
		ComponentKinds:   make(map[string]*ComponentType),
		DiscoveryRules:   make(map[string]ComponentDiscoveryRule),
		Templates:        make(map[string]ModuleTemplate),
		ArtifactMatrices: make(map[string]*ArtifactMatrix),
	}

	// Copy base component kinds, then override
	for k, v := range base.ComponentKinds {
		result.ComponentKinds[k] = v
	}
	for k, v := range override.ComponentKinds {
		result.ComponentKinds[k] = v
	}

	// Copy base discovery rules
	for k, v := range base.DiscoveryRules {
		result.DiscoveryRules[k] = v
	}
	// Override discovery rules
	for k, v := range override.DiscoveryRules {
		result.DiscoveryRules[k] = v
	}

	// Copy base templates
	for k, v := range base.Templates {
		result.Templates[k] = v
	}
	// Override templates
	for k, v := range override.Templates {
		result.Templates[k] = v
	}

	// Copy base artifact matrices
	for k, v := range base.ArtifactMatrices {
		result.ArtifactMatrices[k] = v
	}
	// Override artifact matrices
	for k, v := range override.ArtifactMatrices {
		result.ArtifactMatrices[k] = v
	}

	return result
}

// ExpandModuleFromTemplate expands a module definition using its referenced template.
// Three steps: merge template defaults → substitute params → discover components from rules.
func ExpandModuleFromTemplate(
	mod *Module,
	templates map[string]ModuleTemplate,
	discoveryRules []ComponentDiscoveryRule,
	repoRoot string,
	discoveryVars map[string]string,
	namedRules map[string]ComponentDiscoveryRule,
) error {
	// 1. If module references a template, merge defaults and run template discovery
	if mod.Template != "" {
		tmpl, ok := templates[mod.Template]
		if !ok {
			return fmt.Errorf("module %s: unknown template %q", mod.Moniker, mod.Template)
		}

		// 1a. Merge versioning + depends_on from template (module values win)
		mergeTemplateIntoModule(mod, &tmpl)

		// 1b. Resolve template's discover list into rules and run them
		if len(tmpl.Discover) > 0 {
			templateRules := resolveDiscoverList(tmpl.Discover, namedRules)
			discoverComponentsFromRules(mod, templateRules, repoRoot, discoveryVars)
		}
	}

	// 2. Build parameters and substitute in all fields
	// Discovery vars provide {name_root} for every component (e.g., go_root, typescript_root)
	params := buildModuleParams(mod, discoveryVars)
	substituteModuleParams(mod, params)

	// 3. Discover components from global YAML-driven rules (with template filtering)
	discoverComponentsFromRules(mod, discoveryRules, repoRoot, discoveryVars)

	return nil
}

// resolveDiscoverList resolves a list of discovery rule names into actual rules.
// Unknown names are silently skipped.
func resolveDiscoverList(names []string, namedRules map[string]ComponentDiscoveryRule) []ComponentDiscoveryRule {
	if len(namedRules) == 0 {
		return nil
	}
	var rules []ComponentDiscoveryRule
	for _, name := range names {
		if rule, ok := namedRules[name]; ok {
			rules = append(rules, rule)
		}
	}
	return rules
}

// mergeTemplateIntoModule applies template defaults to module.
// Module values take precedence over template values.
func mergeTemplateIntoModule(mod *Module, tmpl *ModuleTemplate) {
	// Versioning: use template if module doesn't have it
	if mod.Versioning == nil && tmpl.Versioning != nil {
		mod.Versioning = &ModuleVersioning{
			Scheme:      tmpl.Versioning.Scheme,
			Changelog:   tmpl.Versioning.Changelog,
			ReleaseType: tmpl.Versioning.ReleaseType,
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

	// Components: merge template component defaults into module components.
	// Only merges into components the module already declares (template
	// components are defaults, not auto-creation). Root is never overridden
	// from template (it's always module-specific). DockerBuild fields from
	// the template are applied as defaults — module fields take precedence.
	if len(tmpl.Components) > 0 && mod.Components != nil {
		for name, tmplComp := range tmpl.Components {
			if tmplComp == nil {
				continue
			}
			modComp, exists := mod.Components[name]
			if !exists {
				continue
			}
			if modComp == nil {
				modComp = &ComponentEntry{}
				mod.Components[name] = modComp
			}
			if tmplComp.DockerBuild != nil {
				if modComp.DockerBuild == nil {
					// Module has no DockerBuild — use template's entirely
					modComp.DockerBuild = cloneDockerBuildConfig(tmplComp.DockerBuild)
				} else {
					// Module has DockerBuild — merge template fields as defaults
					mergeDockerBuildDefaults(modComp.DockerBuild, tmplComp.DockerBuild)
				}
			}
		}
	}
}

// cloneDockerBuildConfig creates a deep copy of a DockerBuildConfig.
func cloneDockerBuildConfig(src *DockerBuildConfig) *DockerBuildConfig {
	if src == nil {
		return nil
	}
	dst := &DockerBuildConfig{
		Container:  src.Container,
		Context:    src.Context,
		Dockerfile: src.Dockerfile,
		Builder:    src.Builder,
		Load:       src.Load,
		Registry:   src.Registry,
		SBOM:       src.SBOM,
		Provenance: src.Provenance,
	}
	if src.Platforms != nil {
		dst.Platforms = make([]string, len(src.Platforms))
		copy(dst.Platforms, src.Platforms)
	}
	if src.Tags != nil {
		dst.Tags = make([]string, len(src.Tags))
		copy(dst.Tags, src.Tags)
	}
	if src.Push != nil {
		push := *src.Push
		dst.Push = &push
	}
	if src.Cache != nil {
		dst.Cache = &DockerCacheConfig{
			Type:  src.Cache.Type,
			Scope: src.Cache.Scope,
			From:  src.Cache.From,
			To:    src.Cache.To,
			Mode:  src.Cache.Mode,
		}
	}
	return dst
}

// mergeDockerBuildDefaults applies template DockerBuild fields as defaults
// into the module's DockerBuild. Module fields take precedence (non-zero values win).
func mergeDockerBuildDefaults(mod, tmpl *DockerBuildConfig) {
	if mod.Container == "" {
		mod.Container = tmpl.Container
	}
	if mod.Context == "" {
		mod.Context = tmpl.Context
	}
	if mod.Dockerfile == "" {
		mod.Dockerfile = tmpl.Dockerfile
	}
	if mod.Builder == "" {
		mod.Builder = tmpl.Builder
	}
	if len(mod.Platforms) == 0 && len(tmpl.Platforms) > 0 {
		mod.Platforms = make([]string, len(tmpl.Platforms))
		copy(mod.Platforms, tmpl.Platforms)
	}
	if len(mod.Tags) == 0 && len(tmpl.Tags) > 0 {
		mod.Tags = make([]string, len(tmpl.Tags))
		copy(mod.Tags, tmpl.Tags)
	}
	if mod.Push == nil && tmpl.Push != nil {
		push := *tmpl.Push
		mod.Push = &push
	}
	if mod.Registry == "" {
		mod.Registry = tmpl.Registry
	}
	if mod.Cache == nil && tmpl.Cache != nil {
		mod.Cache = &DockerCacheConfig{
			Type:  tmpl.Cache.Type,
			Scope: tmpl.Cache.Scope,
			From:  tmpl.Cache.From,
			To:    tmpl.Cache.To,
			Mode:  tmpl.Cache.Mode,
		}
	}
}

// buildModuleParams creates parameter map for substitution.
// Discovery vars already contain {name_root} for every component (e.g., go_root, core_root),
// so no special inference is needed — component roots ARE the parameters.
func buildModuleParams(mod *Module, discoveryVars map[string]string) map[string]string {
	params := map[string]string{
		"moniker": mod.Moniker,
	}

	// Merge discovery vars (provides owner, component roots, conventions)
	for k, v := range discoveryVars {
		params[k] = v
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
			comp.Patterns.Data = substituteParamsSlice(comp.Patterns.Data, params)
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
