package conf

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// findConfigFile finds the config file by first locating the repository root
// and then looking for configuration files in priority order
func findConfigFile(fileName string) (string, error) {
	// First find the repository root
	repoRoot, err := FindRepositoryRoot()
	if err != nil {
		return "", err // Repository error already wrapped
	}

	// If looking for r2r-cli.yml, use priority-based discovery for user-specific configs
	if fileName == "r2r-cli.yml" {
		candidates := getConfigFileCandidates(repoRoot)

		// Check each candidate file in priority order
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				log.Debug().Str("config", candidate).Msg("Using configuration file")
				return candidate, nil
			} else if os.IsPermission(err) {
				return "", NewConfigFilePermissionError(candidate, err)
			}
		}

		return "", NewConfigFileNotFoundError("r2r-cli.yml", repoRoot)
	}

	// For other filenames, look for the exact file only
	configFilePath := filepath.Join(repoRoot, fileName)
	if _, err := os.Stat(configFilePath); err == nil {
		return configFilePath, nil
	} else if os.IsPermission(err) {
		return "", NewConfigFilePermissionError(configFilePath, err)
	}

	return "", NewConfigFileNotFoundError(fileName, repoRoot)
}

// getConfigFileCandidates returns configuration file paths in priority order
// Priority: R2R_CONFIG_PATH env var first, then user-specific files, then repository default
func getConfigFileCandidates(repoRoot string) []string {
	candidates := []string{}

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

	// Add user-specific files to candidates
	for _, filename := range userSpecificFiles {
		candidates = append(candidates, filepath.Join(repoRoot, filename))
	}

	// 2. Repository default configuration (lowest priority)
	candidates = append(candidates, filepath.Join(repoRoot, "r2r-cli.yml"))

	return candidates
}

// FindRepositoryRoot searches up the directory tree from the current working directory
// until it finds a .git folder or reaches the root of the filesystem
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

// InitConfig initializes the configuration by finding and loading the config file
func InitConfig() {
	// CRITICAL: Block configuration access in test environment
	if os.Getenv("R2R_TESTING") == "true" {
		log.Fatal().Msg("CRITICAL: InitConfig() called in test environment. Tests must use isolated configurations.")
	}

	// Additional check for test binaries
	if strings.Contains(os.Args[0], ".test") || strings.Contains(os.Args[0], "_test") {
		log.Fatal().Msg("CRITICAL: Production configuration access blocked in test binary. Use test-specific configuration.")
	}

	// Load base configuration file
	configFile, err := findConfigFile("r2r-cli.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("Error finding config file. Please run 'r2r init' from the root of your project.")
	}
	err = LoadConfig(configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing config file")
	}

	// Check for and merge local override configurations
	// Priority order (highest to lowest): r2r-cli.local.yml, r2r-cli.personal.yml, r2r-cli.dev.yml
	repoRoot, _ := FindRepositoryRoot()
	if repoRoot != "" {
		overrideFiles := []string{
			"r2r-cli.local.yml",
			"r2r-cli.personal.yml",
			"r2r-cli.dev.yml",
		}

		for _, overrideFile := range overrideFiles {
			overridePath := filepath.Join(repoRoot, overrideFile)
			if _, err := os.Stat(overridePath); err == nil {
				log.Debug().Str("override", overridePath).Msg("Loading configuration override")
				if err := MergeConfigFile(overridePath); err != nil {
					log.Warn().Err(err).Str("file", overridePath).Msg("Failed to load override configuration")
				} else {
					log.Info().Str("file", overrideFile).Msg("Applied configuration override")
				}
			}
		}
	}

	// Check for latest tags after all configs are merged
	// This ensures we check extensions from override files too
	checkLatestTags(&Global)
}
