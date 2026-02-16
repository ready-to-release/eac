package config

// buildDiscoveryVars creates the variable map for template parameter expansion.
// Includes repo-level path variables plus module-specific values.
func buildDiscoveryVars(mod *Module, repoCfg *RepositoryConfig) map[string]string {
	vars := map[string]string{
		"moniker": mod.Moniker,
	}

	// Add repo-level path variables
	if repoCfg != nil {
		vars["specs_root"] = repoCfg.Paths.SpecsRoot
		vars["containers_root"] = repoCfg.Paths.ContainersRoot
		vars["design_dir"] = repoCfg.Conventions.DesignDir
		vars["workspace_dsl"] = repoCfg.Conventions.WorkspaceDSL
		vars["specification"] = repoCfg.Conventions.Specification
		vars["godog_test"] = repoCfg.Conventions.GodogTest
	}

	// Add component roots as variables (e.g., go component root → {go_root})
	// Both the component name and type are registered, so {go_root} works
	// even when the component name is derived from the root path (e.g., "commands").
	if mod.Components != nil {
		for name, entry := range mod.Components {
			if entry != nil && entry.Root != "" {
				vars[name+"_root"] = entry.Root
				// Also register by type if different from name
				if entry.Type != "" && entry.Type != name {
					vars[entry.Type+"_root"] = entry.Root
				}
			}
		}
	}

	return vars
}
