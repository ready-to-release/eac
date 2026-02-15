package riskassess

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testAbsProfilePath returns an absolute path suitable for the current OS.
func testAbsProfilePath() string {
	if runtime.GOOS == "windows" {
		return `C:\path\to\profile.json`
	}
	return "/path/to/profile.json"
}

// testWorkspaceRoot returns a workspace root suitable for the current OS.
func testWorkspaceRoot() string {
	if runtime.GOOS == "windows" {
		return `C:\workspace`
	}
	return "/workspace"
}

func TestParseAssessFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		config      *AssessConfig
		wantErr     bool
		errContains string
		validate    func(t *testing.T, c *AssessConfig)
	}{
		{
			name:   "empty args",
			args:   []string{},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				assert.Empty(t, c.ProfilePath)
			},
		},
		{
			name:   "profile flag with long form absolute path",
			args:   []string{"--profile", testAbsProfilePath()},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				assert.Equal(t, testAbsProfilePath(), c.ProfilePath)
			},
		},
		{
			name:   "profile flag with short form absolute path",
			args:   []string{"-p", testAbsProfilePath()},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				assert.Equal(t, testAbsProfilePath(), c.ProfilePath)
			},
		},
		{
			name:   "profile flag with relative path joins workspace root",
			args:   []string{"--profile", "specs/profile.json"},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				expected := filepath.Join(testWorkspaceRoot(), "specs", "profile.json")
				assert.Equal(t, expected, c.ProfilePath)
			},
		},
		{
			name:        "profile flag without value",
			args:        []string{"--profile"},
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "--profile requires a value",
		},
		{
			name:        "short profile flag without value",
			args:        []string{"-p"},
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "--profile requires a value",
		},
		{
			name:   "max-evidence-age flag with valid duration",
			args:   []string{"--max-evidence-age", "48h"},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				assert.Equal(t, 48*time.Hour, c.MaxEvidenceAge)
			},
		},
		{
			name:   "max-evidence-age flag with minutes",
			args:   []string{"--max-evidence-age", "30m"},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				assert.Equal(t, 30*time.Minute, c.MaxEvidenceAge)
			},
		},
		{
			name:        "max-evidence-age flag without value",
			args:        []string{"--max-evidence-age"},
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "--max-evidence-age requires a value",
		},
		{
			name:        "max-evidence-age with invalid duration",
			args:        []string{"--max-evidence-age", "invalid"},
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "invalid duration",
		},
		{
			name:        "max-evidence-age with negative duration",
			args:        []string{"--max-evidence-age", "-1h"},
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "must be positive",
		},
		{
			name:        "max-evidence-age with zero duration",
			args:        []string{"--max-evidence-age", "0s"},
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "must be positive",
		},
		{
			name:        "max-evidence-age exceeds 30 days",
			args:        []string{"--max-evidence-age", "744h"}, // 31 days
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "too large",
		},
		{
			name:   "max-evidence-age at exactly 30 days",
			args:   []string{"--max-evidence-age", "720h"}, // 30 days
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				assert.Equal(t, 30*24*time.Hour, c.MaxEvidenceAge)
			},
		},
		{
			name:   "debug flag is ignored",
			args:   []string{"--debug"},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				// debug is handled separately, should not error
			},
		},
		{
			name:   "short debug flag is ignored",
			args:   []string{"-d"},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				// debug is handled separately, should not error
			},
		},
		{
			name:   "sequential flag is ignored",
			args:   []string{"--sequential"},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				// sequential is handled separately, should not error
			},
		},
		{
			name:        "unknown flag",
			args:        []string{"--unknown-flag"},
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "unknown flag",
		},
		{
			name:        "unexpected positional argument after flags",
			args:        []string{"--debug", "extra-arg"},
			config:      &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			wantErr:     true,
			errContains: "unexpected positional argument",
		},
		{
			name: "multiple flags combined",
			args: []string{
				"--profile", testAbsProfilePath(),
				"--max-evidence-age", "12h",
				"--debug",
				"--sequential",
			},
			config: &AssessConfig{WorkspaceRoot: testWorkspaceRoot()},
			validate: func(t *testing.T, c *AssessConfig) {
				assert.Equal(t, testAbsProfilePath(), c.ProfilePath)
				assert.Equal(t, 12*time.Hour, c.MaxEvidenceAge)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAssessFlags(tt.config, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, tt.config)
			}
		})
	}
}

func TestAssessConfig_Defaults(t *testing.T) {
	t.Run("default MaxEvidenceAge is 24h", func(t *testing.T) {
		config := &AssessConfig{
			MaxEvidenceAge: 24 * time.Hour,
		}
		assert.Equal(t, 24*time.Hour, config.MaxEvidenceAge)
	})

	t.Run("struct zero values", func(t *testing.T) {
		config := &AssessConfig{}
		assert.Empty(t, config.Modules)
		assert.Empty(t, config.ProfilePath)
		assert.Empty(t, config.WorkspaceRoot)
		assert.Empty(t, config.OutputDir)
		assert.False(t, config.Sequential)
		assert.False(t, config.Debug)
	})
}

func TestModuleAssessmentResult_Fields(t *testing.T) {
	t.Run("zero value result", func(t *testing.T) {
		r := &ModuleAssessmentResult{}
		assert.Empty(t, r.Module)
		assert.Nil(t, r.AssessmentResults)
		assert.Nil(t, r.Evidence)
		assert.Empty(t, r.OutputPath)
		assert.Equal(t, 0, r.Satisfied)
		assert.Equal(t, 0, r.NotSatisfied)
		assert.Nil(t, r.RiskScore)
		assert.Nil(t, r.Error)
		assert.Nil(t, r.Warnings)
	})

	t.Run("result with error", func(t *testing.T) {
		r := &ModuleAssessmentResult{
			Module: "failed-module",
			Error:  assert.AnError,
		}
		assert.Error(t, r.Error)
		assert.Equal(t, "failed-module", r.Module)
	})

	t.Run("result with warnings", func(t *testing.T) {
		r := &ModuleAssessmentResult{
			Module:   "warned-module",
			Warnings: []string{"warning 1", "warning 2"},
		}
		assert.Len(t, r.Warnings, 2)
	})
}
