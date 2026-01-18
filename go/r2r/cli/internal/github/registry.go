package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
)

// RegistryClient handles GitHub Container Registry operations.
type RegistryClient struct {
	token         string
	username      string
	client        *http.Client
	authenticated bool
}

// NewRegistryClient creates a new GitHub registry client
// Supports both authenticated (private packages) and unauthenticated (public packages) access.
func NewRegistryClient() (*RegistryClient, error) {
	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")

	// Only token is required for GitHub API authentication
	// Username is optional (used for some registry operations but not API calls)
	client := &RegistryClient{
		token:    token,
		username: username,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		authenticated: token != "",
	}

	if client.authenticated {
		logging.Debugf("Registry client created with authentication: has_username=%v", username != "")
	} else {
		logging.Debug("Registry client created without authentication (public packages only)")
	}

	return client, nil
}

// Tag represents a container image tag.
type Tag struct {
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TagListResponse represents the GitHub API response for tags.
type TagListResponse struct {
	Tags []Tag `json:"tags"`
}

// ListTags lists all available tags for a given image
// For authenticated clients, uses GitHub API (works with private packages)
// For unauthenticated clients, uses OCI Registry API (public packages only).
func (c *RegistryClient) ListTags(imagePath string) ([]string, error) {
	// Try authenticated GitHub API first if we have credentials
	if c.authenticated {
		tags, err := c.listTagsViaGitHubAPI(imagePath)
		if err == nil {
			return tags, nil
		}
		logging.Debugf("GitHub API failed, falling back to OCI Registry API: %v", err)
	}

	// Fall back to OCI Registry API (works for public packages without auth)
	return c.listTagsViaOCIRegistry(imagePath)
}

// listTagsViaGitHubAPI uses the GitHub API to list tags (requires authentication, works with private packages).
func (c *RegistryClient) listTagsViaGitHubAPI(imagePath string) ([]string, error) {
	// Parse image path to get org/package
	// Example: ghcr.io/ready-to-release/ext-eac -> ready-to-release/ext-eac
	cleanPath := strings.TrimPrefix(imagePath, "ghcr.io/")
	parts := strings.Split(cleanPath, "/")

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid image path: %s (need at least org/package)", imagePath)
	}

	org := parts[0]
	// Package name (may need URL encoding for nested paths)
	packageName := strings.Join(parts[1:], "%2F")

	// GitHub API endpoint for package versions
	url := fmt.Sprintf("https://api.github.com/orgs/%s/packages/container/%s/versions", org, packageName)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read for logging
		logging.Debugf("GitHub API request failed: url=%s status=%d body=%s", url, resp.StatusCode, string(body))
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var versions []struct {
		Metadata struct {
			Container struct {
				Tags []string `json:"tags"`
			} `json:"container"`
		} `json:"metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Collect all tags
	var tags []string
	for _, version := range versions {
		tags = append(tags, version.Metadata.Container.Tags...)
	}

	return tags, nil
}

// listTagsViaOCIRegistry uses the OCI Registry API to list tags (works for public packages without auth).
func (c *RegistryClient) listTagsViaOCIRegistry(imagePath string) ([]string, error) {
	// Parse image path: ghcr.io/ready-to-release/ext-eac -> ready-to-release/ext-eac
	cleanPath := strings.TrimPrefix(imagePath, "ghcr.io/")

	// OCI Registry API endpoint for tags
	url := fmt.Sprintf("https://ghcr.io/v2/%s/tags/list", cleanPath)

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add authentication if available (helps with rate limits)
	if c.authenticated {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Handle 401 - need to get anonymous token for public repos
	if resp.StatusCode == http.StatusUnauthorized {
		return c.listTagsWithAnonymousToken(cleanPath)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read for logging
		logging.Debugf("OCI Registry API request failed: url=%s status=%d body=%s", url, resp.StatusCode, string(body))
		return nil, fmt.Errorf("OCI Registry API returned status %d", resp.StatusCode)
	}

	var tagList struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagList); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return tagList.Tags, nil
}

// listTagsWithAnonymousToken gets an anonymous token and retries the tag list request.
func (c *RegistryClient) listTagsWithAnonymousToken(imagePath string) ([]string, error) {
	// Get anonymous token from ghcr.io
	tokenURL := fmt.Sprintf("https://ghcr.io/token?scope=repository:%s:pull", imagePath)

	tokenReq, err := http.NewRequest("GET", tokenURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}

	tokenResp, err := c.client.Do(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("executing token request: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get anonymous token: status %d", tokenResp.StatusCode)
	}

	var tokenData struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	// Retry with the anonymous token
	url := fmt.Sprintf("https://ghcr.io/v2/%s/tags/list", imagePath)
	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData.Token))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read for logging
		logging.Debugf("OCI Registry API request with token failed: url=%s status=%d body=%s", url, resp.StatusCode, string(body))
		return nil, fmt.Errorf("OCI Registry API returned status %d", resp.StatusCode)
	}

	var tagList struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagList); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return tagList.Tags, nil
}

// GetLatestStableTag finds the latest stable tag (prioritizes SHA for extensions).
func (c *RegistryClient) GetLatestStableTag(imagePath string) (string, error) {
	tags, err := c.ListTags(imagePath)
	if err != nil {
		return "", err
	}

	// For extensions (ext-* packages), prefer SHA tags for pinning
	if strings.Contains(imagePath, "/ext-") {
		// Look for sha-XXX tags first (these are the most stable for pinning)
		shaPattern := regexp.MustCompile(`^sha-[a-f0-9]{7,}$`)
		for _, tag := range tags {
			if shaPattern.MatchString(tag) {
				// Return the first SHA tag found (they're immutable)
				return tag, nil
			}
		}
	}

	// Look for run-XXX tags (these are stable release tags)
	runTagPattern := regexp.MustCompile(`^run-\d+$`)
	var runTags []struct {
		tag string
		num int
	}

	for _, tag := range tags {
		if runTagPattern.MatchString(tag) {
			// Extract the number
			var num int
			_, _ = fmt.Sscanf(tag, "run-%d", &num) //nolint:errcheck // regex already matched, parse will succeed
			runTags = append(runTags, struct {
				tag string
				num int
			}{tag, num})
		}
	}

	if len(runTags) == 0 {
		// No run tags found, look for semantic version tags
		semverPattern := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-.*)?$`)
		var semverTags []string

		for _, tag := range tags {
			if semverPattern.MatchString(tag) {
				semverTags = append(semverTags, tag)
			}
		}

		if len(semverTags) > 0 {
			// Sort and return the latest
			sort.Strings(semverTags)
			return semverTags[len(semverTags)-1], nil
		}

		// No stable tags found
		return "", fmt.Errorf("no stable tags found (sha-XXX, run-XXX or semantic version)")
	}

	// Sort run tags by number and return the highest
	sort.Slice(runTags, func(i, j int) bool {
		return runTags[i].num > runTags[j].num
	})

	return runTags[0].tag, nil
}

