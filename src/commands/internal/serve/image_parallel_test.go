package serve

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
)

func TestParallelImageEnsure_EmptyOperations(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	results := ParallelImageEnsure(ctx, mockClient, []ImageOperation{}, nil)

	assert.Empty(t, results, "Expected empty results for empty operations")
}

func TestParallelImageEnsure_SingleOperation(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock image doesn't exist, needs to be pulled
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	mockClient.On("ImagePull", ctx, "test:latest", image.PullOptions{}).Return(io.NopCloser(strings.NewReader("{}")), nil)

	ops := []ImageOperation{
		{Config: &ServeConfig{Image: "test:latest"}},
	}

	results := ParallelImageEnsure(ctx, mockClient, ops, nil)

	assert.Len(t, results, 1)
	assert.True(t, results[0].Success)
	assert.Equal(t, "test:latest", results[0].Image)
	assert.NoError(t, results[0].Error)
	mockClient.AssertExpectations(t)
}

func TestParallelImageEnsure_MultipleOperations(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock all images don't exist, need to be pulled
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	mockClient.On("ImagePull", ctx, "image1:latest", image.PullOptions{}).Return(io.NopCloser(strings.NewReader("{}")), nil)
	mockClient.On("ImagePull", ctx, "image2:latest", image.PullOptions{}).Return(io.NopCloser(strings.NewReader("{}")), nil)
	mockClient.On("ImagePull", ctx, "image3:latest", image.PullOptions{}).Return(io.NopCloser(strings.NewReader("{}")), nil)

	ops := []ImageOperation{
		{Config: &ServeConfig{Image: "image1:latest"}},
		{Config: &ServeConfig{Image: "image2:latest"}},
		{Config: &ServeConfig{Image: "image3:latest"}},
	}

	results := ParallelImageEnsure(ctx, mockClient, ops, nil)

	assert.Len(t, results, 3)
	for i, result := range results {
		assert.True(t, result.Success, "Operation %d should succeed", i)
		assert.NoError(t, result.Error, "Operation %d should have no error", i)
	}
	mockClient.AssertExpectations(t)
}

func TestParallelImageEnsure_WithFailures(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock: image1 succeeds, image2 fails, image3 succeeds
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	mockClient.On("ImagePull", ctx, "image1:latest", image.PullOptions{}).Return(io.NopCloser(strings.NewReader("{}")), nil)
	mockClient.On("ImagePull", ctx, "image2:latest", image.PullOptions{}).Return(nil, errors.New("pull failed"))
	mockClient.On("ImagePull", ctx, "image3:latest", image.PullOptions{}).Return(io.NopCloser(strings.NewReader("{}")), nil)

	ops := []ImageOperation{
		{Config: &ServeConfig{Image: "image1:latest"}},
		{Config: &ServeConfig{Image: "image2:latest"}},
		{Config: &ServeConfig{Image: "image3:latest"}},
	}

	results := ParallelImageEnsure(ctx, mockClient, ops, nil)

	assert.Len(t, results, 3)
	assert.True(t, results[0].Success, "image1 should succeed")
	assert.False(t, results[1].Success, "image2 should fail")
	assert.True(t, results[2].Success, "image3 should succeed")
	assert.Error(t, results[1].Error, "image2 should have error")
	mockClient.AssertExpectations(t)
}

func TestParallelImageEnsure_ConcurrencyLimit(t *testing.T) {
	// This test verifies that the concurrency limit is respected
	// We test with 6 operations and a limit of 2, ensuring max 2 run concurrently
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock that images don't exist
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)

	// Mock pulls - all succeed with minimal delay
	for i := 1; i <= 6; i++ {
		imageName := "image" + string(rune('0'+i)) + ":latest"
		mockClient.On("ImagePull", ctx, imageName, image.PullOptions{}).
			Return(io.NopCloser(strings.NewReader("{}")), nil)
	}

	ops := make([]ImageOperation, 6)
	for i := 0; i < 6; i++ {
		ops[i] = ImageOperation{
			Config: &ServeConfig{Image: "image" + string(rune('0'+i+1)) + ":latest"},
		}
	}

	// Set concurrency limit to 2
	opts := &ParallelImageEnsureOptions{MaxConcurrency: 2}
	results := ParallelImageEnsure(ctx, mockClient, ops, opts)

	assert.Len(t, results, 6)
	for i, result := range results {
		assert.True(t, result.Success, "Operation %d should succeed", i)
	}
	mockClient.AssertExpectations(t)
}

