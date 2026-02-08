package tool

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	container "github.com/ready-to-release/eac/contracts/container-runtime/0.1.0"
	"github.com/ready-to-release/eac/go/core/environments"
)

// ImageManager handles container image lifecycle (check, build, pull).
// It provides environment-aware image resolution:
//   - Local containers (LocalPath set): Build in devbox, pull from GHCR in CI.
//   - External containers: Try GHCR cache first, fallback to upstream.
//
// When a ContainerPort is injected, all Docker operations go through the port
// interface. Otherwise, falls back to direct Docker CLI calls for backward
// compatibility.
type ImageManager struct {
	workspaceRoot string
	isCI          bool
	ghcrOrg       string
	logWriter     io.Writer
	container     container.ContainerPort // optional: injected container runtime
}

// NewImageManager creates a new image manager.
func NewImageManager(workspaceRoot string, isCI bool, ghcrOrg string, logWriter io.Writer) *ImageManager {
	if logWriter == nil {
		logWriter = io.Discard
	}
	return &ImageManager{
		workspaceRoot: workspaceRoot,
		isCI:          isCI,
		ghcrOrg:       ghcrOrg,
		logWriter:     logWriter,
	}
}

// NewImageManagerWithContainer creates an image manager with an injected container runtime.
// When a ContainerPort is provided, Docker operations (availability check, image exists,
// build, pull) go through the port instead of shelling out to the Docker CLI.
func NewImageManagerWithContainer(workspaceRoot string, isCI bool, ghcrOrg string, logWriter io.Writer, cp container.ContainerPort) *ImageManager {
	mgr := NewImageManager(workspaceRoot, isCI, ghcrOrg, logWriter)
	mgr.container = cp
	return mgr
}

// SetContainer injects a container runtime into an existing ImageManager.
// This is useful when the ImageManager is created before the container runtime
// is available (e.g., lazy initialization).
func (m *ImageManager) SetContainer(cp container.ContainerPort) {
	m.container = cp
}

// ImageOptions controls image resolution behavior.
type ImageOptions struct {
	// ForceRefresh bypasses local cache and always pulls from remote.
	// Default: false (local-first)
	ForceRefresh bool

	// AllowStale permits using local images even if they might be outdated.
	// For external containers with pinned versions, stale = never (pinned = immutable).
	// For local containers, staleness is determined by file modification times.
	// Default: true (use local if available)
	AllowStale bool

	// OfflineMode prevents any network requests. Returns error if image not local.
	// Default: false (network allowed when needed)
	OfflineMode bool
}

// DefaultImageOptions returns the default image resolution options.
func DefaultImageOptions() ImageOptions {
	return ImageOptions{
		ForceRefresh: false,
		AllowStale:   true,
		OfflineMode:  false,
	}
}

// EnsureImage ensures the container image is available using default options.
// Returns the image reference to use for running the container.
// For non-container tools, returns empty string with no error.
func (m *ImageManager) EnsureImage(ctx context.Context, tool *ToolDefinition) (string, error) {
	return m.EnsureImageWithOptions(ctx, tool, DefaultImageOptions())
}

// EnsureImageWithOptions ensures the container image is available with options.
// Returns the image reference to use for running the container.
// For non-container tools, returns empty string with no error.
func (m *ImageManager) EnsureImageWithOptions(ctx context.Context, tool *ToolDefinition, opts ImageOptions) (string, error) {
	if tool.Type != ToolTypeContainer {
		return "", nil
	}

	// Pre-flight check: verify Docker is available
	if err := m.verifyDockerAvailable(); err != nil {
		return "", fmt.Errorf("docker not available: %w", err)
	}

	// For local containers, verify Dockerfile exists before attempting build
	if tool.IsLocalContainer() {
		contextPath := tool.LocalContextPath(m.workspaceRoot)
		dockerfilePath := filepath.Join(contextPath, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			return "", fmt.Errorf("dockerfile not found: %s (tool: %s)", dockerfilePath, tool.ID)
		}
		return m.ensureLocalImageWithOptions(ctx, tool, opts)
	}
	return m.ensureExternalImageWithOptions(ctx, tool, opts)
}

// verifyDockerAvailable checks if Docker daemon is accessible.
// Uses ContainerPort.IsAvailable() when a port is injected, otherwise
// falls back to running "docker info" directly.
func (m *ImageManager) verifyDockerAvailable() error {
	if m.container != nil {
		if !m.container.IsAvailable() {
			return fmt.Errorf("container runtime not accessible (is Docker running?)")
		}
		return nil
	}
	// Fallback: direct CLI check
	cmd := exec.Command("docker", "info") //nolint:gosec // G204: legacy fallback, to be removed when all callers inject ContainerPort
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker daemon not accessible (is Docker running?)")
	}
	return nil
}

