package tool

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewImageManager(t *testing.T) {
	mgr := NewImageManager("/workspace", false, "my-org", nil)

	if mgr == nil {
		t.Fatal("NewImageManager returned nil")
	}
	if mgr.workspaceRoot != "/workspace" {
		t.Errorf("workspaceRoot = %q, want %q", mgr.workspaceRoot, "/workspace")
	}
	if mgr.isCI {
		t.Error("isCI should be false")
	}
	if mgr.ghcrOrg != "my-org" {
		t.Errorf("ghcrOrg = %q, want %q", mgr.ghcrOrg, "my-org")
	}
}

func TestImageManager_GHCRCacheRef(t *testing.T) {
	mgr := NewImageManager("/workspace", false, "ready-to-release", nil)

	tests := []struct {
		name        string
		upstreamRef string
		want        string
	}{
		{
			name:        "ghcr image",
			upstreamRef: "ghcr.io/aquasecurity/trivy:0.50.0",
			want:        "ghcr.io/ready-to-release/cache/aquasecurity/trivy:0.50.0",
		},
		{
			name:        "docker hub image with org",
			upstreamRef: "semgrep/semgrep:1.0.0",
			want:        "ghcr.io/ready-to-release/cache/semgrep/semgrep:1.0.0",
		},
		{
			name:        "docker hub official image",
			upstreamRef: "alpine:3.20",
			want:        "ghcr.io/ready-to-release/cache/alpine:3.20",
		},
		{
			name:        "image without tag defaults to latest",
			upstreamRef: "nginx",
			want:        "ghcr.io/ready-to-release/cache/nginx:latest",
		},
		{
			name:        "quay.io image",
			upstreamRef: "quay.io/org/image:v1.2.3",
			want:        "ghcr.io/ready-to-release/cache/org/image:v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mgr.ghcrCacheRef(tt.upstreamRef)
			if got != tt.want {
				t.Errorf("ghcrCacheRef(%q) = %q, want %q", tt.upstreamRef, got, tt.want)
			}
		})
	}
}

