//go:build L0 && ov

package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidationResult_EmptyIsValid(t *testing.T) {
	result := ConfigValidationResult{
		Valid: true,
	}
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	assert.Empty(t, result.Warnings)
}

func TestConfigIssue_Fields(t *testing.T) {
	issue := ConfigIssue{
		File:    "repository.yml",
		Message: "missing required field",
		Line:    42,
	}
	assert.Equal(t, "repository.yml", issue.File)
	assert.Equal(t, "missing required field", issue.Message)
	assert.Equal(t, 42, issue.Line)
}

func TestConfigFileInfo_Fields(t *testing.T) {
	info := ConfigFileInfo{
		Path:   ".eac/repository.yml",
		Layer:  "user",
		Exists: true,
		Valid:  true,
	}
	assert.Equal(t, ".eac/repository.yml", info.Path)
	assert.Equal(t, "user", info.Layer)
	assert.True(t, info.Exists)
	assert.True(t, info.Valid)
}
