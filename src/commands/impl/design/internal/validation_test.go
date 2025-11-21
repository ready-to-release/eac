package design

import (
	"runtime"
	"strings"
	"testing"
)

// TestValidateModuleName_Valid tests valid module names
func TestValidateModuleName_Valid(t *testing.T) {
	tests := []struct {
		name   string
		module string
	}{
		{"simple name", "src-cli"},
		{"with dash", "my-module"},
		{"with underscore", "my_module"},
		{"with numbers", "module123"},
		{"mixed", "my-module_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModuleName(tt.module)
			if err != nil {
				t.Errorf("ValidateModuleName(%q) returned error: %v", tt.module, err)
			}
		})
	}
}

// TestValidateModuleName_Invalid tests invalid module names
func TestValidateModuleName_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		module      string
		expectError string
		windowsOnly bool // deps:windows - test only runs on Windows
	}{
		{"empty", "", "cannot be empty", false},
		{"path traversal", "../etc/passwd", "cannot contain '..'", false},
		{"path traversal 2", "foo/../bar", "cannot contain '..'", false},
		{"absolute path unix", "/etc/passwd", "cannot be an absolute path", false},
		{"absolute path windows", "C:\\Windows", "cannot be an absolute path", true},
		{"too long", strings.Repeat("a", 256), "too long", false},
		{"null byte", "foo\x00bar", "invalid character", false},
		{"asterisk", "foo*bar", "invalid character", false},
		{"question mark", "foo?bar", "invalid character", false},
		{"quotes", "foo\"bar", "invalid character", false},
		{"less than", "foo<bar", "invalid character", false},
		{"greater than", "foo>bar", "invalid character", false},
		{"pipe", "foo|bar", "invalid character", false},
		{"reserved windows CON", "CON", "reserved Windows name", true},
		{"reserved windows NUL", "NUL", "reserved Windows name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.windowsOnly && runtime.GOOS != "windows" {
				t.Skip("deps:windows - test only runs on Windows")
			}
			err := ValidateModuleName(tt.module)
			if err == nil {
				t.Errorf("ValidateModuleName(%q) expected error, got nil", tt.module)
			} else if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("ValidateModuleName(%q) error = %v, want substring %q", tt.module, err, tt.expectError)
			}
		})
	}
}

// TestCleanModuleName tests module name cleaning
func TestCleanModuleName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no change", "src-cli", "src-cli"},
		{"remove specs prefix", "specs/src-cli", "src-cli"},
		{"remove specs prefix windows", "specs\\src-cli", "src-cli"},
		{"remove design suffix", "src-cli/design", "src-cli"},
		{"remove design suffix windows", "src-cli\\design", "src-cli"},
		{"remove both", "specs/src-cli/.design", "src-cli"},
		{"remove both windows", "specs\\src-cli\\.design", "src-cli"},
		{"remove src prefix", "src/commands", "commands"},
		{"remove src prefix windows", "src\\commands", "commands"},
		{"complex", "specs/src/commands/.design", "commands"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanModuleName(tt.input)
			if result != tt.expected {
				t.Errorf("CleanModuleName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestValidateIdentifier_Valid tests valid identifiers
func TestValidateIdentifier_Valid(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
	}{
		{"simple", "myIdentifier"},
		{"with underscore", "my_identifier"},
		{"with dash", "my-identifier"},
		{"with numbers", "identifier123"},
		{"starts with uppercase", "MyIdentifier"},
		{"mixed", "My_Identifier-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.identifier)
			if err != nil {
				t.Errorf("ValidateIdentifier(%q) returned error: %v", tt.identifier, err)
			}
		})
	}
}

// TestValidateIdentifier_Invalid tests invalid identifiers
func TestValidateIdentifier_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		identifier  string
		expectError string
	}{
		{"empty", "", "cannot be empty"},
		{"starts with number", "123abc", "invalid identifier"},
		{"starts with dash", "-abc", "invalid identifier"},
		{"starts with underscore", "_abc", "invalid identifier"},
		{"contains space", "my identifier", "invalid identifier"},
		{"contains special char", "my@identifier", "invalid identifier"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.identifier)
			if err == nil {
				t.Errorf("ValidateIdentifier(%q) expected error, got nil", tt.identifier)
			} else if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("ValidateIdentifier(%q) error = %v, want substring %q", tt.identifier, err, tt.expectError)
			}
		})
	}
}
