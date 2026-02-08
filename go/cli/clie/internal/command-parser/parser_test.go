//go:build L1
// +build L1

package commandparser

import (
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name              string
		args              []string
		wantBinary        string
		wantGlobalFlags   []string
		wantSubcommand    string
		wantExtension     string
		wantViperArgs     []string
		wantContainerArgs []string
		wantBoundary      int
	}{
		{
			name:              "Empty args",
			args:              []string{},
			wantViperArgs:     []string{},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Binary only",
			args:              []string{"clie"},
			wantBinary:        "clie",
			wantViperArgs:     []string{"clie"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Binary with subcommand",
			args:              []string{"clie", "version"},
			wantBinary:        "clie",
			wantSubcommand:    "version",
			wantViperArgs:     []string{"clie", "version"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Global flags before subcommand",
			args:              []string{"clie", "--clie-debug", "--clie-quiet", "version"},
			wantBinary:        "clie",
			wantGlobalFlags:   []string{"--clie-debug", "--clie-quiet"},
			wantSubcommand:    "version",
			wantViperArgs:     []string{"clie", "--clie-debug", "--clie-quiet", "version"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Run command with extension",
			args:              []string{"clie", "run", "pwsh"},
			wantBinary:        "clie",
			wantSubcommand:    "run",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"clie", "run", "pwsh"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Run command with container args",
			args:              []string{"clie", "run", "python", "script.py", "--verbose"},
			wantBinary:        "clie",
			wantSubcommand:    "run",
			wantExtension:     "python",
			wantViperArgs:     []string{"clie", "run", "python"},
			wantContainerArgs: []string{"script.py", "--verbose"},
			wantBoundary:      3,
		},
		{
			name:              "Metadata command with extension",
			args:              []string{"clie", "metadata", "pwsh"},
			wantBinary:        "clie",
			wantSubcommand:    "metadata",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"clie", "metadata", "pwsh"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Validate with additional args",
			args:              []string{"clie", "validate", "config.yml", "--strict"},
			wantBinary:        "clie",
			wantSubcommand:    "validate",
			wantViperArgs:     []string{"clie", "validate", "config.yml", "--strict"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Complex run with global flags",
			args:              []string{"clie", "--clie-debug", "run", "pwsh", "Write-Host", "Hello"},
			wantBinary:        "clie",
			wantGlobalFlags:   []string{"--clie-debug"},
			wantSubcommand:    "run",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"clie", "--clie-debug", "run", "pwsh"},
			wantContainerArgs: []string{"Write-Host", "Hello"},
			wantBoundary:      4,
		},
		{
			name:              "Run with clie flags mixed with container args",
			args:              []string{"clie", "run", "pwsh", "--clie-debug", "Write-Host", "--help", "Hello"},
			wantBinary:        "clie",
			wantSubcommand:    "run",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"clie", "run", "pwsh", "--clie-debug"},
			wantContainerArgs: []string{"Write-Host", "--help", "Hello"},
			wantBoundary:      4,
		},
		{
			name:              "Run with only clie flags after extension",
			args:              []string{"clie", "run", "python", "--clie-debug", "--clie-quiet"},
			wantBinary:        "clie",
			wantSubcommand:    "run",
			wantExtension:     "python",
			wantViperArgs:     []string{"clie", "run", "python", "--clie-debug", "--clie-quiet"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Run with help flag after extension",
			args:              []string{"clie", "run", "pwsh", "--help"},
			wantBinary:        "clie",
			wantSubcommand:    "run",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"clie", "run", "pwsh"},
			wantContainerArgs: []string{"--help"},
			wantBoundary:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := p.Parse(tt.args)

			if cmd.BinaryName != tt.wantBinary {
				t.Errorf("BinaryName = %v, want %v", cmd.BinaryName, tt.wantBinary)
			}

			if !reflect.DeepEqual(cmd.GlobalFlags, tt.wantGlobalFlags) &&
				!(len(cmd.GlobalFlags) == 0 && len(tt.wantGlobalFlags) == 0) {
				t.Errorf("GlobalFlags = %v, want %v", cmd.GlobalFlags, tt.wantGlobalFlags)
			}

			if cmd.Subcommand != tt.wantSubcommand {
				t.Errorf("Subcommand = %v, want %v", cmd.Subcommand, tt.wantSubcommand)
			}

			if cmd.ExtensionName != tt.wantExtension {
				t.Errorf("ExtensionName = %v, want %v", cmd.ExtensionName, tt.wantExtension)
			}

			if !reflect.DeepEqual(cmd.ViperArgs, tt.wantViperArgs) {
				t.Errorf("ViperArgs = %v, want %v", cmd.ViperArgs, tt.wantViperArgs)
			}

			if !reflect.DeepEqual(cmd.ContainerArgs, tt.wantContainerArgs) {
				t.Errorf("ContainerArgs = %v, want %v", cmd.ContainerArgs, tt.wantContainerArgs)
			}

			if cmd.ArgumentBoundary != tt.wantBoundary {
				t.Errorf("ArgumentBoundary = %v, want %v", cmd.ArgumentBoundary, tt.wantBoundary)
			}
		})
	}
}

