package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/cache"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/conf"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/github"
	"github.com/ready-to-release/eac/go/cli/r2r/internal/logging"
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
func (ch *ContainerHost) EnsureImageExists(imageName, pullPolicy string, loadLocal bool) error {
	// Ensure Docker connectivity before image operations (lazy Ping)
	if err := ch.EnsureConnected(); err != nil {
		return err
	}

	// Apply default if not specified
	if pullPolicy == "" {
		pullPolicy = "AutoDetect"
	}

	// Handle "AutoDetect" policy - choose based on image tag and local availability
	if pullPolicy == "AutoDetect" {
		// First check if image exists locally
		localImageInfo, err := ch.client.ImageInspect(ch.ctx, imageName)
		hasLocalImage := err == nil

		if hasLocalImage {
			logging.Debugf("Local image found: image=%s repoDigests=%d id=%s", imageName, len(localImageInfo.RepoDigests), localImageInfo.ID)
		}

		// Extract tag from image name (format: registry/repo:tag)
		tagIndex := strings.LastIndex(imageName, ":")
		tag := ""
		if tagIndex > 0 && tagIndex < len(imageName)-1 {
			tag = imageName[tagIndex+1:]
		}

		// For development: When loadLocal is true, always prefer local image
		if hasLocalImage && loadLocal {
			// loadLocal explicitly requests using the local image (e.g., for development builds)
			logging.Infof("🏠 Using local development image: %s", imageName)
			logging.Debugf("Using local development image (loadLocal: true): image=%s hasRepoDigests=%v", imageName, len(localImageInfo.RepoDigests) > 0)

			// Cache the local digest for future checks
			ch.cacheImageDigest(imageName, tag, localImageInfo)
			return nil
		} else if hasLocalImage && (tag == "latest" || tag == "main" || tag == "master") {
			// For dynamic tags with local image (not loadLocal mode), check cache TTL
			// This prevents hitting the registry on every command
			registryCache, _ := cache.LoadRegistryCache(ch.rootDir) //nolint:errcheck // nil is handled below
			cacheTTL := 300 // default 5 minutes
			if conf.Global.Registry != nil && conf.Global.Registry.CacheTTL > 0 {
				cacheTTL = conf.Global.Registry.CacheTTL
			}

			if registryCache != nil && !registryCache.IsExpired(cacheTTL) {
				// Cache is still valid, use local image
				logging.Debugf("Using cached image (cache TTL not expired): image=%s cacheTTL=%d", imageName, cacheTTL)
				return nil
			}

			// Cache expired, pull for updates
			pullPolicy = "Always"
			logging.Debugf("Auto-detected pull policy: Always (dynamic tag, cache expired): image=%s tag=%s", imageName, tag)
		} else if hasLocalImage && tag != "" && tag != "latest" && tag != "main" && tag != "master" {
			// For specific version tags, check if it's a local build first (only if loadLocal is true)
			if loadLocal && len(localImageInfo.RepoDigests) == 0 {
				// Local build with version tag
				logging.Infof("🏠 Using local development image: %s", imageName)
				logging.Debugf("Using local development image (AutoDetect: versioned local build): image=%s", imageName)
				return nil
			}
			// For remote images with version tags, use local if present
			// Version tags are immutable by convention, so we can cache aggressively
			logging.Debugf("Using cached image (AutoDetect: version tag): image=%s", imageName)
			return nil
		} else if tag == "latest" || tag == "main" || tag == "master" || tag == "" {
			// For dynamic tags without recent local image, always pull
			pullPolicy = "Always"
			logging.Debugf("Auto-detected pull policy: Always (dynamic tag): image=%s tag=%s", imageName, tag)
		} else {
			// For specific version tags, use IfNotPresent for aggressive caching
			// This includes: v1.0.0, 1.2.3, dev-59-abc123, release-2.0, etc.
			// Version tags are immutable by convention
			pullPolicy = "IfNotPresent"
			logging.Debugf("Auto-detected pull policy: IfNotPresent (version tag - cached aggressively): image=%s tag=%s", imageName, tag)
		}
	}

	// Handle "Never" policy - only use local image
	if pullPolicy == "Never" {
		_, err := ch.client.ImageInspect(ch.ctx, imageName)
		if err != nil {
			return fmt.Errorf("image pull policy is 'Never' but image '%s' not found locally", imageName)
		}
		logging.Debugf("Using local image (pull policy: Never): image=%s", imageName)
		return nil
	}

	// Handle "IfNotPresent" policy - check locally first
	if pullPolicy == "IfNotPresent" {
		_, err := ch.client.ImageInspect(ch.ctx, imageName)
		if err == nil {
			// Image exists locally, no need to pull
			logging.Debugf("Image already exists locally: image=%s", imageName)
			return nil
		}
	}

	// For "Always" policy or when image not found with "IfNotPresent"
	logging.Debugf("Pulling image from registry: image=%s pullPolicy=%s", imageName, pullPolicy)

	// Get GitHub authentication using centralized function
	authConfig, authStr, err := CreateGitHubAuthConfig()
	if err != nil {
		return fmt.Errorf("error creating auth config: %w", err)
	}

	// Check if Docker daemon is running before attempting login
	_, pingErr := ch.client.Ping(ch.ctx)
	if pingErr != nil {
		// Check for common Docker service not running errors
		errStr := pingErr.Error()
		if strings.Contains(errStr, "docker_engine") ||
			strings.Contains(errStr, "cannot connect to the Docker daemon") ||
			strings.Contains(errStr, "Is the docker daemon running") ||
			strings.Contains(errStr, "system cannot find the file specified") {
			return fmt.Errorf("docker service is not running: please start Docker Desktop or the Docker daemon and try again")
		}
		return fmt.Errorf("cannot connect to Docker: %w", pingErr)
	}

	// Log in to registry
	loginResp, err := ch.client.RegistryLogin(ch.ctx, *authConfig)
	if err != nil {
		// Check if this is a Docker service issue
		errStr := err.Error()
		if strings.Contains(errStr, "docker_engine") ||
			strings.Contains(errStr, "cannot connect to the Docker daemon") ||
			strings.Contains(errStr, "system cannot find the file specified") {
			return fmt.Errorf("docker service is not running: please start Docker Desktop or the Docker daemon and try again")
		}
		return fmt.Errorf("error logging in to registry: %w", err)
	}
	logging.Infof("Successfully logged in to registry: status=%s", loginResp.Status)

	// Pull image with user feedback
	logging.Infof("🔍 Contacting registry for %s...", imageName)
	reader, err := ch.client.ImagePull(ch.ctx, imageName, image.PullOptions{
		RegistryAuth: authStr,
	})
	if err != nil {
		return fmt.Errorf("error pulling image: %w", err)
	}
	defer reader.Close()

	// Display progress to user
	if err := DisplayDockerProgress(reader); err != nil {
		return fmt.Errorf("error during image pull: %w", err)
	}

	logging.Infof("Successfully pulled image: image=%s", imageName)

	// Cache the pulled image digest
	tag := ch.extractTag(imageName)
	if pulledImageInfo, err := ch.client.ImageInspect(ch.ctx, imageName); err == nil {
		ch.cacheImageDigest(imageName, tag, pulledImageInfo)
	}

	return nil
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
// For ghcr.io/ready-to-release/ext-eac:0.0.2 -> "eac".
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
