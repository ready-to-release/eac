package contracts

import (
	"testing"
)

// Intent: Tests for TagContract loading and validation logic
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Table-driven tests with clear test case names
//   - Tests cover all TagContract methods
//   - Expected outcomes are explicit
//
// Easy to change:
//   - Tests organized by method
//   - Shared test fixtures
//
// Hard to break:
//   - Tests cover edge cases
//   - Both positive and negative cases

func TestTagContract_GetTagDefinition(t *testing.T) {
	tc := createTestTagContract()

	tests := []struct {
		tag      string
		wantName string
		wantType string
		wantNil  bool
	}{
		{"@ov", "Operational Verification", "verification", false},
		{"@L0", "Very Fast Unit Tests", "taxonomy-level", false},
		{"@deps:docker", "", "system_dependency", false},
		{"@Manual", "", "execution_control", false},
		{"@skip:wip", "", "execution_control", false},
		{"@risk-control:auth-mfa-01", "", "risk_control", false},
		{"@unknown", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			def := tc.GetTagDefinition(tt.tag)

			if tt.wantNil {
				if def != nil {
					t.Errorf("expected nil for tag %s, got %v", tt.tag, def)
				}
				return
			}

			if def == nil {
				t.Errorf("expected non-nil definition for tag %s", tt.tag)
				return
			}

			if tt.wantName != "" && def.Name != tt.wantName {
				t.Errorf("expected name '%s' for tag %s, got '%s'", tt.wantName, tt.tag, def.Name)
			}

			if def.Type != tt.wantType {
				t.Errorf("expected type '%s' for tag %s, got '%s'", tt.wantType, tt.tag, def.Type)
			}
		})
	}
}

func TestTagContract_HasConstraint(t *testing.T) {
	tc := createTestTagContract()

	tests := []struct {
		tag        string
		constraint string
		expected   bool
	}{
		{"@Manual", "mutually_exclusive_with_taxonomy_levels", true},
		{"@Manual", "other_constraint", false},
		{"@ov", "mutually_exclusive_with_taxonomy_levels", false},
		{"@unknown", "any", false},
	}

	for _, tt := range tests {
		name := tt.tag + "_" + tt.constraint
		t.Run(name, func(t *testing.T) {
			result := tc.HasConstraint(tt.tag, tt.constraint)
			if result != tt.expected {
				t.Errorf("HasConstraint(%s, %s) = %v, expected %v", tt.tag, tt.constraint, result, tt.expected)
			}
		})
	}
}

func TestTagContract_ValidateSkipTag(t *testing.T) {
	tc := createTestTagContract()

	tests := []struct {
		tag         string
		expectError bool
	}{
		{"@skip:wip", false},
		{"@skip:broken", false},
		{"@skip:flaky", false},
		{"@skip:infra", false},
		{"@skip:platform", false},
		{"@skip:deprecated", false},
		{"@skip:blocked", false},
		{"@skip:todo", false},
		{"@skip:unknown", true},
		{"@skip:UPPERCASE", true},
		{"@skip:with-dash", true},
		{"@skip:123", true},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			err := tc.validateSkipTag(tt.tag, 1)

			if tt.expectError && err == nil {
				t.Errorf("expected error for tag %s", tt.tag)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for tag %s: %v", tt.tag, err)
			}
		})
	}
}

