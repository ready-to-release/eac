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

	"github.com/rs/zerolog/log"
)

// RegistryClient handles GitHub Container Registry operations
type RegistryClient struct {
	token    string
	username string
	client   *http.Client
}

// NewRegistryClient creates a new GitHub registry client
func NewRegistryClient() (*RegistryClient, error) {
	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")
	
	if token == "" || username == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN and GITHUB_USERNAME environment variables are required")
	}
	
	return &RegistryClient{
		token:    token,
		username: username,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Tag represents a container image tag
type Tag struct {
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TagListResponse represents the GitHub API response for tags
type TagListResponse struct {
	Tags []Tag `json:"tags"`
}

// ListTags lists all available tags for a given image
func (c *RegistryClient) ListTags(imagePath string) ([]string, error) {
	// Parse image path to get org/repo/package
	// Example: ghcr.io/ready-to-release/r2r-cli/extensions/pwsh -> ready-to-release/r2r-cli/extensions/pwsh
	imagePath = strings.TrimPrefix(imagePath, "ghcr.io/")
	parts := strings.Split(imagePath, "/")
	
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid image path: %s", imagePath)
	}
	
	org := parts[0]
	// Package name needs URL encoding: r2r-cli/extensions/pwsh -> r2r-cli%2Fextensions%2Fpwsh
	packageName := strings.Join(parts[1:], "%2F")
	
	// GitHub API endpoint for package versions
	url := fmt.Sprintf("https://api.github.com/orgs/%s/packages/container/%s/versions", org, packageName)
	
	req, err := http.NewRequest("GET", url, nil)
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
		body, _ := io.ReadAll(resp.Body)
		log.Debug().
			Str("url", url).
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("GitHub API request failed")
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

// GetLatestStableTag finds the latest stable tag (prioritizes SHA for extensions)
func (c *RegistryClient) GetLatestStableTag(imagePath string) (string, error) {
	tags, err := c.ListTags(imagePath)
	if err != nil {
		return "", err
	}
	
	// For extensions, prefer SHA tags for pinning
	if strings.Contains(imagePath, "/extensions/") {
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
			fmt.Sscanf(tag, "run-%d", &num)
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

// GetLatestTag gets the most recent tag regardless of pattern
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

// ExtensionInfo represents information about an available extension
type ExtensionInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImagePath   string `json:"image_path"`
}

// ListExtensions discovers available extensions by querying the registry
// Extensions are packages under r2r-cli/extensions/* namespace
func (c *RegistryClient) ListExtensions() ([]ExtensionInfo, error) {
	// Query GitHub API for all container packages in the organization
	url := "https://api.github.com/orgs/ready-to-release/packages?package_type=container&per_page=100"
	
	req, err := http.NewRequest("GET", url, nil)
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
		body, _ := io.ReadAll(resp.Body)
		log.Debug().
			Str("url", url).
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("GitHub API request failed")
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var packages []struct {
		Name string `json:"name"`
		PackageType string `json:"package_type"`
		Visibility string `json:"visibility"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	var extensions []ExtensionInfo
	extensionPrefix := "r2r-cli/extensions/"
	
	for _, pkg := range packages {
		// Filter for extensions only (packages under r2r-cli/extensions/*)
		if strings.HasPrefix(pkg.Name, extensionPrefix) {
			// Extract extension name from package path
			extName := strings.TrimPrefix(pkg.Name, extensionPrefix)
			
			// Skip if it's a nested path (e.g., extensions/foo/bar)
			if strings.Contains(extName, "/") {
				continue
			}
			
			extensions = append(extensions, ExtensionInfo{
				Name:        extName,
				Description: fmt.Sprintf("%s development environment", strings.Title(extName)),
				ImagePath:   fmt.Sprintf("ghcr.io/ready-to-release/%s", pkg.Name),
			})
			log.Debug().
				Str("extension", extName).
				Str("package", pkg.Name).
				Msg("Found extension")
		}
	}
	
	// Sort extensions by name for consistent output
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})
	
	return extensions, nil
}