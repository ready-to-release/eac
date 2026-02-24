// repository_remote.go contains RemoteConfig type and VCS URL derivation logic.
package config

import (
	"path/filepath"
	"strings"
)

// RemoteConfig holds remote VCS repository configuration.
// Only Owner is required - Type defaults to github, URLs are derived if not explicitly set.
type RemoteConfig struct {
	Type        string `yaml:"type,omitempty"`         // VCS provider: github, gitlab, azure-devops, bitbucket (default: github)
	Owner       string `yaml:"owner,omitempty"`        // Organization or username (required)
	RepoName    string `yaml:"repo,omitempty"`         // Repository name - auto-detected from git if empty
	URL         string `yaml:"url,omitempty"`          // Repository URL - derived from type/owner/repo if empty
	PagesURL    string `yaml:"pages_url,omitempty"`    // Documentation site URL - derived from type/owner/repo if empty
	RegistryURL string `yaml:"registry_url,omitempty"` // Container registry URL - derived from type/owner if empty
}

// GetURL returns the repository URL.
// If not explicitly set, derives from type + owner + repo name.
func (r RemoteConfig) GetURL(repoRoot string) string {
	if r.URL != "" {
		return r.URL
	}
	// Try to derive from type + owner + repo name
	if r.Owner != "" {
		repoName := r.GetRepoName(repoRoot)
		if repoName != "" {
			return r.deriveURL(repoName)
		}
	}
	// Fallback: auto-detect full URL from git remote
	return detectGitRemoteURL(repoRoot)
}

// GetRepoName returns the repository name.
// Uses explicit config if set, otherwise auto-detects from git remote.
func (r RemoteConfig) GetRepoName(repoRoot string) string {
	if r.RepoName != "" {
		return r.RepoName
	}
	return detectRepoName(repoRoot)
}

// GetPagesURL returns the documentation site URL.
// If not explicitly set, derives from type + owner + repo name.
func (r RemoteConfig) GetPagesURL(repoRoot string) string {
	if r.PagesURL != "" {
		return r.PagesURL
	}
	if r.Owner == "" {
		return ""
	}
	repoName := r.GetRepoName(repoRoot)
	if repoName == "" {
		return ""
	}
	return r.derivePagesURL(repoName)
}

// GetRegistryURL returns the container registry URL.
// If not explicitly set, derives from type + owner.
func (r RemoteConfig) GetRegistryURL() string {
	if r.RegistryURL != "" {
		return r.RegistryURL
	}
	if r.Owner == "" {
		return ""
	}
	return r.deriveRegistryURL()
}

// GetOwner returns the owner, preferring explicit config over URL parsing.
func (r RemoteConfig) GetOwner() string {
	if r.Owner != "" {
		return r.Owner
	}
	// Fallback: parse from URL if set
	if r.URL == "" {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(r.URL, "https://"), "http://"), "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// deriveURL constructs the repository URL from type, owner, and repo name.
func (r RemoteConfig) deriveURL(repoName string) string {
	switch r.Type {
	case "github", "":
		return "https://github.com/" + r.Owner + "/" + repoName
	case "gitlab":
		return "https://gitlab.com/" + r.Owner + "/" + repoName
	case "azure-devops":
		return "https://dev.azure.com/" + r.Owner + "/_git/" + repoName
	case "bitbucket":
		return "https://bitbucket.org/" + r.Owner + "/" + repoName
	default:
		return ""
	}
}

// derivePagesURL constructs the documentation site URL from type, owner, and repo name.
func (r RemoteConfig) derivePagesURL(repoName string) string {
	switch r.Type {
	case "github", "":
		return "https://" + r.Owner + ".github.io/" + repoName + "/"
	case "gitlab":
		return "https://" + r.Owner + ".gitlab.io/" + repoName + "/"
	default:
		return "" // Azure DevOps and Bitbucket don't have standard pages URLs
	}
}

// deriveRegistryURL constructs the container registry URL from type and owner.
func (r RemoteConfig) deriveRegistryURL() string {
	switch r.Type {
	case "github", "":
		return "ghcr.io/" + r.Owner
	case "gitlab":
		return "registry.gitlab.com/" + r.Owner
	case "azure-devops":
		return r.Owner + ".azurecr.io" // Azure Container Registry uses org as subdomain
	default:
		return ""
	}
}

// detectRepoName gets the repository name from git remote or directory name.
func detectRepoName(repoRoot string) string {
	// Try git remote via injected provider (pure go-git, no exec)
	url, err := resolveGitRemoteURL(repoRoot, "origin")
	if err == nil && url != "" {
		url = strings.TrimSuffix(url, ".git")
		if idx := strings.LastIndex(url, "/"); idx >= 0 {
			return url[idx+1:]
		}
		if idx := strings.LastIndex(url, ":"); idx >= 0 {
			return url[idx+1:]
		}
	}
	// Fallback to directory name
	return filepath.Base(repoRoot)
}

// detectGitRemoteURL attempts to get the remote URL from git.
func detectGitRemoteURL(repoRoot string) string {
	url, err := resolveGitRemoteURL(repoRoot, "origin")
	if err != nil || url == "" {
		return ""
	}
	// Convert SSH URLs to HTTPS
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.Replace(url, "git@github.com:", "https://github.com/", 1)
	}
	if strings.HasPrefix(url, "git@gitlab.com:") {
		url = strings.Replace(url, "git@gitlab.com:", "https://gitlab.com/", 1)
	}
	return strings.TrimSuffix(url, ".git")
}
