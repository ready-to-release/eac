package cmd

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/docker"
	"github.com/ready-to-release/eac/src/cli/internal/extensions"
	"github.com/ready-to-release/eac/src/cli/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	// Prevent help flags from being added to run command
	RunCmd.InitDefaultHelpFlag()
	RunCmd.Flags().Lookup("help").Hidden = true

	// Custom help function for run command
	RunCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Printf("Run an extension using its configured Docker image.\n\n")
		cmd.Printf("Usage:\n  %s\n\n", cmd.UseLine())

		// Try to load configuration and show available extensions
		cmd.Printf("\n\033[1;36mAvailable Extensions:\033[0m\n")
		// Initialize configuration safely (handle errors gracefully)
		defer func() {
			if r := recover(); r != nil {
				cmd.Printf("  \033[1;33m⚠️  Unable to load configuration - run 'r2r init' to initialize\033[0m\n")
			}
		}()

		// Temporarily suppress logger output during help to avoid warning pollution
		originalLevel := log.GetLevel()
		log.SetLevel("error") // Only show errors, suppress warnings
		defer log.SetLevel(originalLevel.String())

		// Try to load config
		conf.InitConfig()

		if len(conf.Global.Extensions) == 0 {
			cmd.Printf("  \033[1;33m⚠️  No extensions configured - check your r2r-cli.yml\033[0m\n")
		} else {
			// Create container host for metadata extraction
			host, err := docker.NewContainerHost()
			if err != nil {
				// Fallback to basic display if Docker is unavailable
				for _, ext := range conf.Global.Extensions {
					description := ext.Description
					if description == "" {
						description = "No description available"
					}
					icon := getExtensionIcon(ext.Name)
					nameColor := getExtensionNameColor(ext.Name)
					cmd.Printf("  %s  %s%-13s\033[0m  \033[0;37m%s\033[0m\n", icon, nameColor, ext.Name, description)
				}
			} else {
				defer host.Close()

				for _, ext := range conf.Global.Extensions {
					description := ext.Description

					// If no description in config, try to get it from extension metadata
					if description == "" {
						if extMetadata := getExtensionDescription(host, ext.Name); extMetadata != "" {
							description = extMetadata
						} else {
							description = "No description available"
						}
					}
					icon := getExtensionIcon(ext.Name)
					nameColor := getExtensionNameColor(ext.Name)
					cmd.Printf("  %s  %s%-13s\033[0m  \033[0;37m%s\033[0m\n", icon, nameColor, ext.Name, description)
				}
			}
		}

		cmd.Printf("\nGlobal Flags:\n")
		cmd.Root().PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if !flag.Hidden {
				cmd.Printf("      --%s   %s\n", flag.Name, flag.Usage)
			}
		})
	})

	RootCmd.AddCommand(RunCmd)
}

// getExtensionDescription attempts to extract description from extension metadata
func getExtensionDescription(host *docker.ContainerHost, extensionName string) string {
	// Find the extension config
	ext, err := host.FindExtension(extensionName)
	if err != nil {
		return ""
	}

	// Try to inspect the image for labels first
	imageInspect, err := host.InspectImage(ext.Image)
	if err == nil && imageInspect.Config != nil && imageInspect.Config.Labels != nil {
		// Common Docker label conventions for descriptions
		labelKeys := []string{
			"org.opencontainers.image.description",
			"org.opencontainers.image.title",
			"description",
			"maintainer.description",
			"extension.description",
		}

		for _, key := range labelKeys {
			if desc, exists := imageInspect.Config.Labels[key]; exists && desc != "" {
				return desc
			}
		}
	}

	return ""
}

