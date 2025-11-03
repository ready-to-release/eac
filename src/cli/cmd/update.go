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

	"github.com/ready-to-release/eac/src/cli/internal/logger"
	"github.com/ready-to-release/eac/src/cli/internal/version"
	"github.com/rs/zerolog"
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
		// Create a pretty console writer for update progress
		prettyWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "", // No timestamp for clean output
			NoColor:    os.Getenv("NO_COLOR") != "",
			FormatLevel: func(i interface{}) string {
				if level, ok := i.(string); ok {
					switch level {
					case "info":
						return "" // No level prefix for clean output
					case "warn":
						return "⚠️ "
					case "error":
						return "❌ "
					case "fatal":
						return "💀 "
					default:
						return ""
					}
				}
				return ""
			},
			FormatMessage: func(i interface{}) string {
				if msg, ok := i.(string); ok {
					// Add appropriate icons for different update steps
					switch {
					case strings.Contains(msg, "Downloading r2r-cli"):
						return "📥 " + msg
					case strings.Contains(msg, "Size:"):
						return "📦 " + msg
					case strings.Contains(msg, "Downloading..."):
						return "⏳ " + msg
					case strings.Contains(msg, "Validating"):
						return "🔍 " + msg
					case strings.Contains(msg, "validation passed"):
						return "✅ " + msg
					case strings.Contains(msg, "Extracting"):
						return "📂 " + msg
					case strings.Contains(msg, "Using executable"):
						return "⚙️ " + msg
					case strings.Contains(msg, "Installing"):
						return "🔧 " + msg
					case strings.Contains(msg, "Successfully updated"):
						return "🎉 " + msg
					case strings.Contains(msg, "Already running"):
						return "ℹ️ " + msg
					default:
						return msg
					}
				}
				return fmt.Sprintf("%s", i)
			},
			FormatFieldName: func(i interface{}) string {
				return "" // Hide all field names for clean output
			},
			FormatFieldValue: func(i interface{}) string {
				return "" // Hide all field values for clean output
			},
			FormatTimestamp: func(i interface{}) string {
				return "" // No timestamp for clean output
			},
		}

		// Create logger with pretty console output
		verboseLogger := &logger.Logger{
			Logger: zerolog.New(prettyWriter).Level(zerolog.InfoLevel),
		}
		// Get latest release info
		release, err := getLatestRelease()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get latest release info")
			os.Exit(1)
		}

		currentVersion := strings.TrimPrefix(version.Version, "v")
		latestVersion := strings.TrimPrefix(release.TagName, "v")

		// Check if update is needed
		if !force && currentVersion == latestVersion {
			verboseLogger.Info().Msg("Already running latest version")
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
			// Try to find Windows ZIP archive first (preferred, matching installer)
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
			// Try legacy naming pattern if not found
			if selectedAsset == nil {
				for _, asset := range release.Assets {
					if strings.Contains(asset.Name, "r2r-windows-amd64") && strings.HasSuffix(asset.Name, ".exe") {
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
			verboseLogger.Error().Msgf("No binary found for %s-%s", runtime.GOOS, runtime.GOARCH)
			verboseLogger.Info().Msg("Available assets:")
			for _, asset := range release.Assets {
				verboseLogger.Info().Msgf("  - %s", asset.Name)
			}
			os.Exit(1)
		}

		// Download new binary using GitHub API asset endpoint (matching installer pattern)
		verboseLogger.Info().Msgf("Downloading r2r-cli %s...", release.TagName)
		verboseLogger.Info().Msgf("Size: %.2f MB", float64(selectedAsset.Size)/1024/1024)

		apiDownloadURL := fmt.Sprintf("https://api.github.com/repos/ready-to-release/r2r-cli/releases/assets/%d", selectedAsset.ID)

		req, err := http.NewRequestWithContext(context.Background(), "GET", apiDownloadURL, nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create download request")
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
			log.Error().Err(err).Msg("Failed to download update")
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error().Msgf("Download failed with status %d", resp.StatusCode)
			os.Exit(1)
		}

		// Create temp file with appropriate extension
		tempSuffix := "r2r-cli-update"
		if strings.HasSuffix(selectedAsset.Name, ".zip") {
			tempSuffix += ".zip"
		}
		tmpFile, err := os.CreateTemp("", tempSuffix)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create temporary file")
			os.Exit(1)
		}
		defer os.Remove(tmpFile.Name())

		// Copy download to temp file
		verboseLogger.Info().Msg("Downloading...")
		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			log.Error().Err(err).Msg("Failed to write update to temp file")
			os.Exit(1)
		}

		if err := tmpFile.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close temp file")
			os.Exit(1)
		}

		// Validate downloaded file (matching installer pattern)
		verboseLogger.Info().Msg("Validating downloaded file...")
		fileInfo, err := os.Stat(tmpFile.Name())
		if err != nil {
			log.Error().Err(err).Msg("Failed to get file info")
			os.Exit(1)
		}

		if fileInfo.Size() == 0 {
			log.Error().Msg("Downloaded file is empty")
			os.Exit(1)
		}

		if fileInfo.Size() < 1000 {
			log.Error().Msgf("Downloaded file is too small (%d bytes) - likely corrupted", fileInfo.Size())
			os.Exit(1)
		}

		// Check file headers (matching installer pattern)
		fileBytes, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			log.Error().Err(err).Msg("Failed to read downloaded file")
			os.Exit(1)
		}

		if len(fileBytes) < 2 {
			log.Error().Msg("Downloaded file is too small to be valid")
			os.Exit(1)
		}

		isZipFile := strings.HasSuffix(selectedAsset.Name, ".zip")
		isTarGzFile := strings.HasSuffix(selectedAsset.Name, ".tar.gz")

		if isZipFile {
			// Check for ZIP header
			if len(fileBytes) < 2 || string(fileBytes[0:2]) != "PK" {
				log.Error().Msg("Downloaded file is not a valid ZIP archive")
				os.Exit(1)
			}
		} else if isTarGzFile {
			// Check for gzip header (1f 8b)
			if len(fileBytes) < 2 || fileBytes[0] != 0x1f || fileBytes[1] != 0x8b {
				log.Error().Msg("Downloaded file is not a valid gzip archive")
				os.Exit(1)
			}
		} else {
			// Check for PE header (Windows executable)
			if runtime.GOOS == "windows" && (len(fileBytes) < 2 || string(fileBytes[0:2]) != "MZ") {
				log.Error().Msg("Downloaded file is not a valid Windows executable")
				os.Exit(1)
			}
		}

		verboseLogger.Info().Msg("File validation passed")

		// Get path to current executable
		exePath, err := os.Executable()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get executable path")
			os.Exit(1)
		}
		exePath, err = filepath.EvalSymlinks(exePath)
		if err != nil {
			log.Error().Err(err).Msg("Failed to resolve executable path")
			os.Exit(1)
		}

		var binaryPath string

		if isZipFile {
			// Extract ZIP file (matching installer pattern)
			verboseLogger.Info().Msg("Extracting ZIP archive...")

			extractDir := filepath.Join(os.TempDir(), fmt.Sprintf("r2r-cli-extract-%d", os.Getpid()))
			if err := os.MkdirAll(extractDir, 0755); err != nil {
				log.Error().Err(err).Msg("Failed to create extraction directory")
				os.Exit(1)
			}
			defer os.RemoveAll(extractDir)

			// Extract ZIP
			reader, err := zip.OpenReader(tmpFile.Name())
			if err != nil {
				log.Error().Err(err).Msg("Failed to open ZIP file")
				os.Exit(1)
			}
			defer reader.Close()

			var foundExe string
			for _, file := range reader.File {
				if strings.HasSuffix(file.Name, ".exe") || (!strings.Contains(file.Name, ".") && !strings.Contains(file.Name, "/")) {
					// Extract this file
					extractPath := filepath.Join(extractDir, filepath.Base(file.Name))

					rc, err := file.Open()
					if err != nil {
						log.Error().Err(err).Msgf("Failed to open file in ZIP: %s", file.Name)
						continue
					}

					outFile, err := os.OpenFile(extractPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
					if err != nil {
						rc.Close()
						log.Error().Err(err).Msgf("Failed to create extracted file: %s", extractPath)
						continue
					}

					_, err = io.Copy(outFile, rc)
					outFile.Close()
					rc.Close()

					if err != nil {
						log.Error().Err(err).Msgf("Failed to extract file: %s", file.Name)
						continue
					}

					// Look for r2r.exe or r2r-cli*.exe
					if strings.Contains(filepath.Base(file.Name), "r2r") || foundExe == "" {
						foundExe = extractPath
					}
				}
			}

			if foundExe == "" {
				log.Error().Msg("No executable found in ZIP archive")
				os.Exit(1)
			}

			binaryPath = foundExe
			verboseLogger.Info().Msgf("Using executable: %s", filepath.Base(foundExe))
		} else if strings.HasSuffix(selectedAsset.Name, ".tar.gz") {
			// Extract tar.gz file for Unix systems
			verboseLogger.Info().Msg("Extracting tar.gz archive...")

			extractDir := filepath.Join(os.TempDir(), fmt.Sprintf("r2r-cli-extract-%d", os.Getpid()))
			if err := os.MkdirAll(extractDir, 0755); err != nil {
				log.Error().Err(err).Msg("Failed to create extraction directory")
				os.Exit(1)
			}
			defer os.RemoveAll(extractDir)

			// Open tar.gz file
			file, err := os.Open(tmpFile.Name())
			if err != nil {
				log.Error().Err(err).Msg("Failed to open tar.gz file")
				os.Exit(1)
			}
			defer file.Close()

			// Create gzip reader
			gzReader, err := gzip.NewReader(file)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create gzip reader")
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
					log.Error().Err(err).Msg("Failed to read tar header")
					os.Exit(1)
				}

				// Look for executable files (r2r-cli* without extension)
				if header.Typeflag == tar.TypeReg {
					name := filepath.Base(header.Name)
					// Check if this looks like the r2r-cli binary
					if strings.Contains(name, "r2r-cli") || strings.Contains(name, "r2r") {
						extractPath := filepath.Join(extractDir, name)

						outFile, err := os.OpenFile(extractPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
						if err != nil {
							log.Error().Err(err).Msgf("Failed to create extracted file: %s", extractPath)
							continue
						}

						if _, err := io.Copy(outFile, tarReader); err != nil {
							outFile.Close()
							log.Error().Err(err).Msgf("Failed to extract file: %s", header.Name)
							continue
						}
						outFile.Close()

						// Make executable
						if err := os.Chmod(extractPath, 0755); err != nil {
							log.Error().Err(err).Msgf("Failed to set executable permissions: %s", extractPath)
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
				log.Error().Msg("No executable found in tar.gz archive")
				os.Exit(1)
			}

			binaryPath = foundExe
			verboseLogger.Info().Msgf("Using executable: %s", filepath.Base(foundExe))
		} else {
			binaryPath = tmpFile.Name()
		}

		// Make binary executable (for Unix systems)
		if runtime.GOOS != "windows" {
			if err := os.Chmod(binaryPath, 0755); err != nil {
				log.Error().Err(err).Msg("Failed to make binary executable")
				os.Exit(1)
			}
		}

		// Replace current executable with new version
		verboseLogger.Info().Msgf("Installing to: %s", exePath)
		if runtime.GOOS == "windows" {
			// Windows requires special handling since files can't be renamed over existing files
			bakPath := exePath + ".bak"
			if err := os.Rename(exePath, bakPath); err != nil {
				log.Error().Err(err).Msg("Failed to rename current executable")
				os.Exit(1)
			}
			if err := copyFile(binaryPath, exePath); err != nil {
				// Try to restore backup on failure
				if restoreErr := os.Rename(bakPath, exePath); restoreErr != nil {
					log.Error().Err(restoreErr).Msg("Failed to restore backup executable")
				}
				log.Error().Err(err).Msg("Failed to copy new executable into place")
				os.Exit(1)
			}
			os.Remove(bakPath)
		} else {
			if err := copyFile(binaryPath, exePath); err != nil {
				log.Error().Err(err).Msg("Failed to replace current executable")
				os.Exit(1)
			}
		}

		verboseLogger.Info().Msgf("Successfully updated to version %s", release.TagName)
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
