//go:build L1 && ov
// +build L1,ov

package domain

import (
	"testing"
)

// Mock validator for testing
type mockValidator struct {
	name string
}

func (m *mockValidator) Validate(output string, context map[string]interface{}) []ValidationError {
	return nil
}

func (m *mockValidator) VerifyImplementation() []ValidationError {
	return nil
}

func TestRegister(t *testing.T) {
	// Clear registry before test
	Clear()

	validator := &mockValidator{name: "test"}
	err := Register("test-validator", validator)
	if err != nil {
		t.Errorf("Expected successful registration, got error: %v", err)
	}

	// Verify it was registered
	if !Has("test-validator") {
		t.Error("Expected validator to be registered")
	}
}

func TestRegister_Duplicate(t *testing.T) {
	// Clear registry before test
	Clear()

	validator1 := &mockValidator{name: "test1"}
	validator2 := &mockValidator{name: "test2"}

	// Register first
	err := Register("test-validator", validator1)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Try to register duplicate
	err = Register("test-validator", validator2)
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}
}

func TestGet(t *testing.T) {
	// Clear registry before test
	Clear()

	validator := &mockValidator{name: "test"}
	Register("test-validator", validator)

	retrieved, err := Get("test-validator")
	if err != nil {
		t.Errorf("Expected to get validator, got error: %v", err)
	}

	if retrieved == nil {
		t.Error("Expected non-nil validator")
	}

	// Verify it's the same instance
	if mock, ok := retrieved.(*mockValidator); ok {
		if mock.name != "test" {
			t.Errorf("Expected validator name 'test', got '%s'", mock.name)
		}
	} else {
		t.Error("Retrieved validator is not mockValidator")
	}
}

func TestGet_NotFound(t *testing.T) {
	// Clear registry before test
	Clear()

	_, err := Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent validator, got nil")
	}
}

func TestHas(t *testing.T) {
	// Clear registry before test
	Clear()

	validator := &mockValidator{name: "test"}
	Register("test-validator", validator)

	if !Has("test-validator") {
		t.Error("Expected Has to return true for registered validator")
	}

	if Has("nonexistent") {
		t.Error("Expected Has to return false for nonexistent validator")
	}
}

func TestList(t *testing.T) {
	// Clear registry before test
	Clear()

	// Register multiple validators
	Register("validator1", &mockValidator{name: "v1"})
	Register("validator2", &mockValidator{name: "v2"})
	Register("validator3", &mockValidator{name: "v3"})

	names := List()
	if len(names) != 3 {
		t.Errorf("Expected 3 validators, got %d", len(names))
	}

	// Verify all names are present (order doesn't matter)
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	expected := []string{"validator1", "validator2", "validator3"}
	for _, name := range expected {
		if !nameMap[name] {
			t.Errorf("Expected to find '%s' in list", name)
		}
	}
}

func TestClear(t *testing.T) {
	// Clear registry before test
	Clear()

	// Register some validators
	Register("validator1", &mockValidator{name: "v1"})
	Register("validator2", &mockValidator{name: "v2"})

	// Verify they exist
	if len(List()) != 2 {
		t.Errorf("Expected 2 validators before clear, got %d", len(List()))
	}

	// Clear
	Clear()

	// Verify empty
	if len(List()) != 0 {
		t.Errorf("Expected 0 validators after clear, got %d", len(List()))
	}

	if Has("validator1") {
		t.Error("Expected validator1 to be removed after clear")
	}
}
