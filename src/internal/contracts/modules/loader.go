package modules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/internal/contracts"
)

// LoadFromWorkspace loads all module contracts for a specific version from the workspace
// This is the main entry point for loading module contracts
func LoadFromWorkspace(workspaceRoot, version string) (*Registry, error) {
	// Create base loader
	loader := contracts.NewLoader(workspaceRoot)

	// Create registry
	registry := NewRegistry(version, workspaceRoot)

	// Construct pattern for module contracts
	// Pattern: contracts/modules/{version}/*.yml
	pattern := filepath.Join("contracts", "modules", version, "*.yml")

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
		if base.Versioning.VersionScheme == "" {
			base.Versioning.VersionScheme = "semver"
		}
		if base.Description == "" {
			base.Description = base.Name
		}
		if base.Source.ChangelogPath == "" {
			if base.Source.Root == "/" {
				base.Source.ChangelogPath = "CHANGELOG.md"
			} else {
				base.Source.ChangelogPath = base.Source.Root + "/CHANGELOG.md"
			}
		}
		// Only apply default includes if:
		// 1. Includes is nil (not set in YAML)
		// 2. OR it's a catch-all singleton (which needs patterns to work)
		if base.Source.Includes == nil {
			base.Source.Includes = []string{"**/*", "*"}
		} else if len(base.Source.Includes) == 0 {
			// includes: [] explicitly set - keep it empty (null filter)
			// UNLESS it's a catch-all singleton which needs patterns
			if base.Source.IsCatchAllSingleton != nil && *base.Source.IsCatchAllSingleton {
				base.Source.Includes = []string{"**/*", "*"}
			}
		}
		// ExcludeChildrenOwnedSource defaults to true
		if base.Source.ExcludeChildrenOwnedSource == nil {
			trueVal := true
			base.Source.ExcludeChildrenOwnedSource = &trueVal
		}
		// DependsOn defaults to empty list
		if base.DependsOn == nil {
			base.DependsOn = []string{}
		}
		// UsedBy defaults to empty list
		if base.UsedBy == nil {
			base.UsedBy = []string{}
		}

		// Validate required fields
		if base.Moniker == "" {
			return contracts.NewContractError("validate", relPath, nil, "moniker field is required")
		}

		// Validate that filename matches moniker
		// Extract filename without extension from relPath
		// relPath is like "contracts/modules/0.1.0/my-module.yml"
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
		if module.Source.IsCatchAllSingleton != nil && *module.Source.IsCatchAllSingleton {
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

// LoadFromWorkspaceLatest loads module contracts using the latest version
// This scans the contracts/modules directory to find the highest version
func LoadFromWorkspaceLatest(workspaceRoot string) (*Registry, error) {
	// For now, default to "0.1.0"
	// TODO: Implement version detection
	return LoadFromWorkspace(workspaceRoot, "0.1.0")
}

// LoadSingleModule loads a single module contract by moniker and version
func LoadSingleModule(workspaceRoot, moniker, version string) (*ModuleContract, error) {
	loader := contracts.NewLoader(workspaceRoot)

	// Construct path to module contract file
	// Path: contracts/modules/{version}/{moniker}.yml
	relPath := filepath.Join("contracts", "modules", version, moniker+".yml")

	// Load the contract
	var base contracts.BaseContract
	if err := loader.LoadYAML(relPath, &base); err != nil {
		return nil, err
	}

	// Apply defaults
	if base.Type == "" {
		base.Type = "no-module-type"
	}
	if base.Parent == "" {
		base.Parent = "."
	}
	if base.Versioning.VersionScheme == "" {
		base.Versioning.VersionScheme = "semver"
	}
	if base.Description == "" {
		base.Description = base.Name
	}
	if base.Source.ChangelogPath == "" {
		if base.Source.Root == "/" {
			base.Source.ChangelogPath = "CHANGELOG.md"
		} else {
			base.Source.ChangelogPath = base.Source.Root + "/CHANGELOG.md"
		}
	}
	if len(base.Source.Includes) == 0 {
		base.Source.Includes = []string{"**/*", "*"}
	}
	// ExcludeChildrenOwnedSource defaults to true
	if base.Source.ExcludeChildrenOwnedSource == nil {
		trueVal := true
		base.Source.ExcludeChildrenOwnedSource = &trueVal
	}
	// DependsOn defaults to empty list
	if base.DependsOn == nil {
		base.DependsOn = []string{}
	}
	// UsedBy defaults to empty list
	if base.UsedBy == nil {
		base.UsedBy = []string{}
	}

	// Validate moniker matches filename
	if base.Moniker != moniker {
		return nil, contracts.NewContractError("validate", relPath, nil,
			fmt.Sprintf("moniker mismatch: expected '%s', got '%s'", moniker, base.Moniker))
	}

	// Create and return module contract
	module := NewModuleContract(base, workspaceRoot)
	return module, nil
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

	if module.Source.Root == "" {
		return fmt.Errorf("source.root is required for module '%s'", module.Moniker)
	}

	// Note: source.includes now has a default, so no need to validate it

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
