package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyInferences_GoTestDefaultsToL1(t *testing.T) {
	inferences := []Inference{
		{
			TestTypes:   []string{"gotest"},
			IfTags:      []string{},
			ThenAddTags: []string{"@L1"},
			Description: "Go tests default to L1",
		},
	}

	tests := []TestReference{
		{
			FilePath: "service_test.go",
			Type:     "gotest",
			TestName: "TestValidateEmail",
			Tags:     []string{}, // No tags
		},
	}

	result := ApplyInferences(tests, inferences)

	assert.Len(t, result, 1)
	assert.Contains(t, result[0].Tags, "@L1")
}

func TestApplyInferences_GodogDefaultsToL2(t *testing.T) {
	inferences := []Inference{
		{
			TestTypes:   []string{"godog"},
			IfTags:      []string{},
			ThenAddTags: []string{"@L2"},
			Description: "Godog tests default to L2",
		},
	}

	tests := []TestReference{
		{
			FilePath: "feature.feature",
			Type:     "godog",
			TestName: "Test scenario",
			Tags:     []string{"@deps:go"}, // Has tag but no level
		},
	}

	result := ApplyInferences(tests, inferences)

	assert.Len(t, result, 1)
	assert.Contains(t, result[0].Tags, "@L2")
	assert.Contains(t, result[0].Tags, "@deps:go")
}

func TestApplyInferences_IVInfersL3(t *testing.T) {
	inferences := []Inference{
		{
			IfTags:      []string{"@iv"},
			ThenAddTags: []string{"@L3"},
			Description: "IV infers L3",
		},
	}

	tests := []TestReference{
		{
			Type:     "godog",
			TestName: "Install check",
			Tags:     []string{"@iv"},
		},
	}

	result := ApplyInferences(tests, inferences)

	assert.Contains(t, result[0].Tags, "@iv")
	assert.Contains(t, result[0].Tags, "@L3")
}

func TestApplyInferences_PVInfersL3(t *testing.T) {
	inferences := []Inference{
		{
			IfTags:      []string{"@pv"},
			ThenAddTags: []string{"@L3"},
			Description: "PV infers L3",
		},
	}

	tests := []TestReference{
		{
			Type:     "godog",
			TestName: "Performance test",
			Tags:     []string{"@pv"},
		},
	}

	result := ApplyInferences(tests, inferences)

	assert.Contains(t, result[0].Tags, "@pv")
	assert.Contains(t, result[0].Tags, "@L3")
}

func TestApplyInferences_PIVInfersL4(t *testing.T) {
	inferences := []Inference{
		{
			IfTags:      []string{"@piv"},
			ThenAddTags: []string{"@L4"},
			Description: "PIV infers L4",
		},
	}

	tests := []TestReference{
		{
			Type:     "godog",
			TestName: "Production smoke test",
			Tags:     []string{"@piv"},
		},
	}

	result := ApplyInferences(tests, inferences)

	assert.Contains(t, result[0].Tags, "@piv")
	assert.Contains(t, result[0].Tags, "@L4")
}

func TestApplyInferences_PPVInfersL4(t *testing.T) {
	inferences := []Inference{
		{
			IfTags:      []string{"@ppv"},
			ThenAddTags: []string{"@L4"},
			Description: "PPV infers L4",
		},
	}

	tests := []TestReference{
		{
			Type:     "godog",
			TestName: "Production performance",
			Tags:     []string{"@ppv"},
		},
	}

	result := ApplyInferences(tests, inferences)

	assert.Contains(t, result[0].Tags, "@ppv")
	assert.Contains(t, result[0].Tags, "@L4")
}

func TestApplyInferences_ExplicitLevelOverridesInference(t *testing.T) {
	inferences := []Inference{
		{
			TestTypes:   []string{"godog"},
			IfTags:      []string{},
			ThenAddTags: []string{"@L2"},
			Description: "Godog defaults to L2",
		},
		{
			IfTags:      []string{"@pv"},
			ThenAddTags: []string{"@L3"},
			Description: "PV infers L3",
		},
	}

	tests := []TestReference{
		{
			Type:     "godog",
			TestName: "Performance at L2",
			Tags:     []string{"@L2", "@pv"}, // Explicit L2
		},
	}

	result := ApplyInferences(tests, inferences)

	// Should keep explicit @L2, not add @L3
	assert.Contains(t, result[0].Tags, "@L2")
	assert.Contains(t, result[0].Tags, "@pv")
	assert.NotContains(t, result[0].Tags, "@L3")
}

func TestApplyInferences_OnlyAppliesToMatchingTestType(t *testing.T) {
	inferences := []Inference{
		{
			TestTypes:   []string{"godog"},
			IfTags:      []string{},
			ThenAddTags: []string{"@L2"},
			Description: "Godog defaults to L2",
		},
	}

	tests := []TestReference{
		{
			Type:     "gotest",
			TestName: "Go test",
			Tags:     []string{},
		},
	}

	result := ApplyInferences(tests, inferences)

	// Should NOT apply godog inference to gotest
	assert.NotContains(t, result[0].Tags, "@L2")
}

func TestDeriveOperationalVerification_NoVerificationTags(t *testing.T) {
	tags := []string{"@L1", "@deps:go"}

	result := DeriveOperationalVerification(tags)

	assert.Contains(t, result, "@ov")
}

func TestDeriveOperationalVerification_HasIV(t *testing.T) {
	tags := []string{"@L3", "@iv"}

	result := DeriveOperationalVerification(tags)

	assert.NotContains(t, result, "@ov")
	assert.Contains(t, result, "@iv")
}

func TestDeriveOperationalVerification_HasPV(t *testing.T) {
	tags := []string{"@L3", "@pv"}

	result := DeriveOperationalVerification(tags)

	assert.NotContains(t, result, "@ov")
	assert.Contains(t, result, "@pv")
}

func TestDeriveOperationalVerification_HasPIV(t *testing.T) {
	tags := []string{"@L4", "@piv"}

	result := DeriveOperationalVerification(tags)

	assert.NotContains(t, result, "@ov")
	assert.Contains(t, result, "@piv")
}

func TestDeriveOperationalVerification_HasPPV(t *testing.T) {
	tags := []string{"@L4", "@ppv"}

	result := DeriveOperationalVerification(tags)

	assert.NotContains(t, result, "@ov")
	assert.Contains(t, result, "@ppv")
}

func TestDeriveOperationalVerification_AlreadyHasOV(t *testing.T) {
	tags := []string{"@L1", "@ov"}

	result := DeriveOperationalVerification(tags)

	// Should not duplicate @ov
	count := 0
	for _, tag := range result {
		if tag == "@ov" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestHasAnyLevelTag(t *testing.T) {
	assert.True(t, hasAnyLevelTag([]string{"@L0", "@deps:go"}))
	assert.True(t, hasAnyLevelTag([]string{"@L1"}))
	assert.True(t, hasAnyLevelTag([]string{"@L2", "@ov"}))
	assert.True(t, hasAnyLevelTag([]string{"@L3"}))
	assert.True(t, hasAnyLevelTag([]string{"@L4"}))
	assert.False(t, hasAnyLevelTag([]string{"@ov", "@deps:go"}))
	assert.False(t, hasAnyLevelTag([]string{}))
}
