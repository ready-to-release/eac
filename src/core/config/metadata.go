package config

// Metadata holds contract version and scope information
// Used by environments and testing-tags configs
type Metadata struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Scope       string `yaml:"scope,omitempty"`
}
