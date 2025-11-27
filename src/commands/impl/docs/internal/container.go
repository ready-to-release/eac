package docs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/src/commands/internal/serve"
	"github.com/ready-to-release/eac/src/core/logging"
	"github.com/ready-to-release/eac/src/core/repository"
	"go.uber.org/zap"
)

const (
	// containerNameBase is the base name for mkdocs containers
	// Actual containers will have port suffix, e.g., cli-mkdocs-9001
	containerNameBase = "cli-mkdocs"
	imageName         = "cli-mkdocs:latest"
	dockerfile        = "containers/mkdocs/.Dockerfile"

	// containerInternalPort is the port MkDocs listens on inside the container
	containerInternalPort = 8000
)

// getRepoRoot returns the repository root directory
func getRepoRoot(logger *logging.Logger) (string, error) {
	logger.Debug("Getting repository root")
	root, err := repository.GetRepositoryRoot("")
	if err != nil {
		logger.Error("Failed to get repository root", zap.Error(err))
		return "", err
	}
	logger.Debug("Repository root found", zap.String("root", root))
	return root, nil
}

// isContainerRunning checks if any MkDocs container is running
func isContainerRunning(cli *client.Client, ctx context.Context, logger *logging.Logger) (bool, *ContainerInfo, error) {
	logger.Debug("Checking if container is running", zap.String("containerName", containerNameBase))

	result, running, err := serve.IsServing(ctx, containerNameBase)
	if err != nil {
		logger.Error("Failed to check container status", zap.Error(err))
		return false, nil, err
	}

	if !running || result == nil {
		logger.Debug("Container is not running")
		return false, nil, nil
	}

	logger.Debug("Container is running",
		zap.String("containerName", result.ContainerName),
		zap.String("url", result.URL),
		zap.Int("port", result.HostPort))

	return true, &ContainerInfo{
		Name: result.ContainerName,
		URL:  result.URL,
		Port: result.HostPort,
	}, nil
}

// startMkDocsContainer starts the MkDocs container
func startMkDocsContainer(cli *client.Client, ctx context.Context, port int, logger *logging.Logger) (*ContainerInfo, error) {
	logger.Debug("Starting MkDocs container", zap.Int("requestedPort", port))

	// Check if container already running
	running, info, err := isContainerRunning(cli, ctx, logger)
	if err != nil {
		return nil, err
	}
	if running {
		logger.Warn("Container is already running", zap.String("url", info.URL))
		return info, fmt.Errorf("container is already running")
	}

	// Get repo root
	repoRoot, err := getRepoRoot(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to determine repository root: %w", err)
	}

	// Build configuration for the serve helper
	dockerfilePath := filepath.Join(repoRoot, "containers", "mkdocs", ".Dockerfile")
	contextPath := filepath.Join(repoRoot, "containers", "mkdocs")

	logger.Debug("Container configuration",
		zap.String("dockerfile", dockerfilePath),
		zap.String("contextPath", contextPath),
		zap.String("contentPath", repoRoot),
		zap.Int("containerPort", containerInternalPort))

	config := &serve.ServeConfig{
		Name:  containerNameBase,
		Image: imageName,
		BuildInfo: &serve.BuildInfo{
			Dockerfile:  dockerfilePath,
			ContextPath: contextPath,
		},
		ContentPath:   repoRoot,
		ContainerPath: "/docs",
		ContainerPort: containerInternalPort,
		Command:       []string{"mkdocs", "serve", "--dev-addr=0.0.0.0:8000"},
		PreferredPort: port,
		RestartPolicy: "unless-stopped",
	}

	// Start the serve container
	logger.Info("Launching container via serve helper", zap.String("image", imageName))
	result, err := serve.StartServe(ctx, config)
	if err != nil {
		logger.Error("Failed to start container", zap.Error(err))
		return nil, err
	}

	logger.Info("Container started successfully",
		zap.String("containerName", result.ContainerName),
		zap.String("url", result.URL),
		zap.Int("hostPort", result.HostPort))

	return &ContainerInfo{
		Name: result.ContainerName,
		URL:  result.URL,
		Port: result.HostPort,
	}, nil
}

// stopMkDocsContainer stops the MkDocs container
func stopMkDocsContainer(cli *client.Client, ctx context.Context, logger *logging.Logger) error {
	logger.Debug("Stopping MkDocs container", zap.String("containerName", containerNameBase))

	err := serve.StopServe(ctx, containerNameBase)
	if err != nil {
		logger.Error("Failed to stop container", zap.Error(err))
		return err
	}

	logger.Info("Container stopped successfully")
	return nil
}

// streamContainerLogs streams container logs to stdout
func streamContainerLogs(cli *client.Client, ctx context.Context, logger *logging.Logger) error {
	logger.Debug("Searching for container to stream logs")

	// Find the container
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		logger.Error("Failed to list containers", zap.Error(err))
		return fmt.Errorf("failed to list containers: %w", err)
	}

	logger.Debug("Found containers", zap.Int("count", len(containers)))

	var containerID string
	for _, c := range containers {
		for _, name := range c.Names {
			cleanName := strings.TrimPrefix(name, "/")
			if cleanName == containerNameBase || strings.HasPrefix(cleanName, containerNameBase+"-") {
				containerID = c.ID
				logger.Debug("Found matching container",
					zap.String("containerID", containerID),
					zap.String("name", cleanName))
				break
			}
		}
		if containerID != "" {
			break
		}
	}

	if containerID == "" {
		logger.Error("Container not found for log streaming")
		return fmt.Errorf("container not found")
	}

	// Stream logs
	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	}

	logger.Debug("Starting log stream", zap.String("containerID", containerID))
	logs, err := cli.ContainerLogs(ctx, containerID, logOptions)
	if err != nil {
		logger.Error("Failed to get container logs", zap.Error(err))
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	logger.Debug("Log stream established, copying to stdout/stderr")

	// Copy logs to stdout and stderr
	// Docker multiplexes stdout and stderr, so we need to demultiplex it
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logs)
	if err != nil {
		logger.Error("Error reading logs", zap.Error(err))
		return fmt.Errorf("error reading logs: %w", err)
	}

	logger.Debug("Log streaming completed")
	return nil
}
