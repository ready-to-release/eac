package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	security "github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/paths"
	"gopkg.in/yaml.v3"
)

// RiskConfig implements security.RiskConfigPort.
// It loads risk profile references and scoring configuration from YAML files.
type RiskConfig struct {
	profile        *ProfileWrapper
	moduleProfiles map[string]*ProfileWrapper
	scoring        *RiskScoringConfig
	catalogURL     string
	configDir      string // Directory containing the config (for relative paths)
}

// Verify RiskConfig implements RiskConfigPort.
var _ security.RiskConfigPort = (*RiskConfig)(nil)

// ProfileConfig holds the profile reference configuration.
type ProfileConfig struct {
	Path       string `yaml:"path"`
	CatalogURL string `yaml:"catalog_url,omitempty"`
}

// RiskConfigYAML is the YAML structure for risk-config.yml.
type RiskConfigYAML struct {
	Profile        ProfileConfig            `yaml:"profile"`
	Scoring        *RiskScoringConfig       `yaml:"scoring,omitempty"`
	ModuleProfiles map[string]ProfileConfig `yaml:"module_profiles,omitempty"`
}

// LoadRiskConfig loads risk configuration from eac-security contract.
// It loads defaults from contracts/security/0.1.0/defaults/risk-config.yml
// and merges with user overrides from .eac/risk-config.yml.
func LoadRiskConfig(repoRoot, configRoot string) (*RiskConfig, error) {
	cfg := &RiskConfig{
		moduleProfiles: make(map[string]*ProfileWrapper),
		scoring:        DefaultRiskScoringConfig(),
	}

	// Load contract defaults
	defaultPath := filepath.Join(repoRoot, "contracts", "security",
		paths.DefaultsVersion, "defaults", "risk-config.yml")
	if err := cfg.loadFromFile(defaultPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("loading risk-config defaults: %w", err)
	}

	// Load user overrides
	userPath := filepath.Join(configRoot, "risk-config.yml")
	if err := cfg.loadFromFile(userPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("loading risk-config overrides: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads a risk config YAML file and merges it into the config.
func (c *RiskConfig) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var yamlCfg RiskConfigYAML
	if err := yaml.Unmarshal(data, &yamlCfg); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	c.configDir = filepath.Dir(path)

	// Load profile
	if yamlCfg.Profile.Path != "" {
		profilePath := c.resolvePath(yamlCfg.Profile.Path)
		profile, err := LoadProfileWrapper(profilePath)
		if err != nil {
			// Profile loading is optional - allow missing files
			// This supports config-first workflow where profile is created later
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("loading profile from %s: %w", profilePath, err)
			}
			// Profile file doesn't exist - that's OK
		} else {
			c.profile = profile
		}
	}

	if yamlCfg.Profile.CatalogURL != "" {
		c.catalogURL = yamlCfg.Profile.CatalogURL
	}

	// Merge scoring config
	if yamlCfg.Scoring != nil {
		c.scoring = mergeRiskScoring(c.scoring, yamlCfg.Scoring)
	}

	// Load module-specific profiles
	for moniker, profileCfg := range yamlCfg.ModuleProfiles {
		if profileCfg.Path != "" {
			profilePath := c.resolvePath(profileCfg.Path)
			profile, err := LoadProfileWrapper(profilePath)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("loading module profile %s from %s: %w", moniker, profilePath, err)
				}
				// Module profile file doesn't exist - that's OK
			} else {
				c.moduleProfiles[moniker] = profile
			}
		}
	}

	return nil
}

// resolvePath resolves a path relative to the config file directory.
func (c *RiskConfig) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(c.configDir, path)
}

// GetProfile returns the solution-wide profile.
func (c *RiskConfig) GetProfile() (security.ProfilePort, error) {
	if c.profile == nil {
		return nil, fmt.Errorf("no risk profile configured")
	}
	return c.profile, nil
}

// GetModuleProfile returns the profile for a module.
// Falls back to solution profile if module has no specific profile.
func (c *RiskConfig) GetModuleProfile(moniker string) (security.ProfilePort, error) {
	if mp, ok := c.moduleProfiles[moniker]; ok {
		return mp, nil
	}
	return c.GetProfile() // Fall back to solution profile
}

// GetCatalogURL returns the catalog URL.
func (c *RiskConfig) GetCatalogURL() string {
	return c.catalogURL
}

// GetScoring returns the scoring configuration.
func (c *RiskConfig) GetScoring() security.RiskScoringPort {
	return c.scoring
}

// ListModuleProfiles returns modules with custom profiles.
func (c *RiskConfig) ListModuleProfiles() []string {
	monikers := make([]string, 0, len(c.moduleProfiles))
	for m := range c.moduleProfiles {
		monikers = append(monikers, m)
	}
	return monikers
}
