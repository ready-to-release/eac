package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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

	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/version"
	
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
	// Add force flag without shorthand to avoid conflicts with existing -f flag
	updateCmd.Flags().BoolVarP(&force, "force", "", false, "Force update even if current version is latest")
	RootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update r2r-cli to the latest version",
	Long:  `Updates r2r-cli to the latest version from GitHub releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get latest release info
		release, err := getLatestRelease()
		if err != nil {
			logging.Errorf("Failed to get latest release info: %v", err)
			os.Exit(1)
		}

		currentVersion := strings.TrimPrefix(version.Version, "v")
		latestVersion := strings.TrimPrefix(release.TagName, "v")

		// Check if update is needed
		if !force && currentVersion == latestVersion {
			logging.Info("Already running latest version")
			return
		}

		// Find correct asset for current platform (matching installer patterns)
		var selectedAsset *struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
			ID          int    `json:"id"`
			Size        int    `json:"size"`
		}

		if runtime.GOOS == "windows" {
			// Find Windows ZIP archive (matching installer)
			for _, asset := range release.Assets {
				if strings.Contains(asset.Name, "r2r-cli-") && strings.Contains(asset.Name, "windows-amd64.zip") {
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
		} else {
			// For Unix systems, look for tar.gz files
			platformSuffix := fmt.Sprintf("%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
			for _, asset := range release.Assets {
				if strings.Contains(asset.Name, "r2r-cli-") && strings.HasSuffix(asset.Name, platformSuffix) {
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
		}

		if selectedAsset == nil {
			logging.Errorf("Error: No binary found for %s-%s", runtime.GOOS, runtime.GOARCH)
			logging.Info("Available assets:")
			for _, asset := range release.Assets {
				logging.Infof("  - %s", asset.Name)
			}
			os.Exit(1)
		}

		// Download new binary using GitHub API asset endpoint (matching installer pattern)
		logging.Infof("Downloading r2r-cli %s...", release.TagName)
		logging.Infof("Size: %.2f MB", float64(selectedAsset.Size)/1024/1024)

		apiDownloadURL := fmt.Sprintf("https://api.github.com/repos/ready-to-release/r2r-cli/releases/assets/%d", selectedAsset.ID)

		req, err := http.NewRequestWithContext(context.Background(), "GET", apiDownloadURL, nil)
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
		req.Header.Set("User-Agent", "r2r-cli-updater/1.0")

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

		// Create temp file with appropriate extension
		tempSuffix := "r2r-cli-update"
		if strings.HasSuffix(selectedAsset.Name, ".zip") {
			tempSuffix += ".zip"
		}
		tmpFile, err := os.CreateTemp("", tempSuffix)
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

		// Validate downloaded file (matching installer pattern)
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

		// Check file headers (matching installer pattern)
		fileBytes, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			logging.Errorf("Failed to read downloaded file: %v", err)
			os.Exit(1)
		}

		if len(fileBytes) < 2 {
			logging.Error("Downloaded file is too small to be valid")
			os.Exit(1)
		}

		isZipFile := strings.HasSuffix(selectedAsset.Name, ".zip")
		isTarGzFile := strings.HasSuffix(selectedAsset.Name, ".tar.gz")

		if isZipFile {
			// Check for ZIP header
			if len(fileBytes) < 2 || string(fileBytes[0:2]) != "PK" {
				logging.Error("Downloaded file is not a valid ZIP archive")
				os.Exit(1)
			}
		} else if isTarGzFile {
			// Check for gzip header (1f 8b)
			if len(fileBytes) < 2 || fileBytes[0] != 0x1f || fileBytes[1] != 0x8b {
				logging.Error("Downloaded file is not a valid gzip archive")
				os.Exit(1)
			}
		} else {
			// Check for PE header (Windows executable)
			if runtime.GOOS == "windows" && (len(fileBytes) < 2 || string(fileBytes[0:2]) != "MZ") {
				logging.Error("Downloaded file is not a valid Windows executable")
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

		var binaryPath string

		if isZipFile {
			// Extract ZIP file (matching installer pattern)
			logging.Info("Extracting ZIP archive...")

			extractDir := filepath.Join(os.TempDir(), fmt.Sprintf("r2r-cli-extract-%d", os.Getpid()))
			if err := os.MkdirAll(extractDir, 0755); err != nil {
				logging.Errorf("Failed to create extraction directory: %v", err)
				os.Exit(1)
			}
			defer os.RemoveAll(extractDir)

			// Extract ZIP
			reader, err := zip.OpenReader(tmpFile.Name())
			if err != nil {
				logging.Errorf("Failed to open ZIP file: %v", err)
				os.Exit(1)
			}
			defer reader.Close()

			var foundExe string
			for _, file := range reader.File {
				if strings.HasSuffix(file.Name, ".exe") || (!strings.Contains(file.Name, ".") && !strings.Contains(file.Name, "/")) {
					// Extract this file using secure path joining to prevent Zip Slip
					extractPath, err := secureJoin(extractDir, file.Name)
					if err != nil {
						logging.Warnf("Skipping unsafe archive entry: %v", err)
						continue
					}

					rc, err := file.Open()
					if err != nil {
						logging.Errorf("Failed to open file in ZIP: file=%s err=%v", file.Name, err)
						continue
					}

					outFile, err := os.OpenFile(extractPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
					if err != nil {
						rc.Close()
						logging.Errorf("Failed to create extracted file: path=%s err=%v", extractPath, err)
						continue
					}

					_, err = io.Copy(outFile, rc)
					outFile.Close()
					rc.Close()

					if err != nil {
						logging.Errorf("Failed to extract file: file=%s err=%v", file.Name, err)
						continue
					}

					// Look for r2r.exe or r2r-cli*.exe
					if strings.Contains(filepath.Base(file.Name), "r2r") || foundExe == "" {
						foundExe = extractPath
					}
				}
			}

			if foundExe == "" {
				logging.Error("No executable found in ZIP archive")
				os.Exit(1)
			}

			binaryPath = foundExe
			logging.Infof("Using executable: %s", filepath.Base(foundExe))
		} else if strings.HasSuffix(selectedAsset.Name, ".tar.gz") {
			// Extract tar.gz file for Unix systems
			logging.Info("Extracting tar.gz archive...")

			extractDir := filepath.Join(os.TempDir(), fmt.Sprintf("r2r-cli-extract-%d", os.Getpid()))
			if err := os.MkdirAll(extractDir, 0755); err != nil {
				logging.Errorf("Failed to create extraction directory: %v", err)
				os.Exit(1)
			}
			defer os.RemoveAll(extractDir)

			// Open tar.gz file
			file, err := os.Open(tmpFile.Name())
			if err != nil {
				logging.Errorf("Failed to open tar.gz file: %v", err)
				os.Exit(1)
			}
			defer file.Close()

			// Create gzip reader
			gzReader, err := gzip.NewReader(file)
			if err != nil {
				logging.Errorf("Failed to create gzip reader: %v", err)
				os.Exit(1)
			}
			defer gzReader.Close()

			// Create tar reader
			tarReader := tar.NewReader(gzReader)

			var foundExe string
			for {
				header, err := tarReader.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					logging.Errorf("Failed to read tar header: %v", err)
					os.Exit(1)
				}

				// Look for executable files (r2r-cli* without extension)
				if header.Typeflag == tar.TypeReg {
					name := filepath.Base(header.Name)
					// Check if this looks like the r2r-cli binary
					if strings.Contains(name, "r2r-cli") || strings.Contains(name, "r2r") {
						// Use secure path joining to prevent Zip Slip
						extractPath, err := secureJoin(extractDir, header.Name)
						if err != nil {
							logging.Warnf("Skipping unsafe archive entry: %v", err)
							continue
						}

						outFile, err := os.OpenFile(extractPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
						if err != nil {
							logging.Errorf("Failed to create extracted file: path=%s err=%v", extractPath, err)
							continue
						}

						if _, err := io.Copy(outFile, tarReader); err != nil {
							outFile.Close()
							logging.Errorf("Failed to extract file: file=%s err=%v", header.Name, err)
							continue
						}
						outFile.Close()

						// Make executable
						if err := os.Chmod(extractPath, 0755); err != nil {
							logging.Errorf("Failed to set executable permissions: path=%s err=%v", extractPath, err)
							continue
						}

						// Use the first r2r-cli binary we find
						if foundExe == "" {
							foundExe = extractPath
						}
					}
				}
			}

			if foundExe == "" {
				logging.Error("No executable found in tar.gz archive")
				os.Exit(1)
			}

			binaryPath = foundExe
			logging.Infof("Using executable: %s", filepath.Base(foundExe))
		} else {
			binaryPath = tmpFile.Name()
		}

		// Make binary executable (for Unix systems)
		if runtime.GOOS != "windows" {
			if err := os.Chmod(binaryPath, 0755); err != nil {
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

	// Set up authenticated GitHub API request (matching installer pattern)
	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		"https://api.github.com/repos/ready-to-release/r2r-cli/releases/latest",
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Add authentication header using Basic auth (matching installer pattern)
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + token))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-Github-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "r2r-cli-updater/1.0")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse response
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// secureJoin safely joins a base directory with an archive entry name,
// preventing Zip Slip (path traversal) attacks.
// Returns error if the resulting path would escape the base directory.
func secureJoin(baseDir, entryName string) (string, error) {
	// Clean the entry name to handle any path separators
	cleanName := filepath.Clean(entryName)

	// Use only the base name to prevent directory traversal
	safeName := filepath.Base(cleanName)

	// Reject if the name is empty, ".", or ".."
	if safeName == "" || safeName == "." || safeName == ".." {
		return "", fmt.Errorf("invalid archive entry name: %q", entryName)
	}

	// Construct the full path
	fullPath := filepath.Join(baseDir, safeName)

	// Verify the path is within baseDir (defense in depth)
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Ensure the resolved path starts with the base directory
	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) && absPath != absBase {
		return "", fmt.Errorf("path traversal detected: %q resolves outside target directory", entryName)
	}

	return fullPath, nil
}

// copyFile copies a file from src to dst (helper function)
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
