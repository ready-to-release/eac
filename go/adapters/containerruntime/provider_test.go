package containerruntime

import (
	"context"
	"sync"
	"testing"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
)

func TestGlobal_ReturnsAdapter(t *testing.T) {
	ResetGlobal()
	t.Cleanup(ResetGlobal)

	adapter := Global()
	if adapter == nil {
		t.Fatal("Global() returned nil")
	}
}

func TestGlobal_ReturnsSameInstance(t *testing.T) {
	ResetGlobal()
	t.Cleanup(ResetGlobal)

	adapter1 := Global()
	adapter2 := Global()

	if adapter1 != adapter2 {
		t.Error("Global() should return the same instance on subsequent calls")
	}
}

func TestGlobal_ThreadSafe(t *testing.T) {
	ResetGlobal()
	t.Cleanup(ResetGlobal)

	var wg sync.WaitGroup
	results := make([]container.ContainerPort, 100)

	for i := range 100 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = Global()
		}(i)
	}

	wg.Wait()

	// All results should be the same instance
	first := results[0]
	for i, r := range results {
		if r != first {
			t.Errorf("result[%d] is a different instance", i)
		}
	}
}

func TestSetGlobal_OverridesDefault(t *testing.T) {
	ResetGlobal()
	t.Cleanup(ResetGlobal)

	mock := PodmanStub()
	SetGlobal(mock)

	adapter := Global()
	if adapter != mock {
		t.Error("SetGlobal should override the default adapter")
	}
}

func TestPodmanStub_ReturnsUnavailable(t *testing.T) {
	stub := PodmanStub()
	if stub.IsAvailable() {
		t.Error("PodmanStub should report unavailable")
	}
}

func TestPodmanStub_ExecuteReturnsError(t *testing.T) {
	stub := PodmanStub()
	_, err := stub.Execute(context.Background(), nil)
	if err == nil {
		t.Error("PodmanStub.Execute should return an error")
	}
}

func TestContainerdStub_ReturnsUnavailable(t *testing.T) {
	stub := ContainerdStub()
	if stub.IsAvailable() {
		t.Error("ContainerdStub should report unavailable")
	}
}

func TestContainerdStub_ExecuteReturnsError(t *testing.T) {
	stub := ContainerdStub()
	_, err := stub.Execute(context.Background(), nil)
	if err == nil {
		t.Error("ContainerdStub.Execute should return an error")
	}
}

func TestContainerdStub_ImageExistsReturnsFalse(t *testing.T) {
	stub := ContainerdStub()
	if stub.ImageExists(context.Background(), "any:image") {
		t.Error("ContainerdStub.ImageExists should return false")
	}
}
