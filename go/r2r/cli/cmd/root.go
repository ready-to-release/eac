package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/version"
	"github.com/spf13/cobra"
)

// RootCmd is the base command for the r2r CLI when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "r2r",
	Short: "Ready to Release - Enterprise-grade automation framework",
	Long:  `r2r CLI standardizes and containerizes development workflows through a portable, scalable automation framework.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Parse the command early to validate structure and detect argument boundaries
		// This doesn't bypass Cobra/Viper - it provides additional validation
		// and helps with argument boundary detection for the run command
		parsedCmd, err := GetParsedCommand()
		if err != nil {
			return fmt.Errorf("command parsing failed: %w", err)
		}

		// Log parsed command details in debug mode
		if logging.IsDebugEnabled() {
			logging.Debugf("Parsed command: subcommand=%s, extension=%s, viper_args=%v, container_args=%v",
				parsedCmd.Subcommand, parsedCmd.ExtensionName, parsedCmd.ViperArgs, parsedCmd.ContainerArgs)
		}

		debug, err := cmd.Flags().GetBool("r2r-debug")
		if err != nil {
			return fmt.Errorf("failed to get r2r-debug flag: %w", err)
		}
		quiet, err := cmd.Flags().GetBool("r2r-quiet")
		if err != nil {
			return fmt.Errorf("failed to get r2r-quiet flag: %w", err)
		}
		if debug && quiet {
			return fmt.Errorf("cannot use both debug and quiet flags")
		}

		// Determine log level from flags and environment variable
		logLevel := "info"

		// Check R2R_LOG_LEVEL environment variable first
		if envLogLevel := os.Getenv("R2R_LOG_LEVEL"); envLogLevel != "" {
			logLevel = envLogLevel
		}

		// Command-line flags override environment variable
		if debug {
			logLevel = "debug"
		} else if quiet {
			logLevel = "error"
		}

		// Set the log level
		if err := logging.SetLevel(logLevel); err != nil {
			return fmt.Errorf("failed to set log level: %w", err)
		}

		// Check if we fixed redirect pollution
		if os.Getenv("R2R_FIXED_REDIRECT") == "true" {
			logging.Warnf("Fixed bash redirect pollution in arguments (removed spurious '2' from '2>&1'): original=%s filtered=%s",
				os.Getenv("R2R_ORIGINAL_ARGS"), os.Getenv("R2R_FILTERED_ARGS"))

			// Clean up env vars
			os.Unsetenv("R2R_FIXED_REDIRECT")
			os.Unsetenv("R2R_ORIGINAL_ARGS")
			os.Unsetenv("R2R_FILTERED_ARGS")
		}

		// Log command execution
		logging.Debugf("Executing command: cmd=%s args=%v version=%s os=%s arch=%s",
			cmd.Name(), args, version.Version, runtime.GOOS, runtime.GOARCH)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			logging.Errorf("Failed to show help: %v", err)
		}
	},
}

func init() {
	// Initialize logger with defaults from environment
	logging.InitFromEnv()

	// Only log initialization in debug/verbose mode
	if os.Getenv("R2R_VERBOSE_LOG") == "true" || os.Getenv("R2R_LOG_LEVEL") == "debug" {
		logging.EnableDebug()
		versionInfo := version.GetInfo()
		logging.Debugf("R2R CLI initialized: version=%s commit=%s timestamp=%s",
			versionInfo.Version, versionInfo.Commit, versionInfo.Timestamp)
	}

	RootCmd.CompletionOptions.DisableDefaultCmd = true

	// Keep normal help functionality for most commands
	// Only the run command will disable help flags since it needs to pass them through

	// Add flags first
	RootCmd.PersistentFlags().Bool("r2r-debug", false, "Enable debug logging")
	RootCmd.PersistentFlags().Bool("r2r-quiet", false, "Disable all output except errors")

	// Now validate version after flags are initialized
	if err := version.Validate(false); err != nil {
		logging.Fatalf("Version validation failed: %v", err)
	}

	// Initialize extension aliases for direct execution (e.g., "r2r eac" instead of "r2r run eac")
	InitializeExtensionAliases()
}

func buildErrorContext(err error) string {
	parts := []string{
		fmt.Sprintf("error=%v", err),
		fmt.Sprintf("command=%s", RootCmd.Name()),
		fmt.Sprintf("args=%v", os.Args[1:]),
		fmt.Sprintf("version=%s", version.Version),
		fmt.Sprintf("os=%s", runtime.GOOS),
		fmt.Sprintf("arch=%s", runtime.GOARCH),
	}

	cmd, _, e := RootCmd.Find(os.Args[1:])
	if e != nil {
		parts = append(parts, "error_type=invalid_command", fmt.Sprintf("attempted_command=%s", strings.Join(os.Args[1:], " ")))
	} else if cmd != nil {
		parts = append(parts, fmt.Sprintf("subcommand=%s", cmd.Name()), fmt.Sprintf("path=%s", cmd.CommandPath()))
	}

	for _, name := range []string{"r2r-debug", "r2r-quiet"} {
		if flag := RootCmd.PersistentFlags().Lookup(name); flag != nil && flag.Changed {
			parts = append(parts, fmt.Sprintf("%s=%s", name, flag.Value.String()))
		}
	}

	if e == nil {
		parts = append(parts, "error_type=execution_failed")
	}
	return strings.Join(parts, " ")
}

func Execute() {
	// Check if --version flag is passed as first argument
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		// Execute the version command instead
		versionCmd.Run(versionCmd, []string{})
		return
	}

	if err := RootCmd.Execute(); err != nil {
		logging.Errorf("Command execution failed: %s", buildErrorContext(err))
		os.Exit(1)
	}
}
