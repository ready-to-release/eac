package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LintProvidersFileName is the name of the lint providers config file.
const LintProvidersFileName = "lint-providers.yml"

// LintProvidersConfig represents the lint-providers.yml configuration.
type LintProvidersConfig struct {
	// LintProviders maps provider name to its configuration
	LintProviders map[string]*LintProvider `yaml:"lint-providers"`
}

// LintProvider defines how a linting tool operates.
type LintProvider struct {
	// Description is a human-readable description of the lint provider
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Command is the executable to run (e.g., "golangci-lint", "eslint")
	Command string `yaml:"command" json:"command"`

	// InputMode specifies how files are passed to the linter:
	// - "packages": Lint by directory/package (e.g., Go modules use ./...)
	// - "files": Lint individual files (e.g., markdown files)
	// - "directory": Lint entire directory
	InputMode string `yaml:"input_mode" json:"input_mode"`

	// ConfigFiles are standard config file names for auto-discovery
	ConfigFiles []string `yaml:"config_files,omitempty" json:"config_files,omitempty"`

	// SystemDependency is a reference to system-dependencies.yml moniker
	SystemDependency string `yaml:"system_dependency" json:"system_dependency"`

	// AppliesTo lists component types this linter can process
	AppliesTo []string `yaml:"applies_to" json:"applies_to"`

	// Options contains optional behavior flags
	Options *LintProviderOptions `yaml:"options,omitempty" json:"options,omitempty"`
}

// LintProviderOptions contains optional lint provider behavior flags.
type LintProviderOptions struct {
	// AllowParallel indicates whether multiple instances can run simultaneously
	AllowParallel bool `yaml:"allow_parallel,omitempty" json:"allow_parallel,omitempty"`

	// AutoFix indicates whether the linter supports --fix or equivalent
	AutoFix bool `yaml:"auto_fix,omitempty" json:"auto_fix,omitempty"`
}

// Get returns a lint provider by name.
func (c *LintProvidersConfig) Get(name string) *LintProvider {
	if c == nil || c.LintProviders == nil {
		return nil
	}
	return c.LintProviders[name]
}

// GetProvidersForComponentType returns all lint providers that can process a component type.
func (c *LintProvidersConfig) GetProvidersForComponentType(compType string) []string {
	if c == nil || c.LintProviders == nil {
		return nil
	}

	var result []string
	for name, provider := range c.LintProviders {
		for _, applies := range provider.AppliesTo {
			if applies == compType {
				result = append(result, name)
				break
			}
		}
	}
	return result
}

// GetDefaultProviderForComponentType returns the first (default) lint provider for a component type.
// Returns empty string if no provider is configured for the component type.
func (c *LintProvidersConfig) GetDefaultProviderForComponentType(compType string) string {
	providers := c.GetProvidersForComponentType(compType)
	if len(providers) == 0 {
		return ""
	}
	return providers[0]
}

// HasProvider returns true if a lint provider with the given name exists.
func (c *LintProvidersConfig) HasProvider(name string) bool {
	return c.Get(name) != nil
}

// GetInputMode returns the input mode for a provider, defaulting to "files" if not specified.
func (p *LintProvider) GetInputMode() string {
	if p.InputMode == "" {
		return "files"
	}
	return p.InputMode
}

// SupportsAutoFix returns true if the lint provider supports auto-fix.
func (p *LintProvider) SupportsAutoFix() bool {
	return p.Options != nil && p.Options.AutoFix
}

// AllowsParallel returns true if multiple instances of this linter can run simultaneously.
func (p *LintProvider) AllowsParallel() bool {
	return p.Options != nil && p.Options.AllowParallel
}

// LoadLintProvidersDefaults loads default lint providers from contract defaults.
// Returns ErrNoDefaults when defaults don't exist - allows tests to work without contracts folder.
func LoadLintProvidersDefaults(repoRoot string) (*LintProvidersConfig, error) {
	cfg, err := cloneLintProvidersDefaults()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading lint-providers defaults: %w", err)
	}
	return cfg, nil
}

// MergeLintProviders merges user lint providers with defaults.
// User entries with same name override defaults, new entries are appended.
func MergeLintProviders(defaults, user *LintProvidersConfig) *LintProvidersConfig {
	if defaults == nil {
		return user
	}
	if user == nil {
		return defaults
	}

	// Start with copy of defaults
	result := &LintProvidersConfig{
		LintProviders: make(map[string]*LintProvider),
	}
	for name, provider := range defaults.LintProviders {
		result.LintProviders[name] = provider
	}

	// Merge user providers (override or add)
	for name, userProvider := range user.LintProviders {
		result.LintProviders[name] = userProvider
	}

	return result
}

// LoadLintProviders loads the lint-providers configuration.
// Merges contract defaults with user config (if exists).
func (c *EACConfig) LoadLintProviders(validateSchema bool) error {
	// Load defaults from contract
	defaults, err := LoadLintProvidersDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return fmt.Errorf("loading lint-providers defaults: %w", err)
	}

	// Check if user config exists
	providersPath := filepath.Join(c.ConfigRoot, LintProvidersFileName)
	if _, err := os.Stat(providersPath); os.IsNotExist(err) {
		// Use defaults only (may be nil in test environments without contract files)
		if defaults == nil {
			// Create empty config to guarantee non-nil
			defaults = &LintProvidersConfig{
				LintProviders: make(map[string]*LintProvider),
			}
		}
		c.LintProviders = defaults
		return nil
	}

	data, err := c.readConfigFile(LintProvidersFileName)
	if err != nil {
		return err
	}

	// Lint providers config is validated structurally by YAML unmarshalling below.
	// JSON schema validation is not applied here because lint-providers.yml is a
	// simple key-value format without a dedicated schema definition.
	_ = validateSchema

	var userCfg LintProvidersConfig
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", LintProvidersFileName, err)
	}

	// Merge defaults with user config
	c.LintProviders = MergeLintProviders(defaults, &userCfg)
	return nil
}
