package docker

import (
	"sync"

	container "github.com/ready-to-release/eac/contracts/container-runtime/0.1.0"
)

var (
	globalContainer   container.ContainerPort
	globalContainerMu sync.RWMutex
)

// GlobalContainer returns the global container adapter.
// Lazily initializes Docker adapter on first call.
func GlobalContainer() container.ContainerPort {
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

	adapter, err := NewContainerAdapter()
	if err != nil {
		// Return a nil-safe adapter that reports unavailable
		return &unavailableAdapter{err: err}
	}
	globalContainer = adapter
	return globalContainer
}

// SetGlobalContainer allows injecting a different adapter (for testing or Podman).
func SetGlobalContainer(c container.ContainerPort) {
	globalContainerMu.Lock()
	defer globalContainerMu.Unlock()
	globalContainer = c
}

// ResetGlobalContainer resets the global container to nil, forcing re-initialization.
// Useful for testing.
func ResetGlobalContainer() {
	globalContainerMu.Lock()
	defer globalContainerMu.Unlock()
	if globalContainer != nil {
		_ = globalContainer.Close() //nolint:errcheck
		globalContainer = nil
	}
}
