package github

import (
	"fmt"
	"os"
	"testing"
)

func TestNewRegistryClient(t *testing.T) {
	// Save original env vars
	origToken := os.Getenv("GITHUB_TOKEN")
	origUsername := os.Getenv("GITHUB_USERNAME")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_USERNAME", origUsername)
	}()

	tests := []struct {
		name     string
		token    string
		username string
		wantErr  bool
	}{
		{
			name:     "Valid credentials",
			token:    "test-token",
			username: "test-user",
			wantErr:  false,
		},
		{
			name:     "Missing token",
			token:    "",
			username: "test-user",
			wantErr:  true,
		},
		{
			name:     "Missing username",
			token:    "test-token",
			username: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GITHUB_TOKEN", tt.token)
			os.Setenv("GITHUB_USERNAME", tt.username)

			client, err := NewRegistryClient()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Errorf("Expected client, got nil")
				}
			}
		})
	}
}

func TestGetLatestStableTag_Patterns(t *testing.T) {
	// Test tag pattern matching (would need mocking for full test)
	testTags := []string{
		"run-123",
		"run-456",
		"run-100",
		"dev-20",
		"sha-abc123",
		"v1.0.0",
		"latest",
		"main",
	}

	// Find the highest run tag
	var highestRun string
	var highestNum int
	for _, tag := range testTags {
		var num int
		if n, _ := fmt.Sscanf(tag, "run-%d", &num); n == 1 {
			if num > highestNum {
				highestNum = num
				highestRun = tag
			}
		}
	}

	if highestRun != "run-456" {
		t.Errorf("Expected run-456, got %s", highestRun)
	}
}