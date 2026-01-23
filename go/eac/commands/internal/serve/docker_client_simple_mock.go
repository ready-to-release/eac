package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// SimpleMockDockerClient is a mock Docker client for BDD tests.
// Unlike MockDockerClient (testify/mock), this mock returns sensible defaults
// for all operations without requiring explicit expectations.
type SimpleMockDockerClient struct {
	mu         sync.Mutex
	images     map[string]bool
	containers map[string]*mockContainerState
}

type mockContainerState struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	State   string    `json:"state"`
	Port    int       `json:"port"`
	Created time.Time `json:"created"`
}

// ensureInitialized ensures maps are initialized to prevent nil panics.
// This is a defensive programming guard in case SimpleMockDockerClient
// is instantiated directly instead of through the factory.
func (m *SimpleMockDockerClient) ensureInitialized() {
	if m.images == nil {
		m.images = make(map[string]bool)
	}
	if m.containers == nil {
		m.containers = make(map[string]*mockContainerState)
	}
}

// ImageList returns images that have been "pulled" or "built".
func (m *SimpleMockDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitialized()

	var summaries []image.Summary
	for img := range m.images {
		summaries = append(summaries, image.Summary{
			RepoTags: []string{img},
			Created:  time.Now().Unix() - 3600, // 1 hour ago
		})
	}
	return summaries, nil
}

// ImageBuild simulates successful image build.
func (m *SimpleMockDockerClient) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitialized()

	for _, tag := range options.Tags {
		m.images[tag] = true
	}

	m.saveStateUnlocked()
	return types.ImageBuildResponse{
		Body: io.NopCloser(strings.NewReader("")),
	}, nil
}

// ImagePull simulates successful image pull.
func (m *SimpleMockDockerClient) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitialized()

	m.images[refStr] = true
	m.saveStateUnlocked()
	return io.NopCloser(strings.NewReader("")), nil
}

// ContainerCreate simulates container creation with unique ID.
func (m *SimpleMockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitialized()

	id := "mock-" + containerName

	// Extract port from bindings
	var port int
	if hostConfig != nil && hostConfig.PortBindings != nil {
		for _, bindings := range hostConfig.PortBindings {
			for _, binding := range bindings {
				// Parse port from string (e.g., "9001" -> 9001)
				if binding.HostPort != "" {
					var err error
					port, err = parseInt(binding.HostPort)
					if err != nil {
						return container.CreateResponse{}, fmt.Errorf("failed to parse port: %w", err)
					}
					break
				}
			}
			if port > 0 {
				break
			}
		}
	}

	m.containers[containerName] = &mockContainerState{
		ID:      id,
		Name:    containerName,
		State:   "created",
		Port:    port,
		Created: time.Now(),
	}

	m.saveStateUnlocked()
	return container.CreateResponse{ID: id}, nil
}

// ContainerStart simulates starting a container.
func (m *SimpleMockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitialized()

	state := m.findContainerByID(containerID)
	if state == nil {
		return fmt.Errorf("container not found: %s", containerID)
	}
	if state.State == "running" {
		return fmt.Errorf("container already running: %s", containerID)
	}
	state.State = "running"
	m.saveStateUnlocked()
	return nil
}

// ContainerList returns running containers.
func (m *SimpleMockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitialized()

	var containers []types.Container
	for _, state := range m.containers {
		if !options.All && state.State != "running" {
			continue
		}

		c := types.Container{
			ID:      state.ID,
			Names:   []string{"/" + state.Name},
			State:   state.State,
			Created: state.Created.Unix(),
		}

		if state.Port > 0 {
			c.Ports = []types.Port{
				{
					PublicPort: uint16(state.Port),
					Type:       "tcp",
				},
			}
		}

		containers = append(containers, c)
	}

	return containers, nil
}

// ContainerStop simulates stopping a container.
func (m *SimpleMockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitialized()

	state := m.findContainerByID(containerID)
	if state == nil {
		return fmt.Errorf("container not found: %s", containerID)
	}
	if state.State == "stopped" {
		return fmt.Errorf("container already stopped: %s", containerID)
	}
	state.State = "stopped"
	m.saveStateUnlocked()
	return nil
}

// ContainerRemove simulates removing a container.
func (m *SimpleMockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureInitialized()

	for name, state := range m.containers {
		if state.ID == containerID {
			delete(m.containers, name)
			m.saveStateUnlocked()
			return nil
		}
	}
	return fmt.Errorf("container not found: %s", containerID)
}

// ContainerWait simulates waiting for container.
func (m *SimpleMockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	waitCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)

	waitCh <- container.WaitResponse{StatusCode: 0}
	close(waitCh)
	close(errCh)

	return waitCh, errCh
}

// ContainerLogs returns empty logs.
func (m *SimpleMockDockerClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

// ContainerAttach returns empty response.
func (m *SimpleMockDockerClient) ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error) {
	return types.HijackedResponse{}, nil
}

// Close is a no-op.
func (m *SimpleMockDockerClient) Close() error {
	// Save state to disk when client closes
	m.saveState()
	return nil
}

// mockState is the serializable state structure for persistence.
type mockState struct {
	Images     map[string]bool                `json:"images"`
	Containers map[string]*mockContainerState `json:"containers"`
}

// getStateFilePath returns the path to the mock state file in the system temp directory.
func getStateFilePath() string {
	return filepath.Join(os.TempDir(), "eac-mock-docker-state.json")
}

// loadState loads mock state from disk if the state file exists.
func (m *SimpleMockDockerClient) loadState() {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := getStateFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		// State file doesn't exist or can't be read - start with empty state
		return
	}

	var state mockState
	if err := json.Unmarshal(data, &state); err != nil {
		// Invalid state file - start with empty state
		return
	}

	m.images = state.Images
	m.containers = state.Containers
}

// saveState persists mock state to disk.
func (m *SimpleMockDockerClient) saveState() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveStateUnlocked()
}

// saveStateUnlocked persists mock state to disk without locking.
// Assumes caller already holds the mutex.
func (m *SimpleMockDockerClient) saveStateUnlocked() {
	state := mockState{
		Images:     m.images,
		Containers: m.containers,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	path := getStateFilePath()
	_ = os.WriteFile(path, data, 0644)
}

// deleteStateFile removes the persisted state file.
func deleteStateFile() {
	path := getStateFilePath()
	_ = os.Remove(path)
}

// Ensure SimpleMockDockerClient implements DockerClient.
var _ DockerClient = (*SimpleMockDockerClient)(nil)

// findContainerByID returns the container state for the given ID, or nil if not found.
// Must be called with mu held.
func (m *SimpleMockDockerClient) findContainerByID(containerID string) *mockContainerState {
	for _, state := range m.containers {
		if state.ID == containerID {
			return state
		}
	}
	return nil
}

// parseInt parses a port string to int, returning an error for invalid formats.
func parseInt(s string) (int, error) {
	var port int
	if _, err := fmt.Sscanf(s, "%d", &port); err != nil {
		return 0, fmt.Errorf("invalid port format: %s", s)
	}
	return port, nil
}
