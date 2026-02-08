//go:build L0
// +build L0

package docker

import (
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupImageTestHost creates a ContainerHost with a MockDockerClient for image-related tests.
// The Ping mock is pre-configured since InspectImage calls EnsureConnected.
func setupImageTestHost() (*ContainerHost, *MockDockerClient) {
	mockClient := new(MockDockerClient)
	mockClient.On("Ping", mock.Anything).Return(types.Ping{}, nil)
	return &ContainerHost{client: mockClient}, mockClient
}

// TestExtractTag verifies tag extraction from image names
func TestExtractTag(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		expected  string
	}{
		{
			name:      "image with version tag",
			imageName: "ghcr.io/ready-to-release/eac-ext:0.0.2",
			expected:  "0.0.2",
		},
		{
			name:      "image with latest tag",
			imageName: "myregistry/myimage:latest",
			expected:  "latest",
		},
		{
			name:      "image with main tag",
			imageName: "ghcr.io/org/image:main",
			expected:  "main",
		},
		{
			name:      "image without tag (defaults to latest)",
			imageName: "ghcr.io/org/image",
			expected:  "latest",
		},
		{
			name:      "image with dev tag",
			imageName: "localhost:5000/myimage:dev-59-abc123",
			expected:  "dev-59-abc123",
		},
		{
			name:      "image with colon in registry",
			imageName: "localhost:5000/image:v1.0.0",
			expected:  "v1.0.0",
		},
		{
			name:      "simple image name",
			imageName: "alpine:3.18",
			expected:  "3.18",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := &ContainerHost{}
			result := ch.extractTag(tt.imageName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractExtensionName verifies extension name extraction from image names
func TestExtractExtensionName(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		expected  string
	}{
		{
			name:      "standard ghcr image with tag",
			imageName: "ghcr.io/ready-to-release/eac-ext:0.0.2",
			expected:  "eac-ext",
		},
		{
			name:      "image without tag",
			imageName: "ghcr.io/ready-to-release/eac-ext",
			expected:  "eac-ext",
		},
		{
			name:      "simple image name with tag",
			imageName: "myimage:latest",
			expected:  "myimage",
		},
		{
			name:      "nested path with tag",
			imageName: "registry.example.com/org/project/submodule:v1.0.0",
			expected:  "submodule",
		},
		{
			name:      "single component",
			imageName: "alpine",
			expected:  "alpine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := &ContainerHost{}
			result := ch.extractExtensionName(tt.imageName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCreateGitHubAuthConfig_EnvVars verifies auth from environment variables
func TestCreateGitHubAuthConfig_EnvVars(t *testing.T) {
	// Save original env
	originalUsername := os.Getenv("GITHUB_USERNAME")
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		os.Setenv("GITHUB_USERNAME", originalUsername)
		os.Setenv("GITHUB_TOKEN", originalToken)
	}()

	// Arrange
	os.Setenv("GITHUB_USERNAME", "testuser")
	os.Setenv("GITHUB_TOKEN", "test-token-123")

	// Act
	authConfig, authStr, err := CreateGitHubAuthConfig()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, authConfig)
	assert.Equal(t, "testuser", authConfig.Username)
	assert.Equal(t, "test-token-123", authConfig.Password)
	assert.Equal(t, "ghcr.io", authConfig.ServerAddress)
	assert.NotEmpty(t, authStr, "Auth string should be base64 encoded")
}

// TestCreateGitHubAuthConfig_DefaultUsername verifies default username when only token provided
func TestCreateGitHubAuthConfig_DefaultUsername(t *testing.T) {
	// Save original env
	originalUsername := os.Getenv("GITHUB_USERNAME")
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		os.Setenv("GITHUB_USERNAME", originalUsername)
		os.Setenv("GITHUB_TOKEN", originalToken)
	}()

	// Arrange - only token, no username
	os.Setenv("GITHUB_USERNAME", "")
	os.Setenv("GITHUB_TOKEN", "test-token-456")

	// Act
	authConfig, authStr, err := CreateGitHubAuthConfig()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, authConfig)
	assert.Equal(t, "github-actions", authConfig.Username, "Should use default username")
	assert.Equal(t, "test-token-456", authConfig.Password)
	assert.NotEmpty(t, authStr)
}

// TestCreateGitHubAuthConfig_NoCredentials verifies error when no credentials available
func TestCreateGitHubAuthConfig_NoCredentials(t *testing.T) {
	// Save original env
	originalUsername := os.Getenv("GITHUB_USERNAME")
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalUsername != "" {
			os.Setenv("GITHUB_USERNAME", originalUsername)
		} else {
			os.Unsetenv("GITHUB_USERNAME")
		}
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	// Arrange - no credentials via env vars
	os.Unsetenv("GITHUB_USERNAME")
	os.Unsetenv("GITHUB_TOKEN")

	// Act
	authConfig, authStr, err := CreateGitHubAuthConfig()

	// The function falls back to GitHub CLI authentication (gh auth).
	// If GitHub CLI is authenticated, we'll get credentials from there.
	// This test verifies behavior when NO authentication sources are available,
	// but we must skip if GitHub CLI provides credentials.
	if err == nil && authConfig != nil {
		t.Skip("GitHub CLI authentication is available - cannot test 'no credentials' scenario")
	}

	// Assert - only reached if no auth sources available
	assert.Error(t, err)
	assert.Nil(t, authConfig)
	assert.Empty(t, authStr)
	assert.Contains(t, err.Error(), "authentication required")
}

// TestInspectImage verifies image inspection
func TestInspectImage(t *testing.T) {
	// Arrange
	ch, mockClient := setupImageTestHost()
	mockClient.On("ImageInspect", mock.Anything, "test-image:latest").Return(
		image.InspectResponse{ID: "sha256:abc123"},
		nil,
	)

	// Act
	result, err := ch.InspectImage("test-image:latest")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "sha256:abc123", result.ID)
	mockClient.AssertExpectations(t)
}

// TestGetImageDigest_WithRepoDigests verifies digest retrieval with RepoDigests
func TestGetImageDigest_WithRepoDigests(t *testing.T) {
	// Arrange
	ch, mockClient := setupImageTestHost()
	mockClient.On("ImageInspect", mock.Anything, "test-image:v1.0.0").Return(
		image.InspectResponse{
			ID:          "sha256:image123",
			RepoDigests: []string{"ghcr.io/org/image@sha256:repo123"},
		},
		nil,
	)

	// Act
	digest, err := ch.GetImageDigest("test-image:v1.0.0")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "ghcr.io/org/image@sha256:repo123", digest, "Should prefer RepoDigests")
	mockClient.AssertExpectations(t)
}

// TestGetImageDigest_WithoutRepoDigests verifies digest fallback to image ID
func TestGetImageDigest_WithoutRepoDigests(t *testing.T) {
	// Arrange
	ch, mockClient := setupImageTestHost()
	mockClient.On("ImageInspect", mock.Anything, "local-build:dev").Return(
		image.InspectResponse{
			ID:          "sha256:local123",
			RepoDigests: []string{}, // No repo digests
		},
		nil,
	)

	// Act
	digest, err := ch.GetImageDigest("local-build:dev")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "sha256:local123", digest, "Should fall back to image ID")
	mockClient.AssertExpectations(t)
}

// TestGetImageDigest_Error verifies error handling
func TestGetImageDigest_Error(t *testing.T) {
	// Arrange
	ch, mockClient := setupImageTestHost()
	mockClient.On("ImageInspect", mock.Anything, "nonexistent:latest").Return(
		image.InspectResponse{},
		assert.AnError,
	)

	// Act
	digest, err := ch.GetImageDigest("nonexistent:latest")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, digest)
	mockClient.AssertExpectations(t)
}
