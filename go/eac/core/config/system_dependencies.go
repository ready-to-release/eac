package config

// SystemDependenciesConfig holds the system dependencies configuration.
type SystemDependenciesConfig struct {
	Dependencies           []SystemDependency  `yaml:"dependencies"`
	CapabilityRequirements map[string][]string `yaml:"capability_requirements,omitempty"`

	// Runtime lookup map (built after load)
	depMap map[string]*SystemDependency
}

// SystemDependency defines a single system dependency.
type SystemDependency struct {
	Moniker     string                 `yaml:"moniker"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Version     string                 `yaml:"version"`
	Phases      []string               `yaml:"phases,omitempty"` // command, build, test, scan, lint
	Verify      SystemDependencyVerify `yaml:"verify"`
}

// AppliesToPhase returns true if this dependency is needed for the given phase.
// If no phases are specified, the dependency is phase-agnostic (e.g., OS platforms).
func (d *SystemDependency) AppliesToPhase(phase string) bool {
	if len(d.Phases) == 0 {
		return true // Phase-agnostic deps always apply
	}
	for _, p := range d.Phases {
		if p == phase {
			return true
		}
	}
	return false
}

// SystemDependencyVerify defines how to verify a dependency is available.
type SystemDependencyVerify struct {
	// Command-based verification
	Command string `yaml:"command,omitempty"`
	Pattern string `yaml:"pattern,omitempty"`

	// Environment variable verification
	EnvVars []string `yaml:"env_vars,omitempty"`
	Require string   `yaml:"require,omitempty"` // "any" or "all"

	// OS platform verification (linux, windows, darwin)
	OSPlatform string `yaml:"os_platform,omitempty"`
}

// IsCommandBased returns true if this verification uses a command.
func (v *SystemDependencyVerify) IsCommandBased() bool {
	return v.Command != ""
}

// IsEnvBased returns true if this verification uses environment variables.
func (v *SystemDependencyVerify) IsEnvBased() bool {
	return len(v.EnvVars) > 0
}

// IsOSPlatformBased returns true if this verification checks OS platform.
func (v *SystemDependencyVerify) IsOSPlatformBased() bool {
	return v.OSPlatform != ""
}

// buildDepMap builds the lookup map for quick access by moniker.
func (c *SystemDependenciesConfig) buildDepMap() {
	c.depMap = make(map[string]*SystemDependency)
	for i := range c.Dependencies {
		dep := &c.Dependencies[i]
		c.depMap[dep.Moniker] = dep
	}
}

// Get returns a system dependency by moniker.
func (c *SystemDependenciesConfig) Get(moniker string) *SystemDependency {
	if c.depMap == nil {
		c.buildDepMap()
	}
	return c.depMap[moniker]
}

// GetMonikers returns all defined dependency monikers.
func (c *SystemDependenciesConfig) GetMonikers() []string {
	monikers := make([]string, len(c.Dependencies))
	for i, dep := range c.Dependencies {
		monikers[i] = dep.Moniker
	}
	return monikers
}

// HasMoniker returns true if the moniker is defined.
func (c *SystemDependenciesConfig) HasMoniker(moniker string) bool {
	return c.Get(moniker) != nil
}

// GetRequiredDeps returns the system dependencies required for the given capabilities.
// Uses the capability_requirements mapping to resolve capabilities to dependency monikers.
func (c *SystemDependenciesConfig) GetRequiredDeps(capabilities []string) []string {
	if c.CapabilityRequirements == nil {
		return nil
	}

	seen := make(map[string]bool)
	var deps []string

	for _, cap := range capabilities {
		if reqs, ok := c.CapabilityRequirements[cap]; ok {
			for _, dep := range reqs {
				if !seen[dep] {
					seen[dep] = true
					deps = append(deps, dep)
				}
			}
		}
	}

	return deps
}

// GetCapabilityDeps returns the system dependencies for a single capability.
func (c *SystemDependenciesConfig) GetCapabilityDeps(capability string) []string {
	if c.CapabilityRequirements == nil {
		return nil
	}
	return c.CapabilityRequirements[capability]
}
