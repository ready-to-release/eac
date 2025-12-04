// repository.go provides repository-wide path configuration loading.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RepositoryFileName is the config file for repository-wide settings
const RepositoryFileName = "repository.yml"

// RepositoryConfig holds repository-wide configuration
type RepositoryConfig struct {
	Paths       PathsConfig       `yaml:"paths"`
	Conventions ConventionsConfig `yaml:"conventions"`
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

// DefaultRepositoryConfig returns the default configuration
func DefaultRepositoryConfig() RepositoryConfig {
	return RepositoryConfig{
		Paths: PathsConfig{
			TestImplRoot: "go/eac/specs/impl",
			SpecsRoot:    "specs",
			Templates:    "templates",
			Out: OutConfig{
				Root:     "out",
				Build:    "out/build",
				Test:     "out/test",
				Logs:     "out/logs",
				Security: "out/security",
				Tools:    "out/tools",
			},
		},
		Conventions: ConventionsConfig{
			GodogTest:   "godog_test.go",
			PackageJSON: "package.json",
			Changelog:   "CHANGELOG.md",
		},
	}
}

// LoadRepositoryConfig loads repository configuration from YAML.
// Returns defaults if the file doesn't exist.
func LoadRepositoryConfig(repoRoot string) (*RepositoryConfig, error) {
	cfg := DefaultRepositoryConfig()

	configPath := filepath.Join(repoRoot, r2rDir, eacDir, RepositoryFileName)
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &cfg, nil
	}
	if err != nil {
		return nil, err
	}

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

// BuildOutputPath returns the path to a module's build output
func (c *RepositoryConfig) BuildOutputPath(moniker string) string {
	return c.Paths.Out.Build + "/" + moniker
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
