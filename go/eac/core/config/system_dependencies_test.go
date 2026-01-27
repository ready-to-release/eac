//go:build L0 && ov

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSystemDependency_AppliesToPhase tests existing phase filtering logic.
func TestSystemDependency_AppliesToPhase(t *testing.T) {
	tests := []struct {
		name   string
		phases []string
		phase  string
		want   bool
	}{
		{
			name:   "no phases = phase-agnostic (applies to all)",
			phases: []string{},
			phase:  "build",
			want:   true,
		},
		{
			name:   "empty phases = phase-agnostic (applies to all)",
			phases: nil,
			phase:  "test",
			want:   true,
		},
		{
			name:   "phase matches first in list",
			phases: []string{"build", "test"},
			phase:  "build",
			want:   true,
		},
		{
			name:   "phase matches last in list",
			phases: []string{"build", "test"},
			phase:  "test",
			want:   true,
		},
		{
			name:   "phase does not match",
			phases: []string{"build", "test"},
			phase:  "command",
			want:   false,
		},
		{
			name:   "single phase matches",
			phases: []string{"scan"},
			phase:  "scan",
			want:   true,
		},
		{
			name:   "single phase does not match",
			phases: []string{"scan"},
			phase:  "lint",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &SystemDependency{
				Moniker: "test-dep",
				Phases:  tt.phases,
			}
			got := dep.AppliesToPhase(tt.phase)
			assert.Equal(t, tt.want, got, "AppliesToPhase() = %v, want %v", got, tt.want)
		})
	}
}

// TestSystemDependency_AppliesToPlatform tests new platform filtering logic.
// These tests will FAIL until the AppliesToPlatform method is implemented.
func TestSystemDependency_AppliesToPlatform(t *testing.T) {
	tests := []struct {
		name      string
		platforms []string
		platform  string
		want      bool
	}{
		{
			name:      "no platforms = platform-agnostic (applies to all)",
			platforms: []string{},
			platform:  "windows",
			want:      true,
		},
		{
			name:      "empty platforms = platform-agnostic (applies to all)",
			platforms: nil,
			platform:  "linux",
			want:      true,
		},
		{
			name:      "platform matches first in list",
			platforms: []string{"linux", "darwin"},
			platform:  "linux",
			want:      true,
		},
		{
			name:      "platform matches last in list",
			platforms: []string{"linux", "darwin"},
			platform:  "darwin",
			want:      true,
		},
		{
			name:      "platform does not match",
			platforms: []string{"linux", "darwin"},
			platform:  "windows",
			want:      false,
		},
		{
			name:      "single platform matches",
			platforms: []string{"windows"},
			platform:  "windows",
			want:      true,
		},
		{
			name:      "single platform does not match",
			platforms: []string{"windows"},
			platform:  "linux",
			want:      false,
		},
		{
			name:      "multiple platforms - middle match",
			platforms: []string{"linux", "windows", "darwin"},
			platform:  "windows",
			want:      true,
		},
		{
			name:      "case sensitive match",
			platforms: []string{"linux"},
			platform:  "Linux",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &SystemDependency{
				Moniker:   "test-dep",
				Platforms: tt.platforms,
			}
			got := dep.AppliesToPlatform(tt.platform)
			assert.Equal(t, tt.want, got, "AppliesToPlatform(%q) = %v, want %v", tt.platform, got, tt.want)
		})
	}
}

// TestSystemDependency_BothFilters tests combined phase and platform filtering.
// These tests will FAIL until both AppliesToPhase and AppliesToPlatform work together.
func TestSystemDependency_BothFilters(t *testing.T) {
	tests := []struct {
		name            string
		phases          []string
		platforms       []string
		checkPhase      string
		checkPlatform   string
		wantPhase       bool
		wantPlatform    bool
		wantDescription string
	}{
		{
			name:            "both empty = applies to all",
			phases:          []string{},
			platforms:       []string{},
			checkPhase:      "build",
			checkPlatform:   "windows",
			wantPhase:       true,
			wantPlatform:    true,
			wantDescription: "universal dependency (Go, Git)",
		},
		{
			name:            "bash: wrong phase, right platform",
			phases:          []string{"command", "test"},
			platforms:       []string{"linux", "darwin"},
			checkPhase:      "build",
			checkPlatform:   "linux",
			wantPhase:       false,
			wantPlatform:    true,
			wantDescription: "should skip due to phase mismatch",
		},
		{
			name:            "bash: right phase, wrong platform",
			phases:          []string{"command", "test"},
			platforms:       []string{"linux", "darwin"},
			checkPhase:      "test",
			checkPlatform:   "windows",
			wantPhase:       true,
			wantPlatform:    false,
			wantDescription: "should skip due to platform mismatch",
		},
		{
			name:            "bash: wrong phase, wrong platform",
			phases:          []string{"command", "test"},
			platforms:       []string{"linux", "darwin"},
			checkPhase:      "build",
			checkPlatform:   "windows",
			wantPhase:       false,
			wantPlatform:    false,
			wantDescription: "should skip due to both mismatches",
		},
		{
			name:            "bash: right phase, right platform",
			phases:          []string{"command", "test"},
			platforms:       []string{"linux", "darwin"},
			checkPhase:      "test",
			checkPlatform:   "linux",
			wantPhase:       true,
			wantPlatform:    true,
			wantDescription: "should verify normally",
		},
		{
			name:            "pwsh: Windows only, build phase",
			phases:          []string{"command", "test"},
			platforms:       []string{"windows"},
			checkPhase:      "build",
			checkPlatform:   "windows",
			wantPhase:       false,
			wantPlatform:    true,
			wantDescription: "should skip due to phase mismatch",
		},
		{
			name:            "pwsh: Windows only, test phase on Linux",
			phases:          []string{"command", "test"},
			platforms:       []string{"windows"},
			checkPhase:      "test",
			checkPlatform:   "linux",
			wantPhase:       true,
			wantPlatform:    false,
			wantDescription: "should skip due to platform mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &SystemDependency{
				Moniker:   "test-dep",
				Phases:    tt.phases,
				Platforms: tt.platforms,
			}

			gotPhase := dep.AppliesToPhase(tt.checkPhase)
			gotPlatform := dep.AppliesToPlatform(tt.checkPlatform)

			assert.Equal(t, tt.wantPhase, gotPhase,
				"%s: AppliesToPhase(%q) = %v, want %v",
				tt.wantDescription, tt.checkPhase, gotPhase, tt.wantPhase)

			assert.Equal(t, tt.wantPlatform, gotPlatform,
				"%s: AppliesToPlatform(%q) = %v, want %v",
				tt.wantDescription, tt.checkPlatform, gotPlatform, tt.wantPlatform)
		})
	}
}

