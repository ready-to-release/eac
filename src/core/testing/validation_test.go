package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTags_AllValid(t *testing.T) {
	tags := []string{"@L1", "@ov", "@dep:go"}

	errors := ValidateTags(tags)

	assert.Empty(t, errors)
}

func TestValidateTags_InvalidTag(t *testing.T) {
	tags := []string{"@L1", "@invalid-tag", "@ov"}

	errors := ValidateTags(tags)

	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "@invalid-tag")
	assert.Contains(t, errors[0], "not defined")
}

func TestValidateTags_MultipleLevelTags(t *testing.T) {
	tags := []string{"@L1", "@L2", "@ov"}

	errors := ValidateTags(tags)

	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "multiple level tags")
}

func TestValidateTags_MultipleVerificationTags(t *testing.T) {
	tags := []string{"@L1", "@iv", "@ov"}

	errors := ValidateTags(tags)

	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "multiple verification tags")
}

func TestValidateTags_ValidRiskTag(t *testing.T) {
	tags := []string{"@L1", "@ov", "@risk:data-loss"}

	errors := ValidateTags(tags)

	assert.Empty(t, errors)
}

func TestValidateTags_EmptyTagList(t *testing.T) {
	tags := []string{}

	errors := ValidateTags(tags)

	assert.Empty(t, errors)
}

func TestValidateTestReference(t *testing.T) {
	test := TestReference{
		FilePath: "suite_test.go",
		Type:     "gotest",
		TestName: "TestExample",
		Tags:     []string{"@L1", "@ov"},
	}

	errors := ValidateTestReference(test)

	assert.Empty(t, errors)
}

func TestValidateTestReference_InvalidType(t *testing.T) {
	test := TestReference{
		FilePath: "test.go",
		Type:     "invalid",
		TestName: "Test",
		Tags:     []string{"@L1"},
	}

	errors := ValidateTestReference(test)

	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "invalid test type")
}

func TestValidateTestReference_InvalidTags(t *testing.T) {
	test := TestReference{
		FilePath: "test.go",
		Type:     "gotest",
		TestName: "Test",
		Tags:     []string{"@L1", "@L2"},
	}

	errors := ValidateTestReference(test)

	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "multiple level tags")
}

func TestIsValidTag_KnownTags(t *testing.T) {
	assert.True(t, IsValidTag("@L0"))
	assert.True(t, IsValidTag("@L1"))
	assert.True(t, IsValidTag("@L2"))
	assert.True(t, IsValidTag("@L3"))
	assert.True(t, IsValidTag("@L4"))
	assert.True(t, IsValidTag("@ov"))
	assert.True(t, IsValidTag("@iv"))
	assert.True(t, IsValidTag("@pv"))
	assert.True(t, IsValidTag("@piv"))
	assert.True(t, IsValidTag("@ppv"))
	assert.True(t, IsValidTag("@dep:docker"))
	assert.True(t, IsValidTag("@dep:git"))
	assert.True(t, IsValidTag("@dep:go"))
	assert.True(t, IsValidTag("@dep:claude"))
	assert.True(t, IsValidTag("@dep:az-cli"))
	assert.True(t, IsValidTag("@requires_isolation"))
}

func TestIsValidTag_RiskPattern(t *testing.T) {
	assert.True(t, IsValidTag("@risk:data-loss"))
	assert.True(t, IsValidTag("@risk:security-vuln"))
	assert.True(t, IsValidTag("@risk:123"))
}

func TestIsValidTag_Invalid(t *testing.T) {
	assert.False(t, IsValidTag("@invalid"))
	assert.False(t, IsValidTag("@L5"))
	assert.False(t, IsValidTag("@dep:unknown"))
	assert.False(t, IsValidTag("not-a-tag"))
}

func TestGetKnownTags(t *testing.T) {
	knownTags := GetKnownTags()

	assert.Contains(t, knownTags, "@L0")
	assert.Contains(t, knownTags, "@L1")
	assert.Contains(t, knownTags, "@ov")
	assert.Contains(t, knownTags, "@dep:docker")
	assert.Contains(t, knownTags, "@requires_isolation")
}
