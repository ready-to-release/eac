package modules

import (
	"fmt"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
)

// LoadFromWorkspace loads all module contracts from the workspace.
// This is the main entry point for loading module domain.
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
		LazyLoad:        true, // We only need modules and component-kinds
	}

	cfg, err := config.Load(opts)
	if err != nil {
		return nil, domain.NewContractError("load", "config", err, "failed to initialize config loader")
	}

	// Load repository config (includes modules)
	if err := cfg.LoadRepository(validate); err != nil {
		repoPath := filepath.Join(config.EACConfigRelPath, config.RepositoryFileName)
		return nil, domain.NewContractError("load", repoPath, err, "failed to load repository.yml")
	}

	// Ensure component kinds map is initialized (populated by LoadBlueprints via LoadRepository)
	if err := cfg.EnsureComponentKinds(validate); err != nil {
		// Component kinds are optional - continue with defaults
	}

	// Apply component-specific defaults with repository path variables
	cfg.Repository.ApplyComponentDefaults(cfg.ComponentKinds, cfg.RepoRoot)

	// Create registry (version kept for internal compatibility)
	registry := NewRegistry("0.1.0", workspaceRoot)

	// Convert config.Module to domain.BaseContract and process
	for _, m := range cfg.Repository.Modules {
		// Convert to BaseContract for ModuleContract creation.
		// Uses ModuleToBaseContract to avoid field-by-field mapping boilerplate.
		base := domain.ModuleToBaseContract(m)

		// Note: Defaults are already applied by config.RepositoryConfig.applyModuleDefaults() and ApplyComponentDefaults()

		// Validate required fields
		if base.Moniker == "" {
			repoPath := filepath.Join(config.EACConfigRelPath, config.RepositoryFileName)
			return nil, domain.NewContractError("validate", repoPath, nil, "moniker field is required")
		}

		// Create module contract
		module := NewModuleContract(base, workspaceRoot)

		// Add to registry
		if err := registry.Add(module); err != nil {
			return nil, domain.NewContractError("add", config.RepositoryFileName, err,
				fmt.Sprintf("failed to add module '%s' to registry: %v", base.Moniker, err))
		}
	}

	// Validate registry has at least one module
	if registry.Count() == 0 {
		return nil, domain.NewContractError("load", config.RepositoryFileName, nil, "no module contracts found")
	}

	return registry, nil
}

// ValidateModuleContract validates a module contract for correctness.
func ValidateModuleContract(module *ModuleContract) error {
	if module.Moniker == "" {
		return fmt.Errorf("moniker is required")
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
