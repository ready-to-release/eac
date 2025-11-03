package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/src/cli/internal/logger"
	"github.com/ready-to-release/eac/src/cli/internal/version"
	"github.com/spf13/cobra"
)

var (
	// Global logger instance
	log *logger.Logger
)

// RootCmd is the base command for the r2r CLI when called without any subcommands
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
		if os.Getenv("R2R_LOG_LEVEL") == "debug" {
			fmt.Fprintf(os.Stderr, "Parsed command: subcommand=%s, extension=%s, viper_args=%v, container_args=%v\n",
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
		if err := log.SetLevel(logLevel); err != nil {
			return fmt.Errorf("failed to set log level: %w", err)
		}

		// Add command context to logger
		ctx := context.Background()
		ctx = logger.ContextWithCommand(ctx, cmd.Name())
		if opID := os.Getenv("R2R_OPERATION_ID"); opID != "" {
			ctx = logger.ContextWithOperationID(ctx, opID)
		}

		// Check if we fixed redirect pollution
		if os.Getenv("R2R_FIXED_REDIRECT") == "true" {
			log.WithContext(ctx).WithFields(map[string]interface{}{
				"original_args": os.Getenv("R2R_ORIGINAL_ARGS"),
				"filtered_args": os.Getenv("R2R_FILTERED_ARGS"),
				"fix_applied":   "redirect_pollution",
			}).Warn().Msg("Fixed bash redirect pollution in arguments (removed spurious '2' from '2>&1')")

			// Clean up env vars
			os.Unsetenv("R2R_FIXED_REDIRECT")
			os.Unsetenv("R2R_ORIGINAL_ARGS")
			os.Unsetenv("R2R_FILTERED_ARGS")
		}

		// Log command execution
		log.WithContext(ctx).WithFields(map[string]interface{}{
			"args":    args,
			"version": version.Version,
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
		}).Debug().Msg("Executing command")

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.WithField("error", err).Error().Msg("Failed to show help")
		}
	},
}

func init() {
	// Initialize logger with default configuration
	logCfg := logger.Config{
		Console: true,
		Level:   "info",
		File:    "", // Disable file logging by default
	}

	// Check if file logging is explicitly requested
	if os.Getenv("R2R_LOG_FILE") != "" {
		logCfg.File = os.Getenv("R2R_LOG_FILE")
	} else if os.Getenv("R2R_ENABLE_FILE_LOG") == "true" {
		logCfg.File = "r2r.log"
	}

	// Initialize logger
	if err := logger.Initialize(logCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log = logger.Get()

	// Only log initialization in debug/verbose mode
	if os.Getenv("R2R_VERBOSE_LOG") == "true" || os.Getenv("R2R_LOG_LEVEL") == "debug" {
		versionInfo := version.GetInfo()
		log.WithFields(map[string]interface{}{
			"version":   versionInfo.Version,
			"commit":    versionInfo.Commit,
			"timestamp": versionInfo.Timestamp,
		}).Debug().Msg("R2R CLI initialized")
	}

	RootCmd.CompletionOptions.DisableDefaultCmd = true

	// Keep normal help functionality for most commands
	// Only the run command will disable help flags since it needs to pass them through

	// Add flags first
	RootCmd.PersistentFlags().Bool("r2r-debug", false, "Enable debug logging")
	RootCmd.PersistentFlags().Bool("r2r-quiet", false, "Disable all output except errors")

	// Now validate version after flags are initialized
	if err := version.Validate(false); err != nil {
		log.WithField("error", err).Fatal().Msg("Version validation failed")
	}

	// Initialize extension aliases for direct execution (e.g., "r2r pwsh" instead of "r2r run pwsh")
	InitializeExtensionAliases()
}

func buildErrorContext(err error) map[string]interface{} {
	fields := map[string]interface{}{
		"error":   err.Error(),
		"command": RootCmd.Name(),
		"args":    os.Args[1:],
		"version": version.Version,
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}

	cmd, _, e := RootCmd.Find(os.Args[1:])
	if e != nil {
		fields["error_type"] = "invalid_command"
		fields["attempted_command"] = strings.Join(os.Args[1:], " ")
	} else if cmd != nil {
		fields["subcommand"] = cmd.Name()
		fields["path"] = cmd.CommandPath()
	}

	for _, name := range []string{"r2r-debug", "r2r-quiet"} {
		if flag := RootCmd.PersistentFlags().Lookup(name); flag != nil && flag.Changed {
			fields[name] = flag.Value.String() == "true"
		}
	}

	if e == nil {
		fields["error_type"] = "execution_failed"
	}
	return fields
}

func Execute() {
	// Check if --version flag is passed as first argument
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		// Execute the version command instead
		versionCmd.Run(versionCmd, []string{})
		return
	}

	if err := RootCmd.Execute(); err != nil {
		fields := buildErrorContext(err)
		log.WithFields(fields).Error().Msg("Command execution failed")
		os.Exit(1)
	}
}
