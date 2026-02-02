// Package github provides GitHub authentication and registry operations.
package github

import (
	"fmt"
	"os/exec"
	"strings"
)

// CLIAuth contains authentication credentials obtained from the GitHub CLI.
type CLIAuth struct {
	Token    string
	Username string
}

// GetCLIAuth attempts to get GitHub authentication from the GitHub CLI (gh).
// It retrieves the token via `gh auth token` and optionally the username
// via `gh auth status`.
//
// Returns an error if the GitHub CLI is not installed, not authenticated,
// or returns an empty token.
//
// Example:
//
//	auth, err := github.GetCLIAuth()
//	if err != nil {
//	    // Fall back to GITHUB_TOKEN env var or prompt user
//	}
//	// Use auth.Token for API calls
func GetCLIAuth() (*CLIAuth, error) {
	// Try to get the token from gh CLI
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub CLI token: %w (is gh installed and authenticated?)", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return nil, fmt.Errorf("GitHub CLI returned empty token")
	}

	auth := &CLIAuth{
		Token: token,
	}

	// Try to get the username from gh CLI status (optional, don't fail if this doesn't work)
	auth.Username = getCLIUsername()

	return auth, nil
}

// getCLIUsername attempts to extract the username from gh auth status output.
// Returns empty string if it fails (this is non-fatal).
func getCLIUsername() string {
	cmd := exec.Command("gh", "auth", "status", "-h", "github.com")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse the output to find the username
	// Example line: "✓ Logged in to github.com account USERNAME (GH_TOKEN)"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Logged in to github.com account") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "account" && i+1 < len(parts) {
					username := parts[i+1]
					// Remove trailing markers like "(GH_TOKEN)" or "(keyring)"
					if idx := strings.Index(username, "("); idx > 0 {
						username = username[:idx]
					}
					return strings.TrimSpace(username)
				}
			}
		}
	}

	return ""
}

// IsCLIAvailable checks if the GitHub CLI is installed and accessible.
func IsCLIAvailable() bool {
	cmd := exec.Command("gh", "--version")
	return cmd.Run() == nil
}

// IsCLIAuthenticated checks if the GitHub CLI is authenticated.
func IsCLIAuthenticated() bool {
	cmd := exec.Command("gh", "auth", "status")
	return cmd.Run() == nil
}
