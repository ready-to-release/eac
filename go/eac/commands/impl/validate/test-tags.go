// Command: validate test-tags
// Short: Validate that all test tags are defined in the tag contract
// Long: Validates that all tags used in test files (Gherkin features) are defined in the tag contract.
// Long:
// Long: This ensures tag filtering and test selection works correctly by preventing
// Long: the use of undefined tags that would be silently ignored by godog.
// Long:
// Long: The validation:
// Long:   - Discovers all Gherkin feature files in the repository
// Long:   - Extracts all tags from features, scenarios, and examples
// Long:   - Loads the tag contract from .r2r/eac/testing-tags.yml
// Long:   - Checks that each tag is defined in the contract
// Long:   - Reports undefined tags with their file locations
// Long:
// Long: Expected Output:
// Long:   Displays undefined tags with file locations (path:line). Tags are validated against
// Long:   .r2r/eac/testing-tags.yml contract. Exit code 0 if all tags defined, 1 if undefined tags found.
// Long:
// Long: Example:
// Long:   validate test-tags
package validate

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(TestTags)
}

// package-level config for use by helper functions.
var eacConfig *config.EACConfig

// TestTags validates that all test tags are defined in the tag contract.
func TestTags() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Load central config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("Error: failed to load config: %v", err)
		return 1
	}
	eacConfig = cfg // Store for helper functions
	repoRoot := cfg.RepoRoot

	tagsConfig := cfg.TestingTags

	// Build set of valid tags from contract
	validTags := make(map[string]bool)
	for _, tag := range tagsConfig.Tags {
		validTags[tag.Tag] = true
	}

	// Discover all feature files
	specsDir := filepath.Join(repoRoot, cfg.Repository.Paths.SpecsRoot)
	var featureFiles []string

	err = filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			featureFiles = append(featureFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Errorf("Error: failed to discover feature files: %v", err)
		return 1
	}

	// Extract tags from all feature files and track undefined tags
	type TagUsage struct {
		Tag      string
		FilePath string
		LineNum  int
	}

	var undefinedTags []TagUsage
	seenUndefined := make(map[string]bool) // Track which undefined tags we've seen

	for _, featurePath := range featureFiles {
		tags, err := extractTagsFromFeature(featurePath)
		if err != nil {
			log.Errorf("Warning: failed to extract tags from %s: %v", featurePath, err)
			continue
		}

		// Check each tag
		for _, tagInfo := range tags {
			// Check if tag is directly in contract (exact match)
			if validTags[tagInfo.Tag] {
				continue
			}

			// Check if tag matches a pattern in the contract
			if isValidPatternTag(tagInfo.Tag, tagsConfig) {
				continue
			}

			// Tag is undefined
			relPath, relErr := filepath.Rel(repoRoot, featurePath)
			if relErr != nil {
				relPath = featurePath
			}
			if !seenUndefined[tagInfo.Tag] {
				undefinedTags = append(undefinedTags, TagUsage{
					Tag:      tagInfo.Tag,
					FilePath: relPath,
					LineNum:  tagInfo.LineNum,
				})
				seenUndefined[tagInfo.Tag] = true
			}
		}
	}

	// Report results
	if len(undefinedTags) == 0 {
		log.Info("✅ All test tags are defined in the tag contract")
		log.Infof("   Validated %d feature files", len(featureFiles))
		log.Infof("   Contract defines %d valid tags", len(validTags))
		return 0
	}

	// Group undefined tags by tag name
	tagsByName := make(map[string][]TagUsage)
	for _, usage := range undefinedTags {
		tagsByName[usage.Tag] = append(tagsByName[usage.Tag], usage)
	}

	// Sort tag names for consistent output
	var tagNames []string
	for tag := range tagsByName {
		tagNames = append(tagNames, tag)
	}
	sort.Strings(tagNames)

	log.Infof("❌ Found %d undefined tag(s) used in %d location(s):\n", len(tagNames), len(undefinedTags))

	for _, tag := range tagNames {
		usages := tagsByName[tag]
		log.Infof("  %s (used in %d file(s)):", tag, len(usages))

		// Show first 3 locations for each tag
		maxToShow := 3
		for i, usage := range usages {
			if i >= maxToShow {
				log.Infof("    ... and %d more location(s)", len(usages)-maxToShow)
				break
			}
			log.Infof("    - %s:%d", usage.FilePath, usage.LineNum)
		}
		log.Info("")
	}

	log.Infof("Fix: Add missing tags to %s/%s", config.EACConfigRelPath, config.TestingTagsFileName)
	return 1
}

