// Package containerruntime provides runtime-agnostic container provider and discovery.
// It exposes a global provider that currently discovers and selects Docker only.
package containerruntime

import (
	"context"
	"fmt"
	"sync"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/adapters/docker"
)

var (
	globalContainer   container.ContainerPort
	globalContainerMu sync.RWMutex
)

// Global returns the global container runtime implementation.
// Lazily initializes on first call using discovery logic (Docker-only for now).
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

	// Discovery (initial): try Docker adapter
	adapter, err := docker.NewContainerAdapter()
	if err != nil {
		return &unavailableAdapter{err: err}
	}
	globalContainer = adapter
	return globalContainer
}

// SetGlobal allows injecting a specific container implementation (tests or overrides).
func SetGlobal(c container.ContainerPort) {
	globalContainerMu.Lock()
	defer globalContainerMu.Unlock()
	globalContainer = c
}

// ResetGlobal closes and clears the global container forcing re-initialization.
func ResetGlobal() {
	globalContainerMu.Lock()
	defer globalContainerMu.Unlock()
	if globalContainer != nil {
		_ = globalContainer.Close() // best-effort
		globalContainer = nil
	}
}

// unavailableAdapter is a ContainerPort that always reports unavailable.
type unavailableAdapter struct {
	err error
}

func (a *unavailableAdapter) Execute(_ context.Context, _ *container.ContainerConfig) (*container.ContainerResult, error) {
	return nil, fmt.Errorf("container runtime not available: %w", a.err)
}

func (a *unavailableAdapter) Build(_ context.Context, _ *container.BuildConfig) error {
	return fmt.Errorf("container runtime not available: %w", a.err)
}

func (a *unavailableAdapter) Pull(_ context.Context, _ string) error {
	return fmt.Errorf("container runtime not available: %w", a.err)
}

func (a *unavailableAdapter) ImageExists(_ context.Context, _ string) bool { return false }

func (a *unavailableAdapter) IsAvailable() bool { return false }

func (a *unavailableAdapter) Close() error { return nil }