// ensureLocalImageWithOptions handles local containers (have LocalPath) with options.
func (m *ImageManager) ensureLocalImageWithOptions(ctx context.Context, tool *ToolDefinition, opts ImageOptions) (string, error) {
	if m.isCI {
		return m.pullCIImageWithOptions(ctx, tool, opts)
	}
	return m.buildLocalImageWithOptions(ctx, tool, opts)
}

// buildLocalImageWithOptions builds from Dockerfile if stale or missing.
func (m *ImageManager) buildLocalImageWithOptions(ctx context.Context, tool *ToolDefinition, opts ImageOptions) (string, error) {
	imageTag := tool.LocalImageTag()
	contextPath := tool.LocalContextPath(m.workspaceRoot)

	// Check if we can use existing local image (unless ForceRefresh is set)
	if !opts.ForceRefresh && m.ImageExists(imageTag) {
		if opts.AllowStale || !m.IsImageStale(imageTag, contextPath) {
			m.log("Container image up-to-date: %s", imageTag)
			return imageTag, nil
		}
	}

	// In offline mode, we can't build (would need to pull base image)
	if opts.OfflineMode {
		if m.ImageExists(imageTag) {
			m.log("Using local image (offline mode): %s", imageTag)
			return imageTag, nil
		}
		return "", fmt.Errorf("image %s not available locally and offline mode is enabled", imageTag)
	}

	m.log("Building container image: %s", imageTag)
	if err := m.dockerBuild(ctx, imageTag, contextPath); err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}

	return imageTag, nil
}

// pullCIImageWithOptions pulls the pre-built image from GHCR in CI with options.
// Uses local-first resolution: checks local images before network requests.
func (m *ImageManager) pullCIImageWithOptions(ctx context.Context, tool *ToolDefinition, opts ImageOptions) (string, error) {
	name := filepath.Base(tool.LocalPath)

	// Try SHA-tagged image first (from current commit's CI build)
	sha := os.Getenv(environments.EnvGitHubSHA)
	if sha != "" && len(sha) >= 7 {
		shaTag := fmt.Sprintf("ghcr.io/%s/%s:sha-%s", m.ghcrOrg, name, sha[:7])

		// Check local first before pulling (unless ForceRefresh is set)
		if !opts.ForceRefresh && m.ImageExists(shaTag) {
			m.log("Using local CI image: %s", shaTag)
			return shaTag, nil
		}

		if !opts.OfflineMode {
			if err := m.dockerPull(ctx, shaTag); err == nil {
				m.log("Pulled CI image: %s", shaTag)
				return shaTag, nil
			}
			m.log("SHA image not found, trying latest: %s", shaTag)
		}
	}

	// Fallback to :latest
	latestTag := fmt.Sprintf("ghcr.io/%s/%s:latest", m.ghcrOrg, name)

	// Check local first before pulling (unless ForceRefresh is set)
	if !opts.ForceRefresh && m.ImageExists(latestTag) {
		m.log("Using local latest image: %s", latestTag)
		return latestTag, nil
	}

	if opts.OfflineMode {
		return "", fmt.Errorf("image %s not available locally and offline mode is enabled", latestTag)
	}

	if err := m.dockerPull(ctx, latestTag); err != nil {
		return "", fmt.Errorf("pull failed for %s: %w", latestTag, err)
	}
	m.log("Pulled latest image: %s", latestTag)
	return latestTag, nil
}

// ensureExternalImageWithOptions handles external containers (no LocalPath) with options.
// Uses local-first resolution: checks local images before network requests.
// For pinned versions (e.g., trivy:0.68.2), if it exists locally, it's immutable.
func (m *ImageManager) ensureExternalImageWithOptions(ctx context.Context, tool *ToolDefinition, opts ImageOptions) (string, error) {
	upstreamRef := tool.FullImage()
	if upstreamRef == "" {
		return "", fmt.Errorf("no image specified for tool %s", tool.ID)
	}

	// Check local for exact upstream image FIRST (unless ForceRefresh is set)
	// For pinned versions, if it exists locally, it's immutable and safe to use
	if !opts.ForceRefresh && m.ImageExists(upstreamRef) {
		m.log("Using local image: %s", upstreamRef)
		return upstreamRef, nil
	}

	// Try GHCR cache (local first, then pull)
	cacheRef := m.ghcrCacheRef(upstreamRef)
	if !opts.ForceRefresh && m.ImageExists(cacheRef) {
		m.log("Using cached image: %s", cacheRef)
		return cacheRef, nil
	}

	// In offline mode, fail if no local image found
	if opts.OfflineMode {
		return "", fmt.Errorf("image %s not available locally and offline mode is enabled", upstreamRef)
	}

	if err := m.dockerPull(ctx, cacheRef); err == nil {
		m.log("Pulled from GHCR cache: %s", cacheRef)
		return cacheRef, nil
	}

	// Cache miss - pull from upstream
	m.log("GHCR cache miss, pulling from upstream: %s", upstreamRef)
	if err := m.dockerPull(ctx, upstreamRef); err != nil {
		return "", fmt.Errorf("pull failed for %s: %w", upstreamRef, err)
	}
	return upstreamRef, nil
}