// TagInfo holds information about a tag found in a feature file.
type TagInfo struct {
	Tag     string
	LineNum int
}

// extractTagsFromFeature extracts all tags from a Gherkin feature file.
func extractTagsFromFeature(filePath string) ([]TagInfo, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var tags []TagInfo

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if line contains tags (starts with @)
		if strings.HasPrefix(line, "@") {
			// Split by spaces to get individual tags
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "@") {
					tags = append(tags, TagInfo{
						Tag:     part,
						LineNum: i + 1, // Line numbers are 1-based
					})
				}
			}
		}
	}

	return tags, nil
}

// isValidPatternTag checks if a tag matches any pattern in the contract.
func isValidPatternTag(tag string, tagsConfig *config.TestingTagsConfig) bool {
	// Only check tags with colons (pattern tags have format @prefix:suffix)
	if !strings.Contains(tag, ":") {
		return false
	}

	parts := strings.SplitN(tag, ":", 2)
	if len(parts) != 2 {
		return false
	}

	prefix := parts[0] + ":"
	suffix := parts[1]

	// Find matching pattern in contract
	for _, contractTag := range tagsConfig.Tags {
		// Check if contract tag is a pattern (has <placeholder>)
		if !strings.Contains(contractTag.Tag, "<") || !strings.Contains(contractTag.Tag, ">") {
			continue
		}

		// Split contract tag to get prefix
		contractParts := strings.SplitN(contractTag.Tag, ":", 2)
		if len(contractParts) != 2 {
			continue
		}

		contractPrefix := contractParts[0] + ":"

		// Check if prefixes match
		if prefix == contractPrefix {
			// For @skip:<reason>, validate against skip_reasons
			if prefix == "@skip:" {
				return isValidSkipReason(suffix, tagsConfig)
			}

			// For @deps:<name>, validate against system deps, OS platforms, and providers
			if prefix == "@deps:" {
				return isValidDepsName(suffix, tagsConfig)
			}

			// For @env:<moniker>, validate against environment contracts
			if prefix == "@env:" {
				return isValidEnvMoniker(suffix)
			}

			// For @depm:<module>, validate against module contracts
			if prefix == "@depm:" {
				return isValidModuleName(suffix)
			}

			// For other patterns, accept any suffix
			return true
		}
	}

	return false
}

// isValidSkipReason checks if a skip reason is defined in the contract.
func isValidSkipReason(reason string, tagsConfig *config.TestingTagsConfig) bool {
	for _, skipReason := range tagsConfig.SkipReasons {
		if skipReason.Code == reason {
			return true
		}
	}
	return false
}

// isValidDepsName checks if a deps name is valid (system dep from system-dependencies.yml or OS platform).
func isValidDepsName(name string, tagsConfig *config.TestingTagsConfig) bool {
	// Check system dependencies from system-dependencies.yml
	if eacConfig != nil && eacConfig.SystemDependencies != nil {
		if eacConfig.SystemDependencies.Get(name) != nil {
			return true
		}
	}

	// Check hardcoded OS platforms (linux, macos, windows)
	switch name {
	case "linux", "macos", "windows":
		return true
	}

	return false
}

// isValidEnvMoniker checks if an environment moniker is defined in environment contracts.
func isValidEnvMoniker(moniker string) bool {
	// Use the already-loaded config
	if eacConfig == nil || eacConfig.Environments == nil {
		log.Errorf("Warning: environments config not loaded")
		return true
	}

	// Check if moniker exists in environment contracts
	for _, env := range eacConfig.Environments.Environments {
		if env.Moniker == moniker {
			return true
		}
	}

	return false
}

// isValidModuleName checks if a module name is defined in module contracts.
func isValidModuleName(moduleName string) bool {
	// Use the already-loaded config
	if eacConfig == nil || eacConfig.Repository == nil {
		log.Errorf("Warning: modules config not loaded")
		return true
	}

	// Check if module exists
	_, found := eacConfig.Repository.GetModule(moduleName)
	return found
}
