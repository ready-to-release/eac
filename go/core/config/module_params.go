package config

import (
	"strings"
)

// SubstituteModuleParams builds parameter map and substitutes {placeholder} in all module fields.
func SubstituteModuleParams(mod *Module, discoveryVars map[string]string) {
	params := buildModuleParams(mod, discoveryVars)
	substituteModuleParams(mod, params)
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

// mergeDockerBuildDefaults applies DockerBuild fields as defaults.
// Module fields take precedence (non-zero values win).
func mergeDockerBuildDefaults(mod, defaults *DockerBuildConfig) {
	if mod.Container == "" {
		mod.Container = defaults.Container
	}
	if mod.Context == "" {
		mod.Context = defaults.Context
	}
	if mod.Dockerfile == "" {
		mod.Dockerfile = defaults.Dockerfile
	}
	if mod.Builder == "" {
		mod.Builder = defaults.Builder
	}
	if len(mod.Platforms) == 0 && len(defaults.Platforms) > 0 {
		mod.Platforms = make([]string, len(defaults.Platforms))
		copy(mod.Platforms, defaults.Platforms)
	}
	if len(mod.Tags) == 0 && len(defaults.Tags) > 0 {
		mod.Tags = make([]string, len(defaults.Tags))
		copy(mod.Tags, defaults.Tags)
	}
	if mod.Push == nil && defaults.Push != nil {
		push := *defaults.Push
		mod.Push = &push
	}
	if mod.Registry == "" {
		mod.Registry = defaults.Registry
	}
	if mod.Cache == nil && defaults.Cache != nil {
		mod.Cache = &DockerCacheConfig{
			Type:  defaults.Cache.Type,
			Scope: defaults.Cache.Scope,
			From:  defaults.Cache.From,
			To:    defaults.Cache.To,
			Mode:  defaults.Cache.Mode,
		}
	}
}

// buildModuleParams creates parameter map for substitution.
func buildModuleParams(mod *Module, discoveryVars map[string]string) map[string]string {
	params := map[string]string{
		"moniker": mod.Moniker,
	}
	for k, v := range discoveryVars {
		params[k] = v
	}
	return params
}

// substituteModuleParams replaces {placeholder} in all string fields.
func substituteModuleParams(mod *Module, params map[string]string) {
	if mod.Versioning != nil {
		mod.Versioning.Changelog = substituteParams(mod.Versioning.Changelog, params)
	}
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
