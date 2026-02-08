// Package container defines the interface contract for OCI container operations.
//
// This package provides a runtime-agnostic interface for container operations.
// Implementations can use Docker, Podman, containerd, or other OCI-compliant runtimes.
//
// The primary interface is ContainerPort, which abstracts:
//   - Container execution (running one-shot tasks)
//   - Image building
//   - Image pulling
//   - Runtime availability checks
//
// Implementations are provided by adapter packages such as:
//   - github.com/ready-to-release/eac/go/eac/adapters/docker (Docker adapter)
package container

import (
	"context"
	"io"
	"time"
)

// ContainerPort defines the interface for OCI container operations.
// Implementations: Docker, Podman, containerd, etc.
type ContainerPort interface {
	// Execute runs a container and waits for completion.
	// This is for one-shot tasks, not long-running services.
	Execute(ctx context.Context, config *ContainerConfig) (*ContainerResult, error)

	// Build builds a container image from a Containerfile/Dockerfile.
	Build(ctx context.Context, config *BuildConfig) error

	// Pull ensures an image is available locally.
	Pull(ctx context.Context, imageRef string) error

	// ImageExists checks if an image exists locally.
	ImageExists(ctx context.Context, imageRef string) bool

	// ManifestExists checks if an image manifest exists in a remote registry.
	ManifestExists(ctx context.Context, imageRef string) (bool, error)

	// ImageCreatedTime returns the creation timestamp of a local image.
	ImageCreatedTime(ctx context.Context, imageRef string) (time.Time, error)

	// IsAvailable checks if the container runtime is available.
	IsAvailable() bool

	// Close releases resources.
	Close() error
}

// ContainerConfig defines configuration for running a container.
type ContainerConfig struct {
	// Image is the container image reference (e.g., "golang:1.21").
	Image string

	// Command is the command to run in the container.
	Command []string

	// Entrypoint overrides the image's default entrypoint.
	Entrypoint []string

	// WorkingDir is the working directory inside the container.
	WorkingDir string

	// Env contains environment variables as key-value pairs.
	Env map[string]string

	// Mounts defines volume mounts.
	Mounts []MountConfig

	// User specifies the user to run as (e.g., "1000:1000").
	User string

	// Privileged enables privileged mode.
	Privileged bool

	// Network specifies the network mode (e.g., "host", "bridge").
	Network string

	// Resources defines resource limits.
	Resources *ResourceConfig

	// Timeout is the maximum time to wait for the container.
	// Zero means no timeout.
	Timeout time.Duration

	// LogWriter receives container output in real-time (combined stdout+stderr).
	// If nil, output is captured in ContainerResult.
	LogWriter io.Writer

	// StdoutWriter receives stdout separately for structured output parsing.
	// When set, stdout is piped here instead of LogWriter.
	// Takes precedence over LogWriter for stdout stream.
	// nil = use LogWriter or capture to ContainerResult.Stdout
	StdoutWriter io.Writer

	// StderrWriter receives stderr separately.
	// When set, stderr is piped here instead of LogWriter.
	// Takes precedence over LogWriter for stderr stream.
	// nil = use LogWriter or capture to ContainerResult.Stderr
	StderrWriter io.Writer

	// StdinReader provides stdin to the container.
	// When set, stdin is piped from this reader.
	// nil = no stdin
	StdinReader io.Reader

	// ContainerName is an optional name for the container.
	// If empty, a unique name is generated.
	ContainerName string
}

// MountConfig defines a volume mount.
type MountConfig struct {
	// Source is the host path to mount.
	Source string

	// Target is the path inside the container.
	Target string

	// ReadOnly makes the mount read-only.
	ReadOnly bool
}

// ResourceConfig defines resource limits.
type ResourceConfig struct {
	// CPUs is the CPU limit (e.g., 2.0 for 2 CPUs).
	CPUs float64

	// Memory is the memory limit (e.g., "4g", "512m").
	Memory string

	// ShmSize is the shared memory size (e.g., "256m").
	ShmSize string
}

// ContainerResult holds execution results.
type ContainerResult struct {
	// ExitCode is the container's exit code.
	ExitCode int

	// Stdout contains captured stdout output (when LogWriter is nil).
	Stdout []byte

	// Stderr contains captured stderr output (when LogWriter is nil).
	Stderr []byte

	// Duration is how long the container ran.
	Duration time.Duration
}

// BuildConfig defines configuration for building an image.
type BuildConfig struct {
	// ContextPath is the build context directory.
	ContextPath string

	// Dockerfile is the path to the Dockerfile/Containerfile.
	// If relative, it's relative to ContextPath.
	Dockerfile string

	// ImageTag is the tag for the built image.
	ImageTag string

	// BuildArgs contains build arguments.
	BuildArgs map[string]string

	// Platforms specifies target platforms for multi-arch builds.
	// Example: ["linux/amd64", "linux/arm64"]
	Platforms []string

	// Push pushes the image after building.
	Push bool

	// Load loads the image into the local daemon (for single-platform builds).
	Load bool

	// CacheFrom specifies cache sources.
	CacheFrom string

	// CacheTo specifies cache destinations.
	CacheTo string

	// LogWriter receives build output.
	LogWriter io.Writer
}
