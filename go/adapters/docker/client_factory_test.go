package docker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDockerClient_UsesRealClientByDefault(t *testing.T) {
	// Ensure mock flag is not set
	os.Unsetenv("R2R_MOCK_DOCKER")

	client, err := NewDockerClient()
	require.NoError(t, err)
	defer client.Close()

	// Should be RealDockerClient
	_, ok := client.(*RealDockerClient)
	assert.True(t, ok, "expected RealDockerClient when R2R_MOCK_DOCKER is not set")
}

func TestNewDockerClient_UsesMockWhenFlagSet(t *testing.T) {
	// Set mock flag
	os.Setenv("R2R_MOCK_DOCKER", "true")
	defer os.Unsetenv("R2R_MOCK_DOCKER")

	client, err := NewDockerClient()
	require.NoError(t, err)
	defer client.Close()

	// Should be SimpleMockDockerClient
	_, ok := client.(*SimpleMockDockerClient)
	assert.True(t, ok, "expected SimpleMockDockerClient when R2R_MOCK_DOCKER=true")
}

func TestShouldUseMockDockerClient(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"not set", "", false},
		{"set to true", "true", true},
		{"set to false", "false", false},
		{"set to other value", "yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("R2R_MOCK_DOCKER")
			} else {
				os.Setenv("R2R_MOCK_DOCKER", tt.envValue)
			}
			defer os.Unsetenv("R2R_MOCK_DOCKER")

			result := shouldUseMockDockerClient()
			assert.Equal(t, tt.expected, result)
		})
	}
}
