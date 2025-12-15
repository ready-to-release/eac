package config

// TestSuitesConfig represents the test-suites.yml configuration
type TestSuitesConfig struct {
	Suites []TestSuiteDef `yaml:"suites"`

	// Internal lookup map (built after load)
	suiteMap map[string]*TestSuiteDef
}

// TestSuiteDef defines a single test suite
type TestSuiteDef struct {
	Moniker       string        `yaml:"moniker"`
	Name          string        `yaml:"name"`
	Description   string        `yaml:"description"`
	Selectors     []SelectorDef `yaml:"selectors"`
	ExtendedSuite bool          `yaml:"extended_suite"` // If true, requires explicit selection (not in default runs)
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

// ListDefault returns monikers of suites that are included in default test runs
// (those without extended_suite: true)
func (c *TestSuitesConfig) ListDefault() []string {
	var defaults []string
	for _, suite := range c.Suites {
		if !suite.ExtendedSuite {
			defaults = append(defaults, suite.Moniker)
		}
	}
	return defaults
}

// ListNonProduction returns monikers of all suites except production-verification (@L4).
// This is used for the "all" composite suite.
func (c *TestSuitesConfig) ListNonProduction() []string {
	var suites []string
	for _, suite := range c.Suites {
		// Skip suites that require @L4 tag (production verification)
		isProduction := false
		for _, sel := range suite.Selectors {
			for _, tag := range sel.RequireTags {
				if tag == "@L4" {
					isProduction = true
					break
				}
			}
			for _, tag := range sel.AnyOfTags {
				if tag == "@L4" {
					isProduction = true
					break
				}
			}
		}
		if !isProduction {
			suites = append(suites, suite.Moniker)
		}
	}
	return suites
}

// GetLTagToSuiteMap returns a mapping from L-level tags (@L0, @L1, etc.) to suite monikers.
// This is derived from each suite's selectors' any_of_tags.
func (c *TestSuitesConfig) GetLTagToSuiteMap() map[string]string {
	tagMap := make(map[string]string)
	for _, suite := range c.Suites {
		for _, sel := range suite.Selectors {
			for _, tag := range sel.AnyOfTags {
				// Only map L-level tags
				if len(tag) >= 2 && tag[0] == '@' && tag[1] == 'L' {
					tagMap[tag] = suite.Moniker
				}
			}
		}
	}
	return tagMap
}

// GetSuiteForLTag returns the suite moniker for a given L-level tag.
// Returns empty string if no suite matches.
func (c *TestSuitesConfig) GetSuiteForLTag(ltag string) string {
	for _, suite := range c.Suites {
		for _, sel := range suite.Selectors {
			for _, tag := range sel.AnyOfTags {
				if tag == ltag {
					return suite.Moniker
				}
			}
		}
	}
	return ""
}

// GetSuiteLTags returns the L-level tags for a given suite moniker.
// Returns nil if suite not found.
func (c *TestSuitesConfig) GetSuiteLTags(moniker string) []string {
	suite := c.Get(moniker)
	if suite == nil {
		return nil
	}
	var ltags []string
	for _, sel := range suite.Selectors {
		for _, tag := range sel.AnyOfTags {
			if len(tag) >= 2 && tag[0] == '@' && tag[1] == 'L' {
				ltags = append(ltags, tag)
			}
		}
	}
	return ltags
}
