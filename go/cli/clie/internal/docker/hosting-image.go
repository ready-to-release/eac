package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/ready-to-release/eac/go/cli/clie/internal/cache"
	"github.com/ready-to-release/eac/go/cli/clie/internal/github"
	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
)

// InspectImage inspects a Docker image and returns the inspection result.
func (ch *ContainerHost) InspectImage(image string) (*image.InspectResponse, error) {
	// Ensure Docker connectivity before inspecting image (lazy Ping)
	if err := ch.EnsureConnected(); err != nil {
		return nil, err
	}

	imageInspect, err := ch.client.ImageInspect(ch.ctx, image)
	if err != nil {
		return nil, fmt.Errorf("error inspecting image: %w", err)
	}
	return &imageInspect, nil
}

// GetImageDigest returns the digest of a Docker image for cache invalidation
// Returns the image ID if no repo digests are available (local images).
func (ch *ContainerHost) GetImageDigest(imageName string) (string, error) {
	inspect, err := ch.InspectImage(imageName)
	if err != nil {
		return "", err
	}

	// Prefer RepoDigests (sha256:...) for pulled images
	if len(inspect.RepoDigests) > 0 {
		return inspect.RepoDigests[0], nil
	}

	// Fall back to image ID for local images
	return inspect.ID, nil
}

// CreateGitHubAuthConfig creates authentication configuration for GitHub Container Registry
// Returns both the registry.AuthConfig and base64-encoded auth string for Docker API calls.
func CreateGitHubAuthConfig() (*registry.AuthConfig, string, error) {
	// Try multiple authentication sources in order of preference

	// 1. First try environment variables (highest priority for CI/CD)
	username := os.Getenv("GITHUB_USERNAME")
	password := os.Getenv("GITHUB_TOKEN")

	// 2. If no token in env, try GitHub CLI authentication
	if password == "" {
		logging.Debug("No GITHUB_TOKEN found, trying GitHub CLI authentication")
		auth, err := github.GetCLIAuth()
		if err == nil && auth.Token != "" {
			password = auth.Token
			if username == "" && auth.Username != "" {
				username = auth.Username
			}
			logging.Debug("Using GitHub CLI authentication for ghcr.io")
		} else if err != nil {
			logging.Debugf("GitHub CLI authentication not available: %v", err)
		}
	}

	// If username is not set but we have a token, use a default username for GitHub Container Registry
	if username == "" && password != "" {
		username = "github-actions" // Generic username that works with personal access tokens and GITHUB_TOKEN
	}

	if password == "" {
		return nil, "", fmt.Errorf("authentication required: GITHUB_TOKEN environment variable must be set or GitHub CLI must be authenticated (run 'gh auth login')")
	}

	// Create authentication config
	authConfig := &registry.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: "ghcr.io",
	}

	// Encode authentication for Docker API
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return nil, "", fmt.Errorf("error encoding auth config: %w", err)
	}
	authStr := base64.StdEncoding.EncodeToString(encodedJSON)

	return authConfig, authStr, nil
}

// EnsureImageExists checks if an image exists locally and pulls it based on the pull policy.
// Policy resolution and image pulling are delegated to focused helper functions in hosting-image-policy.go.
func (ch *ContainerHost) EnsureImageExists(imageName, pullPolicy string, loadLocal bool) error {
	if err := ch.EnsureConnected(); err != nil {
		return err
	}

	if pullPolicy == "" {
		pullPolicy = "AutoDetect"
	}

	// Resolve AutoDetect to a concrete policy
	if pullPolicy == "AutoDetect" {
		resolved, useLocal := ch.resolveAutoDetectPolicy(imageName, loadLocal)
		if useLocal {
			return nil
		}
		pullPolicy = resolved
	}

	// Handle static policies (Never, IfNotPresent)
	handled, err := ch.handleStaticPolicy(imageName, pullPolicy)
	if handled {
		return err
	}

	// Pull from registry (Always policy or IfNotPresent when image not found)
	return ch.pullImage(imageName)
}

// extractTag extracts the tag from an image name (format: registry/repo:tag).
func (ch *ContainerHost) extractTag(imageName string) string {
	tagIndex := strings.LastIndex(imageName, ":")
	if tagIndex > 0 && tagIndex < len(imageName)-1 {
		return imageName[tagIndex+1:]
	}
	return "latest" // Default tag
}

// extractExtensionName extracts the extension name from an image name
// For ghcr.io/ready-to-release/eac-ext:0.0.2 -> "eac".
func (ch *ContainerHost) extractExtensionName(imageName string) string {
	// Remove tag first
	imageWithoutTag := imageName
	if idx := strings.LastIndex(imageName, ":"); idx > 0 {
		imageWithoutTag = imageName[:idx]
	}

	// Extract last path component as extension name
	parts := strings.Split(imageWithoutTag, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// cacheImageDigest saves the image digest to the registry cache for future lookups.
func (ch *ContainerHost) cacheImageDigest(imageName, tag string, imageInfo image.InspectResponse) {
	// Get or create the digest to cache
	var digest string
	if len(imageInfo.RepoDigests) > 0 {
		digest = imageInfo.RepoDigests[0]
	} else {
		digest = imageInfo.ID
	}

	// Extract extension name from image
	extensionName := ch.extractExtensionName(imageName)

	// Load registry cache
	registryCache, err := cache.LoadRegistryCache(ch.rootDir)
	if err != nil {
		logging.Warnf("Failed to load registry cache for digest update: %v", err)
		return
	}

	// Update the digest in cache
	registryCache.SetImageDigest(extensionName, tag, digest)

	// Save cache
	if err := registryCache.SaveRegistryCache(ch.rootDir); err != nil {
		logging.Warnf("Failed to save registry cache after digest update: %v", err)
	} else {
		logging.Debugf("Cached image digest: extension=%s tag=%s digest=%s", extensionName, tag, digest)
	}
}
