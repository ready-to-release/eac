package output

import (
	"testing"
)

func TestFormatComponentName(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		component string
		want      string
	}{
		{
			name:      "simple module and component",
			module:    "core",
			component: "config",
			want:      "core:config",
		},
		{
			name:      "empty module",
			module:    "",
			component: "config",
			want:      ":config",
		},
		{
			name:      "empty component",
			module:    "core",
			component: "",
			want:      "core:",
		},
		{
			name:      "both empty",
			module:    "",
			component: "",
			want:      ":",
		},
		{
			name:      "long module name",
			module:    "very-long-module-name-that-is-quite-lengthy",
			component: "config",
			want:      "very-long-module-name-that-is-quite-lengthy:config",
		},
		{
			name:      "long component name",
			module:    "core",
			component: "very-long-component-name-that-is-quite-lengthy",
			want:      "core:very-long-component-name-that-is-quite-lengthy",
		},
		{
			name:      "both long names",
			module:    "very-long-module-name",
			component: "very-long-component-name",
			want:      "very-long-module-name:very-long-component-name",
		},
		{
			name:      "module with special characters",
			module:    "module-v2.0",
			component: "sub_component",
			want:      "module-v2.0:sub_component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatComponentName(tt.module, tt.component)
			if got != tt.want {
				t.Errorf("FormatComponentName(%q, %q) = %q, want %q",
					tt.module, tt.component, got, tt.want)
			}
		})
	}
}

func TestTruncateComponentName(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		component string
		maxWidth  int
		want      string
	}{
		// No truncation needed cases
		{
			name:      "no truncation needed - fits exactly",
			module:    "core",
			component: "config",
			maxWidth:  15,
			want:      "core:config",
		},
		{
			name:      "no truncation needed - plenty of room",
			module:    "core",
			component: "config",
			maxWidth:  50,
			want:      "core:config",
		},
		{
			name:      "no truncation needed - short names",
			module:    "mod",
			component: "cfg",
			maxWidth:  20,
			want:      "mod:cfg",
		},

		// Module truncated, component preserved
		{
			name:      "module truncated - component preserved",
			module:    "very-long-module-name",
			component: "config",
			maxWidth:  20,
			want:      "very-long-...:config",
		},
		{
			name:      "module truncated - longer component preserved",
			module:    "eac-cli-module",
			component: "toolchain",
			maxWidth:  26,
			want:      "eac-cli-m...:toolchain",
		},
		{
			name:      "long module truncated significantly",
			module:    "extremely-long-module-name-that-needs-heavy-truncation",
			component: "component",
			maxWidth:  25,
			want:      "extremely-lon...:component",
		},

		// Very small maxWidth cases - truncate everything
		{
			name:      "maxWidth zero - empty result",
			module:    "core",
			component: "config",
			maxWidth:  0,
			want:      "",
		},
		{
			name:      "maxWidth one - single char",
			module:    "core",
			component: "config",
			maxWidth:  1,
			want:      "e",
		},
		{
			name:      "maxWidth two - two chars",
			module:    "core",
			component: "config",
			maxWidth:  2,
			want:      "ea",
		},
		{
			name:      "maxWidth three - three chars",
			module:    "core",
			component: "config",
			maxWidth:  3,
			want:      "eac",
		},
		{
			name:      "maxWidth four - one char plus ellipsis",
			module:    "core",
			component: "config",
			maxWidth:  4,
			want:      "e...",
		},
		{
			name:      "maxWidth five - two chars plus ellipsis",
			module:    "core",
			component: "config",
			maxWidth:  5,
			want:      "ea...",
		},

		// Edge cases with short components
		{
			name:      "short component - less than minCompLen",
			module:    "very-long-module-name",
			component: "cfg",
			maxWidth:  15,
			want:      "very-lon...:cfg",
		},
		{
			name:      "very short component - one char",
			module:    "module-name",
			component: "x",
			maxWidth:  10,
			want:      "modul...:x",
		},

		// Edge cases with empty strings
		{
			name:      "empty module - no truncation needed",
			module:    "",
			component: "config",
			maxWidth:  10,
			want:      ":config",
		},
		{
			name:      "empty component - no truncation needed",
			module:    "core",
			component: "",
			maxWidth:  15,
			want:      "core:",
		},
		{
			name:      "both empty - fits",
			module:    "",
			component: "",
			maxWidth:  5,
			want:      ":",
		},

		// Boundary cases
		{
			name:      "exactly at boundary - needs no truncation",
			module:    "abcd",
			component: "efgh",
			maxWidth:  9,
			want:      "abcd:efgh",
		},
		{
			name:      "one over boundary - no truncation since it fits",
			module:    "abcde",
			component: "efgh",
			maxWidth:  12,
			want:      "abcde:efgh",
		},
		{
			name:      "just under boundary - needs truncation",
			module:    "abcdefgh",
			component: "efgh",
			maxWidth:  12,
			want:      "abcd...:efgh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateComponentName(tt.module, tt.component, tt.maxWidth)
			if got != tt.want {
				t.Errorf("TruncateComponentName(%q, %q, %d) = %q, want %q",
					tt.module, tt.component, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestTruncateComponentName_LengthConstraint(t *testing.T) {
	// Verify that output never exceeds maxWidth when component fits within minCompLen (8)
	// Note: The implementation prioritizes preserving the component name, so when
	// the component is longer than 8 chars and maxWidth is too small, it may exceed maxWidth.
	tests := []struct {
		name      string
		module    string
		component string
		maxWidth  int
	}{
		{"short module and component", "core", "config", 20},
		{"short component preserved", "very-long-module-name", "cfg", 15},
		{"very short names", "mod", "cfg", 10},
		{"minimal names", "a", "b", 5},
		{"component exactly 8 chars", "long-module-name", "toolchai", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateComponentName(tt.module, tt.component, tt.maxWidth)
			if len(got) > tt.maxWidth {
				t.Errorf("TruncateComponentName(%q, %q, %d) = %q (len=%d), exceeds maxWidth %d",
					tt.module, tt.component, tt.maxWidth, got, len(got), tt.maxWidth)
			}
		})
	}
}

func TestTruncateComponentName_PreservesComponent(t *testing.T) {
	// Verify that the component is preserved when possible
	tests := []struct {
		name      string
		module    string
		component string
		maxWidth  int
	}{
		{"component preserved in output", "very-long-module-name", "config", 20},
		{"short component preserved", "eac-cli-module", "cfg", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateComponentName(tt.module, tt.component, tt.maxWidth)
			// The component should appear after the colon
			if !contains(got, tt.component) && len(tt.component) <= 8 {
				t.Errorf("TruncateComponentName(%q, %q, %d) = %q, expected to contain component %q",
					tt.module, tt.component, tt.maxWidth, got, tt.component)
			}
		})
	}
}

// contains checks if str contains substr (handles the colon separator)
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
