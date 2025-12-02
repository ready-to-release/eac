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
		LazyLoad:        true, // We only need modules and module-types
	}

	cfg, err := config.Load(opts)
	if err != nil {
		return nil, contracts.NewContractError("load", "config", err, "failed to initialize config loader")
	}

	// Load modules
	if err := cfg.LoadModules(validate); err != nil {
		modulesPath := filepath.Join(config.EACConfigRelPath, config.ModulesFileName)
		return nil, contracts.NewContractError("load", modulesPath, err, "failed to load modules.yml")
	}

	// Load module types for type-specific defaults
	if err := cfg.LoadModuleTypes(validate); err != nil {
		// Module types are optional - log warning but continue
		// This allows backward compatibility when module-types.yml doesn't exist
	}

	// Load repository config for path variables
	if err := cfg.LoadRepository(validate); err != nil {
		// Repository config is optional - continue with defaults
	}

	// Apply type-specific defaults with repository path variables
	cfg.Modules.ApplyTypeDefaults(cfg.ModuleTypes, cfg.Repository)

	// Create registry (version kept for internal compatibility)
	registry := NewRegistry("0.1.0", workspaceRoot)

	// Convert config.Module to contracts.BaseContract and process
	for _, m := range cfg.Modules.Modules {
		// Convert to BaseContract for ModuleContract creation
		base := contracts.BaseContract{
			Moniker:     m.Moniker,
			Name:        m.Name,
			Type:        m.Type,
			Description: m.Description,
			DependsOn:   m.DependsOn,
			Files: contracts.Files{
				Root:      m.Files.Root,
				Source:    m.Files.Source,
				Config:    m.Files.Config,
				Assets:    m.Files.Assets,
				Tests:     m.Files.Tests,
				Exclude:   m.Files.Exclude,
				Changelog: m.Files.Changelog,
				Repo: contracts.RepoPatterns{
					Specs:    m.Files.Repo.Specs,
					TestImpl: m.Files.Repo.TestImpl,
					Design:   m.Files.Repo.Design,
					Other:    m.Files.Repo.Other,
					Exclude:  m.Files.Repo.Exclude,
				},
			},
			Flags: contracts.Flags{},
			Metadata: m.Metadata,
		}
		// Note: Defaults are already applied by config.ModulesConfig.applyDefaults() and ApplyTypeDefaults()

		// Validate required fields
		if base.Moniker == "" {
			modulesPath := filepath.Join(config.EACConfigRelPath, config.ModulesFileName)
			return nil, contracts.NewContractError("validate", modulesPath, nil, "moniker field is required")
		}

		// Create module contract
		module := NewModuleContract(base, workspaceRoot)

		// Add to registry
		if err := registry.Add(module); err != nil {
			return nil, contracts.NewContractError("add", config.ModulesFileName, err,
				fmt.Sprintf("failed to add module '%s' to registry: %v", base.Moniker, err))
		}
	}

	// Validate registry has at least one module
	if registry.Count() == 0 {
		return nil, contracts.NewContractError("load", config.ModulesFileName, nil, "no module contracts found")
	}

	return registry, nil
}

// LoadFromWorkspaceLatest loads module contracts
// Deprecated: Use LoadFromWorkspace instead. Kept for backward compatibility.
func LoadFromWorkspaceLatest(workspaceRoot string) (*Registry, error) {
	return LoadFromWorkspace(workspaceRoot)
}

// ValidateModuleContract validates a module contract for correctness
func ValidateModuleContract(module *ModuleContract) error {
	if module.Moniker == "" {
		return fmt.Errorf("moniker is required")
	}

	if module.Name == "" {
		return fmt.Errorf("name is required for module '%s'", module.Moniker)
	}

	// Note: type now has a default, so no need to validate it

	if module.Files.Root == "" {
		return fmt.Errorf("files.root is required for module '%s'", module.Moniker)
	}

	return nil
}

// ValidateRegistry validates all module contracts in a registry
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
