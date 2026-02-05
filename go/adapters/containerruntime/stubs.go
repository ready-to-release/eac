package containerruntime

import (
	"context"
	"fmt"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
)

// unavailableAdapter returns errors for unavailable runtimes.
// This provides a type-safe placeholder that reports the runtime as unavailable
// rather than panicking on nil pointer access.
type unavailableAdapter struct {
	runtime RuntimeType
	err     error
}

func (a *unavailableAdapter) Execute(_ context.Context, _ *container.ContainerConfig) (*container.ContainerResult, error) {
	return nil, fmt.Errorf("%s runtime not available: %w", a.runtime, a.err)
}

func (a *unavailableAdapter) Build(_ context.Context, _ *container.BuildConfig) error {
	return fmt.Errorf("%s runtime not available: %w", a.runtime, a.err)
}

func (a *unavailableAdapter) Pull(_ context.Context, _ string) error {
	return fmt.Errorf("%s runtime not available: %w", a.runtime, a.err)
}

func (a *unavailableAdapter) ImageExists(_ context.Context, _ string) bool {
	return false
}

func (a *unavailableAdapter) IsAvailable() bool {
	return false
}

func (a *unavailableAdapter) Close() error {
	return nil
}

// PodmanStub returns a stub that reports Podman as not implemented.
// Use this when Podman is selected but not yet implemented.
func PodmanStub() container.ContainerPort {
	return &unavailableAdapter{
		runtime: RuntimePodman,
		err:     fmt.Errorf("podman support not yet implemented"),
	}
}

// ContainerdStub returns a stub that reports containerd as not implemented.
// Use this when containerd/nerdctl is selected but not yet implemented.
func ContainerdStub() container.ContainerPort {
	return &unavailableAdapter{
		runtime: RuntimeContainerd,
		err:     fmt.Errorf("containerd support not yet implemented"),
	}
}
