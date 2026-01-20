package gherkin

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile_Simple(t *testing.T) {
	path := filepath.Join("testdata", "simple.feature")

	scenarios, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(scenarios) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(scenarios))
	}

	// Check first scenario
	s1 := scenarios[0]
	if s1.Name != "Add two numbers" {
		t.Errorf("scenario 1 name: got %q, want %q", s1.Name, "Add two numbers")
	}
	if s1.FeatureName != "test-module_simple-calculator" {
		t.Errorf("scenario 1 feature: got %q, want %q", s1.FeatureName, "test-module_simple-calculator")
	}
	if !containsTag(s1.Tags, "@Manual") {
		t.Error("scenario 1 should have @Manual tag")
	}
	if !containsTag(s1.Tags, "@ov") {
		t.Error("scenario 1 should have @ov tag")
	}
	if len(s1.Steps) != 3 {
		t.Errorf("scenario 1 steps: got %d, want 3", len(s1.Steps))
	}

	// Check second scenario
	s2 := scenarios[1]
	if s2.Name != "Subtract two numbers" {
		t.Errorf("scenario 2 name: got %q, want %q", s2.Name, "Subtract two numbers")
	}
	if !containsTag(s2.Tags, "@L1") {
		t.Error("scenario 2 should have @L1 tag")
	}
	if !containsTag(s2.Tags, "@ov") {
		t.Error("scenario 2 should have @ov tag")
	}
}

func TestParseFile_WithDescription(t *testing.T) {
	path := filepath.Join("testdata", "with-description.feature")

	scenarios, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(scenarios))
	}

	s := scenarios[0]
	if s.Name != "User login with valid credentials" {
		t.Errorf("name: got %q, want %q", s.Name, "User login with valid credentials")
	}

	if s.Description == "" {
		t.Error("description should not be empty")
	}

	// The gherkin library preserves indentation in descriptions
	if !strings.Contains(s.Description, "This scenario tests the happy path") {
		t.Errorf("description should contain expected text, got: %q", s.Description)
	}
	if !strings.Contains(s.Description, "valid username and password") {
		t.Errorf("description should contain expected text, got: %q", s.Description)
	}

	if !containsTag(s.Tags, "@Manual") {
		t.Error("should have @Manual tag")
	}
	if !containsTag(s.Tags, "@ov") {
		t.Error("should have @ov tag")
	}
	if !containsTag(s.Tags, "@control:ac-2") {
		t.Error("should have @control:ac-2 tag")
	}

	if len(s.Steps) != 4 {
		t.Errorf("steps: got %d, want 4", len(s.Steps))
	}
}

func TestParseFile_NonExistent(t *testing.T) {
	path := filepath.Join("testdata", "nonexistent.feature")

	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestParseFile_InvalidGherkin(t *testing.T) {
	// We'll test this by writing empty content or malformed Gherkin
	// For now, we'll skip since gherkin-go might be lenient
	t.Skip("TODO: Add test for invalid Gherkin once we understand gherkin-go error handling")
}

// Helper function to check if a tag exists in a list
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