// GetLatestTag gets the most recent tag regardless of pattern.
func (c *RegistryClient) GetLatestTag(imagePath string) (string, error) {
	// First try to get a stable tag
	tag, err := c.GetLatestStableTag(imagePath)
	if err == nil {
		return tag, nil
	}

	// Fall back to any non-latest tag
	tags, err := c.ListTags(imagePath)
	if err != nil {
		return "", err
	}

	// Filter out unwanted tags
	var candidateTags []string
	for _, tag := range tags {
		if tag != "latest" && tag != "main" && tag != "dev" && tag != "<none>" {
			candidateTags = append(candidateTags, tag)
		}
	}

	if len(candidateTags) == 0 {
		return "", fmt.Errorf("no suitable tags found")
	}

	// Prefer sha- tags over dev- tags
	for _, tag := range candidateTags {
		if strings.HasPrefix(tag, "sha-") {
			return tag, nil
		}
	}

	// Return first available tag
	return candidateTags[0], nil
}

// ExtensionInfo represents information about an available extension.
type ExtensionInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImagePath   string `json:"image_path"`
}

// ListExtensions discovers available extensions by querying the registry
// Extensions are packages with ext-* naming convention.
func (c *RegistryClient) ListExtensions() ([]ExtensionInfo, error) {
	// Query GitHub API for all container packages in the organization
	url := "https://api.github.com/orgs/ready-to-release/packages?package_type=container&per_page=100"

	req, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read for logging
		logging.Debugf("GitHub API request failed: url=%s status=%d body=%s", url, resp.StatusCode, string(body))
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var packages []struct {
		Name        string `json:"name"`
		PackageType string `json:"package_type"`
		Visibility  string `json:"visibility"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var extensions []ExtensionInfo
	extensionPrefix := "ext-"

	for _, pkg := range packages {
		// Filter for extensions only (packages with ext-* prefix)
		if strings.HasPrefix(pkg.Name, extensionPrefix) {
			// Extract extension name (e.g., ext-eac -> eac)
			extName := strings.TrimPrefix(pkg.Name, extensionPrefix)

			// Skip if it's a nested path
			if strings.Contains(extName, "/") {
				continue
			}

			extensions = append(extensions, ExtensionInfo{
				Name:        extName,
				Description: fmt.Sprintf("%s development environment", cases.Title(language.English).String(extName)),
				ImagePath:   fmt.Sprintf("ghcr.io/ready-to-release/%s", pkg.Name),
			})
			logging.Debugf("Found extension: extension=%s package=%s", extName, pkg.Name)
		}
	}

	// Sort extensions by name for consistent output
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})

	return extensions, nil
}
