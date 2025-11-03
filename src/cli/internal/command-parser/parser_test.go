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
			args:              []string{"r2r"},
			wantBinary:        "r2r",
			wantViperArgs:     []string{"r2r"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Binary with subcommand",
			args:              []string{"r2r", "version"},
			wantBinary:        "r2r",
			wantSubcommand:    "version",
			wantViperArgs:     []string{"r2r", "version"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Global flags before subcommand",
			args:              []string{"r2r", "--r2r-debug", "--r2r-quiet", "version"},
			wantBinary:        "r2r",
			wantGlobalFlags:   []string{"--r2r-debug", "--r2r-quiet"},
			wantSubcommand:    "version",
			wantViperArgs:     []string{"r2r", "--r2r-debug", "--r2r-quiet", "version"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Run command with extension",
			args:              []string{"r2r", "run", "pwsh"},
			wantBinary:        "r2r",
			wantSubcommand:    "run",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"r2r", "run", "pwsh"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Run command with container args",
			args:              []string{"r2r", "run", "python", "script.py", "--verbose"},
			wantBinary:        "r2r",
			wantSubcommand:    "run",
			wantExtension:     "python",
			wantViperArgs:     []string{"r2r", "run", "python"},
			wantContainerArgs: []string{"script.py", "--verbose"},
			wantBoundary:      3,
		},
		{
			name:              "Metadata command with extension",
			args:              []string{"r2r", "metadata", "pwsh"},
			wantBinary:        "r2r",
			wantSubcommand:    "metadata",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"r2r", "metadata", "pwsh"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Validate with additional args",
			args:              []string{"r2r", "validate", "config.yml", "--strict"},
			wantBinary:        "r2r",
			wantSubcommand:    "validate",
			wantViperArgs:     []string{"r2r", "validate", "config.yml", "--strict"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Complex run with global flags",
			args:              []string{"r2r", "--r2r-debug", "run", "pwsh", "Write-Host", "Hello"},
			wantBinary:        "r2r",
			wantGlobalFlags:   []string{"--r2r-debug"},
			wantSubcommand:    "run",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"r2r", "--r2r-debug", "run", "pwsh"},
			wantContainerArgs: []string{"Write-Host", "Hello"},
			wantBoundary:      4,
		},
		{
			name:              "Run with r2r flags mixed with container args",
			args:              []string{"r2r", "run", "pwsh", "--r2r-debug", "Write-Host", "--help", "Hello"},
			wantBinary:        "r2r",
			wantSubcommand:    "run",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"r2r", "run", "pwsh", "--r2r-debug"},
			wantContainerArgs: []string{"Write-Host", "--help", "Hello"},
			wantBoundary:      4,
		},
		{
			name:              "Run with only r2r flags after extension",
			args:              []string{"r2r", "run", "python", "--r2r-debug", "--r2r-quiet"},
			wantBinary:        "r2r",
			wantSubcommand:    "run",
			wantExtension:     "python",
			wantViperArgs:     []string{"r2r", "run", "python", "--r2r-debug", "--r2r-quiet"},
			wantContainerArgs: []string{},
			wantBoundary:      -1,
		},
		{
			name:              "Run with help flag after extension",
			args:              []string{"r2r", "run", "pwsh", "--help"},
			wantBinary:        "r2r",
			wantSubcommand:    "run",
			wantExtension:     "pwsh",
			wantViperArgs:     []string{"r2r", "run", "pwsh"},
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
		{"No conflict", []string{"--r2r-debug"}, false},
		{"No conflict multiple", []string{"--r2r-debug"}, false},
		{"Conflict long forms", []string{"--r2r-debug", "--r2r-quiet"}, true},
		{"Conflict short forms", []string{"--r2r-debug", "--r2r-quiet"}, true},
		{"Conflict mixed", []string{"--r2r-debug", "--r2r-quiet"}, true},
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
			args:          []string{"r2r", "version"},
			wantViper:     []string{"r2r", "version"},
			wantContainer: []string{},
		},
		{
			name:          "Run with container args",
			args:          []string{"r2r", "run", "python", "script.py"},
			wantViper:     []string{"r2r", "run", "python"},
			wantContainer: []string{"script.py"},
		},
		{
			name:          "Complex with flags",
			args:          []string{"r2r", "--r2r-debug", "run", "pwsh", "-Command", "Get-Date"},
			wantViper:     []string{"r2r", "--r2r-debug", "run", "pwsh"},
			wantContainer: []string{"-Command", "Get-Date"},
		},
		{
			name:          "Run with mixed r2r flags and container args",
			args:          []string{"r2r", "run", "python", "--r2r-debug", "script.py", "--help"},
			wantViper:     []string{"r2r", "run", "python", "--r2r-debug"},
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
