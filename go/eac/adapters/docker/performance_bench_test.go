//go:build benchmark

package docker

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"

	"github.com/ready-to-release/eac/go/eac/adapters/docker/mocks"
)

// BenchmarkPortAllocation_Sequential benchmarks sequential port scanning from start.
// This represents the old approach before optimization.
func BenchmarkPortAllocation_Sequential(b *testing.B) {
	// Reserve some ports to simulate realistic conditions
	reservedPorts := []int{8000, 8001, 8002, 8003, 8004}
	listeners := make([]net.Listener, 0, len(reservedPorts))
	for _, port := range reservedPorts {
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			listeners = append(listeners, ln)
			defer ln.Close()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Sequential scan from start
		portRange := PortRangeEnd - PortRangeStart + 1
		found := false
		for offset := 0; offset < portRange; offset++ {
			port := PortRangeStart + offset
			if IsPortAvailable(port) {
				found = true
				break
			}
		}
		if !found {
			b.Fatal("No available port found")
		}
	}
}

// BenchmarkPortAllocation_RandomStart benchmarks random start + circular scan.
// This represents the optimized approach.
func BenchmarkPortAllocation_RandomStart(b *testing.B) {
	// Reserve some ports to simulate realistic conditions
	reservedPorts := []int{8000, 8001, 8002, 8003, 8004}
	listeners := make([]net.Listener, 0, len(reservedPorts))
	for _, port := range reservedPorts {
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			listeners = append(listeners, ln)
			defer ln.Close()
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		port, err := FindAvailablePort()
		if err != nil {
			b.Fatal(err)
		}
		if port == 0 {
			b.Fatal("No available port found")
		}
	}
}

// BenchmarkPortReservation_WithLocking benchmarks port reservation with mutex.
func BenchmarkPortReservation_WithLocking(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		port, err := FindAndReservePort()
		if err != nil {
			b.Fatal(err)
		}
		ReleasePort(port)
	}
}

// BenchmarkPortReservation_Parallel benchmarks concurrent port reservations.
func BenchmarkPortReservation_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			port, err := FindAndReservePort()
			if err != nil {
				b.Error(err)
				continue
			}
			ReleasePort(port)
		}
	})
}

// BenchmarkImageEnsure_Sequential benchmarks sequential image operations.
func BenchmarkImageEnsure_Sequential(b *testing.B) {
	ctx := context.Background()
	mockClient := &mocks.MockDockerClient{}

	// Mock image operations
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	for i := 1; i <= 5; i++ {
		imageName := fmt.Sprintf("benchmark-image%d:latest", i)
		mockClient.On("ImagePull", ctx, imageName, image.PullOptions{}).
			Return(io.NopCloser(strings.NewReader("{}")), nil)
	}

	configs := make([]*ServeConfig, 5)
	for i := 0; i < 5; i++ {
		configs[i] = &ServeConfig{
			Image: fmt.Sprintf("benchmark-image%d:latest", i+1),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Sequential execution
		for _, config := range configs {
			if err := ensureImage(ctx, mockClient, config); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkImageEnsure_Parallel benchmarks parallel image operations.
func BenchmarkImageEnsure_Parallel(b *testing.B) {
	ctx := context.Background()
	mockClient := &mocks.MockDockerClient{}

	// Mock image operations
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	for i := 1; i <= 5; i++ {
		imageName := fmt.Sprintf("benchmark-image%d:latest", i)
		mockClient.On("ImagePull", ctx, imageName, image.PullOptions{}).
			Return(io.NopCloser(strings.NewReader("{}")), nil)
	}

	ops := make([]ImageOperation, 5)
	for i := 0; i < 5; i++ {
		ops[i] = ImageOperation{
			Config: &ServeConfig{
				Image: fmt.Sprintf("benchmark-image%d:latest", i+1),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := ParallelImageEnsure(ctx, mockClient, ops, nil)
		for _, result := range results {
			if !result.Success {
				b.Fatal(result.Error)
			}
		}
	}
}

// BenchmarkImageEnsure_ParallelConcurrency1 benchmarks parallel with concurrency=1.
func BenchmarkImageEnsure_ParallelConcurrency1(b *testing.B) {
	ctx := context.Background()
	mockClient := &mocks.MockDockerClient{}

	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	for i := 1; i <= 5; i++ {
		imageName := fmt.Sprintf("benchmark-image%d:latest", i)
		mockClient.On("ImagePull", ctx, imageName, image.PullOptions{}).
			Return(io.NopCloser(strings.NewReader("{}")), nil)
	}

	ops := make([]ImageOperation, 5)
	for i := 0; i < 5; i++ {
		ops[i] = ImageOperation{
			Config: &ServeConfig{
				Image: fmt.Sprintf("benchmark-image%d:latest", i+1),
			},
		}
	}

	opts := &ParallelImageEnsureOptions{MaxConcurrency: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := ParallelImageEnsure(ctx, mockClient, ops, opts)
		for _, result := range results {
			if !result.Success {
				b.Fatal(result.Error)
			}
		}
	}
}

// BenchmarkImageEnsure_ParallelConcurrency3 benchmarks parallel with concurrency=3.
func BenchmarkImageEnsure_ParallelConcurrency3(b *testing.B) {
	ctx := context.Background()
	mockClient := &mocks.MockDockerClient{}

	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	for i := 1; i <= 5; i++ {
		imageName := fmt.Sprintf("benchmark-image%d:latest", i)
		mockClient.On("ImagePull", ctx, imageName, image.PullOptions{}).
			Return(io.NopCloser(strings.NewReader("{}")), nil)
	}

	ops := make([]ImageOperation, 5)
	for i := 0; i < 5; i++ {
		ops[i] = ImageOperation{
			Config: &ServeConfig{
				Image: fmt.Sprintf("benchmark-image%d:latest", i+1),
			},
		}
	}

	opts := &ParallelImageEnsureOptions{MaxConcurrency: 3}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := ParallelImageEnsure(ctx, mockClient, ops, opts)
		for _, result := range results {
			if !result.Success {
				b.Fatal(result.Error)
			}
		}
	}
}

// BenchmarkImageEnsure_ParallelConcurrency5 benchmarks parallel with concurrency=5.
func BenchmarkImageEnsure_ParallelConcurrency5(b *testing.B) {
	ctx := context.Background()
	mockClient := &mocks.MockDockerClient{}

	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	for i := 1; i <= 5; i++ {
		imageName := fmt.Sprintf("benchmark-image%d:latest", i)
		mockClient.On("ImagePull", ctx, imageName, image.PullOptions{}).
			Return(io.NopCloser(strings.NewReader("{}")), nil)
	}

	ops := make([]ImageOperation, 5)
	for i := 0; i < 5; i++ {
		ops[i] = ImageOperation{
			Config: &ServeConfig{
				Image: fmt.Sprintf("benchmark-image%d:latest", i+1),
			},
		}
	}

	opts := &ParallelImageEnsureOptions{MaxConcurrency: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := ParallelImageEnsure(ctx, mockClient, ops, opts)
		for _, result := range results {
			if !result.Success {
				b.Fatal(result.Error)
			}
		}
	}
}
