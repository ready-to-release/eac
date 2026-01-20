package gherkin

import (
	"path/filepath"
	"testing"
)

func TestIsManualScenario(t *testing.T) {
	testCases := []struct {
		name     string
		scenario ScenarioDetail
		want     bool
	}{
		{
			name: "has @Manual tag",
			scenario: ScenarioDetail{
				Tags: []string{"@Manual", "@P0"},
			},
			want: true,
		},
		{
			name: "no @Manual tag",
			scenario: ScenarioDetail{
				Tags: []string{"@Automated", "@L1"},
			},
			want: false,
		},
		{
			name: "empty tags",
			scenario: ScenarioDetail{
				Tags: []string{},
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsManualScenario(tc.scenario)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasTag(t *testing.T) {
	tags := []string{"@Manual", "@P0", "@Feature:Test"}

	if !HasTag(tags, "@Manual") {
		t.Error("should have @Manual tag")
	}

	if HasTag(tags, "@Automated") {
		t.Error("should not have @Automated tag")
	}

	if HasTag([]string{}, "@Manual") {
		t.Error("empty list should not have any tags")
	}
}

func TestFilterScenariosByTag(t *testing.T) {
	scenarios := []ScenarioDetail{
		{Name: "S1", Tags: []string{"@Manual", "@P0"}},
		{Name: "S2", Tags: []string{"@Automated", "@L1"}},
		{Name: "S3", Tags: []string{"@Manual", "@P1"}},
	}

	manual := FilterScenariosByTag(scenarios, "@Manual")
	if len(manual) != 2 {
		t.Errorf("expected 2 manual scenarios, got %d", len(manual))
	}

	automated := FilterScenariosByTag(scenarios, "@Automated")
	if len(automated) != 1 {
		t.Errorf("expected 1 automated scenario, got %d", len(automated))
	}

	none := FilterScenariosByTag(scenarios, "@NonExistent")
	if len(none) != 0 {
		t.Errorf("expected 0 scenarios with nonexistent tag, got %d", len(none))
	}
}

func TestGenerateScenarioID(t *testing.T) {
	testCases := []struct {
		name     string
		scenario ScenarioDetail
		want     string
	}{
		{
			name: "simple names",
			scenario: ScenarioDetail{
				FeatureName: "User Login",
				Name:        "Valid Credentials",
			},
			want: "user-login:valid-credentials",
		},
		{
			name: "with special characters",
			scenario: ScenarioDetail{
				FeatureName: "Test Feature #1",
				Name:        "Scenario @ Test!",
			},
			want: "test-feature-1:scenario-test",
		},
		{
			name: "with numbers",
			scenario: ScenarioDetail{
				FeatureName: "Feature 2.0",
				Name:        "Test 123",
			},
			want: "feature-20:test-123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := GenerateScenarioID(tc.scenario)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectIDCollisions(t *testing.T) {
	scenarios := []ScenarioDetail{
		{
			FeatureName: "Feature 1",
			Name:        "Test Scenario",
			FilePath:    "file1.feature",
		},
		{
			FeatureName: "Feature 1",
			Name:        "Test Scenario",
			FilePath:    "file2.feature",
		},
		{
			FeatureName: "Feature 2",
			Name:        "Unique Scenario",
			FilePath:    "file3.feature",
		},
	}

	collisions := DetectIDCollisions(scenarios)

	if len(collisions) != 1 {
		t.Fatalf("expected 1 collision, got %d", len(collisions))
	}

	expectedID := "feature-1:test-scenario"
	if _, exists := collisions[expectedID]; !exists {
		t.Errorf("expected collision for ID %q", expectedID)
	}

	if len(collisions[expectedID]) != 2 {
		t.Errorf("expected 2 scenarios with collision, got %d", len(collisions[expectedID]))
	}
}

func TestDetectIDCollisions_NoCollisions(t *testing.T) {
	scenarios := []ScenarioDetail{
		{
			FeatureName: "Feature 1",
			Name:        "Scenario 1",
		},
		{
			FeatureName: "Feature 1",
			Name:        "Scenario 2",
		},
		{
			FeatureName: "Feature 2",
			Name:        "Scenario 1",
		},
	}

	collisions := DetectIDCollisions(scenarios)

	if len(collisions) != 0 {
		t.Errorf("expected no collisions, got %d", len(collisions))
	}
}

func TestSlugify(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "Simple Text", want: "simple-text"},
		{input: "With-Hyphens", want: "with-hyphens"},
		{input: "With   Spaces", want: "with-spaces"},
		{input: "Special!@#$%Characters", want: "specialcharacters"},
		{input: "Numbers 123", want: "numbers-123"},
		{input: "   Leading and Trailing   ", want: "leading-and-trailing"},
		{input: "Multiple---Hyphens", want: "multiple-hyphens"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := slugify(tc.input)
			if got != tc.want {
				t.Errorf("slugify(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestIntegration_ParseAndFilter(t *testing.T) {
	// Test integration: parse a file and filter manual scenarios
	path := filepath.Join("testdata", "simple.feature")

	scenarios, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	manual := FilterScenariosByTag(scenarios, "@Manual")
	if len(manual) != 1 {
		t.Errorf("expected 1 manual scenario, got %d", len(manual))
	}

	if len(manual) > 0 && manual[0].Name != "Add two numbers" {
		t.Errorf("expected 'Add two numbers', got %q", manual[0].Name)
	}
}
