// Package testing provides core testing utilities and tag system implementation
package testing

import (
	test "github.com/ready-to-release/eac/contracts/runner/0.1.0/test"
)

// TestReference is the canonical test reference type, defined in contracts/runner.
// All fields are documented in the contract package.
type TestReference = test.TestReference

// TestSuite defines a selector for tests based on tags.
type TestSuite struct {
	Moniker     string        // Canonical identifier (e.g., "commit")
	Name        string        // Human-readable name
	Description string        // What this suite tests
	Selectors   []TagSelector // Tag selection criteria
	Inferences  []Inference   // Tag inference rules
}

// TagSelector specifies criteria for selecting tests.
type TagSelector struct {
	RequireTags []string // AND logic - must have ALL
	AnyOfTags   []string // OR logic - must have at least ONE
	ExcludeTags []string // NOT logic - must NOT have any
}

// Inference defines automatic tag additions based on conditions.
type Inference struct {
	TestTypes   []string // Apply only to these test types (optional)
	IfTags      []string // Condition: has ALL these tags
	ThenAddTags []string // Action: add these tags
	Description string   // Human-readable description
}

// RiskControlRef represents a parsed risk control reference.
type RiskControlRef struct {
	FullTag     string // Complete tag (e.g., "@risk-control:auth-mfa-01")
	ControlName string // Control name (e.g., "auth-mfa")
	ScenarioID  string // Scenario ID (e.g., "01"), empty for GxP controls
	IsGxP       bool   // Is this a GxP risk control?
}
