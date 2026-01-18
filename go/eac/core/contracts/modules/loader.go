package modules

import (
	"fmt"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

// LoadFromWorkspace loads all module contracts from the workspace.
// This is the main entry point for loading module contracts.
// Uses the central config package for loading and schema validation.
func LoadFromWorkspace(workspaceRoot string) (*Registry, error) {
	return loadModules(workspaceRoot, false)
}

// LoadFromWorkspaceNoValidation loads module contracts without schema validation.
// Use this for testing or when schemas are not available.
func LoadFromWorkspaceNoValidation(workspaceRoot string) (*Registry, error) {
	return loadModules(workspaceRoot, true)
}

func loadModules(workspaceRoot string, noValidation bool) (*Registry, error) {
	validate := !noValidation

	// Load using central config
	opts := config.LoadOptions{
		RepoRoot:        workspaceRoot,
		ValidateSchemas: validate,
		LazyLoad:        true, // We only need modules and component-types
	}

	cfg, err := config.Load(opts)
	if err != nil {
		return nil, contracts.NewContractError("load", "config", err, "failed to initialize config loader")
	}

	// Load repository config (includes modules)
	if err := cfg.LoadRepository(validate); err != nil {
		repoPath := filepath.Join(config.EACConfigRelPath, config.RepositoryFileName)
		return nil, contracts.NewContractError("load", repoPath, err, "failed to load repository.yml")
	}

	// Load component types for component-specific defaults
	if err := cfg.LoadComponentTypes(validate); err != nil {
		// Component types are optional - continue with defaults
	}

	// Apply component-specific defaults with repository path variables
	cfg.Repository.ApplyComponentDefaults(cfg.ComponentTypes)

	// Create registry (version kept for internal compatibility)
	registry := NewRegistry("0.1.0", workspaceRoot)

	// Convert config.Module to contracts.BaseContract and process
	for _, m := range cfg.Repository.Modules {
		// Convert to BaseContract for ModuleContract creation
		base := contracts.BaseContract{
			Moniker:       m.Moniker,
			Name:          m.Name,
			Description:   m.Description,
			DependsOn:     m.DependsOn,
			Metadata:      m.Metadata,
			EvidenceBooks: m.EvidenceBooks,
		}

		// Convert Components config
		if m.Components != nil {
			base.Components = make(contracts.ModuleComponents)
			for compName, entry := range m.Components {
				if entry != nil {
					base.Components[compName] = &contracts.ComponentEntry{
						Type:     entry.Type,
						Root:     entry.Root,
						Resolved: entry.Resolved,
					}
					if entry.Patterns != nil {
						base.Components[compName].Patterns = &contracts.ComponentPatterns{
							Source: entry.Patterns.Source,
							Tests:  entry.Patterns.Tests,
							Config: entry.Patterns.Config,
						}
					}
					// Convert Build config
					if entry.Build != nil {
						base.Components[compName].Build = &contracts.ComponentBuild{
							Handler: entry.Build.Handler,
						}
						for _, a := range entry.Build.Artifacts {
							base.Components[compName].Build.Artifacts = append(base.Components[compName].Build.Artifacts, contracts.ComponentArtifact{
								ID:          a.ID,
								Type:        a.Type,
								Pattern:     a.Pattern,
								Compression: a.Compression,
								DeriveFrom:  a.DeriveFrom,
							})
						}
					}
					// Convert DockerBuild from typed struct to map
					if entry.DockerBuild != nil {
						base.Components[compName].DockerBuild = config.DockerBuildConfigToMap(entry.DockerBuild)
					}
				} else {
					base.Components[compName] = nil
				}
			}
		}

		// Convert Versioning config if present
		if m.Versioning != nil {
			base.Versioning = &contracts.ModuleVersioning{
				Scheme:    m.Versioning.Scheme,
				Current:   m.Versioning.Current,
				Changelog: m.Versioning.Changelog,
			}
		}

		// Convert ReleaseBundle config if present
		if m.ReleaseBundle != nil {
			base.ReleaseBundle = &contracts.ReleaseBundle{
				TitleFormat: m.ReleaseBundle.TitleFormat,
				Headline:    m.ReleaseBundle.Headline,
			}
			for _, cat := range m.ReleaseBundle.Categories {
				base.ReleaseBundle.Categories = append(base.ReleaseBundle.Categories, contracts.ReleaseBundleCategory{
					Name:        cat.Name,
					Description: cat.Description,
					Modules:     cat.Modules,
				})
			}
		}

		// Note: Defaults are already applied by config.RepositoryConfig.applyModuleDefaults() and ApplyComponentDefaults()

		// Validate required fields
		if base.Moniker == "" {
			repoPath := filepath.Join(config.EACConfigRelPath, config.RepositoryFileName)
			return nil, contracts.NewContractError("validate", repoPath, nil, "moniker field is required")
		}

		// Create module contract
		module := NewModuleContract(base, workspaceRoot)

		// Add to registry
		if err := registry.Add(module); err != nil {
			return nil, contracts.NewContractError("add", config.RepositoryFileName, err,
				fmt.Sprintf("failed to add module '%s' to registry: %v", base.Moniker, err))
		}
	}

	// Validate registry has at least one module
	if registry.Count() == 0 {
		return nil, contracts.NewContractError("load", config.RepositoryFileName, nil, "no module contracts found")
	}

	return registry, nil
}

// ValidateModuleContract validates a module contract for correctness.
func ValidateModuleContract(module *ModuleContract) error {
	if module.Moniker == "" {
		return fmt.Errorf("moniker is required")
	}

	if module.Name == "" {
		return fmt.Errorf("name is required for module '%s'", module.Moniker)
	}

	// At least one component is required
	if len(module.Components) == 0 {
		return fmt.Errorf("components is required for module '%s'", module.Moniker)
	}

	return nil
}

// ValidateRegistry validates all module contracts in a registry.
func ValidateRegistry(registry *Registry) []error {
	var errors []error

	for _, module := range registry.All() {
		if err := ValidateModuleContract(module); err != nil {
			errors = append(errors, err)
		}
	}

	// Validate dependencies exist
	for _, module := range registry.All() {
		for _, dep := range module.DependsOn {
			if !registry.Has(dep) {
				errors = append(errors, fmt.Errorf("module '%s' depends on non-existent module '%s'",
					module.Moniker, dep))
			}
		}
	}

	return errors
}
