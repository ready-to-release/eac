package config

// TestingTaxonomyConfig represents the testing-taxonomy.yml configuration
type TestingTaxonomyConfig struct {
	TestingTaxonomy Taxonomy `yaml:"testing_taxonomy"`
}

// Taxonomy represents the testing taxonomy definition
type Taxonomy struct {
	Version     string      `yaml:"version"`
	Description string      `yaml:"description"`
	TestLevels  []TestLevel `yaml:"test_levels"`
	Constraints Constraints `yaml:"constraints"`
	Tradeoffs   Tradeoffs   `yaml:"tradeoffs"`
}

// TestLevel represents a single test level (L0-L4, HE2E)
type TestLevel struct {
	Level                string          `yaml:"level"`
	Name                 string          `yaml:"name"`
	ShiftDirection       string          `yaml:"shift_direction"` // LEFT, RIGHT, non-shifted
	ExecutionEnvironment string          `yaml:"execution_environment"`
	Scope                string          `yaml:"scope"`
	ExternalDependencies string          `yaml:"external_dependencies"`
	Determinism          string          `yaml:"determinism"` // Highest, High, Moderate, Lowest
	DomainCoherency      string          `yaml:"domain_coherency"`
	TimeConstraints      TimeConstraints `yaml:"time_constraints"`
	Notes                string          `yaml:"notes,omitempty"`
}

// TimeConstraints defines time expectations for tests
type TimeConstraints struct {
	Preparation string `yaml:"preparation"`
	Execution   string `yaml:"execution"`
}

// Constraints defines taxonomy constraints
type Constraints struct {
	ExecutionRequirements string `yaml:"execution_requirements"`
	TestScope             string `yaml:"test_scope"`
}

// Tradeoffs defines taxonomy tradeoffs
type Tradeoffs struct {
	DeterminismVsCoherency string `yaml:"determinism_vs_coherency"`
}

// GetTestLevel returns a test level by its level identifier
func (c *TestingTaxonomyConfig) GetTestLevel(level string) (*TestLevel, bool) {
	for i := range c.TestingTaxonomy.TestLevels {
		if c.TestingTaxonomy.TestLevels[i].Level == level {
			return &c.TestingTaxonomy.TestLevels[i], true
		}
	}
	return nil, false
}

// GetAllLevels returns all test level identifiers
func (c *TestingTaxonomyConfig) GetAllLevels() []string {
	levels := make([]string, len(c.TestingTaxonomy.TestLevels))
	for i, level := range c.TestingTaxonomy.TestLevels {
		levels[i] = level.Level
	}
	return levels
}

// GetLevelsByShiftDirection returns all levels with a specific shift direction
func (c *TestingTaxonomyConfig) GetLevelsByShiftDirection(direction string) []TestLevel {
	var result []TestLevel
	for _, level := range c.TestingTaxonomy.TestLevels {
		if level.ShiftDirection == direction {
			result = append(result, level)
		}
	}
	return result
}
