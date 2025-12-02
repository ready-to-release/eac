// docker.go - Build handler for Docker build system
package builders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

func init() {
	// Register handler for "docker" build system
	RegisterSystem("docker", BuildDockerModule)
}

// BuildDockerModule builds a Docker container image.
// Behavior is determined by capabilities:
// - go_module → run go mod tidy first
// - container → build Docker image
func BuildDockerModule(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== Building %s: %s ===", module.Type, module.Moniker)

	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	// Get capabilities from contract
	cfg := config.Global()
	hasGoModule := cfg != nil && cfg.ModuleTypes != nil && cfg.ModuleTypes.HasCapability(module.Type, "go_module")

	// Step 1: go mod tidy (if Go module and enabled)
	if hasGoModule && opts.TidyFirst {
		goModPath := filepath.Join(moduleRoot, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			Logln(logWriter, "Running: go mod tidy")
			if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
				Logln(logWriter, "❌ go mod tidy failed")
				return exitCode
			}
		}
	}

	// Find Dockerfile using configured search paths
	var dockerfilePath string
	searchPaths := []string{"containers/{moniker}/Dockerfile", "{root}/Dockerfile"}
	if cfg != nil && cfg.Handlers != nil {
		searchPaths = cfg.Handlers.GetDockerfilePaths()
	}

	for _, pathTemplate := range searchPaths {
		resolvedPath := config.ResolveDockerfilePath(pathTemplate, module.Moniker, module.Files.Root)
		fullPath := filepath.Join(workspaceRoot, resolvedPath)
		if _, err := os.Stat(fullPath); err == nil {
			dockerfilePath = fullPath
			break
		}
	}

	if dockerfilePath == "" {
		Logln(logWriter, "❌ No Dockerfile found")
		return 1
	}

	imageName := fmt.Sprintf("%s:latest", module.Moniker)

	Logln(logWriter, "📦 Building Docker image: %s", imageName)
	Logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	Logln(logWriter, "   Build context: %s", workspaceRoot)

	isCI := os.Getenv("CI") == "true"

	if isCI {
		return buildDockerCI(module, workspaceRoot, outputDir, dockerfilePath, imageName, logWriter)
	}
	return buildDockerLocal(workspaceRoot, outputDir, dockerfilePath, imageName, logWriter)
}

// buildDockerLocal builds a Docker image locally using BuildKit for cache support
func buildDockerLocal(workspaceRoot string, outputDir string, dockerfilePath string, imageName string, logWriter io.Writer) int {
	// Use buildx to enable BuildKit cache mounts (RUN --mount=type=cache)
	// This significantly speeds up subsequent builds by caching Go modules and build artifacts
	exitCode := RunCommandWithLog(workspaceRoot, logWriter,
		"docker", "buildx", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		"--load",
		".")

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Docker build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Docker image built successfully: %s", imageName)

	// Save image info
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\n",
		imageName, dockerfilePath, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
		Logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
	}

	return 0
}

// buildDockerCI builds a Docker image in CI with multi-platform support
func buildDockerCI(module *modules.ModuleContract, workspaceRoot string, outputDir string, dockerfilePath string, imageName string, logWriter io.Writer) int {
	// Get CI platforms from handlers config
	cfg := config.Global()
	ciPlatforms := "linux/amd64,linux/arm64"
	if cfg != nil && cfg.Handlers != nil {
		ciPlatforms = cfg.Handlers.GetCIPlatformsString()
	}

	Logln(logWriter, "\n--- CI Mode: Building single-platform for testing ---")
	exitCode := RunCommandWithLog(workspaceRoot, logWriter,
		"docker", "buildx", "build",
		"--platform", "linux/amd64",
		"-t", imageName,
		"-f", dockerfilePath,
		"--cache-from", "type=gha",
		"--cache-to", "type=gha,mode=max",
		"--load",
		".")

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Docker build failed")
		return exitCode
	}
	Logln(logWriter, "✅ Single-platform image built successfully: %s", imageName)

	// Export multi-platform for release
	Logln(logWriter, "\n--- CI Mode: Building multi-platform for release ---")
	ociArchivePath := filepath.Join(outputDir, fmt.Sprintf("%s-ci-test.tar", module.Moniker))

	exitCode = RunCommandWithLog(workspaceRoot, logWriter,
		"docker", "buildx", "build",
		"--platform", ciPlatforms,
		"-t", imageName,
		"-f", dockerfilePath,
		"--cache-from", "type=gha",
		"-o", fmt.Sprintf("type=oci,dest=%s", ociArchivePath),
		".")

	if exitCode != 0 {
		Logln(logWriter, "\n❌ Multi-platform build failed")
		return exitCode
	}

	Logln(logWriter, "✅ Multi-platform image exported: %s", ociArchivePath)

	// Compress the OCI archive
	Logln(logWriter, "Compressing OCI archive...")
	exitCode = RunCommandWithLog(outputDir, logWriter, "gzip", filepath.Base(ociArchivePath))
	if exitCode != 0 {
		Logln(logWriter, "⚠️  Warning: failed to compress archive")
	}

	// Save image info
	imageInfoPath := filepath.Join(outputDir, "image-info.txt")
	imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\nPlatforms: linux/amd64,linux/arm64\nOCI Archive: %s.gz\n",
		imageName, dockerfilePath, time.Now().Format(time.RFC3339), ociArchivePath)

	if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
		Logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
	}

	return 0
}
