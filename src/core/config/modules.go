package config

import "fmt"

// ModulesConfig represents the modules.yml configuration
type ModulesConfig struct {
	Modules []Module `yaml:"modules"`
}

// Module represents a single module definition
type Module struct {
	Moniker     string   `yaml:"moniker"`
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Parent      string   `yaml:"parent"`
	DependsOn   []string `yaml:"depends_on"`
	Files       Files    `yaml:"files"`
	Flags       Flags    `yaml:"flags"`
}

// Files defines file ownership patterns for a module
type Files struct {
	Root      string   `yaml:"root"`
	Source    []string `yaml:"source"`
	Config    []string `yaml:"config"`
	Assets    []string `yaml:"assets"`
	Tests     []string `yaml:"tests"`
	Exclude   []string `yaml:"exclude"`
	Changelog string   `yaml:"changelog"`
	Repo      RepoFiles `yaml:"repo"`
}

// RepoFiles defines repository-level file ownership
type RepoFiles struct {
	Specs   []string `yaml:"specs"`
	Other   []string `yaml:"other"`
	Exclude []string `yaml:"exclude"`
}

// Flags defines module behavior flags
type Flags struct {
	CatchAll         bool `yaml:"catch_all"`
	OwnChildrenFiles bool `yaml:"own_children_files"`
}

// applyDefaults applies default values to all modules
func (c *ModulesConfig) applyDefaults() {
	for i := range c.Modules {
		m := &c.Modules[i]

		if m.Type == "" {
			m.Type = "no-module-type"
		}
		if m.Parent == "" {
			m.Parent = "."
		}
		if m.Description == "" {
			m.Description = m.Name
		}
		if m.DependsOn == nil {
			m.DependsOn = []string{}
		}
		if m.Files.Changelog == "" {
			m.Files.Changelog = "CHANGELOG.md"
		}
		if m.Files.Repo.Specs == nil {
			m.Files.Repo.Specs = []string{fmt.Sprintf("specs/%s/**", m.Moniker)}
		}
	}
}

// GetModule returns a module by moniker
func (c *ModulesConfig) GetModule(moniker string) (*Module, bool) {
	for i := range c.Modules {
		if c.Modules[i].Moniker == moniker {
			return &c.Modules[i], true
		}
	}
	return nil, false
}

// GetModulesByType returns all modules of a specific type
func (c *ModulesConfig) GetModulesByType(moduleType string) []Module {
	var result []Module
	for _, m := range c.Modules {
		if m.Type == moduleType {
			result = append(result, m)
		}
	}
	return result
}

// GetModulesByParent returns all modules with a specific parent
func (c *ModulesConfig) GetModulesByParent(parent string) []Module {
	var result []Module
	for _, m := range c.Modules {
		if m.Parent == parent {
			result = append(result, m)
		}
	}
	return result
}

// GetCatchAllModule returns the catch-all module if one exists
func (c *ModulesConfig) GetCatchAllModule() (*Module, bool) {
	for i := range c.Modules {
		if c.Modules[i].Flags.CatchAll {
			return &c.Modules[i], true
		}
	}
	return nil, false
}

// AllMonikers returns a list of all module monikers
func (c *ModulesConfig) AllMonikers() []string {
	monikers := make([]string, len(c.Modules))
	for i, m := range c.Modules {
		monikers[i] = m.Moniker
	}
	return monikers
}