// getExtensionIcon returns an appropriate icon for the extension based on its name/type
func getExtensionIcon(extensionName string) string {
	iconMap := map[string]string{
		"pwsh":       "💙", // PowerShell blue
		"powershell": "💙",
		"python":     "🐍", // Python snake
		"py":         "🐍",
		"node":       "💚", // Node.js green
		"nodejs":     "💚",
		"js":         "💛", // JavaScript yellow
		"javascript": "💛",
		"go":         "🔵", // Go blue
		"golang":     "🔵",
		"rust":       "🦀", // Rust crab
		"rs":         "🦀",
		"docker":     "🐳", // Docker whale
		"java":       "☕", // Java coffee
		"dotnet":     "🟣", // .NET purple
		"csharp":     "🟣",
		"ruby":       "💎", // Ruby gem
		"php":        "🟦", // PHP blue
		"cpp":        "⚡", // C++ lightning
		"c++":        "⚡",
		"typescript": "🔷", // TypeScript blue diamond
		"ts":         "🔷",
		"bash":       "🐚", // Bash shell
		"sh":         "🐚",
		"sql":        "🗄️",  // SQL database
		"database":   "🗄️",
		"terraform":  "🟦", // Terraform blue
		"ansible":    "🔴", // Ansible red
		"kubernetes": "⚙️",  // Kubernetes gear
		"k8s":        "⚙️",
	}

	// Check for exact match first
	if icon, exists := iconMap[extensionName]; exists {
		return icon
	}

	// Check for partial matches
	name := strings.ToLower(extensionName)
	for key, icon := range iconMap {
		if strings.Contains(name, key) {
			return icon
		}
	}

	// Default icon for unknown extensions
	return "📦"
}

// getExtensionNameColor returns ANSI color codes for extension names based on their type
func getExtensionNameColor(extensionName string) string {
	colorMap := map[string]string{
		"pwsh":       "\033[1;34m", // Bright blue for PowerShell
		"powershell": "\033[1;34m",
		"python":     "\033[1;33m", // Bright yellow for Python
		"py":         "\033[1;33m",
		"node":       "\033[1;32m", // Bright green for Node.js
		"nodejs":     "\033[1;32m",
		"js":         "\033[1;33m", // Bright yellow for JavaScript
		"javascript": "\033[1;33m",
		"go":         "\033[1;36m", // Bright cyan for Go
		"golang":     "\033[1;36m",
		"rust":       "\033[1;31m", // Bright red for Rust
		"rs":         "\033[1;31m",
		"docker":     "\033[1;36m", // Bright cyan for Docker
		"java":       "\033[1;31m", // Bright red for Java
		"dotnet":     "\033[1;35m", // Bright magenta for .NET
		"csharp":     "\033[1;35m",
		"ruby":       "\033[1;31m", // Bright red for Ruby
		"php":        "\033[1;35m", // Bright magenta for PHP
		"cpp":        "\033[1;36m", // Bright cyan for C++
		"c++":        "\033[1;36m",
		"typescript": "\033[1;34m", // Bright blue for TypeScript
		"ts":         "\033[1;34m",
		"bash":       "\033[1;32m", // Bright green for Bash
		"sh":         "\033[1;32m",
		"sql":        "\033[1;33m", // Bright yellow for SQL
		"database":   "\033[1;33m",
		"terraform":  "\033[1;35m", // Bright magenta for Terraform
		"ansible":    "\033[1;31m", // Bright red for Ansible
		"kubernetes": "\033[1;36m", // Bright cyan for Kubernetes
		"k8s":        "\033[1;36m",
	}

	// Check for exact match first
	if color, exists := colorMap[extensionName]; exists {
		return color
	}

	// Check for partial matches
	name := strings.ToLower(extensionName)
	for key, color := range colorMap {
		if strings.Contains(name, key) {
			return color
		}
	}

	// Default color for unknown extensions
	return "\033[1;37m" // Bright white
}

