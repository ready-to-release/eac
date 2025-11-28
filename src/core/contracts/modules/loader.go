package modules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts"
)

// LoadFromWorkspace loads all module contracts from the workspace.
// This is the main entry point for loading module contracts.
func LoadFromWorkspace(workspaceRoot string) (*Registry, error) {
	// Create base loader
	loader := contracts.NewLoader(workspaceRoot)

	// Create registry (version kept for internal compatibility)
	registry := NewRegistry("0.1.0", workspaceRoot)

	// Construct pattern for module contracts
	pattern := filepath.Join(contracts.EACConfigRelPath, "modules", "*.yml")

	// Load all matching YAML files
	err := loader.LoadYAMLPattern(pattern, func(relPath string) error {
		// Skip definitions.yml as it's metadata, not a module contract
		if strings.HasSuffix(relPath, "definitions.yml") {
			// Actually, based on the contract, definitions IS a module
			// Let's load it but we can check IsDefinitionsFile() later if needed
		}

		// Parse the module contract
		var base contracts.BaseContract
		if err := loader.LoadYAML(relPath, &base); err != nil {
			return err
		}

		// Apply defaults
		if base.Type == "" {
			base.Type = "no-module-type"
		}
		if base.Parent == "" {
			base.Parent = "."
		}
		if base.Description == "" {
			base.Description = base.Name
		}
		// DependsOn defaults to empty list
		if base.DependsOn == nil {
			base.DependsOn = []string{}
		}
		// Files.Changelog default
		if base.Files.Changelog == "" {
			base.Files.Changelog = "CHANGELOG.md"
		}
		// Files.Repo.Specs default: specs/<moniker>/**
		if base.Files.Repo.Specs == nil {
			base.Files.Repo.Specs = []string{fmt.Sprintf("specs/%s/**", base.Moniker)}
		}

		// Validate required fields
		if base.Moniker == "" {
			return contracts.NewContractError("validate", relPath, nil, "moniker field is required")
		}

		// Validate that filename matches moniker
		filename := filepath.Base(relPath)
		expectedFilename := base.Moniker + ".yml"
		if filename != expectedFilename {
			return contracts.NewContractError("validate", relPath, nil,
				fmt.Sprintf("filename mismatch: expected '%s', got '%s' (moniker: '%s')",
					expectedFilename, filename, base.Moniker))
		}

		// Create module contract
		module := NewModuleContract(base, workspaceRoot)

		// Add to registry
		if err := registry.Add(module); err != nil {
			return contracts.NewContractError("add", relPath, err, fmt.Sprintf("failed to add module to registry: %v", err))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Validate registry has at least one module
	if registry.Count() == 0 {
		return nil, contracts.NewContractError("load", pattern, nil, "no module contracts found")
	}

	// Validate only one catch-all singleton exists
	catchAllModules := []*ModuleContract{}
	for _, module := range registry.All() {
		if module.Flags.CatchAll {
			catchAllModules = append(catchAllModules, module)
		}
	}
	if len(catchAllModules) > 1 {
		monikers := []string{}
		for _, m := range catchAllModules {
			monikers = append(monikers, m.Moniker)
		}
		return nil, contracts.NewContractError("validate", pattern, nil,
			fmt.Sprintf("multiple catch-all singleton modules found: %v (only one allowed)", monikers))
	}

	// Validate parent chains for all modules
	for _, module := range registry.All() {
		if err := ValidateParentChain(module, registry); err != nil {
			return nil, contracts.NewContractError("validate", "", err,
				fmt.Sprintf("invalid parent chain for module '%s': %v", module.Moniker, err))
		}
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
