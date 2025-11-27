//go:build L0
// +build L0

package docker

import (
	"fmt"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestParallelEnsureImages_Success(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock successful image inspections (images already exist locally)
	mockClient.On("ImageInspect", mock.Anything, "image1:latest").
		Return(image.InspectResponse{ID: "sha256:abc123"}, nil).Once()
	mockClient.On("ImageInspect", mock.Anything, "image2:latest").
		Return(image.InspectResponse{ID: "sha256:def456"}, nil).Once()
	mockClient.On("ImageInspect", mock.Anything, "image3:v1.0.0").
		Return(image.InspectResponse{
			ID:          "sha256:ghi789",
			RepoDigests: []string{"image3@sha256:ghi789"},
		}, nil).Once()

	requests := []ImagePullRequest{
		{ImageName: "image1:latest", PullPolicy: "IfNotPresent", LoadLocal: false, Index: 0},
		{ImageName: "image2:latest", PullPolicy: "IfNotPresent", LoadLocal: false, Index: 1},
		{ImageName: "image3:v1.0.0", PullPolicy: "IfNotPresent", LoadLocal: false, Index: 2},
	}

	results := host.ParallelEnsureImages(requests, nil)

	assert.Len(t, results, 3)
	for _, result := range results {
		assert.True(t, result.Success, "Expected success for image: %s", result.ImageName)
		assert.NoError(t, result.Error)
	}
}

func TestParallelEnsureImages_WithFailures(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock: First image exists, second fails, third exists
	mockClient.On("ImageInspect", mock.Anything, "image1:latest").
		Return(image.InspectResponse{ID: "sha256:abc123"}, nil).Once()
	mockClient.On("ImageInspect", mock.Anything, "missing:latest").
		Return(image.InspectResponse{}, fmt.Errorf("not found")).Once()
	mockClient.On("ImageInspect", mock.Anything, "image3:v1.0.0").
		Return(image.InspectResponse{
			ID:          "sha256:ghi789",
			RepoDigests: []string{"image3@sha256:ghi789"},
		}, nil).Once()

	requests := []ImagePullRequest{
		{ImageName: "image1:latest", PullPolicy: "IfNotPresent", LoadLocal: false, Index: 0},
		{ImageName: "missing:latest", PullPolicy: "Never", LoadLocal: false, Index: 1},
		{ImageName: "image3:v1.0.0", PullPolicy: "IfNotPresent", LoadLocal: false, Index: 2},
	}

	results := host.ParallelEnsureImages(requests, nil)

	assert.Len(t, results, 3)
	assert.True(t, results[0].Success)
	assert.False(t, results[1].Success)
	assert.Error(t, results[1].Error)
	assert.True(t, results[2].Success)
}

func TestParallelEnsureImages_EmptyList(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	requests := []ImagePullRequest{}
	results := host.ParallelEnsureImages(requests, nil)

	assert.Empty(t, results)
}

func TestParallelEnsureImages_CustomConcurrency(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock 5 images
	for i := 1; i <= 5; i++ {
		imageName := fmt.Sprintf("image%d:latest", i)
		mockClient.On("ImageInspect", mock.Anything, imageName).
			Return(image.InspectResponse{ID: fmt.Sprintf("sha256:abc%d", i)}, nil).Once()
	}

	requests := make([]ImagePullRequest, 5)
	for i := 0; i < 5; i++ {
		requests[i] = ImagePullRequest{
			ImageName:  fmt.Sprintf("image%d:latest", i+1),
			PullPolicy: "IfNotPresent",
			LoadLocal:  false,
			Index:      i,
		}
	}

	opts := &ParallelImagePullOptions{
		MaxConcurrency: 2, // Limit to 2 concurrent pulls
	}

	results := host.ParallelEnsureImages(requests, opts)

	assert.Len(t, results, 5)
	for _, result := range results {
		assert.True(t, result.Success)
	}
}

func TestEnsureMultipleImages_Convenience(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock 3 images
	mockClient.On("ImageInspect", mock.Anything, "image1:latest").
		Return(image.InspectResponse{ID: "sha256:abc123"}, nil).Once()
	mockClient.On("ImageInspect", mock.Anything, "image2:latest").
		Return(image.InspectResponse{ID: "sha256:def456"}, nil).Once()
	mockClient.On("ImageInspect", mock.Anything, "image3:latest").
		Return(image.InspectResponse{ID: "sha256:ghi789"}, nil).Once()

	imageNames := []string{"image1:latest", "image2:latest", "image3:latest"}
	results := host.EnsureMultipleImages(imageNames, "IfNotPresent", false)

	assert.Len(t, results, 3)
	for _, result := range results {
		assert.True(t, result.Success)
		assert.NoError(t, result.Error)
	}
}

func TestParallelEnsureImages_PreservesIndex(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock images
	mockClient.On("ImageInspect", mock.Anything, "imageA:latest").
		Return(image.InspectResponse{ID: "sha256:aaa"}, nil).Once()
	mockClient.On("ImageInspect", mock.Anything, "imageB:latest").
		Return(image.InspectResponse{ID: "sha256:bbb"}, nil).Once()
	mockClient.On("ImageInspect", mock.Anything, "imageC:latest").
		Return(image.InspectResponse{ID: "sha256:ccc"}, nil).Once()

	requests := []ImagePullRequest{
		{ImageName: "imageA:latest", PullPolicy: "IfNotPresent", LoadLocal: false, Index: 10},
		{ImageName: "imageB:latest", PullPolicy: "IfNotPresent", LoadLocal: false, Index: 20},
		{ImageName: "imageC:latest", PullPolicy: "IfNotPresent", LoadLocal: false, Index: 30},
	}

	results := host.ParallelEnsureImages(requests, nil)

	assert.Len(t, results, 3)
	assert.Equal(t, 10, results[0].Index)
	assert.Equal(t, 20, results[1].Index)
	assert.Equal(t, 30, results[2].Index)
}
