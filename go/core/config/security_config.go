package config

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/paths"
	"gopkg.in/yaml.v3"

	security "github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces"
)

// SecurityConfig implements security.SecurityConfigPort.
// It loads scanner definitions and policies from the eac-security contract.
type SecurityConfig struct {
	scanners map[string]*security.ScannerDefinition
	policies *security.PoliciesConfig
	risk     *RiskConfig
}

// Verify SecurityConfig implements SecurityConfigPort.
var _ security.SecurityConfigPort = (*SecurityConfig)(nil)

// LoadSecurityConfig loads the security configuration from the eac-security contract.
// It loads scanners.yml, policies.yml, and risk-config.yml, merging defaults with user overrides.
func LoadSecurityConfig(repoRoot, configRoot string) (*SecurityConfig, error) {
	cfg := &SecurityConfig{
		scanners: make(map[string]*security.ScannerDefinition),
	}

	// Load scanners
	if err := cfg.loadScanners(repoRoot, configRoot); err != nil {
		return nil, err
	}

	// Load policies
	if err := cfg.loadPolicies(repoRoot, configRoot); err != nil {
		return nil, err
	}

	// Load risk configuration
	if err := cfg.loadRisk(repoRoot, configRoot); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadRisk loads risk configuration from risk-config.yml.
func (c *SecurityConfig) loadRisk(repoRoot, configRoot string) error {
	risk, err := LoadRiskConfig(repoRoot, configRoot)
	if err != nil {
		return err
	}
	c.risk = risk
	return nil
}

// Risk returns the risk configuration.
// Returns nil if risk config was not loaded.
func (c *SecurityConfig) Risk() security.RiskConfigPort {
	if c.risk == nil {
		return nil
	}
	return c.risk
}

// loadScanners loads scanner definitions from scanners.yml.
func (c *SecurityConfig) loadScanners(repoRoot, configRoot string) error {
	// Load contract defaults
	defaultPath := filepath.Join(repoRoot, "contracts", "security", paths.DefaultsVersion, "defaults", "scanners.yml")
	defaults, err := loadScannersFile(defaultPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Load user overrides
	userPath := filepath.Join(configRoot, "scanners.yml")
	user, err := loadScannersFile(userPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Merge: start with defaults, override with user
	if defaults != nil {
		for id, scanner := range defaults.Scanners {
			scanner.IDValue = id // Set ID from map key
			c.scanners[id] = scanner
		}
	}
	if user != nil {
		for id, scanner := range user.Scanners {
			scanner.IDValue = id
			c.scanners[id] = scanner
		}
	}

	return nil
}

// loadPolicies loads scanner policies from policies.yml.
func (c *SecurityConfig) loadPolicies(repoRoot, configRoot string) error {
	// Load contract defaults
	defaultPath := filepath.Join(repoRoot, "contracts", "security", paths.DefaultsVersion, "defaults", "policies.yml")
	defaults, err := loadPoliciesFile(defaultPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Load user overrides
	userPath := filepath.Join(configRoot, "policies.yml")
	user, err := loadPoliciesFile(userPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Merge policies
	c.policies = mergePolicies(defaults, user)
	if c.policies == nil {
		// Fallback to empty policies
		c.policies = &security.PoliciesConfig{
			ComponentScanners: make(map[string][]string),
			Default:           []string{"trivy-sbom", "trivy-vuln"},
		}
	}

	return nil
}

// GetScanner returns a scanner definition by ID.
func (c *SecurityConfig) GetScanner(id string) (security.ScannerPort, bool) {
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
		return []string{"trivy-sbom", "trivy-vuln"}
	}

	// Try component-specific policy first
	if scanners, ok := c.policies.ComponentScanners[componentType]; ok {
		return scanners
	}

	// Fall back to default
	if len(c.policies.Default) > 0 {
		return c.policies.Default
	}

	return []string{"trivy-sbom", "trivy-vuln"}
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

// Helper functions

func loadScannersFile(path string) (*security.ScannersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg security.ScannersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadPoliciesFile(path string) (*security.PoliciesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg security.PoliciesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func mergePolicies(defaults, user *security.PoliciesConfig) *security.PoliciesConfig {
	if defaults == nil {
		return user
	}
	if user == nil {
		return defaults
	}

	result := &security.PoliciesConfig{
		ComponentScanners: make(map[string][]string),
		SkipModules:       defaults.SkipModules,
		Default:           defaults.Default,
	}

	// Copy defaults
	for k, v := range defaults.ComponentScanners {
		result.ComponentScanners[k] = v
	}

	// Override with user values
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
