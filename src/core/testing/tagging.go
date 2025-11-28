// Package testing provides core testing utilities and tag system implementation
package testing

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/contracts"
	"gopkg.in/yaml.v3"
)

// Metadata holds contract version and scope information
type Metadata struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Scope       string `yaml:"scope"`
}

// Tag represents a single test tag definition
type Tag struct {
	Tag         string `yaml:"tag"`
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Pattern     string `yaml:"pattern,omitempty"` // Optional regex pattern for validation
}

// TagType defines a category of tags
type TagType struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// SkipReason represents a valid skip reason code
type SkipReason struct {
	Code        string `yaml:"code"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// SystemDependency represents a system dependency name
type SystemDependency struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// OSPlatform represents an OS platform name
type OSPlatform struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Inferred    bool   `yaml:"inferred,omitempty"`
}

// TagContract represents the complete tagging system contract
type TagContract struct {
	Metadata           Metadata           `yaml:"metadata"`
	SystemDependencies []SystemDependency `yaml:"system_dependency_names"`
	OSPlatforms        []OSPlatform       `yaml:"os_platform_names"`
	Tags               []Tag              `yaml:"tags"`
	Types              []TagType          `yaml:"types"`
	SkipReasons        []SkipReason       `yaml:"skip_reasons"`
}

// findRepositoryRoot finds the git repository root by walking up directories
// This is a lightweight implementation that doesn't require go-git
func findRepositoryRoot(startPath string) (string, error) {
	// Check for Docker R2R mode
	if os.Getenv("DOCKER_R2R_MODE") == "true" {
		return "/var/task", nil
	}

	// Check for repository root override
	if repoRoot := os.Getenv("R2R_REPO_ROOT"); repoRoot != "" {
		return filepath.Clean(repoRoot), nil
	}

	// Use current directory if no path provided
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk up looking for .git directory
	currentPath := absPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() || info.Mode().IsRegular() {
				return currentPath, nil
			}
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return "", fmt.Errorf("not a git repository (or any parent up to mount point)")
		}
		currentPath = parentPath
	}
}

// LoadTagContract reads and parses the tag contract from the contracts directory
func LoadTagContract() (*TagContract, error) {
	// Get repository root
	repoRoot, err := findRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	contractPath := filepath.Join(repoRoot, contracts.EACConfigRelPath, "testing", "tags.yml")
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tag contract from %s: %w", contractPath, err)
	}

	var contract TagContract
	if err := yaml.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("failed to parse tag contract: %w", err)
	}

	return &contract, nil
}

// GetTagsByType returns all tags of a specific type
func (c *TagContract) GetTagsByType(tagType string) []Tag {
	var tags []Tag
	for _, tag := range c.Tags {
		if tag.Type == tagType {
			tags = append(tags, tag)
		}
	}
	return tags
}

// GetTag returns a specific tag by its tag string
func (c *TagContract) GetTag(tagString string) (*Tag, error) {
	for _, tag := range c.Tags {
		if tag.Tag == tagString {
			return &tag, nil
		}
	}
	return nil, fmt.Errorf("tag not found: %s", tagString)
}

// GetDependencyTags returns all system dependency tags
func (c *TagContract) GetDependencyTags() []Tag {
	return c.GetTagsByType("system_dependency")
}

// GetLevelTags returns all taxonomy level tags
func (c *TagContract) GetLevelTags() []Tag {
	return c.GetTagsByType("taxonomy-level")
}

// GetVerificationTags returns all verification type tags
func (c *TagContract) GetVerificationTags() []Tag {
	return c.GetTagsByType("verification")
}

// GetSafetyTags returns all safety tags
func (c *TagContract) GetSafetyTags() []Tag {
	return c.GetTagsByType("safety")
}

// ValidateTag checks if a tag is defined in the contract
func (c *TagContract) ValidateTag(tagString string) bool {
	_, err := c.GetTag(tagString)
	return err == nil
}

// GetSkipReasons returns a map of valid skip reason codes
func (c *TagContract) GetSkipReasons() map[string]SkipReason {
	reasons := make(map[string]SkipReason)
	for _, reason := range c.SkipReasons {
		reasons[reason.Code] = reason
	}
	return reasons
}

// ValidateSkipReason checks if a skip reason code is valid
func (c *TagContract) ValidateSkipReason(code string) (SkipReason, bool) {
	reasons := c.GetSkipReasons()
	reason, ok := reasons[code]
	return reason, ok
}

// BuildGodogSkipTagFilter builds a Godog tag filter expression that excludes all @skip:<reason> tags
// Returns: "~@skip:wip && ~@skip:broken && ~@skip:flaky && ..." based on skip_reasons in contract
func (c *TagContract) BuildGodogSkipTagFilter() string {
	if len(c.SkipReasons) == 0 {
		return ""
	}

	filter := ""
	for i, reason := range c.SkipReasons {
		if i > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("~@skip:%s", reason.Code)
	}
	return filter
}

// GetSkipTagsForSuite returns skip tags as a slice suitable for adding to TestSuite selectors
// Returns: []string{"@skip:wip", "@skip:broken", ...} plus "@pending"
func (c *TagContract) GetSkipTagsForSuite() []string {
	tags := make([]string, 0, len(c.SkipReasons)+1)
	for _, reason := range c.SkipReasons {
		tags = append(tags, fmt.Sprintf("@skip:%s", reason.Code))
	}
	tags = append(tags, "@pending")
	return tags
}
