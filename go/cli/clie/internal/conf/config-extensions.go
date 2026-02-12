// Package conf provides configuration handling for the clie CLI.
package conf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/cli/clie/internal/cache"
	"github.com/ready-to-release/eac/go/cli/clie/internal/github"
	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
	"github.com/ready-to-release/eac/go/cli/clie/internal/session"
	"github.com/ready-to-release/eac/go/cli/clie/internal/envconsts"
)

var (
	reRunTagFallback    = regexp.MustCompile(`^run-\d+$`)
	reSemverFallback    = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)
)

func detectCIEnvironment() bool {
	// Common CI environment variables
	ciVars := []string{
		"CI",                     // Generic CI indicator (GitHub Actions, GitLab CI, CircleCI, etc.)
		"GITHUB_ACTIONS",         // GitHub Actions
		"GITLAB_CI",              // GitLab CI
		"JENKINS_URL",            // Jenkins
		"TEAMCITY_VERSION",       // TeamCity
		"BUILDKITE",              // Buildkite
		"DRONE",                  // Drone
		"TRAVIS",                 // Travis CI
		"CIRCLECI",               // CircleCI
		"APPVEYOR",               // AppVeyor
		"CODEBUILD_BUILD_ID",     // AWS CodeBuild
		"BITBUCKET_BUILD_NUMBER", // Bitbucket Pipelines
		"AZURE_PIPELINES",        // Azure DevOps
		"BUILD_ID",               // Various CI systems
		"CONTINUOUS_INTEGRATION", // Generic
	}

	for _, v := range ciVars {
		if val := os.Getenv(v); val != "" && val != "false" && val != "0" {
			return true
		}
	}

	return false
}

// ValidatePinnedExtensions validates that extensions use pinned tags in CI environments.
// Returns an error if running in CI and extensions have unpinned tags.
func ValidatePinnedExtensions(cfg *Config, isCI bool) ([]string, error) {
	registryCache, cacheTTL := loadRegistryCacheForValidation(cfg)

	unpinnedExtensions := collectUnpinnedExtensions(cfg, registryCache, cacheTTL)

	saveRegistryCache(registryCache)

	// In CI, return error if any extensions are unpinned
	if isCI && len(unpinnedExtensions) > 0 {
		errMsg := fmt.Sprintf("extensions MUST be pinned in CI:\n  - %s", strings.Join(unpinnedExtensions, "\n  - "))
		if os.Getenv(envconsts.EnvCLIETesting) == "true" {
			logging.Errorf("Extensions MUST be pinned in CI:\n  - %s", strings.Join(unpinnedExtensions, "\n  - "))
		}
		return unpinnedExtensions, fmt.Errorf("%s", errMsg)
	}

	return unpinnedExtensions, nil
}

// loadRegistryCacheForValidation loads the registry cache and determines the TTL.
func loadRegistryCacheForValidation(cfg *Config) (*cache.RegistryCache, int) {
	repoRoot := RootDir
	if repoRoot == "" {
		repoRoot, _ = FindRepositoryRoot() //nolint:errcheck // fall through with empty repoRoot
	}

	registryCache, _ := cache.LoadRegistryCache(repoRoot) //nolint:errcheck // empty cache on error
	if registryCache == nil {
		registryCache, _ = cache.LoadRegistryCache(repoRoot) //nolint:errcheck // empty cache on error
	}

	cacheTTL := 300 // default 5 minutes
	if cfg.Registry != nil && cfg.Registry.CacheTTL > 0 {
		cacheTTL = cfg.Registry.CacheTTL
	}

	return registryCache, cacheTTL
}

// collectUnpinnedExtensions checks all extensions for unpinned tags and resolves
// their latest pinned versions from cache or GHCR registry.
func collectUnpinnedExtensions(cfg *Config, registryCache *cache.RegistryCache, cacheTTL int) []string {
	needsRefresh := registryCache.IsExpired(cacheTTL)
	var unpinnedExtensions []string

	for _, ext := range cfg.Extensions {
		if ext.LoadLocal {
			continue
		}
		if !hasLatestTag(ext.Image) {
			continue
		}

		baseImage, extensionName := parseExtensionImage(ext.Image)
		pinnedVersion := resolvePinnedVersion(baseImage, extensionName, registryCache, needsRefresh)

		unpinnedExtensions = append(unpinnedExtensions,
			fmt.Sprintf("'%s' must be pinned, latest is: %s", ext.Name, pinnedVersion))
	}

	return unpinnedExtensions
}