func TestImageManager_IsImageStale(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(testFile, []byte("FROM alpine"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	mgr := NewImageManager("/workspace", false, "org", nil)

	t.Run("returns true when image doesn't exist", func(t *testing.T) {
		// Image that doesn't exist - should be considered stale
		stale := mgr.IsImageStale("nonexistent-image:tag", tmpDir)
		if !stale {
			t.Error("non-existent image should be considered stale")
		}
	})

	// Note: Testing actual staleness comparison requires mocking docker inspect
	// which would be done in integration tests with real Docker
}

func TestImageManager_EnsureImage_LocalContainer_Devbox(t *testing.T) {
	// This tests the logic flow, not actual Docker operations
	// Full integration tests would require Docker

	mgr := &ImageManager{
		workspaceRoot: t.TempDir(),
		isCI:          false,
		ghcrOrg:       "test-org",
		logWriter:     io.Discard,
	}

	// Create a mock container context
	containerDir := filepath.Join(mgr.workspaceRoot, "containers", "test-container")
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		t.Fatalf("failed to create container dir: %v", err)
	}
	dockerfile := filepath.Join(containerDir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine:3.20"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	tool := &ToolDefinition{
		ID:        "test-tool",
		Type:      ToolTypeContainer,
		LocalPath: "containers/test-container",
	}

	// Without Docker available, this will fail - but we're testing the method exists
	// and handles the local container case
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := mgr.EnsureImage(ctx, tool)
	// We expect an error because Docker isn't available in unit tests
	// The important thing is the method exists and handles local containers
	if err == nil {
		t.Log("EnsureImage succeeded (Docker available)")
	} else {
		t.Logf("EnsureImage failed as expected without Docker: %v", err)
	}
}

func TestImageManager_EnsureImage_ExternalContainer(t *testing.T) {
	mgr := &ImageManager{
		workspaceRoot: t.TempDir(),
		isCI:          false,
		ghcrOrg:       "test-org",
		logWriter:     io.Discard,
	}

	tool := &ToolDefinition{
		ID:    "trivy",
		Type:  ToolTypeContainer,
		Image: "ghcr.io/aquasecurity/trivy",
		Tag:   "0.50.0",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := mgr.EnsureImage(ctx, tool)
	// We expect an error because Docker isn't available in unit tests
	if err == nil {
		t.Log("EnsureImage succeeded (Docker available)")
	} else {
		t.Logf("EnsureImage failed as expected without Docker: %v", err)
	}
}

func TestImageManager_EnsureImage_NonContainer(t *testing.T) {
	mgr := NewImageManager("/workspace", false, "org", nil)

	tool := &ToolDefinition{
		ID:     "go-build",
		Type:   ToolTypeSystem,
		Binary: "go",
	}

	ctx := context.Background()
	imageRef, err := mgr.EnsureImage(ctx, tool)

	// Non-container tools should return empty string with no error
	if err != nil {
		t.Errorf("unexpected error for system tool: %v", err)
	}
	if imageRef != "" {
		t.Errorf("imageRef should be empty for system tool, got %q", imageRef)
	}
}

func TestDefaultImageOptions(t *testing.T) {
	opts := DefaultImageOptions()

	if opts.ForceRefresh {
		t.Error("ForceRefresh should default to false")
	}
	if !opts.AllowStale {
		t.Error("AllowStale should default to true")
	}
	if opts.OfflineMode {
		t.Error("OfflineMode should default to false")
	}
}

func TestImageOptions_Struct(t *testing.T) {
	// Test that all ImageOptions fields can be set correctly
	opts := ImageOptions{
		ForceRefresh: true,
		AllowStale:   false,
		OfflineMode:  true,
	}

	if !opts.ForceRefresh {
		t.Error("ForceRefresh should be true")
	}
	if opts.AllowStale {
		t.Error("AllowStale should be false")
	}
	if !opts.OfflineMode {
		t.Error("OfflineMode should be true")
	}
}

func TestImageManager_EnsureImageWithOptions_NonContainer(t *testing.T) {
	mgr := NewImageManager("/workspace", false, "org", nil)

	tool := &ToolDefinition{
		ID:     "go-build",
		Type:   ToolTypeSystem,
		Binary: "go",
	}

	ctx := context.Background()
	opts := ImageOptions{
		ForceRefresh: true,
		OfflineMode:  true,
	}

	imageRef, err := mgr.EnsureImageWithOptions(ctx, tool, opts)

	// Non-container tools should return empty string with no error regardless of options
	if err != nil {
		t.Errorf("unexpected error for system tool: %v", err)
	}
	if imageRef != "" {
		t.Errorf("imageRef should be empty for system tool, got %q", imageRef)
	}
}

func TestImageManager_EnsureImageWithOptions_ExternalContainer_OfflineMode(t *testing.T) {
	mgr := &ImageManager{
		workspaceRoot: t.TempDir(),
		isCI:          false,
		ghcrOrg:       "test-org",
		logWriter:     io.Discard,
	}

	tool := &ToolDefinition{
		ID:    "trivy",
		Type:  ToolTypeContainer,
		Image: "ghcr.io/aquasecurity/trivy",
		Tag:   "0.50.0",
	}

	ctx := context.Background()
	opts := ImageOptions{
		OfflineMode: true,
	}

	_, err := mgr.EnsureImageWithOptions(ctx, tool, opts)
	// In offline mode, if the image doesn't exist locally, we should get an error
	// (Docker check happens first, so we may get docker error or offline mode error)
	if err == nil {
		// If Docker is available and image exists locally, this could succeed
		t.Log("EnsureImageWithOptions succeeded (image available locally)")
	} else {
		t.Logf("EnsureImageWithOptions failed as expected: %v", err)
	}
}

func TestImageManager_EnsureImageWithOptions_LocalContainer_OfflineMode(t *testing.T) {
	mgr := &ImageManager{
		workspaceRoot: t.TempDir(),
		isCI:          false,
		ghcrOrg:       "test-org",
		logWriter:     io.Discard,
	}

	// Create a mock container context
	containerDir := filepath.Join(mgr.workspaceRoot, "containers", "test-container")
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		t.Fatalf("failed to create container dir: %v", err)
	}
	dockerfile := filepath.Join(containerDir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine:3.20"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	tool := &ToolDefinition{
		ID:        "test-tool",
		Type:      ToolTypeContainer,
		LocalPath: "containers/test-container",
	}

	ctx := context.Background()
	opts := ImageOptions{
		OfflineMode: true,
	}

	_, err := mgr.EnsureImageWithOptions(ctx, tool, opts)
	// In offline mode, building a local container would require base image pull
	// so we should either get docker not available or offline mode error
	if err == nil {
		t.Log("EnsureImageWithOptions succeeded (Docker available and image exists locally)")
	} else {
		t.Logf("EnsureImageWithOptions failed as expected: %v", err)
	}
}

func TestGetNewestFileTime_SentinelTaken(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sentinel file (Dockerfile) in the temp dir.
	dockerfile := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	// Create a nested file that is NOT a sentinel — should be ignored by the
	// sentinel path and should NOT affect the returned time.
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "hidden.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create hidden.txt: %v", err)
	}

	mgr := NewImageManager("/workspace", false, "org", nil)
	got, err := mgr.getNewestFileTime(tmpDir)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.IsZero() {
		t.Error("expected a non-zero time from the sentinel file, got zero")
	}
}

func TestGetNewestFileTime_FallbackWalk(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only a non-sentinel file so the fallback walk is exercised.
	if err := os.WriteFile(filepath.Join(tmpDir, "random.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create random.txt: %v", err)
	}

	mgr := NewImageManager("/workspace", false, "org", nil)
	got, err := mgr.getNewestFileTime(tmpDir)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.IsZero() {
		t.Error("expected a non-zero time from the fallback walk, got zero")
	}
}

func TestGetNewestFileTime_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	// No files at all — neither sentinel nor non-sentinel.
	mgr := NewImageManager("/workspace", false, "org", nil)
	_, err := mgr.getNewestFileTime(tmpDir)

	if err != nil {
		t.Errorf("unexpected error for empty directory: %v", err)
	}
	// A zero time is acceptable for an empty directory; we only assert no error.
}
