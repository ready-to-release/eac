package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/cache"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/conf"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/docker"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/extensions"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
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
		originalLevel := logging.GetLevel()
		_ = logging.SetLevel("error") //nolint:errcheck // error level always valid
		defer func() { _ = logging.SetLevel(originalLevel) }()

		// Try to load config
		conf.InitConfig()

		if len(conf.Global.Extensions) == 0 {
			cmd.Printf("  \033[1;33m⚠️  No extensions configured - check your .r2r/r2r-cli.yml\033[0m\n")
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

// getExtensionDescription attempts to extract description from extension metadata.
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

// getExtensionIcon returns an appropriate icon for the extension based on its name/type.
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
		"sql":        "🗄️", // SQL database
		"database":   "🗄️",
		"terraform":  "🟦",  // Terraform blue
		"ansible":    "🔴",  // Ansible red
		"kubernetes": "⚙️", // Kubernetes gear
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

// getExtensionNameColor returns ANSI color codes for extension names based on their type.
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
			_ = cmd.Help() //nolint:errcheck // output error is not actionable
			return
		}

		// If no arguments provided, show help with available extensions
		if len(args) == 0 {
			_ = cmd.Help() //nolint:errcheck // output error is not actionable
			return
		}

		// Get parsed command for proper argument boundary detection
		parsedCmd, _ := GetParsedCommand() //nolint:errcheck // nil is handled below

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

		logging.Debugf("Running extension: extension=%s args=%v parsed_boundary=%d", extensionName, containerArgs, parsedCmd.ArgumentBoundary)

		// If no arguments are provided, switch to interactive mode
		// This makes "r2r pwsh" behave like "r2r interactive pwsh"
		if len(containerArgs) == 0 {
			logging.Debug("No arguments provided, switching to interactive mode")
			// Call the interactive command directly
			InteractiveCmd.Run(cmd, []string{extensionName})
			return
		}

		conf.InitConfig()

		// Create extension installer
		logging.Debug("Creating extension installer")
		installer, err := extensions.NewInstaller()
		if err != nil {
			logging.Errorf("Failed to create extension installer: %v", err)
			os.Exit(1)
		}
		defer installer.Close()

		// Get the container host for running
		host := installer.GetContainerHost()

		// Validate extensions
		logging.Debug("Validating extensions")
		if err := host.ValidateExtensions(); err != nil {
			logging.Errorf("Extension validation failed: %v", err)
			os.Exit(1)
		}

		logging.Debugf("Root directory found: root_dir=%s", host.GetRootDir())

		// Debug: List all available extensions before searching
		logging.Debugf("Available extensions in config: extension_count=%d", len(conf.Global.Extensions))
		for _, ext := range conf.Global.Extensions {
			logging.Debugf("Extension found in config: name=%s image=%s", ext.Name, ext.Image)
		}

		// Find extension
		logging.Debugf("Finding extension: extension=%v", extensionName)
		ext, err := host.FindExtension(extensionName)
		if err != nil {
			logging.Errorf("Extension '%s' not found", extensionName)
			// Ensure output is flushed before exit
			os.Stdout.Sync()
			os.Stderr.Sync()
			os.Exit(1)
		}
		logging.Debugf("Loading extension image: image=%s", ext.Image)

		// Before-snapshot removed for startup performance (~200ms savings)
		// Orphan detection now uses after-snapshot comparison with extension image only
		var beforeSnapshot map[string]string

		// Ensure image exists locally using installer
		logging.Debugf("Ensuring image exists: image=%s pull_policy=%s", ext.Image, ext.ImagePullPolicy)
		if _, err := installer.EnsureExtensionImage(extensionName); err != nil {
			logging.Errorf("Error ensuring image exists: %v", err)
			os.Exit(1)
		}

		// Inspect image
		logging.Debugf("Inspecting image: image=%v", ext.Image)
		imageInspect, err := host.InspectImage(ext.Image)
		if err != nil {
			logging.Errorf("Failed to inspect image '%s': %v", ext.Image, err)
			os.Exit(1)
		}

		// Get extension metadata for volume mounts and env vars
		var volumeRequests []cache.VolumeRequest
		extMeta, err := host.GetExtensionMetadata(ext)
		if err != nil {
			logging.Debugf("Failed to get extension metadata, continuing without metadata")
		} else if extMeta != nil {
			if len(extMeta.Volumes) > 0 {
				volumeRequests = extMeta.Volumes
				logging.Debugf("Loaded volume requests from extension metadata: extension=%s volumes=%d", ext.Name, len(volumeRequests))
			}
			// Merge env vars from metadata into extension config
			docker.MergeMetadataEnv(ext, extMeta)
		}

		// Create container configuration
		containerConfig := host.CreateContainerConfig(ext, docker.ModeRun, containerArgs, imageInspect)
		hostConfig := host.CreateHostConfig(ext, volumeRequests)

		// Create container
		logging.Debug("Creating container")
		containerID, err := host.CreateContainer(containerConfig, hostConfig)
		if err != nil {
			logging.Errorf("Failed to create container: %v", err)
			os.Exit(1)
		}
		logging.Debugf("Container created: container_id=%v", containerID)

		// Attach to container for input/output FIRST
		logging.Debugf("Attaching to container: container_id=%v", containerID)
		attachResp, err := host.AttachToContainer(containerID)
		if err != nil {
			logging.Errorf("Failed to attach to container %s: %v", containerID, err)
			os.Exit(1)
		}
		defer attachResp.Close()

		// Set up wait for container AFTER attach but BEFORE starting it
		logging.Debugf("Setting up container wait: container_id=%v", containerID)
		statusCh, errCh := host.WaitForContainer(containerID)

		// Start container
		logging.Debugf("Starting container: container_id=%v", containerID)
		if err := host.StartContainer(containerID); err != nil {
			logging.Errorf("Failed to start container %s: %v", containerID, err)
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

			logging.Debugf("Received interrupt signal, stopping container gracefully: signal=%s container_id=%s", sig.String(), containerID)

			// Start cleanup in a separate goroutine to avoid blocking
			go func() {
				// If we're running in Docker (Docker-in-Docker), clean up child containers first
				if docker.IsRunningInContainer() {
					logging.Debug("Detected Docker-in-Docker, cleaning up child containers")
					if err := host.CleanupChildContainers(); err != nil {
						logging.Warnf("Failed to clean up some child containers: error=%v", err)
					}
				}

				// Try to stop the container gracefully
				stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := host.StopContainerWithContext(stopCtx, containerID); err != nil {
					logging.Warnf("Failed to stop container gracefully, forcing termination: container_id=%s error=%v", containerID, err)

					// Force stop if graceful stop failed
					if err := host.StopContainer(containerID); err != nil {
						logging.Errorf("Failed to force stop container: %v", err)
					}
				} else {
					logging.Debugf("Container stopped gracefully: container_id=%s", containerID)
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
			// TTY mode - apply ANSI filter for interactive sessions
			// When TTY is enabled, Docker doesn't multiplex the stream
			go func() {
				// Wrap stdout with ANSI filter to remove problematic sequences
				ansiFilter := docker.NewAnsiFilter(os.Stdout)
				_, err := io.Copy(ansiFilter, attachResp.Reader)
				done <- err
			}()
		} else {
			// Non-TTY mode - use stdcopy to demultiplex the stream
			// Apply ANSI filter to remove terminal escape sequences from container output
			go func() {
				stdoutFilter := docker.NewAnsiFilter(os.Stdout)
				stderrFilter := docker.NewAnsiFilter(os.Stderr)
				_, err := stdcopy.StdCopy(stdoutFilter, stderrFilter, attachResp.Reader)
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
						_ = conn.CloseWrite() //nolint:errcheck // cleanup operation
					}
				}()

				// Copy stdin to the connection
				_, err := io.Copy(attachResp.Conn, os.Stdin)
				if err != nil && !errors.Is(err, io.EOF) {
					logging.Debugf("stdin copy error: %v", err)
				}
			}()
		}

		// Wait for container to finish (wait channels already set up before start)
		logging.Debugf("Waiting for container to finish: container_id=%v", containerID)

		// Wait for container completion
		var containerExitCode int64

		// Wait for container to exit first, then wait for I/O to complete
		select {
		case status := <-statusCh:
			logging.Debugf("Container finished: container_id=%s status_code=%d", containerID, status.StatusCode)
			containerExitCode = status.StatusCode
		case err := <-errCh:
			if err != nil {
				errStr := err.Error()
				if !strings.Contains(errStr, "No such container") && errStr != "" {
					logging.Errorf("Error waiting for container: container_id=%s error=%s", containerID, errStr)
					os.Exit(1)
				}
			}
		}

		if shuttingDown {
			os.Exit(0)
		}

		// Container has exited - now wait for I/O to complete
		// The I/O goroutine will receive EOF when Docker closes the stream
		ioErr := <-done
		if ioErr != nil && !errors.Is(ioErr, io.EOF) {
			logging.Debugf("I/O error: error=%v", ioErr)
		}
		logging.Debug("I/O copy completed")

		// Check for new containers only when auto-remove is enabled (orphan cleanup)
		// Skipped when beforeSnapshot is nil (startup optimization) since we can't detect "new" containers
		if ext.AutoRemoveChildren {
			afterSnapshot, err := host.GetContainerSnapshot()
			if err != nil {
				logging.Debugf("Failed to take container snapshot after run: error=%v", err)
			} else {
				// Get expected host images from extension metadata (for serve commands, etc.)
				var expectedHostImages []string
				if extMeta != nil {
					expectedHostImages = extMeta.ExpectedHostImages
				}
				host.WarnAboutNewContainers(beforeSnapshot, afterSnapshot, ext.Image, ext.AutoRemoveChildren, expectedHostImages)
			}
		}

		// Clean up any child containers if we're in Docker-in-Docker
		if docker.IsRunningInContainer() {
			logging.Debug("Cleaning up any remaining child containers before exit")
			if err := host.CleanupChildContainers(); err != nil {
				logging.Warnf("Failed to clean up some child containers: error=%v", err)
			}
		}

		// Exit with the same code as the container (unless we were interrupted)
		if !shuttingDown && containerExitCode != 0 {
			os.Exit(int(containerExitCode))
		}
	},
}
