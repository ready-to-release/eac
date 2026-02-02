// Package interfaces provides type definitions for testing configuration.
package interfaces

// SuiteDefinition defines a test suite configuration.
// This is the concrete implementation of SuitePort.
type SuiteDefinition struct {
	// MonikerValue is the unique suite identifier (e.g., "unit").
	MonikerValue string `yaml:"moniker"`

	// NameValue is the human-readable suite name.
	NameValue string `yaml:"name"`

	// DescriptionValue describes the suite's purpose.
	DescriptionValue string `yaml:"description"`

	// ExtendedSuite marks this as an extended suite (not run by default).
	ExtendedSuite bool `yaml:"extended_suite,omitempty"`

	// SelectorsValue defines tag filters for this suite.
	SelectorsValue []SelectorDef `yaml:"selectors"`
}

// Moniker returns the suite identifier.
func (s *SuiteDefinition) Moniker() string { return s.MonikerValue }

// Name returns the suite name.
func (s *SuiteDefinition) Name() string { return s.NameValue }

// Description returns the suite description.
func (s *SuiteDefinition) Description() string { return s.DescriptionValue }

// IsExtended returns true if this is an extended suite.
func (s *SuiteDefinition) IsExtended() bool { return s.ExtendedSuite }

// Selectors returns the tag selectors.
func (s *SuiteDefinition) Selectors() []SelectorDef { return s.SelectorsValue }

// TagDefinition defines a test tag configuration.
// This is the concrete implementation of TagPort.
type TagDefinition struct {
	// TagValue is the tag value (e.g., "@L0", "@deps:<name>").
	TagValue string `yaml:"tag"`

	// NameValue is the human-readable tag name.
	NameValue string `yaml:"name,omitempty"`

	// DescriptionValue describes the tag's purpose.
	DescriptionValue string `yaml:"description,omitempty"`

	// TypeValue is the tag type (e.g., "taxonomy-level").
	TypeValue string `yaml:"type"`

	// PatternValue is the regex pattern for parameterized tags.
	PatternValue string `yaml:"pattern,omitempty"`

	// LevelValue is the tag scope level.
	LevelValue string `yaml:"level,omitempty"`

	// Example provides a usage example.
	Example string `yaml:"example,omitempty"`

	// Examples provides multiple usage examples.
	Examples []string `yaml:"examples,omitempty"`

	// Note provides additional guidance.
	Note string `yaml:"note,omitempty"`

	// ConstraintValue describes any tag constraints.
	ConstraintValue string `yaml:"constraint,omitempty"`
}

// Tag returns the tag value.
func (t *TagDefinition) Tag() string { return t.TagValue }

// Name returns the tag name.
func (t *TagDefinition) Name() string { return t.NameValue }

// Description returns the tag description.
func (t *TagDefinition) Description() string { return t.DescriptionValue }

// Type returns the tag type.
func (t *TagDefinition) Type() string { return t.TypeValue }

// Pattern returns the regex pattern.
func (t *TagDefinition) Pattern() string { return t.PatternValue }

// Level returns the tag level.
func (t *TagDefinition) Level() string { return t.LevelValue }

// Constraint returns any tag constraint.
func (t *TagDefinition) Constraint() string { return t.ConstraintValue }

// SuitesConfig holds the suite definitions from suites.yml.
type SuitesConfig struct {
	// Suites is the list of test suite definitions.
	Suites []*SuiteDefinition `yaml:"suites"`
}

// TagsConfig holds the tag definitions from tags.yml.
type TagsConfig struct {
	// Tags is the list of tag definitions.
	Tags []*TagDefinition `yaml:"tags"`

	// Types defines valid tag types.
	Types []TagType `yaml:"types,omitempty"`

	// SkipReasons defines valid skip reason codes.
	SkipReasons []SkipReason `yaml:"skip_reasons,omitempty"`
}

// TagType defines a tag type classification.
type TagType struct {
	// Type is the type identifier (e.g., "taxonomy-level").
	Type string `yaml:"type"`

	// Description describes the tag type.
	Description string `yaml:"description"`
}

// TaxonomyLevel constants for well-known taxonomy levels.
const (
	LevelL0   = "@L0"   // Very fast unit tests
	LevelL1   = "@L1"   // Fast unit tests
	LevelL2   = "@L2"   // Emulated system tests
	LevelL3   = "@L3"   // Production-like system tests
	LevelL4   = "@L4"   // Production system tests
	LevelHE2E = "@HE2E" // Horizontal end-to-end tests
)

// VerificationType constants for well-known verification types.
const (
	VerificationOV  = "@ov"  // Operational verification
	VerificationIV  = "@iv"  // Installation verification
	VerificationPV  = "@pv"  // Performance verification
	VerificationPIV = "@piv" // Production installation verification
	VerificationPPV = "@ppv" // Production performance verification
)

// TagTypeConstants for well-known tag types.
const (
	TagTypeSystemDependency = "system_dependency"
	TagTypeModuleDependency = "module_dependency"
	TagTypeEnvironment      = "environment"
	TagTypeTaxonomyLevel    = "taxonomy-level"
	TagTypeVerification     = "verification"
	TagTypeExecutionControl = "execution_control"
	TagTypeOSCALControl     = "oscal_control"
	TagTypeGxPRegulatory    = "gxp_regulatory"
)