var RunCmd = &cobra.Command{
	Use:                "run <extension> [args...]",
	Short:              "Run an extension from the config",
	Long:               `Run an extension using its configured Docker image.`,
	DisableFlagParsing: true, // Don't parse flags - pass them through to the extension
	Run: func(cmd *cobra.Command, args []string) {
		// Handle help flag manually since DisableFlagParsing is true
		if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
			cmd.Help()
			return
		}

		// If no arguments provided, show help with available extensions
		if len(args) == 0 {
			cmd.Help()
			return
		}

		// Create context with command info
		ctx := context.Background()
		ctx = logger.ContextWithCommand(ctx, "run")
		ctx = logger.ContextWithComponent(ctx, "docker")

		// Get logger instance
		log := logger.WithContext(ctx)

		// Get parsed command for proper argument boundary detection
		parsedCmd, _ := GetParsedCommand()

		// Use parsed command data if available, fallback to args
		extensionName := args[0]
		containerArgs := args[1:]

		// If we have a parsed command, use its container args which properly
		// handle the boundary between Viper and container arguments
		if parsedCmd != nil && parsedCmd.Subcommand == "run" {
			if parsedCmd.ExtensionName != "" {
				extensionName = parsedCmd.ExtensionName
			}
			if len(parsedCmd.ContainerArgs) > 0 || parsedCmd.ArgumentBoundary > 0 {
				// Use parsed container args (may be empty if no args after extension)
				containerArgs = parsedCmd.ContainerArgs
			}
		}

		log.WithFields(map[string]interface{}{
			"extension":      extensionName,
			"args":           containerArgs,
			"parsed_boundary": parsedCmd.ArgumentBoundary,
		}).Info().Msg("Running extension")

		// If no arguments are provided, switch to interactive mode
		// This makes "r2r pwsh" behave like "r2r interactive pwsh"
		if len(containerArgs) == 0 {
			log.Info().Msg("No arguments provided, switching to interactive mode")
			// Call the interactive command directly
			InteractiveCmd.Run(cmd, []string{extensionName})
			return
		}

		conf.InitConfig()

		// Create extension installer
		log.Debug().Msg("Creating extension installer")
		installer, err := extensions.NewInstaller()
		if err != nil {
			log.Error().Msgf("Failed to create extension installer: %v", err)
			os.Exit(1)
		}
		defer installer.Close()

		// Get the container host for running
		host := installer.GetContainerHost()

		// Validate extensions
		log.Debug().Msg("Validating extensions")
		if err := host.ValidateExtensions(); err != nil {
			log.Error().Msgf("Extension validation failed: %v", err)
			os.Exit(1)
		}

		log.WithField("root_dir", host.GetRootDir()).Debug().Msg("Root directory found")

		// Debug: List all available extensions before searching
		log.Debug().Int("extension_count", len(conf.Global.Extensions)).Msg("Available extensions in config")
		for _, ext := range conf.Global.Extensions {
			log.Debug().Str("name", ext.Name).Str("image", ext.Image).Msg("Extension found in config")
		}

		// Find extension
		log.WithField("extension", extensionName).Debug().Msg("Finding extension")
		ext, err := host.FindExtension(extensionName)
		if err != nil {
			log.Error().Msgf("Extension '%s' not found", extensionName)
			// Ensure output is flushed before exit
			os.Stdout.Sync()
			os.Stderr.Sync()
			os.Exit(1)
		}
		log.WithField("image", ext.Image).Info().Msg("Loading extension image")

		// Take snapshot of running containers before starting
		beforeSnapshot, err := host.GetContainerSnapshot()
		if err != nil {
			log.WithField("error", err.Error()).Debug().Msg("Failed to take container snapshot before run")
			beforeSnapshot = make(map[string]string) // Continue with empty snapshot
		}

		// Ensure image exists locally using installer
		log.WithFields(map[string]interface{}{
			"image":       ext.Image,
			"pull_policy": ext.ImagePullPolicy,
		}).Debug().Msg("Ensuring image exists")
		if _, err := installer.EnsureExtensionImage(extensionName); err != nil {
			log.Error().Msgf("Error ensuring image exists: %v", err)
			os.Exit(1)
		}

		// Inspect image
		log.WithField("image", ext.Image).Debug().Msg("Inspecting image")
		imageInspect, err := host.InspectImage(ext.Image)
		if err != nil {
			log.Error().Msgf("Failed to inspect image '%s': %v", ext.Image, err)
			os.Exit(1)
		}

		// Create container configuration
		containerConfig := host.CreateContainerConfig(ext, docker.ModeRun, containerArgs, imageInspect)
		hostConfig := host.CreateHostConfig()

		// Create container
		log.Debug().Msg("Creating container")
		containerID, err := host.CreateContainer(containerConfig, hostConfig)
		if err != nil {
			log.Error().Msgf("Failed to create container: %v", err)
			os.Exit(1)
		}
		log.WithField("container_id", containerID).Debug().Msg("Container created")

		// Attach to container for input/output FIRST
		log.WithField("container_id", containerID).Debug().Msg("Attaching to container")
		attachResp, err := host.AttachToContainer(containerID)
		if err != nil {
			log.Error().Msgf("Failed to attach to container %s: %v", containerID, err)
			os.Exit(1)
		}
		defer attachResp.Close()

		// Set up wait for container AFTER attach but BEFORE starting it
		log.WithField("container_id", containerID).Debug().Msg("Setting up container wait")
		statusCh, errCh := host.WaitForContainer(containerID)

		// Start container
		log.WithField("container_id", containerID).Debug().Msg("Starting container")
		if err := host.StartContainer(containerID); err != nil {
			log.Error().Msgf("Failed to start container %s: %v", containerID, err)
			os.Exit(1)
		}

		// Set up signal handling for graceful shutdown
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		// Track if we're shutting down
		shuttingDown := false

		go func() {
			sig := <-signalChan
			shuttingDown = true

			log.WithFields(map[string]interface{}{
				"signal":       sig.String(),
				"container_id": containerID,
			}).Info().Msg("Received interrupt signal, stopping container gracefully")

			// Start cleanup in a separate goroutine to avoid blocking
			go func() {
				// If we're running in Docker (Docker-in-Docker), clean up child containers first
				if docker.IsRunningInContainer() {
					log.Info().Msg("Detected Docker-in-Docker, cleaning up child containers")
					if err := host.CleanupChildContainers(); err != nil {
						log.WithField("error", err.Error()).Warn().Msg("Failed to clean up some child containers")
					}
				}

				// Try to stop the container gracefully
				stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := host.StopContainerWithContext(stopCtx, containerID); err != nil {
					log.WithFields(map[string]interface{}{
						"container_id": containerID,
						"error":        err.Error(),
					}).Warn().Msg("Failed to stop container gracefully, forcing termination")

					// Force stop if graceful stop failed
					if err := host.StopContainer(containerID); err != nil {
						log.Error().Msgf("Failed to force stop container: %v", err)
					}
				} else {
					log.WithField("container_id", containerID).Info().Msg("Container stopped gracefully")
				}
			}()

			// Give cleanup a moment to start, then exit
			time.Sleep(100 * time.Millisecond)
			os.Exit(130) // Standard exit code for SIGINT
		}()

		// Copy stdin/stdout/stderr in goroutines
		// At this point we know we have arguments (command mode)
		// since interactive mode is handled above
		done := make(chan error, 1)
		if containerConfig.Tty {
			// TTY mode with commands - apply ANSI filter
			// When TTY is enabled, Docker doesn't multiplex the stream
			go func() {
				// Wrap stdout with ANSI filter to remove problematic sequences
				ansiFilter := docker.NewAnsiFilter(os.Stdout)
				_, err := io.Copy(ansiFilter, attachResp.Reader)
				done <- err
			}()
		} else {
			// Non-TTY mode - use stdcopy to demultiplex the stream
			// This removes the 8-byte headers that cause control characters
			go func() {
				_, err := stdcopy.StdCopy(os.Stdout, os.Stderr, attachResp.Reader)
				done <- err
			}()
		}

		// Copy stdin to container if OpenStdin is enabled
		// This should work in both TTY and non-TTY modes
		if containerConfig.OpenStdin {
			go func() {
				defer func() {
					// Close stdin side of the connection when we're done
					if conn, ok := attachResp.Conn.(interface {
						CloseWrite() error
					}); ok {
						conn.CloseWrite()
					}
				}()

				// Copy stdin to the connection
				_, err := io.Copy(attachResp.Conn, os.Stdin)
				if err != nil && err != io.EOF {
					log.Debug().Err(err).Msg("stdin copy error")
				}
			}()
		}

		// Wait for container to finish (wait channels already set up before start)
		log.WithField("container_id", containerID).Debug().Msg("Waiting for container to finish")

		// Wait for container completion
		var containerExitCode int64

		// Wait for both container exit and I/O completion
		// We need to handle both to ensure we get all output
		containerDone := false
		ioDone := false

		for !containerDone || !ioDone {
			select {
			case status := <-statusCh:
				if !containerDone {
					containerDone = true
					log.Debug().Msg("Received status from container")
					if shuttingDown {
						log.WithFields(map[string]interface{}{
							"container_id": containerID,
							"status_code":  status.StatusCode,
						}).Info().Msg("Container stopped by user interrupt")
						os.Exit(0)
					}
					log.WithFields(map[string]interface{}{
						"container_id": containerID,
						"status_code":  status.StatusCode,
					}).Info().Msg("Container finished")
					containerExitCode = status.StatusCode
				}
			case err, ok := <-errCh:
				// Docker's ContainerWait error channel behavior with AutoRemove containers:
				// When a container with AutoRemove:true exits quickly, Docker removes it immediately.
				// This creates a race condition where ContainerWait might return:
				// 1. "No such container" error - container was removed before wait completed
				// 2. An error object that is not nil but has an empty message - Docker's way of
				//    signaling "wait is done but container is gone"
				// 3. nil error - normal completion
				// We must handle all three cases to avoid spurious failures in CI/CD environments
				if !containerDone && ok {
					if err != nil {
						errStr := err.Error()
						// Check if this is the "No such container" error from AutoRemove
						if strings.Contains(errStr, "No such container") {
							log.WithFields(map[string]interface{}{
								"container_id": containerID,
							}).Debug().Msg("Container already removed (AutoRemove)")
							containerDone = true
						} else if errStr != "" {
							// Real error with non-empty message - this is an actual failure
							log.WithFields(map[string]interface{}{
								"container_id": containerID,
								"error":        errStr,
							}).Error().Msg("Error waiting for container")
							os.Exit(1)
						} else {
							// Error with empty message - Docker's signal that wait completed but
							// can't provide status (container was auto-removed)
							log.Debug().Msg("Container wait completed (empty error)")
							containerDone = true
						}
					} else {
						// Nil error means the wait completed successfully
						log.Debug().Msg("Container wait completed (nil error)")
						containerDone = true
					}
				}
			case ioErr := <-done:
				if !ioDone {
					ioDone = true
					if ioErr != nil && ioErr != io.EOF {
						log.WithField("error", ioErr.Error()).Debug().Msg("I/O error")
					}
					log.Debug().Msg("I/O copy completed")
				}
			}
		}

		// Check for new containers that appeared during execution
		afterSnapshot, err := host.GetContainerSnapshot()
		if err != nil {
			log.WithField("error", err.Error()).Debug().Msg("Failed to take container snapshot after run")
		} else {
			host.WarnAboutNewContainers(beforeSnapshot, afterSnapshot, ext.Image, ext.AutoRemoveChildren)
		}

		// Clean up any child containers if we're in Docker-in-Docker
		if docker.IsRunningInContainer() {
			log.Debug().Msg("Cleaning up any remaining child containers before exit")
			if err := host.CleanupChildContainers(); err != nil {
				log.WithField("error", err.Error()).Warn().Msg("Failed to clean up some child containers")
			}
		}

		// Exit with the same code as the container (unless we were interrupted)
		if !shuttingDown && containerExitCode != 0 {
			os.Exit(int(containerExitCode))
		}

	},
}