func TestParser_IsValidExtensionName(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid simple", "pwsh", true},
		{"Valid with hyphen", "my-extension", true},
		{"Valid with underscore", "my_ext", true},
		{"Valid with number", "ext123", true},
		{"Valid mixed", "My_Ext-123", true},
		{"Invalid empty", "", false},
		{"Invalid starts with number", "123ext", false},
		{"Invalid starts with hyphen", "-ext", false},
		{"Invalid special char", "ext@name", false},
		{"Invalid space", "ext name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.IsValidExtensionName(tt.input); got != tt.want {
				t.Errorf("IsValidExtensionName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParser_HasConflictingGlobalFlags(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name  string
		flags []string
		want  bool
	}{
		{"No conflict", []string{"--clie-debug"}, false},
		{"No conflict multiple", []string{"--clie-debug"}, false},
		{"Conflict long forms", []string{"--clie-debug", "--clie-quiet"}, true},
		{"Conflict short forms", []string{"--clie-debug", "--clie-quiet"}, true},
		{"Conflict mixed", []string{"--clie-debug", "--clie-quiet"}, true},
		{"Empty", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.HasConflictingGlobalFlags(tt.flags); got != tt.want {
				t.Errorf("HasConflictingGlobalFlags(%v) = %v, want %v", tt.flags, got, tt.want)
			}
		})
	}
}

func TestParser_SplitArguments(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name          string
		args          []string
		wantViper     []string
		wantContainer []string
	}{
		{
			name:          "No container args",
			args:          []string{"clie", "version"},
			wantViper:     []string{"clie", "version"},
			wantContainer: []string{},
		},
		{
			name:          "Run with container args",
			args:          []string{"clie", "run", "python", "script.py"},
			wantViper:     []string{"clie", "run", "python"},
			wantContainer: []string{"script.py"},
		},
		{
			name:          "Complex with flags",
			args:          []string{"clie", "--clie-debug", "run", "pwsh", "-Command", "Get-Date"},
			wantViper:     []string{"clie", "--clie-debug", "run", "pwsh"},
			wantContainer: []string{"-Command", "Get-Date"},
		},
		{
			name:          "Run with mixed clie flags and container args",
			args:          []string{"clie", "run", "python", "--clie-debug", "script.py", "--help"},
			wantViper:     []string{"clie", "run", "python", "--clie-debug"},
			wantContainer: []string{"script.py", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper, container := p.SplitArguments(tt.args)

			if !reflect.DeepEqual(viper, tt.wantViper) {
				t.Errorf("SplitArguments() viper = %v, want %v", viper, tt.wantViper)
			}

			if !reflect.DeepEqual(container, tt.wantContainer) {
				t.Errorf("SplitArguments() container = %v, want %v", container, tt.wantContainer)
			}
		})
	}
}