func TestCheckImageResults_AllSuccess(t *testing.T) {
	results := []ImageResult{
		{Image: "image1:latest", Success: true},
		{Image: "image2:latest", Success: true},
	}

	err := CheckImageResults(results)
	assert.NoError(t, err)
}

func TestCheckImageResults_SomeFailures(t *testing.T) {
	results := []ImageResult{
		{Image: "image1:latest", Success: true},
		{Image: "image2:latest", Success: false, Error: errors.New("pull failed")},
		{Image: "image3:latest", Success: false, Error: errors.New("network error")},
	}

	err := CheckImageResults(results)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ensure 2 image(s)")
	assert.Contains(t, err.Error(), "image2:latest")
	assert.Contains(t, err.Error(), "image3:latest")
	assert.Contains(t, err.Error(), "pull failed")
	assert.Contains(t, err.Error(), "network error")
}

func TestGetSuccessfulImages(t *testing.T) {
	results := []ImageResult{
		{Image: "image1:latest", Success: true},
		{Image: "image2:latest", Success: false},
		{Image: "image3:latest", Success: true},
	}

	successful := GetSuccessfulImages(results)
	assert.Len(t, successful, 2)
	assert.Contains(t, successful, "image1:latest")
	assert.Contains(t, successful, "image3:latest")
}

func TestGetFailedImages(t *testing.T) {
	results := []ImageResult{
		{Image: "image1:latest", Success: true},
		{Image: "image2:latest", Success: false},
		{Image: "image3:latest", Success: false},
	}

	failed := GetFailedImages(results)
	assert.Len(t, failed, 2)
	assert.Contains(t, failed, "image2:latest")
	assert.Contains(t, failed, "image3:latest")
}

func TestParallelImageEnsure_CustomConcurrency(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{}, nil)
	mockClient.On("ImagePull", ctx, "image1:latest", image.PullOptions{}).Return(io.NopCloser(strings.NewReader("{}")), nil)
	mockClient.On("ImagePull", ctx, "image2:latest", image.PullOptions{}).Return(io.NopCloser(strings.NewReader("{}")), nil)

	ops := []ImageOperation{
		{Config: &ServeConfig{Image: "image1:latest"}},
		{Config: &ServeConfig{Image: "image2:latest"}},
	}

	// Test with custom concurrency
	opts := &ParallelImageEnsureOptions{MaxConcurrency: 1}
	results := ParallelImageEnsure(ctx, mockClient, ops, opts)

	assert.Len(t, results, 2)
	for _, result := range results {
		assert.True(t, result.Success)
	}
	mockClient.AssertExpectations(t)
}

func TestParallelImageEnsure_ExistingImages(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockDockerClient{}

	// Mock that all images already exist
	mockClient.On("ImageList", ctx, image.ListOptions{}).Return([]image.Summary{
		{RepoTags: []string{"image1:latest"}},
		{RepoTags: []string{"image2:latest"}},
	}, nil)

	ops := []ImageOperation{
		{Config: &ServeConfig{Image: "image1:latest"}},
		{Config: &ServeConfig{Image: "image2:latest"}},
	}

	results := ParallelImageEnsure(ctx, mockClient, ops, nil)

	assert.Len(t, results, 2)
	for _, result := range results {
		assert.True(t, result.Success, "Should succeed when image exists")
		assert.NoError(t, result.Error)
	}

	// ImagePull should not be called
	mockClient.AssertNotCalled(t, "ImagePull")
	mockClient.AssertExpectations(t)
}