func TestTagContract_PatternMatching(t *testing.T) {
	tc := createTestTagContract()

	tests := []struct {
		tag      string
		matches  bool
		errorMsg string
	}{
		// @skip:<reason> pattern: ^@skip:[a-z]+$
		{"@skip:wip", true, ""},
		{"@skip:abc", true, ""},
		{"@skip:WIP", false, "uppercase not allowed"},
		{"@skip:a1b", false, "numbers not allowed"},
		{"@skip:", false, "empty reason"},

		// @risk-control:<name>-<id> pattern: ^@risk-control:[a-z0-9-]+-[0-9]{2}$
		{"@risk-control:auth-mfa-01", true, ""},
		{"@risk-control:test-99", true, ""},
		{"@risk-control:a-00", true, ""},
		{"@risk-control:auth-mfa", false, "missing id"},
		{"@risk-control:auth-mfa-1", false, "id must be 2 digits"},
		{"@risk-control:AUTH-MFA-01", false, "uppercase not allowed"},

		// @risk-control:gxp-<name> pattern: ^@risk-control:gxp-[a-z0-9-]+$
		{"@risk-control:gxp-test", true, ""},
		{"@risk-control:gxp-account-lockout", true, ""},
		{"@risk-control:gxp-123", true, ""},
		{"@risk-control:gxp-", false, "empty name"},
		{"@risk-control:gxp-TEST", false, "uppercase not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			isKnown := tc.IsKnownTag(tt.tag)

			if tt.matches && !isKnown {
				t.Errorf("expected tag %s to be known (match pattern): %s", tt.tag, tt.errorMsg)
			}
			if !tt.matches && isKnown {
				t.Errorf("expected tag %s to NOT match pattern: %s", tt.tag, tt.errorMsg)
			}
		})
	}
}

func TestTagContract_InitializeWithInvalidPattern(t *testing.T) {
	tc := &TagContract{
		Tags: []TagDefinition{
			{Tag: "@bad-pattern", Pattern: "[invalid(regex"},
		},
	}

	err := tc.Initialize()
	if err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestTagContract_GetTagsByType(t *testing.T) {
	tc := createTestTagContract()

	// Verify tags are grouped correctly by type
	verificationCount := len(tc.tagsByType["verification"])
	if verificationCount != 5 {
		t.Errorf("expected 5 verification tags, got %d", verificationCount)
	}

	taxonomyCount := len(tc.tagsByType["taxonomy-level"])
	if taxonomyCount != 6 {
		t.Errorf("expected 6 taxonomy-level tags, got %d", taxonomyCount)
	}

	systemDepCount := len(tc.tagsByType["system_dependency"])
	if systemDepCount != 3 {
		t.Errorf("expected 3 system_dependency tags, got %d", systemDepCount)
	}
}

func TestTagContract_TagLookup(t *testing.T) {
	tc := createTestTagContract()

	// Exact match tags should be in lookup
	exactTags := []string{"@ov", "@iv", "@L0", "@Manual", "@gxp"}
	for _, tag := range exactTags {
		if _, ok := tc.tagLookup[tag]; !ok {
			t.Errorf("expected tag %s in tagLookup", tag)
		}
	}

	// Pattern tags should NOT be in lookup (they have <placeholders>)
	patternTags := []string{"@skip:<reason>", "@risk-control:<name>-<id>"}
	for _, tag := range patternTags {
		if _, ok := tc.tagLookup[tag]; ok {
			t.Errorf("pattern tag %s should NOT be in tagLookup", tag)
		}
	}
}

func TestTagContract_CompiledPatterns(t *testing.T) {
	tc := createTestTagContract()

	// Pattern-based tags should have compiled patterns
	patternTags := []string{"@skip:<reason>", "@risk-control:<name>-<id>", "@risk-control:gxp-<name>"}
	for _, tag := range patternTags {
		if _, ok := tc.compiledPatterns[tag]; !ok {
			t.Errorf("expected compiled pattern for %s", tag)
		}
	}
}

func TestTagContract_EmptyContract(t *testing.T) {
	tc := &TagContract{}
	err := tc.Initialize()
	if err != nil {
		t.Errorf("unexpected error initializing empty contract: %v", err)
	}

	// Should return empty slices, not nil
	if tc.GetVerificationTags() == nil {
		t.Error("GetVerificationTags should return empty slice, not nil")
	}
	if tc.GetTaxonomyLevelTags() == nil {
		t.Error("GetTaxonomyLevelTags should return empty slice, not nil")
	}
	if tc.GetValidSkipReasons() == nil {
		t.Error("GetValidSkipReasons should return empty slice, not nil")
	}

	// Unknown tags should not be known
	if tc.IsKnownTag("@any") {
		t.Error("empty contract should not know any tags")
	}
}