// parseExtensionImage extracts the base image and extension name from an image reference.
func parseExtensionImage(image string) (baseImage, extensionName string) {
	baseImage = image
	if idx := strings.LastIndex(baseImage, ":"); idx > 0 {
		baseImage = baseImage[:idx]
	}

	parts := strings.Split(baseImage, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if strings.HasPrefix(lastPart, "ext-") {
			extensionName = strings.TrimPrefix(lastPart, "ext-")
		} else {
			extensionName = lastPart
		}
	}

	return baseImage, extensionName
}

// resolvePinnedVersion looks up the latest pinned version from cache or fetches from GHCR.
func resolvePinnedVersion(baseImage, extensionName string, registryCache *cache.RegistryCache, needsRefresh bool) string {
	if extensionName == "" {
		return "sha-<unavailable>"
	}

	// Check cache first
	if !needsRefresh {
		if extCache, ok := registryCache.GetExtension(extensionName); ok {
			return extCache.LatestSHA
		}
	}

	// Fetch from GHCR
	logging.Debugf("Fetching latest tags from GHCR: extension=%s baseImage=%s", extensionName, baseImage)
	pinnedVersion := fetchAndCacheExtensionTags(baseImage, extensionName, registryCache)
	logging.Debugf("Fetched pin version: extension=%s pinnedVersion=%s", extensionName, pinnedVersion)

	if pinnedVersion == "" {
		return "sha-<unavailable>"
	}
	return pinnedVersion
}

// saveRegistryCache persists the registry cache to disk (best-effort).
func saveRegistryCache(registryCache *cache.RegistryCache) {
	repoRoot := RootDir
	if repoRoot == "" {
		repoRoot, _ = FindRepositoryRoot() //nolint:errcheck // fall through with empty repoRoot
	}

	logging.Debugf("Saving registry cache: extensionsInCache=%d", len(registryCache.Extensions))
	if err := registryCache.SaveRegistryCache(repoRoot); err != nil {
		logging.Errorf("Failed to save registry cache: %v", err)
	} else {
		logging.Debug("Registry cache saved successfully")
	}
}

// checkLatestTags checks for usage of "latest" Docker image tags and logs warnings.
func checkLatestTags(cfg *Config) {
	// Check if running in CI environment
	isCI := detectCIEnvironment()

	// Determine if we should suppress warnings (but still update cache)
	suppressWarnings := false
	var sessionID string

	// Check warning suppression for non-CI environments
	if !isCI {
		// Check explicit suppression
		if os.Getenv(envconsts.EnvCLIESkipPinWarning) == "true" {
			suppressWarnings = true
		} else {
			// Get session identifier (parent process ID works across platforms)
			sessionID = session.GetIdentifier()

			// Check if warnings were shown recently (within last hour) for this session
			warningFile := filepath.Join(os.TempDir(), fmt.Sprintf("nncli-%s-warning-disabled", sessionID))
			if stat, err := os.Stat(warningFile); err == nil {
				// If file exists and was modified within the last hour, suppress warnings
				if time.Since(stat.ModTime()) < time.Hour {
					suppressWarnings = true
				}
			}
		}
	}
	// In CI, we NEVER suppress - always check and fail on unpinned extensions

	// Validate pinned extensions
	unpinnedExtensions, err := ValidatePinnedExtensions(cfg, isCI)
	// Handle validation result
	if err != nil {
		// In CI with unpinned extensions - fatal error (unless in test)
		if os.Getenv(envconsts.EnvCLIETesting) != "true" {
			logging.Fatalf(": %v", err)
		}
		return
	}

	// Show warnings in non-CI if not suppressed
	if !isCI && !suppressWarnings && len(unpinnedExtensions) > 0 {
		for _, msg := range unpinnedExtensions {
			logging.Warn(msg)
		}

		// Create/touch warning file to indicate warnings were shown for this session
		if sessionID == "" {
			sessionID = session.GetIdentifier()
		}
		warningFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("nncli-%s-warning-disabled", sessionID))
		if file, err := os.Create(warningFilePath); err == nil {
			file.Close()
			logging.Debugf("Created warning flag file: file=%s session=%s", warningFilePath, sessionID)
		}

		// Also set environment variable for current process tree
		os.Setenv(envconsts.EnvCLIESkipPinWarning, "true")
	}
}

