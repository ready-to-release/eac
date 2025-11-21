// Package testing provides core testing utilities and tag system implementation
package testing

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/repository"
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

// TagContract represents the complete tagging system contract
type TagContract struct {
	Metadata    Metadata     `yaml:"metadata"`
	Tags        []Tag        `yaml:"tags"`
	Types       []TagType    `yaml:"types"`
	SkipReasons []SkipReason `yaml:"skip_reasons"`
}

// LoadTagContract reads and parses the tag contract from the contracts directory
func LoadTagContract() (*TagContract, error) {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Read the contract file from contracts/testing/0.1.0/tags.yml
	contractPath := filepath.Join(repoRoot, "contracts", "testing", "0.1.0", "tags.yml")
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
