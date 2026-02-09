package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
	"github.com/ready-to-release/eac/go/cli/clie/internal/version"

	"github.com/spf13/cobra"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
		ID          int    `json:"id"`
		Size        int    `json:"size"`
	} `json:"assets"`
}

var force bool

func init() {
	// Add force flag to update self command
	updateSelfCmd.Flags().BoolVarP(&force, "force", "", false, "Force update even if current version is latest")

	// Add update self as subcommand of update
	updateCmd.AddCommand(updateSelfCmd)

	// Add update to root
	RootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update commands",
	Long:  `Parent command for update operations (update design, etc). Run 'clie update --help' for available subcommands.`,
}

var updateSelfCmd = &cobra.Command{
	Use:   "self",
	Short: "Update clie to the latest version",
	Long:  `Updates clie to the latest version from GitHub releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get latest release info
		release, err := getLatestRelease()
		if err != nil {
			logging.Errorf("Failed to get latest release info: %v", err)
			os.Exit(1)
		}

		currentVersion := strings.TrimPrefix(version.Version, "v")
		// Tag format is clie/x.y.z - extract just the version
		latestVersion := strings.TrimPrefix(release.TagName, "clie/")
		latestVersion = strings.TrimPrefix(latestVersion, "v")

		// Check if update is needed
		if !force && currentVersion == latestVersion {
			logging.Info("Already running latest version")
			return
		}

		// Find correct asset for current platform (raw executable)
		var selectedAsset *struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
			ID          int    `json:"id"`
			Size        int    `json:"size"`
		}

		// Build expected asset name for current platform
		// Binary names are: clie-{os}-{arch}[.exe]
		var expectedName string
		if runtime.GOOS == "windows" {
			expectedName = fmt.Sprintf("clie-%s-%s.exe", runtime.GOOS, runtime.GOARCH)
		} else {
			expectedName = fmt.Sprintf("clie-%s-%s", runtime.GOOS, runtime.GOARCH)
		}

		for _, asset := range release.Assets {
			// Match exact name (avoid -upx variants)
			if asset.Name == expectedName {
				selectedAsset = &struct {
					Name        string `json:"name"`
					DownloadURL string `json:"browser_download_url"`
					ID          int    `json:"id"`
					Size        int    `json:"size"`
				}{
					Name:        asset.Name,
					DownloadURL: asset.DownloadURL,
					ID:          asset.ID,
					Size:        asset.Size,
				}
				break
			}
		}

		if selectedAsset == nil {
			logging.Errorf("Error: No binary found matching '%s'", expectedName)
			logging.Info("Available assets:")
			for _, asset := range release.Assets {
				logging.Infof("  - %s", asset.Name)
			}
			os.Exit(1)
		}

		// Download new binary using GitHub API asset endpoint (matching installer pattern)
		logging.Infof("Downloading clie %s...", release.TagName)
		logging.Infof("Size: %.2f MB", float64(selectedAsset.Size)/1024/1024)

		apiDownloadURL := fmt.Sprintf("https://api.github.com/repos/ready-to-release/clie/releases/assets/%d", selectedAsset.ID)

		req, err := http.NewRequestWithContext(context.Background(), "GET", apiDownloadURL, http.NoBody)
		if err != nil {
			logging.Errorf("Failed to create download request: %v", err)
			os.Exit(1)
		}

		// Add authentication headers (matching installer pattern)
		username := os.Getenv("GITHUB_USERNAME")
		token := os.Getenv("GITHUB_TOKEN")
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + token))
		req.Header.Set("Authorization", "Basic "+auth)
		req.Header.Set("Accept", "application/octet-stream")
		req.Header.Set("User-Agent", "clie-updater/1.0")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logging.Errorf("Failed to download update: %v", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logging.Errorf("Download failed with status %d", resp.StatusCode)
			os.Exit(1)
		}

		// Create temp file for download
		tmpFile, err := os.CreateTemp("", "clie-update")
		if err != nil {
			logging.Errorf("Failed to create temporary file: %v", err)
			os.Exit(1)
		}
		defer os.Remove(tmpFile.Name())

		// Copy download to temp file
		logging.Info("Downloading...")
		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			logging.Errorf("Failed to write update to temp file: %v", err)
			os.Exit(1)
		}

		if err := tmpFile.Close(); err != nil {
			logging.Errorf("Failed to close temp file: %v", err)
			os.Exit(1)
		}

		// Validate downloaded file
		logging.Info("Validating downloaded file...")
		fileInfo, err := os.Stat(tmpFile.Name())
		if err != nil {
			logging.Errorf("Failed to get file info: %v", err)
			os.Exit(1)
		}

		if fileInfo.Size() == 0 {
			logging.Error("Downloaded file is empty")
			os.Exit(1)
		}

		if fileInfo.Size() < 1000 {
			logging.Errorf("Downloaded file is too small (%d bytes) - likely corrupted", fileInfo.Size())
			os.Exit(1)
		}

		// Check executable header
		fileBytes := make([]byte, 4)
		f, err := os.Open(tmpFile.Name())
		if err != nil {
			logging.Errorf("Failed to open downloaded file: %v", err)
			os.Exit(1)
		}
		_, err = f.Read(fileBytes)
		f.Close()
		if err != nil {
			logging.Errorf("Failed to read file header: %v", err)
			os.Exit(1)
		}

		if runtime.GOOS == "windows" {
			// Check for PE header (MZ)
			if string(fileBytes[0:2]) != "MZ" {
				logging.Error("Downloaded file is not a valid Windows executable")
				os.Exit(1)
			}
		} else {
			// Check for ELF header
			if string(fileBytes[0:4]) != "\x7fELF" {
				logging.Error("Downloaded file is not a valid ELF executable")
				os.Exit(1)
			}
		}

		logging.Info("File validation passed")

		// Get path to current executable
		exePath, err := os.Executable()
		if err != nil {
			logging.Errorf("Failed to get executable path: %v", err)
			os.Exit(1)
		}
		exePath, err = filepath.EvalSymlinks(exePath)
		if err != nil {
			logging.Errorf("Failed to resolve executable path: %v", err)
			os.Exit(1)
		}

		binaryPath := tmpFile.Name()

		// Make binary executable (for Unix systems)
		if runtime.GOOS != "windows" {
			if err := os.Chmod(binaryPath, 0o755); err != nil { //nolint:gosec // G302: executable binaries require 0755
				logging.Errorf("Failed to make binary executable: %v", err)
				os.Exit(1)
			}
		}

		// Replace current executable with new version
		logging.Infof("Installing to: %s", exePath)
		if runtime.GOOS == "windows" {
			// Windows requires special handling since files can't be renamed over existing files
			bakPath := exePath + ".bak"
			if err := os.Rename(exePath, bakPath); err != nil {
				logging.Errorf("Failed to rename current executable: %v", err)
				os.Exit(1)
			}
			if err := copyFile(binaryPath, exePath); err != nil {
				// Try to restore backup on failure
				if restoreErr := os.Rename(bakPath, exePath); restoreErr != nil {
					logging.Errorf("Failed to restore backup executable: %v", restoreErr)
				}
				logging.Errorf("Failed to copy new executable into place: %v", err)
				os.Exit(1)
			}
			os.Remove(bakPath)
		} else {
			if err := copyFile(binaryPath, exePath); err != nil {
				logging.Errorf("Failed to replace current executable: %v", err)
				os.Exit(1)
			}
		}

		logging.Infof("Successfully updated to version %s", release.TagName)
	},
}

func getLatestRelease() (*Release, error) {
	// Check authentication (matching installer pattern)
	username := os.Getenv("GITHUB_USERNAME")
	token := os.Getenv("GITHUB_TOKEN")

	if username == "" || token == "" {
		return nil, fmt.Errorf("GitHub authentication required. Please set GITHUB_USERNAME and GITHUB_TOKEN environment variables")
	}

	// Query all releases and filter by clie/ tag prefix
	// This is necessary because we use --latest=false in a monorepo
	req, err := http.NewRequestWithContext(context.Background(), "GET", "https://api.github.com/repos/ready-to-release/clie/releases?per_page=50", http.NoBody)
	if err != nil {
		return nil, err
	}

	// Add authentication header using Basic auth (matching installer pattern)
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + token))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-Github-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "clie-updater/1.0")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse response as array of releases
	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	// Filter releases by clie/ tag prefix and find highest semver
	var latestRelease *Release
	var latestVersion [3]int

	for i := range releases {
		release := &releases[i]
		tag := release.TagName

		// Only consider clie/ prefixed tags
		if !strings.HasPrefix(tag, "clie/") {
			continue
		}

		// Extract version from tag (e.g., "clie/1.0.0" -> "1.0.0")
		versionStr := strings.TrimPrefix(tag, "clie/")
		versionStr = strings.TrimPrefix(versionStr, "v")

		// Parse semver
		parts := strings.Split(versionStr, ".")
		if len(parts) < 3 {
			continue
		}

		var ver [3]int
		for j := 0; j < 3 && j < len(parts); j++ {
			_, _ = fmt.Sscanf(parts[j], "%d", &ver[j]) //nolint:errcheck // invalid parts default to 0
		}

		// Compare versions (major.minor.patch)
		if latestRelease == nil ||
			ver[0] > latestVersion[0] ||
			(ver[0] == latestVersion[0] && ver[1] > latestVersion[1]) ||
			(ver[0] == latestVersion[0] && ver[1] == latestVersion[1] && ver[2] > latestVersion[2]) {
			latestRelease = release
			latestVersion = ver
		}
	}

	if latestRelease == nil {
		return nil, fmt.Errorf("no clie releases found")
	}

	return latestRelease, nil
}

// copyFile copies a file from src to dst (helper function).
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
