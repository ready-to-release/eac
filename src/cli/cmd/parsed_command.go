package cmd

import (
	"os"
	"sync"

	commandparser "github.com/ready-to-release/eac/src/cli/internal/command-parser"
)

var (
	// Global parsed command instance - populated during PersistentPreRun
	parsedCommand     *commandparser.ParsedCommand
	parsedCommandOnce sync.Once
	parsedCommandErr  error
)

// GetParsedCommand returns the parsed command structure
// This is populated during the root command's PersistentPreRun
func GetParsedCommand() (*commandparser.ParsedCommand, error) {
	parsedCommandOnce.Do(func() {
		parser := commandparser.NewParser()
		parsedCommand = parser.Parse(os.Args)
	})
	return parsedCommand, parsedCommandErr
}

// GetContainerArgs returns the container arguments from the parsed command
// This is used by the run command to get arguments that should be passed to the container
func GetContainerArgs() []string {
	cmd, err := GetParsedCommand()
	if err != nil || cmd == nil {
		return []string{}
	}
	return cmd.ContainerArgs
}

// GetViperArgs returns the arguments that should be processed by Viper/Cobra
func GetViperArgs() []string {
	cmd, err := GetParsedCommand()
	if err != nil || cmd == nil {
		return os.Args
	}
	return cmd.ViperArgs
}

// IsRunCommand checks if the current command is a run command
func IsRunCommand() bool {
	cmd, err := GetParsedCommand()
	if err != nil || cmd == nil {
		return false
	}
	return cmd.Subcommand == "run"
}

// GetExtensionName returns the extension name from the parsed command
func GetExtensionName() string {
	cmd, err := GetParsedCommand()
	if err != nil || cmd == nil {
		return ""
	}
	return cmd.ExtensionName
}
