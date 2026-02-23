// buildx.go - Build handler for buildx capability (uses docker_build config from module or type)
package builders

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/tool"
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
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

func (h *BuildxHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string, logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(BuildOptions)
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
	dockerBuild := getDockerBuildConfig(concrete, opts.Component)
	if dockerBuild == nil {
		Logln(logWriter, "❌ No docker_build configuration found for module %s", moniker)
		return 1
	}

	// Adjust for local builds: single platform, local tags, no push
	isCI := environments.IsCI()
	if !isCI {
		dockerBuild = adjustForLocalBuild(dockerBuild, module, logWriter)
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

// adjustForLocalBuild creates a local-friendly docker build config.
// Local builds: single platform, local tags, no push, no registry cache, no SBOM/provenance.
func adjustForLocalBuild(dockerBuild *config.DockerBuildConfig, module core.ModuleContractPort, logWriter io.Writer) *config.DockerBuildConfig {
	localPlatform := detectLocalPlatform()
	Logln(logWriter, "🏠 Local build detected - adjusting build settings")
	Logln(logWriter, "   Platform: %s (auto-detected)", localPlatform)

	localTags := []string{
		fmt.Sprintf("%s:local", dockerBuild.Container),
	}
	if hash, err := module.GetContentHash(); err == nil && hash != "" {
		localTags = append(localTags, fmt.Sprintf("%s:%s", dockerBuild.Container, hash))
	}

	return &config.DockerBuildConfig{
		Container:  dockerBuild.Container,
		Context:    dockerBuild.Context,
		Dockerfile: dockerBuild.Dockerfile,
		Platforms:  []string{localPlatform},
		Tags:       localTags,
		Load:       true,
		Push:       config.BoolPtr(false),
		Cache:      nil,
		SBOM:       false,
		Provenance: false,
		Registry:   "",
	}
}

// buildDockerBuildArgs constructs the docker buildx build command arguments.
func buildDockerBuildArgs(dockerBuild *config.DockerBuildConfig, buildContext, dockerfilePath string, tags []string, moniker string, opts BuildOptions, logWriter io.Writer) []string {
	args := []string{"buildx", "build"}

	// Resolve builder:
	// - Explicit config: always use it
	// - CI: omit --builder so docker uses the default buildx builder
	//   (set up by docker/setup-buildx-action with docker-container driver for multi-platform)
	// - Local: resolve the docker-driver builder for the active Docker context
	//   to avoid user-configured buildx defaults that may use a slower docker-container driver
	builder := dockerBuild.Builder
	if builder == "" && !environments.IsCI() {
		builder = detectDockerBuilder()
	}
	if builder != "" {
		args = append(args, "--builder", builder)
	}

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

// authenticateRegistry authenticates to a Docker registry via the tool executor.
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

	return RunCommandWithStdin(
		context.Background(), "", logWriter,
		strings.NewReader(token),
		"docker", "login", registry, "-u", actor, "--password-stdin",
	)
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
// Supported variables: {moniker}, {container}, {owner}, {short_sha}.
func expandTemplate(template, moniker string) string {
	result := template
	result = strings.ReplaceAll(result, "{moniker}", moniker)
	result = strings.ReplaceAll(result, "{container}", moniker) // container defaults to moniker

	// Expand {owner} from CI environment
	if owner := os.Getenv("GITHUB_REPOSITORY_OWNER"); owner != "" {
		result = strings.ReplaceAll(result, "{owner}", owner)
	}

	// Get short SHA if available
	if sha := os.Getenv("GITHUB_SHA"); sha != "" && len(sha) >= 7 {
		result = strings.ReplaceAll(result, "{short_sha}", sha[:7])
	}

	return result
}

// getDockerBuildConfig gets docker_build config from a named component, falling back to container type.
func getDockerBuildConfig(module *modules.ModuleContract, componentName string) *config.DockerBuildConfig {
	// Try named component first
	if componentName != "" {
		if pkg, ok := module.Components[componentName]; ok && pkg != nil && pkg.DockerBuild != nil {
			return pkg.DockerBuild
		}
	}

	// Fall back to first component with type "container"
	if _, entry := module.Components.GetFirstByType("container"); entry != nil && entry.DockerBuild != nil {
		return entry.DockerBuild
	}
	return nil
}

var _ build.BuilderPort = (*BuildxHandler)(nil)
