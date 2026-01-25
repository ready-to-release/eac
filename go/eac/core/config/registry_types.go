// registry_types.go defines container registry configuration types.
// These types support cleanup policies for container images in registries like GHCR.
package config

// RegistriesConfig maps registry hostnames to their configurations.
// Example keys: "ghcr.io", "docker.io", "gcr.io"
type RegistriesConfig map[string]*RegistryConfig

// RegistryConfig contains configuration for a container registry.
type RegistryConfig struct {
	// Org is the organization name in the registry.
	// For GHCR, this is the GitHub organization or user.
	Org string `yaml:"org" json:"org"`

	// Cleanup defines the cleanup policy for container images.
	// If nil, cleanup is disabled for this registry.
	Cleanup *RegistryCleanupPolicy `yaml:"cleanup,omitempty" json:"cleanup,omitempty"`
}

// RegistryCleanupPolicy defines how old container images are cleaned up.
// Uses two protection mechanisms:
//  1. Image tag patterns - checked locally against OCI/Docker tags
//  2. GitHub Release API - correlates images with GitHub Releases
type RegistryCleanupPolicy struct {
	// Enabled controls whether cleanup is active for this registry.
	// Default: true
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Policy specifies the cleanup strategy.
	// Currently only "keep-latest-n" is supported.
	// Default: "keep-latest-n"
	Policy string `yaml:"policy" json:"policy"`

	// Keep specifies the number of versions to retain (for keep-latest-n policy).
	// Only applies to versions matching prune patterns.
	// Default: 10
	Keep int `yaml:"keep" json:"keep"`

	// ImageTags contains rules based on container image tags (OCI/Docker tags).
	// These are the tags visible on the image like "v1.0.0", "sha-abc123", "latest".
	ImageTags *ImageTagRules `yaml:"image_tags,omitempty" json:"image_tags,omitempty"`

	// GitHubReleases contains rules using GitHub Release API.
	// Correlates container images with GitHub Releases.
	GitHubReleases *GitHubReleaseRules `yaml:"github_releases,omitempty" json:"github_releases,omitempty"`

	// MinAgeDays specifies the minimum age in days before a version can be pruned.
	// Versions created more recently than this are protected.
	// Default: 7
	MinAgeDays int `yaml:"min_age_days" json:"min_age_days"`

	// Legacy fields for backwards compatibility (deprecated, use ImageTags instead)
	PreservePatterns []string `yaml:"preserve_patterns,omitempty" json:"preserve_patterns,omitempty"`
	PrunePatterns    []string `yaml:"prune_patterns,omitempty" json:"prune_patterns,omitempty"`
}

// ImageTagRules defines protection rules based on container image tags.
// These are OCI/Docker tags visible on the image (not git tags).
type ImageTagRules struct {
	// Preserve lists image tag patterns that must NEVER be deleted.
	// Uses glob patterns. Examples:
	//   - "v*" matches "v1.0.0", "v2.3.4-rc1"
	//   - "latest" matches exactly "latest"
	//   - "[0-9]*.[0-9]*.[0-9]*" matches semver like "1.0.0"
	Preserve []string `yaml:"preserve" json:"preserve"`

	// Prune lists image tag patterns eligible for cleanup.
	// Only images with tags matching these patterns are candidates for deletion.
	// Examples:
	//   - "sha-*" matches CI builds like "sha-abc1234"
	//   - "pr-*" matches PR builds like "pr-123"
	//   - "ci" matches exactly "ci"
	Prune []string `yaml:"prune" json:"prune"`
}

// GitHubReleaseRules defines configuration for GitHub Release API correlation.
// Images matching GitHub Releases are ALWAYS protected - this is a non-configurable
// safety feature. Released packages must never be auto-deleted.
type GitHubReleaseRules struct {
	// TagFormat specifies the expected format of release tags on images.
	// Use {module} and {version} placeholders.
	// Example: "{module}/{version}" matches "ext-eac/1.0.0"
	// Default: "{module}/{version}"
	TagFormat string `yaml:"tag_format" json:"tag_format"`
}

