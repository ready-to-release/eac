// buildx.go - Build handler for buildx capability (uses docker_build config from module or type)
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"gopkg.in/yaml.v3"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&BuildxHandler{})
}

// BuildxHandler builds Docker images using docker_build config from module or type.
// This handler uses docker buildx for multi-platform builds, registry cache, SBOM, and provenance.
type BuildxHandler struct{}

func (h *BuildxHandler) Name() string { return "buildx" }

func (h *BuildxHandler) Capabilities() []string { return []string{"buildx"} }

func (h *BuildxHandler) Requirements() []string { return []string{"docker"} }

func (h *BuildxHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available")
	}
	return nil
}

// IsContainer returns false as buildx runs docker commands on the host.
func (h *BuildxHandler) IsContainer() bool { return false }

// IsHostInstalled returns true as buildx uses the local docker CLI.
func (h *BuildxHandler) IsHostInstalled() bool { return true }

func (h *BuildxHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{fmt.Sprintf("docker-image:%s", module.GetMoniker())}
}

func (h *BuildxHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	concrete := adapters.UnwrapModule(module)
	if concrete == nil {
		Logln(logWriter, "Error: invalid module type")
		return 1
	}
	moniker := module.GetMoniker()
	Logln(logWriter, "\n=== Building Docker Image: %s ===", moniker)

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
	dockerBuild := getDockerBuildConfig(concrete, opts.Component, logWriter)
	if dockerBuild == nil {
		Logln(logWriter, "❌ No docker_build configuration found for module %s", moniker)
		return 1
	}

	// Detect CI environment and adjust build settings for local builds
	isCI := os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"
	if !isCI {
		Logln(logWriter, "🏠 Local build detected - adjusting build settings")

		// Auto-detect current platform for local builds
		localPlatform := detectLocalPlatform()
		Logln(logWriter, "   Platform: %s (auto-detected)", localPlatform)

		// Use :local tag plus content hash - matches tool system expectations (LocalImageTag())
		// Content hash tag changes only when source files change
		localTags := []string{
			fmt.Sprintf("%s:local", dockerBuild.Container),
		}
		// Add content-hash tag for cache validation
		if hash, err := module.GetContentHash(); err == nil && hash != "" {
			localTags = append(localTags, fmt.Sprintf("%s:%s", dockerBuild.Container, hash))
		}

		// For local builds: use simple tags and local-friendly settings
		dockerBuild = &config.DockerBuildConfig{
			Container:  dockerBuild.Container,
			Context:    dockerBuild.Context,     // Keep original context from config
			Dockerfile: dockerBuild.Dockerfile,  // Keep original Dockerfile path
			Platforms:  []string{localPlatform}, // Auto-detect platform
			Tags:       localTags,               // Use local-friendly tags
			Load:       true,                    // Load into local Docker instead of push
			Push:       config.BoolPtr(false),    // Don't push to registry
			Cache:      nil,                     // Skip registry cache for local builds
			SBOM:       false,                   // Skip SBOM for local builds
			Provenance: false,                   // Skip provenance for local builds
			Registry:   "",
		}
	}

	// Expand template variables in docker_build config
	container := expandTemplate(dockerBuild.Container, moniker)
	buildContext := filepath.Join(workspaceRoot, expandTemplate(dockerBuild.Context, moniker))
	dockerfilePath := expandTemplate(dockerBuild.Dockerfile, moniker)
	if dockerfilePath == "" {
		dockerfilePath = filepath.Join(buildContext, "Dockerfile")
	} else {
		dockerfilePath = filepath.Join(workspaceRoot, dockerfilePath)
	}

	// Expand tags
	var tags []string
	for _, tagTemplate := range dockerBuild.Tags {
		tag := expandTemplate(tagTemplate, moniker)
		tags = append(tags, tag)
	}

	Logln(logWriter, "📦 Building Docker image")
	Logln(logWriter, "   Container: %s", container)
	Logln(logWriter, "   Context: %s", buildContext)
	Logln(logWriter, "   Dockerfile: %s", dockerfilePath)
	Logln(logWriter, "   Tags: %v", tags)

	// Authenticate to registry if push is enabled
	if dockerBuild.ShouldPush() && dockerBuild.Registry != "" {
		if exitCode := authenticateRegistry(dockerBuild.Registry, logWriter); exitCode != 0 {
			return exitCode
		}
	}

	// Build the image
	args := buildDockerBuildArgs(dockerBuild, buildContext, dockerfilePath, tags, moniker, opts, logWriter)

	Logln(logWriter, "\nExecuting: docker %s", strings.Join(args, " "))

	if exitCode := RunCommandWithLog(context.Background(), workspaceRoot, logWriter, "docker", args...); exitCode != 0 {
		Logln(logWriter, "\n❌ Docker build failed")
		return exitCode
	}

	Logln(logWriter, "\n✅ Docker image built successfully")
	Logln(logWriter, "   Tags: %v", tags)
	if dockerBuild.ShouldPush() {
		Logln(logWriter, "   Pushed to: %s", dockerBuild.Registry)
	}
	if dockerBuild.Load {
		Logln(logWriter, "   Loaded locally")
	}

	return 0
}

