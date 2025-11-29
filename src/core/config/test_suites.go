package config

// TestSuitesConfig represents the test-suites.yml configuration
type TestSuitesConfig struct {
	Metadata TestSuitesMetadata `yaml:"metadata"`
	Suites   []TestSuiteDef     `yaml:"suites"`

	// Internal lookup map (built after load)
	suiteMap map[string]*TestSuiteDef
}

// TestSuitesMetadata contains schema metadata
type TestSuitesMetadata struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// TestSuiteDef defines a single test suite
type TestSuiteDef struct {
	Moniker     string        `yaml:"moniker"`
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Selectors   []SelectorDef `yaml:"selectors"`
}

// SelectorDef specifies criteria for selecting tests based on tags
type SelectorDef struct {
	RequireTags []string `yaml:"require_tags"`
	AnyOfTags   []string `yaml:"any_of_tags"`
	ExcludeTags []string `yaml:"exclude_tags"`
}

// buildSuiteMap constructs the internal lookup map
func (c *TestSuitesConfig) buildSuiteMap() {
	c.suiteMap = make(map[string]*TestSuiteDef, len(c.Suites))
	for i := range c.Suites {
		c.suiteMap[c.Suites[i].Moniker] = &c.Suites[i]
	}
}

// Get retrieves a suite definition by its moniker
func (c *TestSuitesConfig) Get(moniker string) *TestSuiteDef {
	if c.suiteMap == nil {
		c.buildSuiteMap()
	}
	return c.suiteMap[moniker]
}

// List returns all available suite monikers
func (c *TestSuitesConfig) List() []string {
	monikers := make([]string, len(c.Suites))
	for i, suite := range c.Suites {
		monikers[i] = suite.Moniker
	}
	return monikers
}

// GetAll returns all suite definitions
func (c *TestSuitesConfig) GetAll() []TestSuiteDef {
	return c.Suites
}
