package conf

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/r2r/internal/logging"
)

// findConfigFile finds the config file by first locating the repository root
// and then looking for configuration files in priority order.
func findConfigFile(fileName string) (string, error) {
	// First find the repository root
	repoRoot, err := FindRepositoryRoot()
	if err != nil {
		return "", err // Repository error already wrapped
	}

	// Config files are located in .r2r directory
	r2rDir := filepath.Join(repoRoot, ".r2r")

	// If looking for r2r-cli.yml, use priority-based discovery for user-specific configs
	if fileName == "r2r-cli.yml" {
		candidates := getConfigFileCandidates(repoRoot)

		// Check each candidate file in priority order
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				logging.Debugf("Using configuration file: config=%s", candidate)
				return candidate, nil
			} else if os.IsPermission(err) {
				return "", NewConfigFilePermissionError(candidate, err)
			}
		}

		return "", NewConfigFileNotFoundError(".r2r/r2r-cli.yml", repoRoot)
	}

	// For other filenames, look in .r2r directory
	configFilePath := filepath.Join(r2rDir, fileName)
	if _, err := os.Stat(configFilePath); err == nil {
		return configFilePath, nil
	} else if os.IsPermission(err) {
		return "", NewConfigFilePermissionError(configFilePath, err)
	}

	return "", NewConfigFileNotFoundError(fileName, r2rDir)
}

// getConfigFileCandidates returns configuration file paths in priority order
// Priority: R2R_CONFIG_PATH env var first, then user-specific files, then repository default
// All config files are located in .r2r directory.
func getConfigFileCandidates(repoRoot string) []string {
	candidates := []string{}
	r2rDir := filepath.Join(repoRoot, ".r2r")

	// 0. R2R_CONFIG_PATH environment variable (highest priority)
	if configPath := os.Getenv("R2R_CONFIG_PATH"); configPath != "" {
		// Support both absolute and relative paths
		if filepath.IsAbs(configPath) {
			candidates = append(candidates, configPath)
		} else {
			candidates = append(candidates, filepath.Join(repoRoot, configPath))
		}
	}

	// 1. User-specific configuration files (highest to lowest priority)
	userSpecificFiles := []string{
		"r2r-cli.local.yml",    // Local development overrides (highest priority)
		"r2r-cli.personal.yml", // Personal user customizations
		"r2r-cli.dev.yml",      // Development environment settings
	}

	// Add username-specific config if we can get the current user
	if currentUser, err := user.Current(); err == nil {
		userSpecificFiles = append(userSpecificFiles, fmt.Sprintf("r2r-cli.%s.yml", currentUser.Username))
	}

	// Add user-specific files to candidates (in .r2r directory)
	for _, filename := range userSpecificFiles {
		candidates = append(candidates, filepath.Join(r2rDir, filename))
	}

	// 2. Repository default configuration (lowest priority)
	candidates = append(candidates, filepath.Join(r2rDir, "r2r-cli.yml"))

	return candidates
}

// FindRepositoryRoot searches up the directory tree from the current working directory
// until it finds a .git folder or reaches the root of the filesystem.
func FindRepositoryRoot() (string, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	startDir := currentDir // Keep track of starting directory for error message

	// Loop until we reach the root of the filesystem
	for {
		// Check if the .git directory exists in the current directory
		gitPath := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentDir, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break // Reached the root of the filesystem
		}
		currentDir = parentDir
	}

	return "", NewRepositoryNotFoundError(startDir)
}

// InitConfig initializes the configuration by finding and loading the config file.
func InitConfig() {
	// CRITICAL: Block configuration access in test environment
	if os.Getenv("R2R_TESTING") == "true" {
		logging.Fatal("CRITICAL: InitConfig() called in test environment. Tests must use isolated configurations.")
	}

	// Additional check for test binaries
	if strings.Contains(os.Args[0], ".test") || strings.Contains(os.Args[0], "_test") {
		logging.Fatal("CRITICAL: Production configuration access blocked in test binary. Use test-specific configuration.")
	}

	// Load base configuration file
	configFile, err := findConfigFile("r2r-cli.yml")
	if err != nil {
		logging.Fatalf("Error finding config file. Please run 'r2r init' from the root of your project.: %v", err)
	}
	err = LoadConfig(configFile)
	if err != nil {
		logging.Fatalf("Error parsing config file: %v", err)
	}

	// Check for and merge local override configurations
	// Priority order (highest to lowest): r2r-cli.local.yml, r2r-cli.personal.yml, r2r-cli.dev.yml
	// All override files are in .r2r directory
	repoRoot, _ := FindRepositoryRoot() //nolint:errcheck // empty string is valid fallback
	RootDir = repoRoot // Store root directory for cache operations
	if repoRoot != "" {
		r2rDir := filepath.Join(repoRoot, ".r2r")
		overrideFiles := []string{
			"r2r-cli.local.yml",
			"r2r-cli.personal.yml",
			"r2r-cli.dev.yml",
		}

		for _, overrideFile := range overrideFiles {
			overridePath := filepath.Join(r2rDir, overrideFile)
			if _, err := os.Stat(overridePath); err == nil {
				logging.Debugf("Loading configuration override: override=%s", overridePath)
				if err := MergeConfigFile(overridePath); err != nil {
					logging.Warnf("Failed to load override configuration: file=%s err=%v", overridePath, err)
				} else {
					logging.Debugf("Applied configuration override: file=%s", overrideFile)
				}
			}
		}
	}

	// Tag checking is now opt-in for startup performance (~2000ms savings)
	// - Set R2R_CHECK_TAGS=true to enable tag checking on startup
	// - CI environments still validate pinned tags via ValidatePinnedExtensions()
	if os.Getenv("R2R_CHECK_TAGS") == "true" {
		checkLatestTags(&Global)
	}
}
