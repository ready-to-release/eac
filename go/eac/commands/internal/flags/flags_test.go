package flags

import (
	"strings"
	"testing"
)

func TestParseDebugFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "no debug flag",
			args: []string{"arg1", "arg2"},
			want: false,
		},
		{
			name: "debug flag long form",
			args: []string{"arg1", "--debug", "arg2"},
			want: true,
		},
		{
			name: "debug flag short form",
			args: []string{"arg1", "-d", "arg2"},
			want: true,
		},
		{
			name: "debug flag at start",
			args: []string{"--debug", "arg1", "arg2"},
			want: true,
		},
		{
			name: "debug flag at end",
			args: []string{"arg1", "arg2", "-d"},
			want: true,
		},
		{
			name: "empty args",
			args: []string{},
			want: false,
		},
		{
			name: "similar but not debug",
			args: []string{"--debugger", "-debug"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDebugFlag(tt.args)
			if got != tt.want {
				t.Errorf("ParseDebugFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		flag      string
		shorthand string
		want      bool
	}{
		{
			name:      "flag present - long form",
			args:      []string{"--all", "arg1"},
			flag:      "--all",
			shorthand: "-a",
			want:      true,
		},
		{
			name:      "flag present - short form",
			args:      []string{"-a", "arg1"},
			flag:      "--all",
			shorthand: "-a",
			want:      true,
		},
		{
			name:      "flag not present",
			args:      []string{"arg1", "arg2"},
			flag:      "--all",
			shorthand: "-a",
			want:      false,
		},
		{
			name:      "no shorthand",
			args:      []string{"--verbose"},
			flag:      "--verbose",
			shorthand: "",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasFlag(tt.args, tt.flag, tt.shorthand)
			if got != tt.want {
				t.Errorf("HasFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFlagValue(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		prefix string
		want   string
	}{
		{
			name:   "flag with value",
			args:   []string{"--target=main", "arg1"},
			prefix: "--target",
			want:   "main",
		},
		{
			name:   "flag not present",
			args:   []string{"arg1", "arg2"},
			prefix: "--target",
			want:   "",
		},
		{
			name:   "flag without value",
			args:   []string{"--target", "arg1"},
			prefix: "--target",
			want:   "",
		},
		{
			name:   "empty value",
			args:   []string{"--target="},
			prefix: "--target",
			want:   "",
		},
		{
			name:   "value with equals",
			args:   []string{"--config=/path/to=file"},
			prefix: "--config",
			want:   "/path/to=file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFlagValue(tt.args, tt.prefix)
			if got != tt.want {
				t.Errorf("GetFlagValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPositionalArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "mixed args and flags",
			args: []string{"arg1", "--flag", "arg2", "-f"},
			want: []string{"arg1", "arg2"},
		},
		{
			name: "only positional args",
			args: []string{"arg1", "arg2", "arg3"},
			want: []string{"arg1", "arg2", "arg3"},
		},
		{
			name: "only flags",
			args: []string{"--flag1", "-f", "--flag2"},
			want: nil,
		},
		{
			name: "empty args",
			args: []string{},
			want: nil,
		},
		{
			name: "empty string in args",
			args: []string{"arg1", "", "arg2"},
			want: []string{"arg1", "arg2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPositionalArgs(tt.args)
			if len(got) != len(tt.want) {
				t.Errorf("GetPositionalArgs() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetPositionalArgs()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}


func TestValidateFlags(t *testing.T) {
	// Define test flag set
	testFlags := []FlagDefinition{
		{Name: "--module", Shorthand: "-m", HasValue: true, ValueType: "string"},
		{Name: "--debug", Shorthand: "-d", HasValue: false, ValueType: "bool"},
		{Name: "--ai-token", HasValue: true, ValueType: "string", Aliases: []string{"--api-token", "--token"}},
		{Name: "--force", HasValue: false, ValueType: "bool"},
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid flags - long form",
			args:    []string{"--module", "test", "--debug"},
			wantErr: false,
		},
		{
			name:    "valid flags - short form",
			args:    []string{"-m", "test", "-d"},
			wantErr: false,
		},
		{
			name:    "valid flags - mixed",
			args:    []string{"--module", "test", "-d", "--force"},
			wantErr: false,
		},
		{
			name:    "valid flags - with equals",
			args:    []string{"--module=test", "--debug"},
			wantErr: false,
		},
		{
			name:    "positional args ignored",
			args:    []string{"--module", "test", "positional", "--debug"},
			wantErr: false,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown", "test"},
			wantErr: true,
			errMsg:  "Unknown flag: --unknown",
		},
		{
			name:    "typo - api-token instead of ai-token",
			args:    []string{"--api-token", "sk-ant-xxx"},
			wantErr: true,
			errMsg:  "--ai-token (instead of --api-token)",
		},
		{
			name:    "typo - token instead of ai-token",
			args:    []string{"--token", "sk-ant-xxx"},
			wantErr: true,
			errMsg:  "--ai-token (instead of --token)",
		},
		{
			name:    "unknown short flag",
			args:    []string{"-x"},
			wantErr: true,
			errMsg:  "Unknown flag: -x",
		},
		{
			name:    "empty args",
			args:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFlags(tt.args, testFlags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateFlags() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestSuggestSimilarFlags(t *testing.T) {
	testFlags := []FlagDefinition{
		{Name: "--ai-token", Aliases: []string{"--api-token"}},
		{Name: "--ai-provider", Aliases: []string{"--api-provider"}},
		{Name: "--module"},
		{Name: "--debug"},
	}

	tests := []struct {
		name         string
		unknownFlag  string
		wantContains string
	}{
		{
			name:         "exact alias match",
			unknownFlag:  "--api-token",
			wantContains: "--ai-token (instead of --api-token)",
		},
		{
			name:         "similar prefix",
			unknownFlag:  "--ai",
			wantContains: "--ai-token",
		},
		{
			name:         "no suggestions",
			unknownFlag:  "--xyz",
			wantContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := suggestSimilarFlags(tt.unknownFlag, testFlags)
			if tt.wantContains == "" {
				if len(suggestions) > 0 {
					t.Errorf("suggestSimilarFlags() = %v, want no suggestions", suggestions)
				}
				return
			}

			found := false
			for _, s := range suggestions {
				if strings.Contains(s, tt.wantContains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("suggestSimilarFlags() = %v, want suggestion containing %v", suggestions, tt.wantContains)
			}
		})
	}
}

func TestBuildUnknownFlagError(t *testing.T) {
	testFlags := []FlagDefinition{
		{Name: "--module", Shorthand: "-m", HasValue: true, ValueType: "string"},
		{Name: "--debug", Shorthand: "-d", HasValue: false, ValueType: "bool"},
	}

	err := buildUnknownFlagError("--unknown", testFlags)
	if err == nil {
		t.Fatal("buildUnknownFlagError() returned nil, want error")
	}

	errMsg := err.Error()

	// Check error message contains expected parts
	expectedParts := []string{
		"Unknown flag: --unknown",
		"Valid flags:",
		"--module",
		"--debug",
		"-m",
		"-d",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("buildUnknownFlagError() error missing %q, got:\n%s", part, errMsg)
		}
	}
}