// fetchAndCacheExtensionTags fetches tags from GHCR and updates cache.
func fetchAndCacheExtensionTags(baseImage, extensionName string, registryCache *cache.RegistryCache) string {
	logging.Debugf("fetchAndCacheExtensionTags called: baseImage=%s extensionName=%s", baseImage, extensionName)
	client, err := github.NewRegistryClient()
	if err != nil {
		logging.Debugf("Failed to create registry client: %v", err)
		return ""
	}

	// Get the latest stable tag
	latestTag, err := client.GetLatestStableTag(baseImage)
	if err != nil {
		logging.Debugf("Failed to get latest stable tag, trying any tag: baseImage=%s err=%v", baseImage, err)
		// Try any tag
		latestTag, err = client.GetLatestTag(baseImage)
		if err != nil {
			logging.Debugf("Failed to get any tag: baseImage=%s err=%v", baseImage, err)
			return ""
		}
	}
	logging.Debugf("Got latest tag: latestTag=%s", latestTag)

	// Get all tags for caching
	allTags, _ := client.ListTags(baseImage) //nolint:errcheck // best-effort tag retrieval

	// Update cache
	registryCache.SetExtension(extensionName, latestTag, allTags)
	logging.Debugf("Updated cache with extension data: extensionName=%s latestTag=%s tagsCount=%d cacheExtensions=%d",
		extensionName, latestTag, len(allTags), len(registryCache.Extensions))

	return latestTag
}

