// Package containerruntime provides runtime selection for container operations.
//
// This package abstracts container runtime selection, allowing the tool system
// to work with different runtimes (Docker, Podman, containerd) without requiring
// full implementations for all of them.
//
// Currently, Docker is the only working runtime. Podman and containerd stubs
// return "not implemented" errors but provide the type structure for future
// implementations.
package containerruntime

import (
	"sync"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/adapters/docker"
)

// RuntimeType identifies a container runtime.
type RuntimeType string

const (
	// RuntimeDocker is the Docker container runtime.
	RuntimeDocker RuntimeType = "docker"
	// RuntimePodman is the Podman container runtime.
	RuntimePodman RuntimeType = "podman"
	// RuntimeContainerd is the containerd runtime with nerdctl CLI.
	RuntimeContainerd RuntimeType = "containerd"
)

var (
	globalContainer   container.ContainerPort
	globalContainerMu sync.RWMutex
)

// Global returns the global container runtime (Docker only for now).
// Lazily initializes on first call. Thread-safe.
func Global() container.ContainerPort {
	globalContainerMu.RLock()
	c := globalContainer
	globalContainerMu.RUnlock()

	if c != nil {
		return c
	}

	globalContainerMu.Lock()
	defer globalContainerMu.Unlock()

	// Double-check after acquiring write lock
	if globalContainer != nil {
		return globalContainer
	}

	// Delegate to docker package's global container
	globalContainer = docker.GlobalContainer()
	return globalContainer
}

// SetGlobal injects a container implementation (for tests or alternative runtimes).
func SetGlobal(c container.ContainerPort) {
	globalContainerMu.Lock()
	defer globalContainerMu.Unlock()
	globalContainer = c
}

// ResetGlobal clears the global container, forcing re-initialization on next call.
// This resets both this package's reference and the underlying docker package's
// global to ensure a completely clean state (useful for testing).
func ResetGlobal() {
	globalContainerMu.Lock()
	defer globalContainerMu.Unlock()
	// Clear our reference without calling Close() since we delegate to docker's
	// GlobalContainer(), which owns the lifecycle. We reset the docker global
	// below, which handles closing properly.
	globalContainer = nil
	docker.ResetGlobalContainer()
}
