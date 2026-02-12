package docker

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/ready-to-release/eac/go/cli/clie/internal/cache"
	"github.com/ready-to-release/eac/go/cli/clie/internal/conf"
	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
)

// resolveAutoDetectPolicy determines the actual pull policy when AutoDetect is configured.
// Returns the resolved policy ("Always", "IfNotPresent", or "") and whether we should return early (use local image).
func (ch *ContainerHost) resolveAutoDetectPolicy(imageName string, loadLocal bool) (resolvedPolicy string, useLocal bool) {
	localImageInfo, err := ch.client.ImageInspect(ch.ctx, imageName)
	hasLocalImage := err == nil

	if hasLocalImage {
		logging.Debugf("Local image found: image=%s repoDigests=%d id=%s", imageName, len(localImageInfo.RepoDigests), localImageInfo.ID)
	}

	tag := ch.extractTag(imageName)

	// For development: When loadLocal is true, always prefer local image
	if hasLocalImage && loadLocal {
		logging.Infof("🏠 Using local development image: %s", imageName)
		logging.Debugf("Using local development image (loadLocal: true): image=%s hasRepoDigests=%v", imageName, len(localImageInfo.RepoDigests) > 0)
		ch.cacheImageDigest(imageName, tag, localImageInfo)
		return "", true
	}

	isDynamicTag := tag == "latest" || tag == "main" || tag == "master" || tag == ""

	if hasLocalImage && isDynamicTag && tag != "" {
		// For dynamic tags with local image (not loadLocal mode), check cache TTL
		registryCache, _ := cache.LoadRegistryCache(ch.rootDir) //nolint:errcheck // nil is handled below
		cacheTTL := 300 // default 5 minutes
		if conf.Global.Registry != nil && conf.Global.Registry.CacheTTL > 0 {
			cacheTTL = conf.Global.Registry.CacheTTL
		}

		if registryCache != nil && !registryCache.IsExpired(cacheTTL) {
			logging.Debugf("Using cached image (cache TTL not expired): image=%s cacheTTL=%d", imageName, cacheTTL)
			return "", true
		}

		logging.Debugf("Auto-detected pull policy: Always (dynamic tag, cache expired): image=%s tag=%s", imageName, tag)
		return "Always", false
	}

	if hasLocalImage && !isDynamicTag {
		// For specific version tags with local image
		if loadLocal && len(localImageInfo.RepoDigests) == 0 {
			logging.Infof("🏠 Using local development image: %s", imageName)
			logging.Debugf("Using local development image (AutoDetect: versioned local build): image=%s", imageName)
			return "", true
		}
		logging.Debugf("Using cached image (AutoDetect: version tag): image=%s", imageName)
		return "", true
	}

	if isDynamicTag {
		logging.Debugf("Auto-detected pull policy: Always (dynamic tag): image=%s tag=%s", imageName, tag)
		return "Always", false
	}

	logging.Debugf("Auto-detected pull policy: IfNotPresent (version tag - cached aggressively): image=%s tag=%s", imageName, tag)
	return "IfNotPresent", false
}

// handleStaticPolicy handles "Never" and "IfNotPresent" pull policies.
// Returns true if the image was found locally and no pull is needed.
func (ch *ContainerHost) handleStaticPolicy(imageName, pullPolicy string) (handled bool, err error) {
	if pullPolicy == "Never" {
		_, err := ch.client.ImageInspect(ch.ctx, imageName)
		if err != nil {
			return true, fmt.Errorf("image pull policy is 'Never' but image '%s' not found locally", imageName)
		}
		logging.Debugf("Using local image (pull policy: Never): image=%s", imageName)
		return true, nil
	}

	if pullPolicy == "IfNotPresent" {
		_, err := ch.client.ImageInspect(ch.ctx, imageName)
		if err == nil {
			logging.Debugf("Image already exists locally: image=%s", imageName)
			return true, nil
		}
	}

	return false, nil
}

// pullImage pulls an image from the registry with authentication.
func (ch *ContainerHost) pullImage(imageName string) error {
	logging.Debugf("Pulling image from registry: image=%s", imageName)

	authConfig, authStr, err := CreateGitHubAuthConfig()
	if err != nil {
		return fmt.Errorf("error creating auth config: %w", err)
	}

	if err := ch.EnsureConnected(); err != nil {
		return err
	}

	loginResp, err := ch.client.RegistryLogin(ch.ctx, *authConfig)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "docker_engine") ||
			strings.Contains(errStr, "cannot connect to the Docker daemon") ||
			strings.Contains(errStr, "system cannot find the file specified") {
			return fmt.Errorf("docker service is not running: please start Docker Desktop or the Docker daemon and try again")
		}
		return fmt.Errorf("error logging in to registry: %w", err)
	}
	logging.Infof("Successfully logged in to registry: status=%s", loginResp.Status)

	logging.Infof("🔍 Contacting registry for %s...", imageName)
	reader, err := ch.client.ImagePull(ch.ctx, imageName, image.PullOptions{
		RegistryAuth: authStr,
	})
	if err != nil {
		return fmt.Errorf("error pulling image: %w", err)
	}
	defer reader.Close()

	if err := DisplayDockerProgress(reader); err != nil {
		return fmt.Errorf("error during image pull: %w", err)
	}

	logging.Infof("Successfully pulled image: image=%s", imageName)

	tag := ch.extractTag(imageName)
	if pulledImageInfo, err := ch.client.ImageInspect(ch.ctx, imageName); err == nil {
		ch.cacheImageDigest(imageName, tag, pulledImageInfo)
	}

	return nil
}
