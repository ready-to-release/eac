package config

// ModuleTypesConfig represents the module-types.yml configuration
type ModuleTypesConfig struct {
	Types []ModuleTypeDef `yaml:"types"`

	// Internal lookup map (built after load)
	typeMap map[string]*ModuleTypeDef
}

// ModuleTypeDef defines a module type
type ModuleTypeDef struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	BuildSystem  string            `yaml:"build_system"`
	Capabilities []string          `yaml:"capabilities"`
	Defaults     *TypeDefaults     `yaml:"defaults,omitempty"`
}

// TypeDefaults contains default values for modules of this type.
// Supports variable substitution: {moniker}, {root}, {type}
type TypeDefaults struct {
	Files *FilesDefaults `yaml:"files,omitempty"`
	Repo  *RepoDefaults  `yaml:"repo,omitempty"`
	Flags *FlagsDefaults `yaml:"flags,omitempty"`
}

// FilesDefaults contains default file patterns for a module type
type FilesDefaults struct {
	Source    []string          `yaml:"source,omitempty"`
	Config    []string          `yaml:"config,omitempty"`
	Assets    []string          `yaml:"assets,omitempty"`
	Tests     []string          `yaml:"tests,omitempty"`
	Changelog string            `yaml:"changelog,omitempty"`
	Workflows *WorkflowDefaults `yaml:"workflows,omitempty"`
}

// WorkflowDefaults contains default workflow file paths
type WorkflowDefaults struct {
	CI      string `yaml:"ci,omitempty"`
	Release string `yaml:"release,omitempty"`
}

// RepoDefaults contains default repo-level configurations
type RepoDefaults struct {
	Specs    []string `yaml:"specs,omitempty"`
	TestImpl string   `yaml:"test_impl,omitempty"`
	Design   string   `yaml:"design,omitempty"`
}

// FlagsDefaults contains default flag values
type FlagsDefaults struct {
	CatchAll         *bool `yaml:"catch_all,omitempty"`
	OwnChildrenFiles *bool `yaml:"own_children_files,omitempty"`
}

// buildTypeMap builds the internal lookup map
func (c *ModuleTypesConfig) buildTypeMap() {
	c.typeMap = make(map[string]*ModuleTypeDef, len(c.Types))
	for i := range c.Types {
		c.typeMap[c.Types[i].Name] = &c.Types[i]
	}
}

// Get returns a module type definition by name
func (c *ModuleTypesConfig) Get(typeName string) *ModuleTypeDef {
	if c.typeMap == nil {
		c.buildTypeMap()
	}
	return c.typeMap[typeName]
}

// HasCapability checks if a module type has a specific capability
func (c *ModuleTypesConfig) HasCapability(typeName, capability string) bool {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return false
	}
	for _, cap := range typeDef.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// GetBuildSystem returns the build system for a module type
func (c *ModuleTypesConfig) GetBuildSystem(typeName string) string {
	typeDef := c.Get(typeName)
	if typeDef == nil {
		return "none"
	}
	return typeDef.BuildSystem
}

// GetTypesWithCapability returns all type names that have the given capability
func (c *ModuleTypesConfig) GetTypesWithCapability(capability string) []string {
	var result []string
	for _, t := range c.Types {
		for _, cap := range t.Capabilities {
			if cap == capability {
				result = append(result, t.Name)
				break
			}
		}
	}
	return result
}

// GetTypesByBuildSystem returns all type names that use the given build system
func (c *ModuleTypesConfig) GetTypesByBuildSystem(buildSystem string) []string {
	var result []string
	for _, t := range c.Types {
		if t.BuildSystem == buildSystem {
			result = append(result, t.Name)
		}
	}
	return result
}

// HasCapability checks if this type definition has a specific capability
func (t *ModuleTypeDef) HasCapability(capability string) bool {
	for _, cap := range t.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}
