package cmd

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/docker"
	"github.com/ready-to-release/eac/src/cli/internal/logger"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(InteractiveCmd)
}

var InteractiveCmd = &cobra.Command{
	Use:   "interactive <extension>",
	Short: "Start an extension container in interactive mode",
	Long:  `Start an extension container in interactive mode with shell access.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf.InitConfig()

		// Create container host
		host, err := docker.NewContainerHost()
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		defer host.Close()

		// Validate extensions
		if err := host.ValidateExtensions(); err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}

		cmd.Println("Root directory found:", host.GetRootDir())

		// Find extension
		extensionName := args[0]
		ext, err := host.FindExtension(extensionName)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		cmd.Println("Loading extension image:", ext.Image)

		// Ensure image exists locally (pull if necessary)
		if err := host.EnsureImageExists(ext.Image, ext.ImagePullPolicy, ext.LoadLocal); err != nil {
			cmd.PrintErrf("Error ensuring image exists: %v\n", err)
			os.Exit(1)
		}

		// Inspect image
		imageInspect, err := host.InspectImage(ext.Image)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}

		// Create container configuration
		containerConfig := host.CreateContainerConfig(ext, docker.ModeInteractive, nil, imageInspect)
		hostConfig := host.CreateHostConfig()

		// Create and start container
		containerID, err := host.CreateContainer(containerConfig, hostConfig)
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}

		if err := host.StartContainer(containerID); err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}

		cmd.Printf("Starting interactive session for extension '%s'...\n", extensionName)
		cmd.Println("Type 'exit' to quit the interactive session.")

		// Set up signal handling for graceful shutdown
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-signalChan

			// Create context for logging
			ctx := context.Background()
			ctx = logger.ContextWithCommand(ctx, "interactive")
			ctx = logger.ContextWithComponent(ctx, "docker")
			log := logger.WithContext(ctx)

			log.WithFields(map[string]interface{}{
				"signal":       sig.String(),
				"container_id": containerID,
			}).Info().Msg("Received interrupt signal, stopping container gracefully")

			// Try to stop the container gracefully with a timeout
			stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := host.StopContainerWithContext(stopCtx, containerID); err != nil {
				log.Warn().Msgf("Failed to stop container gracefully: %v, forcing termination", err)
				// Force stop if graceful stop failed
				host.StopContainer(containerID)
			} else {
				log.Info().Msg("Container stopped gracefully")
			}
			os.Exit(0)
		}()

		// For containers with entrypoint, attach to the running container
		// For containers without entrypoint, exec into the shell
		if len(imageInspect.Config.Entrypoint) > 0 {
			// Container has entrypoint - attach to it directly
			execCmd := exec.Command("docker", "attach", containerID)
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				cmd.PrintErrln("Error attaching to container:", err)
				os.Exit(1)
			}
		} else {
			// Container has no entrypoint - exec into shell
			execCmd := exec.Command("docker", "exec", "-it", containerID, "/bin/sh")
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			if err := execCmd.Run(); err != nil {
				cmd.PrintErrln("Error running interactive session:", err)
				os.Exit(1)
			}
		}

		cmd.Println("Interactive session ended.")
	},
}