// TestSystemDependency_RealWorldScenarios tests realistic dependency configurations.
// These tests document expected behavior for common scenarios.
func TestSystemDependency_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name            string
		dependency      SystemDependency
		scenario        string
		checkPhase      string
		checkPlatform   string
		shouldApply     bool
		reasonIfSkipped string
	}{
		{
			name: "Go - universal build dependency",
			dependency: SystemDependency{
				Moniker:   "go",
				Phases:    []string{"build", "test"},
				Platforms: []string{}, // Empty = all platforms
			},
			scenario:      "Windows build",
			checkPhase:    "build",
			checkPlatform: "windows",
			shouldApply:   true,
		},
		{
			name: "Go - universal build dependency",
			dependency: SystemDependency{
				Moniker:   "go",
				Phases:    []string{"build", "test"},
				Platforms: []string{},
			},
			scenario:      "Linux build",
			checkPhase:    "build",
			checkPlatform: "linux",
			shouldApply:   true,
		},
		{
			name: "Bash - Unix runtime only",
			dependency: SystemDependency{
				Moniker:   "bash",
				Phases:    []string{"command", "test"},
				Platforms: []string{"linux", "darwin"},
			},
			scenario:        "Windows build",
			checkPhase:      "build",
			checkPlatform:   "windows",
			shouldApply:     false,
			reasonIfSkipped: "wrong phase + wrong platform",
		},
		{
			name: "Bash - Unix runtime only",
			dependency: SystemDependency{
				Moniker:   "bash",
				Phases:    []string{"command", "test"},
				Platforms: []string{"linux", "darwin"},
			},
			scenario:        "Linux build",
			checkPhase:      "build",
			checkPlatform:   "linux",
			shouldApply:     false,
			reasonIfSkipped: "wrong phase (not in build)",
		},
		{
			name: "Bash - Unix runtime only",
			dependency: SystemDependency{
				Moniker:   "bash",
				Phases:    []string{"command", "test"},
				Platforms: []string{"linux", "darwin"},
			},
			scenario:      "Linux test",
			checkPhase:    "test",
			checkPlatform: "linux",
			shouldApply:   true,
		},
		{
			name: "Bash - Unix runtime only",
			dependency: SystemDependency{
				Moniker:   "bash",
				Phases:    []string{"command", "test"},
				Platforms: []string{"linux", "darwin"},
			},
			scenario:        "Windows test",
			checkPhase:      "test",
			checkPlatform:   "windows",
			shouldApply:     false,
			reasonIfSkipped: "wrong platform (Windows, not Unix)",
		},
		{
			name: "PowerShell - Windows runtime only",
			dependency: SystemDependency{
				Moniker:   "pwsh",
				Phases:    []string{"command", "test"},
				Platforms: []string{"windows"},
			},
			scenario:        "Linux build",
			checkPhase:      "build",
			checkPlatform:   "linux",
			shouldApply:     false,
			reasonIfSkipped: "wrong phase + wrong platform",
		},
		{
			name: "PowerShell - Windows runtime only",
			dependency: SystemDependency{
				Moniker:   "pwsh",
				Phases:    []string{"command", "test"},
				Platforms: []string{"windows"},
			},
			scenario:      "Windows test",
			checkPhase:    "test",
			checkPlatform: "windows",
			shouldApply:   true,
		},
		{
			name: "Docker - all platforms, build phase",
			dependency: SystemDependency{
				Moniker:   "docker",
				Phases:    []string{"build", "test"},
				Platforms: []string{}, // Empty = all platforms
			},
			scenario:      "macOS build",
			checkPhase:    "build",
			checkPlatform: "darwin",
			shouldApply:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" - "+tt.scenario, func(t *testing.T) {
			phaseMatch := tt.dependency.AppliesToPhase(tt.checkPhase)
			platformMatch := tt.dependency.AppliesToPlatform(tt.checkPlatform)
			applies := phaseMatch && platformMatch

			assert.Equal(t, tt.shouldApply, applies,
				"Scenario '%s': expected shouldApply=%v (phase=%v, platform=%v). Reason: %s",
				tt.scenario, tt.shouldApply, phaseMatch, platformMatch, tt.reasonIfSkipped)
		})
	}
}
