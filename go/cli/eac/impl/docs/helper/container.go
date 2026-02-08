package docs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/builders/mkdocs"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

const (
	// defaultContainerNameBase is the fallback base name for mkdocs containers.
	defaultContainerNameBase = "mkdocs-render-oci"
	// defaultImageName is the fallback Docker image name (uses :local tag for local builds).
	defaultImageName = "mkdocs-render-oci:local"
	// defaultDockerfile is the fallback Dockerfile path.
	defaultDockerfile = "containers/mkdocs-render-oci/Dockerfile"

	// containerInternalPort is the port MkDocs listens on inside the container.
	containerInternalPort = 8000
)

// getDockerImageConfig returns the Docker image configuration for mkdocs-render-oci type.
// Uses hardcoded defaults since module types no longer define docker image config.
func getDockerImageConfig() (containerNameBase, imageName, dockerfile string) {
	containerNameBase = defaultContainerNameBase
	imageName = defaultImageName
	dockerfile = defaultDockerfile
	return containerNameBase, imageName, dockerfile
}

// getRepoRoot returns the repository root directory.
func getRepoRoot() (string, error) {
	log.Debugf("Getting repository root")
	root, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Failed to get repository root: error=%v", err)
		return "", err
	}
	log.Debugf("Repository root found: root=%s", root)
	return root, nil
}

// isContainerRunning checks if any MkDocs container is running.
func isContainerRunning(cli docker.DockerClient, ctx context.Context) (bool, *ContainerInfo, error) {
	containerNameBase, _, _ := getDockerImageConfig()
	log.Debugf("Checking if container is running: containerName=%s", containerNameBase)

	result, running, err := docker.IsServing(ctx, containerNameBase)
	if err != nil {
		log.Errorf("Failed to check container status: error=%v", err)
		return false, nil, err
	}

	if !running || result == nil {
		log.Debugf("Container is not running")
		return false, nil, nil
	}

	log.Debugf("Container is running: containerName=%s, url=%s, port=%d", result.ContainerName, result.URL, result.HostPort)

	return true, &ContainerInfo{
		Name: result.ContainerName,
		URL:  result.URL,
		Port: result.HostPort,
	}, nil
}

// startMkDocsContainer starts the MkDocs container.
func startMkDocsContainer(cli docker.DockerClient, ctx context.Context, port int) (*ContainerInfo, error) {
	log.Debugf("Starting MkDocs container: requestedPort=%d", port)

	// Check if container already running
	running, info, err := isContainerRunning(cli, ctx)
	if err != nil {
		return nil, err
	}
	if running {
		log.Warnf("Container is already running: url=%s", info.URL)
		return info, fmt.Errorf("container is already running")
	}

	// Get repo root
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to determine repository root: %w", err)
	}

	// Generate mkdocs.yml from site template
	configDir := paths.ServeOutputPath(repoRoot)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		log.Errorf("Failed to create serve config directory: error=%v", err)
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := paths.MkDocsConfigPath(configDir)
	// docs_dir is relative to config file location (out/serve/mkdocs.yml)
	// So we need ../../docs to get back to repo root's docs/ directory
	configOpts := mkdocs.ConfigOptions{
		SiteName:     "Documentation",
		DocsDir:      "../../docs",
		OutputFormat: "site",
	}
	if err := mkdocs.WriteMkDocsConfig(repoRoot, configPath, configOpts); err != nil {
		log.Errorf("Failed to generate mkdocs.yml: error=%v", err)
		return nil, fmt.Errorf("failed to generate mkdocs.yml: %w", err)
	}
	log.Debugf("Generated mkdocs.yml from template: configPath=%s", configPath)

	// Copy mkdocs macros script to serve directory as main.py
	// mkdocs-macros will automatically find main.py in the same directory as mkdocs.yml
	macrosSource := filepath.Join(repoRoot, "containers", "mkdocs-render-oci", "mkdocs_macros.py")
	macrosTarget := filepath.Join(configDir, "main.py")
	macrosData, err := os.ReadFile(macrosSource)
	if err == nil {
		if err := os.WriteFile(macrosTarget, macrosData, 0o644); err != nil {
			log.Warnf("Failed to copy mkdocs macros script: error=%v", err)
		} else {
			log.Debugf("Copied mkdocs macros to main.py: target=%s", macrosTarget)
		}
	} else {
		log.Debugf("Mkdocs macros script not found (optional): path=%s", macrosSource)
	}

	// Calculate relative config path for Docker
	relConfigPath, relErr := filepath.Rel(repoRoot, configPath)
	if relErr != nil {
		relConfigPath = configPath // Fallback to absolute path
	}
	dockerConfigPath := strings.ReplaceAll(relConfigPath, "\\", "/")

	// Get Docker configuration from module types
	containerNameBase, imageName, dockerfile := getDockerImageConfig()

	// Build configuration for the serve helper
	dockerfilePath := filepath.Join(repoRoot, dockerfile)
	contextPath := filepath.Dir(dockerfilePath)

	log.Debugf("Container configuration: dockerfile=%s, contextPath=%s, contentPath=%s, configPath=%s, containerPort=%d", dockerfilePath, contextPath, repoRoot, dockerConfigPath, containerInternalPort)

	serveConfig := &docker.ServeConfig{
		Name:  containerNameBase,
		Image: imageName,
		BuildInfo: &docker.BuildInfo{
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
	log.Infof("Launching container via serve helper: image=%s", imageName)
	result, err := docker.StartServe(ctx, serveConfig)
	if err != nil {
		log.Errorf("Failed to start container: error=%v", err)
		return nil, err
	}

	log.Infof("Container started successfully: containerName=%s, url=%s, hostPort=%d", result.ContainerName, result.URL, result.HostPort)

	return &ContainerInfo{
		Name: result.ContainerName,
		URL:  result.URL,
		Port: result.HostPort,
	}, nil
}

// stopMkDocsContainer stops the MkDocs container.
func stopMkDocsContainer(cli docker.DockerClient, ctx context.Context) error {
	containerNameBase, _, _ := getDockerImageConfig()
	log.Debugf("Stopping MkDocs container: containerName=%s", containerNameBase)

	err := docker.StopServe(ctx, containerNameBase)
	if err != nil {
		log.Errorf("Failed to stop container: error=%v", err)
		return err
	}

	log.Info("Container stopped successfully")
	return nil
}

// streamContainerLogs streams container logs to stdout.
func streamContainerLogs(cli docker.DockerClient, ctx context.Context) error {
	containerNameBase, _, _ := getDockerImageConfig()
	log.Debugf("Streaming container logs: containerName=%s", containerNameBase)

	// Use the adapters/docker log streaming function
	err := docker.StreamContainerLogs(ctx, containerNameBase)
	if err != nil {
		log.Errorf("Failed to stream container logs: error=%v", err)
		return err
	}

	log.Debugf("Log streaming completed")
	return nil
}
