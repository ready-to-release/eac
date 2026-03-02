package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	scanner "github.com/ready-to-release/eac/contracts/scanner/0.1.0"
)

// defaultScannerIDs is the fallback scanner list when no policies are configured.
var defaultScannerIDs = []string{"trivy-sbom", "trivy-vuln"}

// SecurityConfig implements scanner.SecurityConfigPort.
// It loads scanner definitions and policies from the eac-security contract.
type SecurityConfig struct {
	scanners map[string]*scanner.ScannerDefinition
	policies *scanner.PoliciesConfig
	risk     *RiskConfig
}

// Verify SecurityConfig implements SecurityConfigPort.
var _ scanner.SecurityConfigPort = (*SecurityConfig)(nil)

// LoadSecurityConfig loads the security configuration from the embedded scanner contract
// and merges with user overrides from configRoot.
func LoadSecurityConfig(configRoot string) (*SecurityConfig, error) {
	cfg := &SecurityConfig{
		scanners: make(map[string]*scanner.ScannerDefinition),
	}

	if err := cfg.loadScanners(configRoot); err != nil {
		return nil, err
	}

	if err := cfg.loadPolicies(configRoot); err != nil {
		return nil, err
	}

	risk, err := LoadRiskConfig(configRoot)
	if err != nil {
		return nil, err
	}
	cfg.risk = risk

	return cfg, nil
}

// Risk returns the risk configuration.
// Returns nil if risk config was not loaded.
func (c *SecurityConfig) Risk() scanner.RiskConfigPort {
	if c.risk == nil {
		return nil
	}
	return c.risk
}

// loadScanners loads scanner definitions from embedded defaults and user overrides.
func (c *SecurityConfig) loadScanners(configRoot string) error {
	defaults, err := loadEmbedded[scanner.ScannersConfig]("scanners.yml")
	if err != nil {
		return fmt.Errorf("loading embedded scanners defaults: %w", err)
	}

	user, err := loadFile[scanner.ScannersConfig](filepath.Join(configRoot, "scanners.yml"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Merge: start with defaults, override with user
	if defaults != nil {
		for id, scannerDef := range defaults.Scanners {
			scannerDef.IDValue = id
			c.scanners[id] = scannerDef
		}
	}
	if user != nil {
		for id, scannerDef := range user.Scanners {
			scannerDef.IDValue = id
			c.scanners[id] = scannerDef
		}
	}

	return nil
}

// loadPolicies loads scanner policies from embedded defaults and user overrides.
func (c *SecurityConfig) loadPolicies(configRoot string) error {
	defaults, err := loadEmbedded[scanner.PoliciesConfig]("policies.yml")
	if err != nil {
		return fmt.Errorf("loading embedded policies defaults: %w", err)
	}

	user, err := loadFile[scanner.PoliciesConfig](filepath.Join(configRoot, "policies.yml"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	c.policies = mergePolicies(defaults, user)
	if c.policies == nil {
		c.policies = &scanner.PoliciesConfig{
			ComponentScanners: make(map[string][]string),
			Default:           defaultScannerIDs,
		}
	}

	return nil
}

// GetScanner returns a scanner definition by ID.
func (c *SecurityConfig) GetScanner(id string) (scanner.ScannerPort, bool) {
	scanner, ok := c.scanners[id]
	if !ok {
		return nil, false
	}
	return scanner, true
}

// ListScanners returns all available scanner IDs.
func (c *SecurityConfig) ListScanners() []string {
	ids := make([]string, 0, len(c.scanners))
	for id := range c.scanners {
		ids = append(ids, id)
	}
	return ids
}

// GetDefaultScanners returns the default scanner IDs for a component type.
func (c *SecurityConfig) GetDefaultScanners(componentType string) []string {
	if c.policies == nil {
		return defaultScannerIDs
	}

	if scanners, ok := c.policies.ComponentScanners[componentType]; ok {
		return scanners
	}

	if len(c.policies.Default) > 0 {
		return c.policies.Default
	}

	return defaultScannerIDs
}

// ShouldSkipModule returns true if the module should be skipped during scanning.
func (c *SecurityConfig) ShouldSkipModule(moniker string) bool {
	if c.policies == nil {
		return false
	}
	for _, skip := range c.policies.SkipModules {
		if skip == moniker {
			return true
		}
	}
	return false
}

// Generic YAML loading helpers

func loadEmbedded[T any](filename string) (*T, error) {
	data, err := scanner.FS.ReadFile(scanner.DefaultPath(filename))
	if err != nil {
		return nil, err
	}
	var cfg T
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadFile[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg T
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func mergePolicies(defaults, user *scanner.PoliciesConfig) *scanner.PoliciesConfig {
	if defaults == nil {
		return user
	}
	if user == nil {
		return defaults
	}

	result := &scanner.PoliciesConfig{
		ComponentScanners: make(map[string][]string),
		SkipModules:       defaults.SkipModules,
		Default:           defaults.Default,
	}

	for k, v := range defaults.ComponentScanners {
		result.ComponentScanners[k] = v
	}

	for k, v := range user.ComponentScanners {
		result.ComponentScanners[k] = v
	}

	if len(user.SkipModules) > 0 {
		result.SkipModules = user.SkipModules
	}

	if len(user.Default) > 0 {
		result.Default = user.Default
	}

	return result
}