// ghcrCacheRef converts an upstream image ref to GHCR cache ref.
// Example: "ghcr.io/aquasecurity/trivy:0.50.0" → "ghcr.io/{org}/cache/aquasecurity/trivy:0.50.0".
func (m *ImageManager) ghcrCacheRef(upstreamRef string) string {
	// Parse image reference
	ref := upstreamRef
	tag := "latest"
	if idx := strings.LastIndex(ref, ":"); idx != -1 {
		tag = ref[idx+1:]
		ref = ref[:idx]
	}

	// Strip registry prefix if present
	path := ref
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		// Check if first part looks like a registry (contains a dot)
		if strings.Contains(parts[0], ".") {
			path = parts[1]
		}
	}

	return fmt.Sprintf("ghcr.io/%s/cache/%s:%s", m.ghcrOrg, path, tag)
}

// ImageExists checks if a Docker image exists locally.
// Uses ContainerPort.ImageExists() when a port is injected, otherwise
// falls back to "docker images -q".
func (m *ImageManager) ImageExists(imageRef string) bool {
	if m.container != nil {
		return m.container.ImageExists(context.Background(), imageRef)
	}
	// Fallback: direct CLI check
	cmd := exec.Command("docker", "images", "-q", imageRef) //nolint:gosec // G204: legacy fallback
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) != ""
}

// IsImageStale checks if any file in the context directory was modified after the image was created.
func (m *ImageManager) IsImageStale(imageRef, contextPath string) bool {
	imageTime, err := m.getImageCreatedTime(imageRef)
	if err != nil {
		return true // Can't determine, assume stale
	}

	newestFile, err := m.getNewestFileTime(contextPath)
	if err != nil {
		return true
	}

	return newestFile.After(imageTime)
}

func (m *ImageManager) getImageCreatedTime(imageRef string) (time.Time, error) {
	if m.container != nil {
		return m.container.ImageCreatedTime(context.Background(), imageRef)
	}
	cmd := exec.Command("docker", "inspect", "--format", "{{.Created}}", imageRef) //nolint:gosec // G204: legacy fallback, see Step 3D
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	timeStr := strings.TrimSpace(string(output))
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		t, err = time.Parse(time.RFC3339, timeStr)
	}
	return t, err
}

func (m *ImageManager) getNewestFileTime(dir string) (time.Time, error) {
	var newest time.Time
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})
	return newest, err
}

// dockerBuild builds a Docker image from a Dockerfile.
// Uses ContainerPort.Build() when a port is injected, otherwise
// falls back to running "docker build" directly.
func (m *ImageManager) dockerBuild(ctx context.Context, tag, contextPath string) error {
	if m.container != nil {
		return m.container.Build(ctx, &container.BuildConfig{
			ContextPath: contextPath,
			Dockerfile:  filepath.Join(contextPath, "Dockerfile"),
			ImageTag:    tag,
			LogWriter:   m.logWriter,
		})
	}
	// Fallback: direct CLI call
	dockerfile := filepath.Join(contextPath, "Dockerfile")
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", tag, "-f", dockerfile, contextPath) //nolint:gosec // G204: legacy fallback
	cmd.Stdout = m.logWriter
	cmd.Stderr = m.logWriter
	return cmd.Run()
}

// dockerPull pulls a Docker image from a registry.
// Uses ContainerPort.Pull() when a port is injected, otherwise
// falls back to running "docker pull" directly.
func (m *ImageManager) dockerPull(ctx context.Context, imageRef string) error {
	if m.container != nil {
		return m.container.Pull(ctx, imageRef)
	}
	// Fallback: direct CLI call
	cmd := exec.CommandContext(ctx, "docker", "pull", imageRef) //nolint:gosec // G204: legacy fallback
	cmd.Stdout = m.logWriter
	cmd.Stderr = m.logWriter
	return cmd.Run()
}

func (m *ImageManager) log(format string, args ...interface{}) {
	fmt.Fprintf(m.logWriter, "   "+format+"\n", args...)
}
