package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/validation/formats/gherkin"
)

// validatePath ensures the path is within the repository (prevents path traversal attacks).
func validatePath(path, repoRoot string) error {
	// Clean paths
	cleanPath := filepath.Clean(path)
	cleanRoot := filepath.Clean(repoRoot)

	// Make path absolute
	absPath := cleanPath
	if !filepath.IsAbs(cleanPath) {
		absPath = filepath.Join(cleanRoot, cleanPath)
		absPath = filepath.Clean(absPath)
	}

	absRoot := cleanRoot
	if !filepath.IsAbs(cleanRoot) {
		var err error
		absRoot, err = filepath.Abs(cleanRoot)
		if err != nil {
			return fmt.Errorf("failed to resolve repository root: %w", err)
		}
	}

	// Ensure path is within repository
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// If the relative path starts with "..", it's trying to escape the repository
	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return fmt.Errorf("path must be within repository (attempted: %s)", path)
	}

	return nil
}

// validateGherkinFile validates a single Gherkin specification file.
func validateGherkinFile(filePath, repoRoot string, checkTags bool) ([]domain.ValidationError, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Load contract and validator
	loader := ai.NewContractLoader(repoRoot, ai.TypeSpecs, paths.DefaultsVersion)

	contractData, err := loader.LoadContract()
	if err != nil {
		return nil, fmt.Errorf("failed to load contract: %w", err)
	}

	// Load tags config for advanced tag validation (only when checkTags is enabled)
	var tagsConfig *config.TestingTagsConfig
	if checkTags {
		cfg, cfgErr := config.Load(config.DefaultLoadOptions())
		if cfgErr == nil {
			tagsConfig = cfg.TestingTags
		}
		// Config load errors are ignored - validation continues without advanced tag checks
	}

	// Create validator with tags config support
	validator := gherkin.NewValidator(contractData, tagsConfig)

	// Validate content
	errors := validator.Validate(string(content), nil)

	return errors, nil
}

// validateDirectory validates all .feature files in a directory (recursive).
func validateDirectory(dirPath, repoRoot string, quiet, checkTags bool, format string) ([]*ValidationResult, error) {
	var results []*ValidationResult

	log.Debugf("Walking directory for .feature files: dir=%s", dirPath)

	// Walk directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debugf("Error accessing path: path=%s, error=%v", path, err)
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .feature files
		if !strings.HasSuffix(path, ".feature") {
			return nil
		}

		log.Debugf("Validating file: path=%s", relativePath(path, repoRoot))

		// Validate file
		errors, validateErr := validateGherkinFile(path, repoRoot, checkTags)
		if validateErr != nil {
			// Log error but continue processing other files
			log.Debugf("Failed to validate file: path=%s, error=%v", path, validateErr)
			log.Errorf("Warning: failed to validate %s: %v", path, validateErr)
			return nil
		}

		criticalCount := domain.CountCriticalErrors(errors)
		result := &ValidationResult{
			Path:   path,
			Valid:  criticalCount == 0,
			Errors: errors,
		}

		results = append(results, result)

		log.Debugf("File validation result: path=%s, valid=%v, criticalErrors=%d, totalErrors=%d", relativePath(path, repoRoot), result.Valid, criticalCount, len(errors))

		// In quiet mode, only show progress for invalid files
		// Skip progress output entirely when format is JSON to avoid corrupting JSON output
		if format != "json" && (!quiet || !result.Valid) {
			if result.Valid && len(result.Errors) == 0 {
				log.Infof("✅ %s", relativePath(path, repoRoot))
			} else if result.Valid {
				log.Infof("✅ %s (%d warning(s))", relativePath(path, repoRoot), len(result.Errors))
			} else {
				log.Infof("❌ %s", relativePath(path, repoRoot))
			}
		}

		return nil
	})
	if err != nil {
		log.Debugf("Directory walk failed: error=%v", err)
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	log.Debugf("Directory walk complete: filesProcessed=%d", len(results))

	return results, nil
}
