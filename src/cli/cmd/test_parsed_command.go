//go:build L1

package cmd

import (
	"os"
	"sync"
	"testing"

	commandparser "github.com/ready-to-release/eac/src/cli/internal/command-parser"
)

func TestParsedCommandIntegration(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedSubcmd    string
		expectedExtension string
		expectedContainer []string
		expectedBoundary  int
	}{
		{
			name:              "simple version",
			args:              []string{"r2r", "version"},
			expectedSubcmd:    "version",
			expectedExtension: "",
			expectedContainer: []string{},
			expectedBoundary:  -1,
		},
		{
			name:              "run with extension",
			args:              []string{"r2r", "run", "pwsh"},
			expectedSubcmd:    "run",
			expectedExtension: "pwsh",
			expectedContainer: []string{},
			expectedBoundary:  -1,
		},
		{
			name:              "run with container args",
			args:              []string{"r2r", "run", "pwsh", "Get-Date"},
			expectedSubcmd:    "run",
			expectedExtension: "pwsh",
			expectedContainer: []string{"Get-Date"},
			expectedBoundary:  3,
		},
		{
			name:              "run with complex container args",
			args:              []string{"r2r", "--debug", "run", "python", "script.py", "--arg1", "value"},
			expectedSubcmd:    "run",
			expectedExtension: "python",
			expectedContainer: []string{"script.py", "--arg1", "value"},
			expectedBoundary:  4,
		},
		{
			name:              "metadata with extension",
			args:              []string{"r2r", "metadata", "pwsh", "--json"},
			expectedSubcmd:    "metadata",
			expectedExtension: "pwsh",
			expectedContainer: []string{},
			expectedBoundary:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the global parsed command
			parsedCommand = nil
			parsedCommandOnce = sync.Once{}
			parsedCommandErr = nil

			// Set test args
			oldArgs := os.Args
			os.Args = tt.args
			defer func() { os.Args = oldArgs }()

			// Get parsed command
			parsed, err := GetParsedCommand()
			if err != nil {
				t.Fatalf("Failed to parse command: %v", err)
			}

			if parsed.Subcommand != tt.expectedSubcmd {
				t.Errorf("Subcommand: expected %s, got %s", tt.expectedSubcmd, parsed.Subcommand)
			}

			if parsed.ExtensionName != tt.expectedExtension {
				t.Errorf("Extension: expected %s, got %s", tt.expectedExtension, parsed.ExtensionName)
			}

			if len(parsed.ContainerArgs) != len(tt.expectedContainer) {
				t.Errorf("Container args count: expected %d, got %d",
					len(tt.expectedContainer), len(parsed.ContainerArgs))
			}

			for i, expected := range tt.expectedContainer {
				if i >= len(parsed.ContainerArgs) || parsed.ContainerArgs[i] != expected {
					t.Errorf("Container arg[%d]: expected %s, got %s",
						i, expected, parsed.ContainerArgs[i])
				}
			}

			if parsed.ArgumentBoundary != tt.expectedBoundary {
				t.Errorf("Boundary: expected %d, got %d", tt.expectedBoundary, parsed.ArgumentBoundary)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// Set up a test parsed command
	parsedCommand = &commandparser.ParsedCommand{
		BinaryName:       "r2r",
		Subcommand:       "run",
		ExtensionName:    "pwsh",
		ContainerArgs:    []string{"Get-Date", "-Format", "yyyy-MM-dd"},
		ViperArgs:        []string{"r2r", "run", "pwsh"},
		ArgumentBoundary: 3,
	}
	parsedCommandOnce = sync.Once{}
	parsedCommandOnce.Do(func() {}) // Mark as done

	t.Run("GetContainerArgs", func(t *testing.T) {
		args := GetContainerArgs()
		if len(args) != 3 {
			t.Errorf("Expected 3 container args, got %d", len(args))
		}
		if args[0] != "Get-Date" {
			t.Errorf("First container arg should be 'Get-Date', got %s", args[0])
		}
	})

	t.Run("GetViperArgs", func(t *testing.T) {
		args := GetViperArgs()
		if len(args) != 3 {
			t.Errorf("Expected 3 Viper args, got %d", len(args))
		}
		if args[2] != "pwsh" {
			t.Errorf("Last Viper arg should be 'pwsh', got %s", args[2])
		}
	})

	t.Run("IsRunCommand", func(t *testing.T) {
		if !IsRunCommand() {
			t.Error("Should identify as run command")
		}
	})

	t.Run("GetExtensionName", func(t *testing.T) {
		if GetExtensionName() != "pwsh" {
			t.Errorf("Extension name should be 'pwsh', got %s", GetExtensionName())
		}
	})
}
