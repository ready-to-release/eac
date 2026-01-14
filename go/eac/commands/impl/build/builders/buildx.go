// buildx.go - Build handler for buildx capability (uses docker_build config from module or type)
package builders

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"gopkg.in/yaml.v3"
)

func init() {
	RegisterHandler(&BuildxHandler{})
}

// BuildxHandler builds Docker images using docker_build config from module or type.
// This handler uses docker buildx for multi-platform builds, registry cache, SBOM, and provenance.
type BuildxHandler struct{}

func (h *BuildxHandler) Name() string { return "buildx" }

func (h *BuildxHandler) Capabilities() []string { return []string{"buildx"} }

func (h *BuildxHandler) Requirements() []string { return []string{"docker"} }

func (h *BuildxHandler) ValidateModule(module *modules.ModuleContract, workspaceRoot string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available")
	}
	return nil
}

func (h *BuildxHandler) ListArtifacts(module *modules.ModuleContract, workspaceRoot string) []string {
	return []string{fmt.Sprintf("docker-image:%s", module.Moniker)}
}

func (h *BuildxHandler) Build(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== Building Docker Image: %s ===", module.Moniker)

	// Check if Docker is available
	if !IsDockerAvailable() {
		Logln(logWriter, "❌ Docker is not available")
		if IsDockerInDocker() {
			Logln(logWriter, "   Container detected but Docker socket not mounted")
			Logln(logWriter, "   Ensure the Docker socket is mounted (-v /var/run/docker.sock:/var/run/docker.sock)")
		} else {
			Logln(logWriter, "   Ensure Docker is installed and the daemon is running")
		}
		return 1
	}

	// Get docker_build config from module first, then fall back to module type
	dockerBuild := getDockerBuildConfig(module, logWriter)
	if dockerBuild == nil {
		Logln(logWriter, "❌ No docker_build configuration found for module %s", module.Moniker)
		return 1
	}

	// Detect CI environment and adjust build settings for local builds
	isCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"
	if !isCI {
		Logln(logWriter, "🏠 Local build detected - adjusting build settings")

		// Auto-detect current platform for local builds
		localPlatform := detectLocalPlatform()
		Logln(logWriter, "   Platform: %s (auto-detected)", localPlatform)

		// Use simple local tags instead of CI tags
		localTags := []string{
			fmt.Sprintf("%s:local", dockerBuild.Container),
			fmt.Sprintf("%s:dev", dockerBuild.Container),
		}

		// For local builds: use simple tags and local-friendly settings
		dockerBuild = &config.DockerBuildConfig{
			Container:  dockerBuild.Container,
			Context:    dockerBuild.Context,     // Keep original context from config
			Dockerfile: dockerBuild.Dockerfile,  // Keep original Dockerfile path
			Platforms:  []string{localPlatform}, // Auto-detect platform
			Tags:       localTags,               // Use local-friendly tags
			Load:       true,                    // Load into local Docker instead of push
			Push:       false,                   // Don't push to registry
			Cache:      nil,                     // Skip registry cache for local builds
			SBOM:       false,                   // Skip SBOM for local builds
			Provenance: false,                   // Skip provenance for local builds
			Registry:   "",
		}
	}

	// Expand template variables in docker_build config
	container := expandTemplate(dockerBuild.Container, module.Moniker)
	context := filepath.Join(workspaceRoot, expandTemplate(dockerBuild.Context, module.Moniker))
	dockerfilePath := expandTemplate(dockerBuild.Dockerfile, module.Moniker)
	if dockerfilePath == "" {
		dockerfilePath = filepath.Join(context, "Dockerfile")
	} else {
		dockerfilePath = filepath.Join(workspaceRoot, dockerfilePath)
	}

	// Expand tags
	var tags []string
	for _, tagTemplate := range dockerBuild.Tags {
		tag := expandTemplate(tagTemplate, module.Moniker)
		tags = append(tags, tag)
	}

	Logln(logWriter, "📦 Building Docker image")
	Logln(logWriter, "   Container: %s", container)
	Logln(logWriter, "   Context: %s", context)
	Logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	Logln(logWriter, "   Tags: %v", tags)

	// Authenticate to registry if push is enabled
	if dockerBuild.Push && dockerBuild.Registry != "" {
		if exitCode := authenticateRegistry(dockerBuild.Registry, logWriter); exitCode != 0 {
			return exitCode
		}
	}

	// Build the image
	args := buildDockerBuildArgs(dockerBuild, context, dockerfilePath, tags, module.Moniker)

	Logln(logWriter, "\nExecuting: docker %s", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		Logln(logWriter, "\n❌ Docker build failed: %v", err)
		return 1
	}

	Logln(logWriter, "\n✅ Docker image built successfully")
	Logln(logWriter, "   Tags: %v", tags)
	if dockerBuild.Push {
		Logln(logWriter, "   Pushed to: %s", dockerBuild.Registry)
	}
	if dockerBuild.Load {
		Logln(logWriter, "   Loaded locally")
	}

	return 0
}

