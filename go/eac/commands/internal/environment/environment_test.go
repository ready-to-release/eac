package environment

import (
	"os"
	"testing"
)

func TestDetect_LocalConsole(t *testing.T) {
	// Clear CI-related env vars
	os.Unsetenv("CI")
	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("GITLAB_CI")
	os.Unsetenv("R2R_TEST_RUN_ID")
	os.Unsetenv("GODOG_FORMAT")
	os.Unsetenv("R2R_MOCK_SECURITY")

	env := Detect()

	if env.IsCI {
		t.Error("IsCI should be false without CI env vars")
	}
	if env.IsTestContext {
		t.Error("IsTestContext should be false without test env vars")
	}
	// Note: IsContainer depends on logging.GetExecutionContext() which we can't easily mock
}

func TestDetect_CI(t *testing.T) {
	// Clear all first
	os.Unsetenv("CI")
	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("GITLAB_CI")

	tests := []struct {
		name   string
		envVar string
	}{
		{"CI", "CI"},
		{"GITHUB_ACTIONS", "GITHUB_ACTIONS"},
		{"GITLAB_CI", "GITLAB_CI"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVar, "true")
			defer os.Unsetenv(tt.envVar)

			env := Detect()
			if !env.IsCI {
				t.Errorf("IsCI should be true with %s set", tt.envVar)
			}
		})
	}
}

func TestDetectWithTestVars(t *testing.T) {
	// Clear env vars
	os.Unsetenv("CUSTOM_TEST_VAR")

	env := DetectWithTestVars("CUSTOM_TEST_VAR")
	if env.IsTestContext {
		t.Error("IsTestContext should be false without custom env var")
	}

	os.Setenv("CUSTOM_TEST_VAR", "true")
	defer os.Unsetenv("CUSTOM_TEST_VAR")

	env = DetectWithTestVars("CUSTOM_TEST_VAR")
	if !env.IsTestContext {
		t.Error("IsTestContext should be true with custom env var set")
	}
}

func TestShouldUseTUI(t *testing.T) {
	tests := []struct {
		name     string
		env      Env
		expected bool
	}{
		{"LocalConsole", Env{IsLocalConsole: true}, true},
		{"CI", Env{IsCI: true}, false},
		{"Container", Env{IsContainer: true}, false},
		{"TestContext", Env{IsTestContext: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.env.ShouldUseTUI(); got != tt.expected {
				t.Errorf("ShouldUseTUI() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateTUI(t *testing.T) {
	tests := []struct {
		name          string
		env           Env
		explicitlySet bool
		useTUI        bool
		wantErr       bool
	}{
		{"Not explicitly set", Env{IsCI: true}, false, true, false},
		{"Explicitly disabled", Env{IsCI: true}, true, false, false},
		{"CI with TUI", Env{IsCI: true}, true, true, true},
		{"Container with TUI", Env{IsContainer: true}, true, true, true},
		{"LocalConsole with TUI", Env{IsLocalConsole: true}, true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.ValidateTUI(tt.explicitlySet, tt.useTUI)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTUI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContextName(t *testing.T) {
	tests := []struct {
		name     string
		env      Env
		expected string
	}{
		{"CI", Env{IsCI: true}, "CI"},
		{"Container", Env{IsContainer: true}, "container"},
		{"TestContext", Env{IsTestContext: true}, "test"},
		{"Local", Env{IsLocalConsole: true}, "local"},
		{"Default", Env{}, "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.env.ContextName(); got != tt.expected {
				t.Errorf("ContextName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
