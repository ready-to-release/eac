package config

// SecurityToolsFileName is the config file for security tools
const SecurityToolsFileName = "security-tools.yml"

// SecurityToolsConfig holds security tool configuration
type SecurityToolsConfig struct {
	DockerImages    DockerImagesConfig  `yaml:"docker_images"`
	DefaultScanners map[string][]string `yaml:"default_scanners,omitempty"`
	SkipModules     []string            `yaml:"skip_modules,omitempty"`
}

// DockerImagesConfig holds Docker image specifications
type DockerImagesConfig struct {
	Trivy   DockerImage `yaml:"trivy"`
	Semgrep DockerImage `yaml:"semgrep"`
	ZAP     DockerImage `yaml:"zap"`
}

// DockerImage represents a versioned Docker image
type DockerImage struct {
	Image       string `yaml:"image"`
	Tag         string `yaml:"tag"`
	Description string `yaml:"description,omitempty"`
}

// FullImage returns the complete Docker image reference (image:tag)
func (d *DockerImage) FullImage() string {
	return d.Image + ":" + d.Tag
}

// GetDefaultScanners returns the default scanners for a given module type.
// Falls back to "default" key if the specific module type is not configured.
// Returns a slice of scanner type strings (e.g., ["sbom", "vuln", "secrets"]).
func (c *SecurityToolsConfig) GetDefaultScanners(moduleType string) []string {
	if c.DefaultScanners == nil {
		return defaultScannerList()
	}

	// Try module-specific scanners first
	if scanners, ok := c.DefaultScanners[moduleType]; ok && len(scanners) > 0 {
		return scanners
	}

	// Fall back to default
	if scanners, ok := c.DefaultScanners["default"]; ok && len(scanners) > 0 {
		return scanners
	}

	return defaultScannerList()
}

// defaultScannerList returns the built-in default scanner list
func defaultScannerList() []string {
	return []string{"sbom", "vuln", "secrets"}
}

// GetSkipModules returns the list of modules to skip during scanning.
// Returns an empty slice if no skip list is configured.
func (c *SecurityToolsConfig) GetSkipModules() []string {
	if c.SkipModules == nil {
		return []string{}
	}
	return c.SkipModules
}

// ShouldSkipModule returns true if the given module moniker should be skipped.
func (c *SecurityToolsConfig) ShouldSkipModule(moniker string) bool {
	for _, skip := range c.SkipModules {
		if skip == moniker {
			return true
		}
	}
	return false
}

// DefaultSecurityToolsConfig returns default configuration
func DefaultSecurityToolsConfig() SecurityToolsConfig {
	return SecurityToolsConfig{
		DockerImages: DockerImagesConfig{
			Trivy: DockerImage{
				Image: "ghcr.io/aquasecurity/trivy",
				Tag:   "latest",
			},
			Semgrep: DockerImage{
				Image: "semgrep/semgrep",
				Tag:   "latest",
			},
			ZAP: DockerImage{
				Image: "zaproxy/zap-stable",
				Tag:   "latest",
			},
		},
		DefaultScanners: map[string][]string{
			"default": {"sbom", "vuln", "secrets"},
		},
		SkipModules: []string{},
	}
}
