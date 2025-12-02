package config

// SystemDependenciesConfig holds the system dependencies configuration
type SystemDependenciesConfig struct {
	Dependencies []SystemDependency `yaml:"dependencies"`

	// Runtime lookup map (built after load)
	depMap map[string]*SystemDependency
}

// SystemDependency defines a single system dependency
type SystemDependency struct {
	Moniker     string                   `yaml:"moniker"`
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description,omitempty"`
	Version     string                   `yaml:"version"`
	Verify      SystemDependencyVerify   `yaml:"verify"`
}

// SystemDependencyVerify defines how to verify a dependency is available
type SystemDependencyVerify struct {
	// Command-based verification
	Command string `yaml:"command,omitempty"`
	Pattern string `yaml:"pattern,omitempty"`

	// Environment variable verification
	EnvVars []string `yaml:"env_vars,omitempty"`
	Require string   `yaml:"require,omitempty"` // "any" or "all"
}

// IsCommandBased returns true if this verification uses a command
func (v *SystemDependencyVerify) IsCommandBased() bool {
	return v.Command != ""
}

// IsEnvBased returns true if this verification uses environment variables
func (v *SystemDependencyVerify) IsEnvBased() bool {
	return len(v.EnvVars) > 0
}

// buildDepMap builds the lookup map for quick access by moniker
func (c *SystemDependenciesConfig) buildDepMap() {
	c.depMap = make(map[string]*SystemDependency)
	for i := range c.Dependencies {
		dep := &c.Dependencies[i]
		c.depMap[dep.Moniker] = dep
	}
}

// Get returns a system dependency by moniker
func (c *SystemDependenciesConfig) Get(moniker string) *SystemDependency {
	if c.depMap == nil {
		c.buildDepMap()
	}
	return c.depMap[moniker]
}

// GetMonikers returns all defined dependency monikers
func (c *SystemDependenciesConfig) GetMonikers() []string {
	monikers := make([]string, len(c.Dependencies))
	for i, dep := range c.Dependencies {
		monikers[i] = dep.Moniker
	}
	return monikers
}

// HasMoniker returns true if the moniker is defined
func (c *SystemDependenciesConfig) HasMoniker(moniker string) bool {
	return c.Get(moniker) != nil
}
