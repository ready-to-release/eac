// Package testing provides core testing utilities and tag system implementation
package testing

// TestReference identifies a specific test with its tags
type TestReference struct {
	FilePath string   // Path to test file
	Type     string   // "godog", "gotest"
	TestName string   // Name of test/scenario
	Tags     []string // All effective tags (inherited + explicit + inferred)

	// Execution control
	IsIgnored bool // Has @ignore tag
	IsManual  bool // Has @Manual tag

	// Risk control linkage
	RiskControls []string // All @risk-control:<name>-<id> references

	// GxP regulatory classification
	IsGxP            bool // Has @gxp tag
	IsCriticalAspect bool // Has @critical-aspect tag
}

// TestSuite defines a selector for tests based on tags
type TestSuite struct {
	Moniker     string        // Canonical identifier (e.g., "pre-commit")
	Name        string        // Human-readable name
	Description string        // What this suite tests
	Selectors   []TagSelector // Tag selection criteria
	Inferences  []Inference   // Tag inference rules
}

// TagSelector specifies criteria for selecting tests
type TagSelector struct {
	RequireTags []string // AND logic - must have ALL
	AnyOfTags   []string // OR logic - must have at least ONE
	ExcludeTags []string // NOT logic - must NOT have any
}

// Inference defines automatic tag additions based on conditions
type Inference struct {
	TestTypes   []string // Apply only to these test types (optional)
	IfTags      []string // Condition: has ALL these tags
	ThenAddTags []string // Action: add these tags
	Description string   // Human-readable description
}

// RiskControlRef represents a parsed risk control reference
type RiskControlRef struct {
	FullTag     string // Complete tag (e.g., "@risk-control:auth-mfa-01")
	ControlName string // Control name (e.g., "auth-mfa")
	ScenarioID  string // Scenario ID (e.g., "01"), empty for GxP controls
	IsGxP       bool   // Is this a GxP risk control?
}