// buildDockerBuildArgs constructs the docker buildx build command arguments
func buildDockerBuildArgs(dockerBuild *config.DockerBuildConfig, context, dockerfilePath string, tags []string, moniker string) []string {
	args := []string{"buildx", "build"}

	// Platforms (multi-arch)
	if len(dockerBuild.Platforms) > 0 {
		args = append(args, "--platform", strings.Join(dockerBuild.Platforms, ","))
	}

	// Tags
	for _, tag := range tags {
		args = append(args, "--tag", tag)
	}

	// Dockerfile
	args = append(args, "--file", dockerfilePath)

	// Load or Push
	if dockerBuild.Load {
		args = append(args, "--load")
	}
	if dockerBuild.Push {
		args = append(args, "--push")
	}

	// Cache configuration
	if dockerBuild.Cache != nil {
		cacheArgs := buildCacheArgs(dockerBuild.Cache, moniker)
		args = append(args, cacheArgs...)
	}

	// SBOM
	if dockerBuild.SBOM {
		args = append(args, "--sbom=true")
	}

	// Provenance
	if dockerBuild.Provenance {
		args = append(args, "--provenance=true")
	}

	// Build context (must be last)
	args = append(args, context)

	return args
}

// buildCacheArgs constructs cache arguments for docker buildx
func buildCacheArgs(cache *config.DockerCacheConfig, moniker string) []string {
	var args []string

	switch cache.Type {
	case "gha":
		// GitHub Actions cache
		scope := cache.Scope
		if scope == "" {
			scope = moniker
		}
		args = append(args, "--cache-from", fmt.Sprintf("type=gha,scope=%s", scope))

		mode := cache.Mode
		if mode == "" {
			mode = "min"
		}
		args = append(args, "--cache-to", fmt.Sprintf("type=gha,mode=%s,scope=%s", mode, scope))

	case "registry":
		// Registry cache
		if cache.From != "" {
			from := expandTemplate(cache.From, moniker)
			args = append(args, "--cache-from", fmt.Sprintf("type=registry,ref=%s", from))
		}
		if cache.To != "" {
			to := expandTemplate(cache.To, moniker)
			mode := cache.Mode
			if mode == "" {
				mode = "min"
			}
			args = append(args, "--cache-to", fmt.Sprintf("type=registry,mode=%s,ref=%s", mode, to))
		}
	}

	return args
}

// authenticateRegistry authenticates to a Docker registry
func authenticateRegistry(registry string, logWriter io.Writer) int {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		Logln(logWriter, "❌ GITHUB_TOKEN environment variable not set (required for registry push)")
		return 1
	}

	actor := os.Getenv("GITHUB_ACTOR")
	if actor == "" {
		actor = "github-actions"
	}

	Logln(logWriter, "🔐 Authenticating to %s...", registry)

	cmd := exec.Command("docker", "login", registry, "-u", actor, "--password-stdin")
	cmd.Stdin = strings.NewReader(token)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Run(); err != nil {
		Logln(logWriter, "❌ Registry authentication failed: %v", err)
		return 1
	}

	return 0
}

// detectLocalPlatform detects the current platform for local Docker builds
// Returns platform in Docker format: "linux/amd64", "linux/arm64", etc.
func detectLocalPlatform() string {
	// Run 'docker version' to get the platform Docker is configured for
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Os}}/{{.Server.Arch}}")
	output, err := cmd.Output()
	if err == nil {
		platform := strings.TrimSpace(string(output))
		if platform != "" && platform != "/" {
			return platform
		}
	}

	// Fallback: use linux/amd64 as default since most Docker Desktop instances run Linux containers
	return "linux/amd64"
}

// expandTemplate expands template variables in a string
// Supported variables: {moniker}, {container}, {short_sha}
func expandTemplate(template, moniker string) string {
	result := template
	result = strings.ReplaceAll(result, "{moniker}", moniker)
	result = strings.ReplaceAll(result, "{container}", moniker) // container defaults to moniker

	// Get short SHA if available
	if sha := os.Getenv("GITHUB_SHA"); sha != "" && len(sha) >= 7 {
		result = strings.ReplaceAll(result, "{short_sha}", sha[:7])
	}

	return result
}

// getDockerBuildConfig gets docker_build config from module first, then falls back to module type
func getDockerBuildConfig(module *modules.ModuleContract, logWriter io.Writer) *config.DockerBuildConfig {
	// First, check if module has docker_build config
	if module.DockerBuild != nil && len(module.DockerBuild) > 0 {
		// Convert map[string]interface{} to DockerBuildConfig via YAML round-trip
		dockerCfg, err := convertDockerBuildConfig(module.DockerBuild)
		if err != nil {
			Logln(logWriter, "⚠️  Failed to parse module docker_build config: %v", err)
		} else {
			return dockerCfg
		}
	}

	// Fall back to module type config
	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		return nil
	}

	return cfg.ModuleTypes.GetDockerBuildConfig(module.Type)
}

// convertDockerBuildConfig converts map[string]interface{} to DockerBuildConfig via YAML
func convertDockerBuildConfig(m map[string]interface{}) (*config.DockerBuildConfig, error) {
	// Marshal to YAML, then unmarshal to typed struct
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	var dockerCfg config.DockerBuildConfig
	if err := yaml.Unmarshal(data, &dockerCfg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &dockerCfg, nil
}
