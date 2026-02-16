package services

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
)

// ============================================================================
// Config Adapter - wraps *config.EACConfig to implement ConfigPort
// ============================================================================

// configAdapter wraps *config.EACConfig to implement core.ConfigPort.
type configAdapter struct {
	cfg *config.EACConfig
}

// Compile-time check
var _ core.ConfigPort = (*configAdapter)(nil)

func (a *configAdapter) GetRepoRoot() string   { return a.cfg.RepoRoot }
func (a *configAdapter) GetConfigRoot() string { return a.cfg.ConfigRoot }

func (a *configAdapter) GetRepository() core.RepositoryConfigPort {
	if a.cfg.Repository == nil {
		return nil
	}
	return &repositoryAdapter{cfg: a.cfg}
}

func (a *configAdapter) GetEnvironments() core.EnvironmentsConfigPort {
	if a.cfg.Environments == nil {
		return nil
	}
	return &environmentsAdapter{envs: a.cfg.Environments}
}

func (a *configAdapter) GetTestingTags() core.TestingTagsConfigPort {
	if a.cfg.TestingTags == nil {
		return nil
	}
	return &testingTagsAdapter{tags: a.cfg.TestingTags}
}

func (a *configAdapter) GetTestSuites() core.TestSuitesConfigPort {
	if a.cfg.TestSuites == nil {
		return nil
	}
	return &testSuitesAdapter{suites: a.cfg.TestSuites}
}

func (a *configAdapter) GetComponentKinds() core.ComponentKindsConfigPort {
	if a.cfg.ComponentKinds == nil {
		return nil
	}
	return &componentKindsAdapter{kinds: a.cfg.ComponentKinds}
}

// repositoryAdapter wraps config.RepositoryConfig
type repositoryAdapter struct {
	cfg *config.EACConfig
}

func (a *repositoryAdapter) AllMonikers() []string {
	monikers := make([]string, len(a.cfg.Repository.Modules))
	for i := range a.cfg.Repository.Modules {
		monikers[i] = a.cfg.Repository.Modules[i].Moniker
	}
	return monikers
}

func (a *repositoryAdapter) GetModule(moniker string) (core.ModuleContractPort, bool) {
	for i := range a.cfg.Repository.Modules {
		if a.cfg.Repository.Modules[i].Moniker == moniker {
			return &moduleAdapter{module: &a.cfg.Repository.Modules[i], wsRoot: a.cfg.RepoRoot}, true
		}
	}
	return nil, false
}

func (a *repositoryAdapter) GetBuildOutputPath(workspaceRoot, moniker string) string {
	return paths.BuildOutputPath(workspaceRoot, moniker)
}

func (a *repositoryAdapter) GetTestOutputPath(workspaceRoot, moniker string) string {
	return paths.TestOutputPath(workspaceRoot, moniker)
}

func (a *repositoryAdapter) GetScanOutputPath(workspaceRoot, moniker string) string {
	return paths.UnitScanOutputPath(workspaceRoot, moniker, "")
}

func (a *repositoryAdapter) GetLintOutputPath(workspaceRoot, moniker string) string {
	return paths.LintOutputPath(workspaceRoot, moniker)
}

func (a *repositoryAdapter) TestOutputDir() string {
	return a.cfg.Repository.Paths.Out.Test
}

func (a *repositoryAdapter) GetBuildArtifacts(moniker string, includeAll bool) []core.ArtifactPort {
	artifacts := a.cfg.GetBuildArtifacts(moniker, includeAll)
	result := make([]core.ArtifactPort, len(artifacts))
	for i := range artifacts {
		result[i] = &artifactAdapter{a: &artifacts[i]}
	}
	return result
}

func (a *repositoryAdapter) GetBuildArtifactIDs(moniker string, buildAll bool) []string {
	return a.cfg.GetBuildArtifactIDs(moniker, buildAll)
}

func (a *repositoryAdapter) GetBooksByModule(moniker string) []core.BookConfigPort {
	books := a.cfg.GetBooksByModule(moniker)
	result := make([]core.BookConfigPort, len(books))
	for i := range books {
		result[i] = &bookAdapter{b: books[i]}
	}
	return result
}

func (a *repositoryAdapter) GetDefaultBooksByModule(moniker string) []core.BookConfigPort {
	books := a.cfg.GetDefaultBooksByModule(moniker)
	result := make([]core.BookConfigPort, len(books))
	for i := range books {
		result[i] = &bookAdapter{b: books[i]}
	}
	return result
}

// moduleAdapter wraps config.Module for minimal interface compliance
type moduleAdapter struct {
	module *config.Module
	wsRoot string
}

func (a *moduleAdapter) GetMoniker() string                       { return a.module.Moniker }
func (a *moduleAdapter) GetName() string                          { return a.module.Moniker }
func (a *moduleAdapter) GetDescription() string                   { return a.module.Description }
func (a *moduleAdapter) GetGroup() string                          { return a.module.Group }
func (a *moduleAdapter) HasComponent(compType string) bool        { return a.module.Components != nil && a.module.Components[compType] != nil }
func (a *moduleAdapter) GetComponentRoot(compType string) string  {
	if a.module.Components == nil {
		return ""
	}
	if entry := a.module.Components[compType]; entry != nil {
		return entry.Root
	}
	return ""
}
func (a *moduleAdapter) GetComponentRoots() map[string]string {
	if a.module.Components == nil {
		return nil
	}
	roots := make(map[string]string)
	for name, entry := range a.module.Components {
		if entry != nil {
			roots[name] = entry.Root
		}
	}
	return roots
}
func (a *moduleAdapter) GetComponentTypesDisplay() string { return "" }
func (a *moduleAdapter) GetComponentGroup(componentName string) string {
	if a.module.Components == nil {
		return ""
	}
	entry := a.module.Components[componentName]
	if entry == nil {
		return ""
	}
	return entry.ComponentGroup
}
func (a *moduleAdapter) GetComponentAmp(compName, op string) float64 { return 1.0 }
func (a *moduleAdapter) GetDependsOn() []string           { return a.module.DependsOn }
func (a *moduleAdapter) GetVersioningScheme() string {
	if a.module.Versioning == nil {
		return ""
	}
	return a.module.Versioning.Scheme
}
func (a *moduleAdapter) GetReleaseType() string {
	if a.module.Versioning == nil {
		return ""
	}
	return a.module.Versioning.ReleaseType
}
func (a *moduleAdapter) GetChangelog() string {
	if a.module.Versioning == nil {
		return ""
	}
	return a.module.Versioning.Changelog
}
func (a *moduleAdapter) HasVersioning() bool { return a.module.Versioning != nil }
func (a *moduleAdapter) GetMetadata() map[string]interface{} {
	if a.module.Metadata == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range a.module.Metadata {
		result[k] = v
	}
	return result
}
func (a *moduleAdapter) GetContentHash() (string, error) { return "", nil }

// artifactAdapter wraps config.Artifact
type artifactAdapter struct {
	a *config.Artifact
}

func (a *artifactAdapter) GetID() string      { return a.a.ID }
func (a *artifactAdapter) GetType() string    { return a.a.Type }
func (a *artifactAdapter) GetPattern() string { return a.a.Pattern }

// bookAdapter wraps config.Book
type bookAdapter struct {
	b *config.Book
}

func (a *bookAdapter) GetName() string { return a.b.Name }
func (a *bookAdapter) GetType() string { return a.b.Template } // Template serves as type
func (a *bookAdapter) GetRoot() string { return "" }           // Books don't have root in this sense
