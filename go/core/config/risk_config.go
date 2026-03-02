package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	scanner "github.com/ready-to-release/eac/contracts/scanner/0.1.0"
	"gopkg.in/yaml.v3"
)

// RiskConfig implements scanner.RiskConfigPort.
// It loads risk profile references and scoring configuration from YAML files.
type RiskConfig struct {
	profile        *ProfileWrapper
	moduleProfiles map[string]*ProfileWrapper
	scoring        *RiskScoringConfig
	catalogURL     string
	configDir      string // Directory containing the config (for relative paths)
}

// Verify RiskConfig implements RiskConfigPort.
var _ scanner.RiskConfigPort = (*RiskConfig)(nil)

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

// LoadRiskConfig loads risk configuration from the embedded scanner contract
// and merges with user overrides from configRoot/risk-config.yml.
func LoadRiskConfig(configRoot string) (*RiskConfig, error) {
	cfg := &RiskConfig{
		moduleProfiles: make(map[string]*ProfileWrapper),
		scoring:        DefaultRiskScoringConfig(),
	}

	// Load contract defaults from embedded filesystem
	if err := cfg.loadFromEmbeddedDefaults("risk-config.yml"); err != nil {
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

	// Load main profile
	if yamlCfg.Profile.Path != "" {
		profile, err := c.loadOptionalProfile(yamlCfg.Profile.Path)
		if err != nil {
			return fmt.Errorf("loading profile from %s: %w", c.resolvePath(yamlCfg.Profile.Path), err)
		}
		if profile != nil {
			c.profile = profile
		}
	}

	if yamlCfg.Profile.CatalogURL != "" {
		c.catalogURL = yamlCfg.Profile.CatalogURL
	}

	if yamlCfg.Scoring != nil {
		c.scoring = mergeRiskScoring(c.scoring, yamlCfg.Scoring)
	}

	// Load module-specific profiles
	for moniker, profileCfg := range yamlCfg.ModuleProfiles {
		if profileCfg.Path != "" {
			profile, err := c.loadOptionalProfile(profileCfg.Path)
			if err != nil {
				return fmt.Errorf("loading module profile %s from %s: %w", moniker, c.resolvePath(profileCfg.Path), err)
			}
			if profile != nil {
				c.moduleProfiles[moniker] = profile
			}
		}
	}

	return nil
}

// loadOptionalProfile loads a profile, returning nil if the file doesn't exist.
// Returns an error only for non-missing-file failures (e.g. invalid JSON).
func (c *RiskConfig) loadOptionalProfile(relativePath string) (*ProfileWrapper, error) {
	profilePath := c.resolvePath(relativePath)
	profile, err := LoadProfileWrapper(profilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return profile, nil
}

// loadFromEmbeddedDefaults loads risk config defaults from the embedded scanner contract.
// Only scoring and catalog URL are loaded. Profile loading is skipped because
// embedded defaults reference relative paths that cannot be resolved without a filesystem context.
func (c *RiskConfig) loadFromEmbeddedDefaults(filename string) error {
	data, err := scanner.FS.ReadFile(scanner.DefaultPath(filename))
	if err != nil {
		return err
	}

	var yamlCfg RiskConfigYAML
	if err := yaml.Unmarshal(data, &yamlCfg); err != nil {
		return fmt.Errorf("parsing embedded %s: %w", filename, err)
	}

	if yamlCfg.Profile.CatalogURL != "" {
		c.catalogURL = yamlCfg.Profile.CatalogURL
	}

	if yamlCfg.Scoring != nil {
		c.scoring = mergeRiskScoring(c.scoring, yamlCfg.Scoring)
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
func (c *RiskConfig) GetProfile() (scanner.ProfilePort, error) {
	if c.profile == nil {
		return nil, fmt.Errorf("no risk profile configured")
	}
	return c.profile, nil
}

// GetModuleProfile returns the profile for a module.
// Falls back to solution profile if module has no specific profile.
func (c *RiskConfig) GetModuleProfile(moniker string) (scanner.ProfilePort, error) {
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
func (c *RiskConfig) GetScoring() scanner.RiskScoringPort {
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