// buildDockerBuildArgs constructs the docker buildx build command arguments.
func buildDockerBuildArgs(dockerBuild *config.DockerBuildConfig, buildContext, dockerfilePath string, tags []string, moniker string, opts BuildOptions, logWriter io.Writer) []string {
	args := []string{"buildx", "build"}

	// Use specified builder or default to "default" (docker driver).
	// The docker driver builds directly in Docker daemon - no slow tarball export.
	// Use a docker-container driver only for multi-platform builds with QEMU.
	builder := dockerBuild.Builder
	if builder == "" {
		builder = "default"
	}
	args = append(args, "--builder", builder)

	// Apply cache flags from CacheConfig
	if opts.CacheConfig != nil {
		// --skip-cache=local:layer -> --no-cache (bypass BuildKit layer cache)
		if opts.CacheConfig.ShouldForceNoCacheDocker() {
			args = append(args, "--no-cache")
			Logln(logWriter, "   Cache: --no-cache (local:layer skipped)")
		}
		// --skip-cache=local:registry -> --pull (force fresh base images)
		if opts.CacheConfig.ShouldForcePull() {
			args = append(args, "--pull")
			Logln(logWriter, "   Cache: --pull (local:registry skipped)")
		}
	}

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
	if dockerBuild.ShouldPush() {
		args = append(args, "--push")
	}

	// Cache configuration
	// --skip-cache=remote:layer disables registry cache
	skipRemoteLayer := opts.CacheConfig != nil && opts.CacheConfig.ShouldSkipRemoteLayer()
	if dockerBuild.Cache != nil && !skipRemoteLayer {
		cacheArgs := buildCacheArgs(dockerBuild.Cache, moniker)
		args = append(args, cacheArgs...)
	} else if skipRemoteLayer && dockerBuild.Cache != nil {
		Logln(logWriter, "   Cache: registry cache disabled (remote:layer skipped)")
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
	args = append(args, buildContext)

	return args
}

// buildCacheArgs constructs cache arguments for docker buildx.
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

// authenticateRegistry authenticates to a Docker registry.
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

	// TODO: migrate to tool executor once stdin support is added to ExecutionContext
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
	// Run 'docker version' via tool executor to get the platform Docker is configured for
	toolDef := tool.GlobalRegistry().GetOrAdhoc("docker")
	execCtx := &tool.ExecutionContext{
		ArgsOverrides: []string{"version", "--format", "{{.Server.Os}}/{{.Server.Arch}}"},
	}
	result, err := tool.GlobalExecutor().Execute(context.Background(), toolDef, execCtx)
	if err == nil && result.ExitCode == 0 {
		platform := strings.TrimSpace(string(result.Stdout))
		if platform != "" && platform != "/" {
			return platform
		}
	}

	// Fallback: use linux/amd64 as default since most Docker Desktop instances run Linux containers
	return "linux/amd64"
}

// expandTemplate expands template variables in a string
// Supported variables: {moniker}, {container}, {short_sha}.
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

// getDockerBuildConfig gets docker_build config from a named component, falling back to "dockerfile".
func getDockerBuildConfig(module *modules.ModuleContract, componentName string, logWriter io.Writer) *config.DockerBuildConfig {
	// Try the specific component name first (for multi-dockerfile modules like oci-tools)
	if componentName != "" {
		if pkg := module.Components[componentName]; pkg != nil {
			if pkg.DockerBuild != nil && len(pkg.DockerBuild) > 0 {
				dockerCfg, err := convertDockerBuildConfig(pkg.DockerBuild)
				if err != nil {
					Logln(logWriter, "⚠️  Failed to parse docker_build config from component %s: %v", componentName, err)
				} else {
					return dockerCfg
				}
			}
		}
	}

	// Fall back to legacy "dockerfile" key (backward compatibility for single-container modules)
	dockerfilePackage := module.Components["dockerfile"]
	if dockerfilePackage == nil {
		return nil
	}
	if dockerfilePackage.DockerBuild == nil || len(dockerfilePackage.DockerBuild) == 0 {
		return nil
	}

	dockerCfg, err := convertDockerBuildConfig(dockerfilePackage.DockerBuild)
	if err != nil {
		Logln(logWriter, "⚠️  Failed to parse dockerfile package docker_build config: %v", err)
		return nil
	}
	return dockerCfg
}

// convertDockerBuildConfig converts map[string]interface{} to DockerBuildConfig via YAML.
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
