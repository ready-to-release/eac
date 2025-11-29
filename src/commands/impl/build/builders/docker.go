// docker.go - Build functions for Docker module types
package builders

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// BuildDockerDefault is the default build handler for Docker modules.
func BuildDockerDefault(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== Building Docker module: %s (type: %s) ===", module.Moniker, module.Type)
	Logln(logWriter, "ℹ️  Docker modules are built via docker build command")
	return 0
}

// BuildR2RExtension builds an R2R CLI extension as a Docker image
func BuildR2RExtension(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== Building R2R extension: %s ===", module.Moniker)

	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	// Step 1: go mod tidy (if enabled)
	if opts.TidyFirst {
		goModPath := filepath.Join(moduleRoot, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			Logln(logWriter, "Running: go mod tidy (in %s)", module.Files.Root)
			if exitCode := RunCommandWithLog(moduleRoot, logWriter, "go", "mod", "tidy"); exitCode != 0 {
				Logln(logWriter, "❌ go mod tidy failed")
				return exitCode
			}
		}
	}

	// Extract extension name from moniker
	extensionName := module.Moniker
	if len(module.Moniker) > 4 && module.Moniker[:4] == "ext-" {
		extensionName = module.Moniker[4:]
	}

	// Dockerfile is in containers/{moniker}/Dockerfile
	dockerfilePath := filepath.Join(workspaceRoot, "containers", module.Moniker, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		Logln(logWriter, "❌ No Dockerfile found at: %s", dockerfilePath)
		return 1
	}

	imageName := fmt.Sprintf("ext-%s:latest", extensionName)

	Logln(logWriter, "📦 Building Docker image: %s", imageName)
	Logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	Logln(logWriter, "   Build context: %s", workspaceRoot)

	isCI := os.Getenv("CI") == "true"

	if isCI {
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
			Logln(logWriter, "\n❌ Docker build failed (see errors above)")
			return exitCode
		}
		Logln(logWriter, "✅ Single-platform image built successfully: %s", imageName)

		// Export multi-platform for release
		Logln(logWriter, "\n--- CI Mode: Building multi-platform for release ---")
		ociArchivePath := filepath.Join(outputDir, fmt.Sprintf("ext-%s-ci-test.tar", extensionName))

		exitCode = RunCommandWithLog(workspaceRoot, logWriter,
			"docker", "buildx", "build",
			"--platform", "linux/amd64,linux/arm64",
			"-t", imageName,
			"-f", dockerfilePath,
			"--cache-from", "type=gha",
			"-o", fmt.Sprintf("type=oci,dest=%s", ociArchivePath),
			".")

		if exitCode != 0 {
			Logln(logWriter, "\n❌ Multi-platform build failed (see errors above)")
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
	} else {
		// Local build
		exitCode := RunCommandWithLog(workspaceRoot, logWriter,
			"docker", "build",
			"-t", imageName,
			"-f", dockerfilePath,
			".")

		if exitCode != 0 {
			Logln(logWriter, "\n❌ Docker build failed (see errors above)")
			return exitCode
		}

		Logln(logWriter, "✅ Docker image built successfully: %s", imageName)

		imageInfoPath := filepath.Join(outputDir, "image-info.txt")
		imageInfo := fmt.Sprintf("Image: %s\nDockerfile: %s\nBuild Date: %s\n",
			imageName, dockerfilePath, time.Now().Format(time.RFC3339))

		if err := os.WriteFile(imageInfoPath, []byte(imageInfo), 0644); err != nil {
			Logln(logWriter, "⚠️  Warning: could not save image info: %v", err)
		}
	}

	return 0
}
