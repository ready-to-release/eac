package config

// SecurityToolsFileName is the config file for security tools
const SecurityToolsFileName = "security-tools.yml"

// SecurityToolsConfig holds security tool configuration
type SecurityToolsConfig struct {
	DockerImages DockerImagesConfig `yaml:"docker_images"`
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
	}
}