// DefaultRegistryCleanupPolicy returns sensible default values.
func DefaultRegistryCleanupPolicy() *RegistryCleanupPolicy {
	return &RegistryCleanupPolicy{
		Enabled: true,
		Policy:  "keep-latest-n",
		Keep:    10,
		ImageTags: &ImageTagRules{
			Preserve: []string{"v*", "latest", "[0-9]*.[0-9]*.[0-9]*"},
			Prune:    []string{"sha-*", "dev-*", "pr-*", "ci"},
		},
		GitHubReleases: &GitHubReleaseRules{
			TagFormat: "{module}/{version}",
		},
		MinAgeDays: 7,
	}
}

// GetRegistryConfig returns the configuration for a specific registry.
// Returns nil if the registry is not configured.
func (c *RepositoryConfig) GetRegistryConfig(hostname string) *RegistryConfig {
	if c.Registries == nil {
		return nil
	}
	return c.Registries[hostname]
}

// IsCleanupEnabled returns true if cleanup is enabled for the registry.
// Returns false if registry is not configured or cleanup is disabled.
func (r *RegistryConfig) IsCleanupEnabled() bool {
	if r == nil || r.Cleanup == nil {
		return false
	}
	return r.Cleanup.Enabled
}

// GetCleanupPolicy returns the cleanup policy, applying defaults where needed.
// Returns nil if registry or cleanup is not configured.
func (r *RegistryConfig) GetCleanupPolicy() *RegistryCleanupPolicy {
	if r == nil || r.Cleanup == nil {
		return nil
	}

	policy := *r.Cleanup

	// Apply defaults for zero values
	if policy.Policy == "" {
		policy.Policy = "keep-latest-n"
	}
	if policy.Keep == 0 {
		policy.Keep = 10
	}
	if policy.MinAgeDays == 0 {
		policy.MinAgeDays = 7
	}

	// Handle new ImageTags structure or fall back to legacy fields
	if policy.ImageTags == nil {
		policy.ImageTags = &ImageTagRules{}
	}
	if len(policy.ImageTags.Preserve) == 0 {
		// Use legacy field or default
		if len(policy.PreservePatterns) > 0 {
			policy.ImageTags.Preserve = policy.PreservePatterns
		} else {
			policy.ImageTags.Preserve = []string{"v*", "latest", "[0-9]*.[0-9]*.[0-9]*"}
		}
	}
	if len(policy.ImageTags.Prune) == 0 {
		// Use legacy field or default
		if len(policy.PrunePatterns) > 0 {
			policy.ImageTags.Prune = policy.PrunePatterns
		} else {
			policy.ImageTags.Prune = []string{"sha-*", "dev-*", "pr-*", "ci"}
		}
	}

	// Handle GitHubReleases defaults
	if policy.GitHubReleases == nil {
		policy.GitHubReleases = &GitHubReleaseRules{
			TagFormat: "{module}/{version}",
		}
	}

	return &policy
}

// GetPreservePatterns returns the image tag patterns to preserve.
// Uses ImageTags.Preserve if set, falls back to legacy PreservePatterns.
func (p *RegistryCleanupPolicy) GetPreservePatterns() []string {
	if p.ImageTags != nil && len(p.ImageTags.Preserve) > 0 {
		return p.ImageTags.Preserve
	}
	if len(p.PreservePatterns) > 0 {
		return p.PreservePatterns
	}
	return []string{"v*", "latest", "[0-9]*.[0-9]*.[0-9]*"}
}

// GetPrunePatterns returns the image tag patterns eligible for pruning.
// Uses ImageTags.Prune if set, falls back to legacy PrunePatterns.
func (p *RegistryCleanupPolicy) GetPrunePatterns() []string {
	if p.ImageTags != nil && len(p.ImageTags.Prune) > 0 {
		return p.ImageTags.Prune
	}
	if len(p.PrunePatterns) > 0 {
		return p.PrunePatterns
	}
	return []string{"sha-*", "dev-*", "pr-*", "ci"}
}

