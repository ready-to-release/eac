package validate

import "testing"

func TestSemverRegex(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		// Valid versions
		{"simple version", "1.2.3", true},
		{"zero major", "0.1.0", true},
		{"zero minor", "1.0.0", true},
		{"zero patch", "1.2.0", true},
		{"all zeros", "0.0.0", true},
		{"large numbers", "123.456.789", true},
		{"single digits", "1.0.1", true},

		// Invalid versions - prefix
		{"v prefix lowercase", "v1.2.3", false},
		{"V prefix uppercase", "V1.2.3", false},

		// Invalid versions - format
		{"missing patch", "1.2", false},
		{"missing minor and patch", "1", false},
		{"extra component", "1.2.3.4", false},
		{"leading zero major", "01.2.3", false},
		{"leading zero minor", "1.02.3", false},
		{"leading zero patch", "1.2.03", false},
		{"negative major", "-1.2.3", false},
		{"non-numeric", "a.b.c", false},
		{"spaces", "1 .2.3", false},
		{"empty string", "", false},
		{"with prerelease", "1.2.3-alpha", false},
		{"with build metadata", "1.2.3+build", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := semverRegex.MatchString(tt.version)
			if got != tt.valid {
				t.Errorf("semverRegex.MatchString(%q) = %v, want %v", tt.version, got, tt.valid)
			}
		})
	}
}