// getActualImageVersion tries to suggest a proper version tag.
func getActualImageVersion(image string) string {
	// Extract base image name without tag
	baseImage := image
	if idx := strings.LastIndex(image, ":"); idx > 0 {
		baseImage = baseImage[:idx]
	}

	// Extract extension name from image path (e.g., "ghcr.io/ready-to-release/eac-ext" -> "eac")
	var extensionName string
	parts := strings.Split(baseImage, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// For ext-<name> format, extract the name
		if strings.HasPrefix(lastPart, "ext-") {
			extensionName = strings.TrimPrefix(lastPart, "ext-")
		}
	}

	// Determine cache TTL (default 300 seconds = 5 minutes)
	cacheTTL := 300
	if Global.Registry != nil && Global.Registry.CacheTTL > 0 {
		cacheTTL = Global.Registry.CacheTTL
	}

	// Get repository root for cache operations
	repoRoot := RootDir
	if repoRoot == "" {
		repoRoot, _ = FindRepositoryRoot() //nolint:errcheck // empty string is valid fallback
	}

	// Try to use cached data first
	registryCache, _ := cache.LoadRegistryCache(repoRoot) //nolint:errcheck // nil is handled below
	if registryCache != nil && extensionName != "" {
		if !registryCache.IsExpired(cacheTTL) {
			if latestSHA, ok := registryCache.GetLatestSHA(extensionName); ok {
				cachePath := cache.GetRegistryCachePath(repoRoot)
				logging.Debugf("Using cached SHA tag from registry cache: extension=%s cached_sha=%s cache_path=%s cache_age_seconds=%d",
					extensionName, latestSHA, cachePath, int(time.Since(registryCache.UpdatedAt).Seconds()))
				return latestSHA
			}
		}
	}

	// Cache miss or expired, try to query GitHub registry
	client, err := github.NewRegistryClient()
	if err == nil {
		logging.Debugf("Cache miss or expired, querying GitHub Container Registry: image=%s", baseImage)

		// Try to get the latest stable tag from the registry
		latestTag, err := client.GetLatestStableTag(baseImage)
		if err == nil && latestTag != "" {
			// Update cache if we have an extension name
			if extensionName != "" && registryCache != nil {
				// Also get all tags for caching
				allTags, _ := client.ListTags(baseImage) //nolint:errcheck // best-effort tag retrieval
				registryCache.SetExtension(extensionName, latestTag, allTags)
				_ = registryCache.SaveRegistryCache(repoRoot) //nolint:errcheck // best-effort cache save
			}

			logging.Debugf("Found latest stable tag from GitHub registry: image=%s latest_tag=%s", baseImage, latestTag)
			return latestTag
		}

		// Fall back to any available tag
		anyTag, err := client.GetLatestTag(baseImage)
		if err == nil && anyTag != "" {
			// Update cache if we have an extension name
			if extensionName != "" && registryCache != nil {
				allTags, _ := client.ListTags(baseImage) //nolint:errcheck // best-effort tag retrieval
				registryCache.SetExtension(extensionName, anyTag, allTags)
				_ = registryCache.SaveRegistryCache(repoRoot) //nolint:errcheck // best-effort cache save
			}

			logging.Debugf("Found available tag from GitHub registry: image=%s tag=%s", baseImage, anyTag)
			return anyTag
		}
	}

	// Fall back to checking local Docker images if registry query fails
	cmd := exec.Command("docker", "images", baseImage, "--format", "{{.Tag}}")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		lines := strings.Split(string(output), "\n")

		// Look for run-XXX tags first (stable releases)
		runTagPattern := reRunTagFallback
		var bestRunTag string
		var bestRunNum int

		for _, line := range lines {
			tag := strings.TrimSpace(line)
			if runTagPattern.MatchString(tag) {
				var num int
				_, _ = fmt.Sscanf(tag, "run-%d", &num) //nolint:errcheck // invalid format defaults to 0
				if num > bestRunNum {
					bestRunNum = num
					bestRunTag = tag
				}
			}
		}

		if bestRunTag != "" {
			return bestRunTag
		}

		// Look for stable semantic version tags (e.g., 1.0.0, v1.2.3)
		for _, line := range lines {
			tag := strings.TrimSpace(line)
			if tag != "" && reSemverFallback.MatchString(tag) {
				return tag
			}
		}

		// Look for sha- tags (commit-based, stable)
		for _, line := range lines {
			tag := strings.TrimSpace(line)
			if tag != "" && strings.HasPrefix(tag, "sha-") {
				return tag
			}
		}
	}

	// For extension images, provide the specific GitHub packages URL if we couldn't get a tag
	if extensionName != "" {
		return fmt.Sprintf("run-XXX  # Check for stable versions at https://github.com/ready-to-release/eac/go/cli/clie/pkgs/container/clie%%2Fextensions%%2F%s", extensionName)
	}

	// This shouldn't happen for our extension images, but handle gracefully
	return "run-XXX  # Check your registry for available stable version tags"
}

// hasLatestTag checks if a Docker image reference uses an unpinned tag (latest, main, master).
func hasLatestTag(image string) bool {
	if image == "" {
		return false
	}

	// Check for explicit unpinned tags
	if strings.HasSuffix(image, ":latest") || strings.HasSuffix(image, ":main") || strings.HasSuffix(image, ":master") {
		return true
	}

	// Check for implicit latest (no tag specified)
	// Docker images have format: [registry/]namespace/name[:tag]
	// If there's no colon after the last slash, it's implicit latest
	lastSlash := strings.LastIndex(image, "/")
	colonIndex := strings.LastIndex(image, ":")

	// If no colon, or colon appears before the last slash (part of registry URL), it's implicit latest
	if colonIndex == -1 || (lastSlash != -1 && colonIndex < lastSlash) {
		return true
	}

	return false
}

// validateMemoryLimit validates Docker memory limit format (e.g., "512MB", "1GB")
