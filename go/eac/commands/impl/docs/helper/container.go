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
	"github.com/ready-to-release/eac/go/eac/commands/impl/build/books"
	"github.com/ready-to-release/eac/go/eac/commands/internal/serve"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"go.uber.org/zap"
)

const (
	// defaultContainerNameBase is the fallback base name for mkdocs containers
	defaultContainerNameBase = "cli-mkdocs-site"
	// defaultImageName is the fallback Docker image name
	defaultImageName = "cli-mkdocs-site:latest"
	// defaultDockerfile is the fallback Dockerfile path
	defaultDockerfile = "containers/mkdocs-site/Dockerfile"

	// containerInternalPort is the port MkDocs listens on inside the container
	containerInternalPort = 8000
)

// getDockerImageConfig returns the Docker image configuration for mkdocs-site type.
// Falls back to defaults if config is not available.
func getDockerImageConfig() (containerNameBase, imageName, dockerfile string) {
	containerNameBase = defaultContainerNameBase
	imageName = defaultImageName
	dockerfile = defaultDockerfile

	cfg := config.Global()
	if cfg != nil && cfg.ModuleTypes != nil {
		if img := cfg.ModuleTypes.GetDockerImageName("mkdocs-site"); img != "" {
			imageName = img
			// Derive container name from image (e.g., "cli-mkdocs-site:latest" -> "cli-mkdocs-site")
			if idx := strings.Index(img, ":"); idx > 0 {
				containerNameBase = img[:idx]
			}
		}
		if dir := cfg.ModuleTypes.GetDockerContainerDir("mkdocs-site"); dir != "" {
			dockerfile = filepath.Join(dir, "Dockerfile")
		}
	}
	return
}

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
	containerNameBase, _, _ := getDockerImageConfig()
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

	// Generate mkdocs.yml from site template
	configDir := paths.ServeOutputPath(repoRoot)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Error("Failed to create serve config directory", zap.Error(err))
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := paths.MkDocsConfigPath(configDir)
	// docs_dir is relative to config file location (out/serve/mkdocs.yml)
	// So we need ../../docs to get back to repo root's docs/ directory
	configOpts := books.ConfigOptions{
		SiteName:     "Documentation",
		DocsDir:      "../../docs",
		OutputFormat: "site",
	}
	if err := books.WriteMkDocsConfig(repoRoot, configPath, configOpts); err != nil {
		logger.Error("Failed to generate mkdocs.yml", zap.Error(err))
		return nil, fmt.Errorf("failed to generate mkdocs.yml: %w", err)
	}
	logger.Debug("Generated mkdocs.yml from template", zap.String("configPath", configPath))

	// Copy mkdocs macros script to serve directory as main.py
	// mkdocs-macros will automatically find main.py in the same directory as mkdocs.yml
	macrosSource := filepath.Join(repoRoot, "containers", "mkdocs-site", "mkdocs_macros.py")
	macrosTarget := filepath.Join(configDir, "main.py")
	macrosData, err := os.ReadFile(macrosSource)
	if err == nil {
		if err := os.WriteFile(macrosTarget, macrosData, 0644); err != nil {
			logger.Warn("Failed to copy mkdocs macros script", zap.Error(err))
		} else {
			logger.Debug("Copied mkdocs macros to main.py", zap.String("target", macrosTarget))
		}
	} else {
		logger.Debug("Mkdocs macros script not found (optional)", zap.String("path", macrosSource))
	}

	// Calculate relative config path for Docker
	relConfigPath, _ := filepath.Rel(repoRoot, configPath)
	dockerConfigPath := strings.ReplaceAll(relConfigPath, "\\", "/")

	// Get Docker configuration from module types
	containerNameBase, imageName, dockerfile := getDockerImageConfig()

	// Build configuration for the serve helper
	dockerfilePath := filepath.Join(repoRoot, dockerfile)
	contextPath := filepath.Dir(dockerfilePath)

	logger.Debug("Container configuration",
		zap.String("dockerfile", dockerfilePath),
		zap.String("contextPath", contextPath),
		zap.String("contentPath", repoRoot),
		zap.String("configPath", dockerConfigPath),
		zap.Int("containerPort", containerInternalPort))

	serveConfig := &serve.ServeConfig{
		Name:  containerNameBase,
		Image: imageName,
		BuildInfo: &serve.BuildInfo{
			Dockerfile:  dockerfilePath,
			ContextPath: contextPath,
		},
		ContentPath:   repoRoot,
		ContainerPath: "/docs",
		ContainerPort: containerInternalPort,
		Command:       []string{"mkdocs", "serve", "-f", dockerConfigPath, "--dev-addr=0.0.0.0:8000"},
		PreferredPort: port,
		RestartPolicy: "unless-stopped",
	}

	// Start the serve container
	logger.Info("Launching container via serve helper", zap.String("image", imageName))
	result, err := serve.StartServe(ctx, serveConfig)
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
	containerNameBase, _, _ := getDockerImageConfig()
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
	containerNameBase, _, _ := getDockerImageConfig()
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
