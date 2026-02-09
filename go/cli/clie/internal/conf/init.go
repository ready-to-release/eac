package conf

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
	"github.com/ready-to-release/eac/go/cli/clie/internal/envconsts"
)

// findConfigFile finds the config file by first locating the repository root
// and then looking for configuration files in priority order.
func findConfigFile(fileName string) (string, error) {
	// First find the repository root
	repoRoot, err := FindRepositoryRoot()
	if err != nil {
		return "", err // Repository error already wrapped
	}

	// Config files are located in .clie directory
	clieDir := filepath.Join(repoRoot, ".clie")

	// If looking for clie.yml, use priority-based discovery for user-specific configs
	if fileName == "clie.yml" {
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

		return "", NewConfigFileNotFoundError(".clie/clie.yml", repoRoot)
	}

	// For other filenames, look in .clie directory
	configFilePath := filepath.Join(clieDir, fileName)
	if _, err := os.Stat(configFilePath); err == nil {
		return configFilePath, nil
	} else if os.IsPermission(err) {
		return "", NewConfigFilePermissionError(configFilePath, err)
	}

	return "", NewConfigFileNotFoundError(fileName, clieDir)
}

// getConfigFileCandidates returns configuration file paths in priority order
// Priority: CLIE_CONFIG_PATH env var first, then user-specific files, then repository default
// All config files are located in .clie directory.
func getConfigFileCandidates(repoRoot string) []string {
	candidates := []string{}
	clieDir := filepath.Join(repoRoot, ".clie")

	// 0. CLIE_CONFIG_PATH environment variable (highest priority)
	if configPath := os.Getenv(envconsts.EnvCLIEConfigPath); configPath != "" {
		// Support both absolute and relative paths
		if filepath.IsAbs(configPath) {
			candidates = append(candidates, configPath)
		} else {
			candidates = append(candidates, filepath.Join(repoRoot, configPath))
		}
	}

	// 1. User-specific configuration files (highest to lowest priority)
	userSpecificFiles := []string{
		"clie.local.yml",    // Local development overrides (highest priority)
		"clie.personal.yml", // Personal user customizations
		"clie.dev.yml",      // Development environment settings
	}

	// Add username-specific config if we can get the current user
	if currentUser, err := user.Current(); err == nil {
		userSpecificFiles = append(userSpecificFiles, fmt.Sprintf("clie.%s.yml", currentUser.Username))
	}

	// Add user-specific files to candidates (in .clie directory)
	for _, filename := range userSpecificFiles {
		candidates = append(candidates, filepath.Join(clieDir, filename))
	}

	// 2. Repository default configuration (lowest priority)
	candidates = append(candidates, filepath.Join(clieDir, "clie.yml"))

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
	if os.Getenv(envconsts.EnvCLIETesting) == "true" {
		logging.Fatal("CRITICAL: InitConfig() called in test environment. Tests must use isolated configurations.")
	}

	// Additional check for test binaries
	if strings.Contains(os.Args[0], ".test") || strings.Contains(os.Args[0], "_test") {
		logging.Fatal("CRITICAL: Production configuration access blocked in test binary. Use test-specific configuration.")
	}

	// Load base configuration file
	configFile, err := findConfigFile("clie.yml")
	if err != nil {
		logging.Fatalf("Error finding config file. Please run 'clie init' from the root of your project.: %v", err)
	}
	err = LoadConfig(configFile)
	if err != nil {
		logging.Fatalf("Error parsing config file: %v", err)
	}

	// Check for and merge local override configurations
	// Priority order (highest to lowest): clie.local.yml, clie.personal.yml, clie.dev.yml
	// All override files are in .clie directory
	repoRoot, _ := FindRepositoryRoot() //nolint:errcheck // empty string is valid fallback
	RootDir = repoRoot // Store root directory for cache operations
	if repoRoot != "" {
		clieDir := filepath.Join(repoRoot, ".clie")
		overrideFiles := []string{
			"clie.local.yml",
			"clie.personal.yml",
			"clie.dev.yml",
		}

		for _, overrideFile := range overrideFiles {
			overridePath := filepath.Join(clieDir, overrideFile)
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
	// - Set CLIE_CHECK_TAGS=true to enable tag checking on startup
	// - CI environments still validate pinned tags via ValidatePinnedExtensions()
	if os.Getenv(envconsts.EnvCLIECheckTags) == "true" {
		checkLatestTags(&Global)
	}
}
