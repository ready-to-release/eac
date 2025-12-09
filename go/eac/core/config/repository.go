// repository.go provides unified repository configuration loading.
// This is the single source of truth for all repository configuration including modules.
package config

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"gopkg.in/yaml.v3"
)

// RepositoryFileName is the config file for all repository settings
const RepositoryFileName = "repository.yml"

// RepositoryConfig holds all repository configuration including modules.
// This is the unified config loaded from .r2r/eac/repository.yml
type RepositoryConfig struct {
	// Repository settings
	Repository RepositorySettings `yaml:"repository"`

	// Path configuration
	Paths PathsConfig `yaml:"paths"`

	// Filename conventions
	Conventions ConventionsConfig `yaml:"conventions"`

	// Module definitions (previously in separate modules.yml)
	Modules []Module `yaml:"modules"`
}

// RepositorySettings holds repository-level configuration
type RepositorySettings struct {
	Type             string           `yaml:"type"`               // mono, poly, adjunct
	TrunkBranch      string           `yaml:"trunk_branch"`       // main branch name
	MaxBranchAgeDays int              `yaml:"max_branch_age_days"` // max age for feature branches
	Schemes          []string         `yaml:"schemes"`            // valid versioning schemes
	PR               PRConfig         `yaml:"pr"`                 // PR workflow config
	Versioning       VersioningConfig `yaml:"versioning"`         // versioning constraints
}

// PRConfig holds pull request workflow configuration
type PRConfig struct {
	DeleteBranchOnMerge bool   `yaml:"delete_branch_on_merge"`
	MergeStrategy       string `yaml:"merge_strategy"` // squash, merge, rebase
}

// VersioningConfig holds repository-wide versioning constraints
type VersioningConfig struct {
	Constraint string `yaml:"constraint"` // unrestricted, patch-only, calver-only
}

// PathsConfig defines repository-specific directory structures
type PathsConfig struct {
	TestImplRoot string    `yaml:"test_impl_root"`
	SpecsRoot    string    `yaml:"specs_root"`
	Templates    string    `yaml:"templates"`
	Out          OutConfig `yaml:"out"`
}

// OutConfig defines output directory structure
type OutConfig struct {
	Root     string `yaml:"root"`
	Build    string `yaml:"build"`
	Test     string `yaml:"test"`
	Logs     string `yaml:"logs"`
	Security string `yaml:"security"`
	Tools    string `yaml:"tools"` // CI tools like the commands binary (not build outputs)
}

// ConventionsConfig defines conventional filenames
type ConventionsConfig struct {
	GodogTest   string `yaml:"godog_test"`
	PackageJSON string `yaml:"package_json"`
	Changelog   string `yaml:"changelog"`
}

// loadRepositoryConfigUnmerged loads repository configuration from user's YAML file only.
// WARNING: This returns UNMERGED config without defaults. Do NOT use directly.
// Use config.Load().Repository instead to get properly merged config with defaults.
// This is only exported for use by the config package's merge logic.
func loadRepositoryConfigUnmerged(repoRoot string) (*RepositoryConfig, error) {
	configPath := filepath.Join(paths.EACConfigPath(repoRoot), RepositoryFileName)
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &RepositoryConfig{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg RepositoryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// TestImplPath returns the full path to a module's test implementation
func (c *RepositoryConfig) TestImplPath(moniker string) string {
	return c.Paths.TestImplRoot + "/" + moniker
}

// SpecsPath returns the full path to a module's specs directory
func (c *RepositoryConfig) SpecsPath(moniker string) string {
	return c.Paths.SpecsRoot + "/" + moniker
}

// BuildOutputPath returns the relative path to a module's build output.
// For absolute paths, use paths.BuildOutputPath(workspaceRoot, moniker) instead.
func (c *RepositoryConfig) BuildOutputPath(moniker string) string {
	return c.Paths.Out.Build + "/" + moniker
}

// BuildOutputPathAbs returns the absolute path to a module's build output
func (c *RepositoryConfig) BuildOutputPathAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Build, moniker)
}

// TestOutputPath returns the path to a test suite's output
func (c *RepositoryConfig) TestOutputPath(suiteName string) string {
	return c.Paths.Out.Test + "/" + suiteName
}

// LogsPath returns the path to logs for a command
func (c *RepositoryConfig) LogsPath(command string) string {
	return c.Paths.Out.Logs + "/" + command
}

// ToolsPath returns the path to the tools directory
func (c *RepositoryConfig) ToolsPath() string {
	return c.Paths.Out.Tools
}

// IsGodogTestFile checks if a filename is the godog test file
func (c *RepositoryConfig) IsGodogTestFile(filename string) bool {
	return filename == c.Conventions.GodogTest
}

// GetPathVariables returns a map of path variables for template substitution
func (c *RepositoryConfig) GetPathVariables() map[string]string {
	return map[string]string{
		"test_impl_root": c.Paths.TestImplRoot,
		"specs_root":     c.Paths.SpecsRoot,
		"templates":      c.Paths.Templates,
		"out_root":       c.Paths.Out.Root,
		"out_build":      c.Paths.Out.Build,
		"out_test":       c.Paths.Out.Test,
		"out_logs":       c.Paths.Out.Logs,
		"out_security":   c.Paths.Out.Security,
		"out_tools":      c.Paths.Out.Tools,
	}
}

// GetModule returns a module by moniker
func (c *RepositoryConfig) GetModule(moniker string) (*Module, bool) {
	for i := range c.Modules {
		if c.Modules[i].Moniker == moniker {
			return &c.Modules[i], true
		}
	}
	return nil, false
}

// GetByMoniker returns a module by moniker, or nil if not found
func (c *RepositoryConfig) GetByMoniker(moniker string) *Module {
	m, ok := c.GetModule(moniker)
	if !ok {
		return nil
	}
	return m
}

// GetModulesByType returns all modules of a specific type
func (c *RepositoryConfig) GetModulesByType(moduleType string) []Module {
	var result []Module
	for _, m := range c.Modules {
		if m.Type == moduleType {
			result = append(result, m)
		}
	}
	return result
}

// AllMonikers returns a list of all module monikers
func (c *RepositoryConfig) AllMonikers() []string {
	monikers := make([]string, len(c.Modules))
	for i, m := range c.Modules {
		monikers[i] = m.Moniker
	}
	return monikers
}
