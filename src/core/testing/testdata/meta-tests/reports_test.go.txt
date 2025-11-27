package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTraceabilityMatrix(t *testing.T) {
	tests := []TestReference{
		{
			TestName:     "Test login",
			RiskControls: []string{"@risk-control:auth-mfa-01", "@risk-control:gxp-audit"},
			IsManual:     false,
		},
		{
			TestName:     "Test account lockout",
			RiskControls: []string{"@risk-control:gxp-account-lockout"},
			IsManual:     true,
		},
	}

	report := GenerateTraceabilityMatrix(tests)

	assert.Contains(t, report, "# Risk Traceability Matrix")
	assert.Contains(t, report, "| Test | Risk Control | Type |")
	assert.Contains(t, report, "Test login")
	assert.Contains(t, report, "@risk-control:auth-mfa-01")
	assert.Contains(t, report, "@risk-control:gxp-audit")
	assert.Contains(t, report, "Automated")
	assert.Contains(t, report, "Test account lockout")
	assert.Contains(t, report, "@risk-control:gxp-account-lockout")
	assert.Contains(t, report, "Manual")
}

func TestGenerateTraceabilityMatrix_NoRiskControls(t *testing.T) {
	tests := []TestReference{
		{
			TestName:     "Simple test",
			RiskControls: []string{},
		},
	}

	report := GenerateTraceabilityMatrix(tests)

	assert.Contains(t, report, "# Risk Traceability Matrix")
	assert.NotContains(t, report, "Simple test")
}

func TestGenerateGxPReport(t *testing.T) {
	tests := []TestReference{
		{
			FilePath:         "auth.feature",
			TestName:         "Login test",
			Tags:             []string{"@gxp", "@L2"},
			IsGxP:            true,
			IsCriticalAspect: false,
			IsManual:         false,
			RiskControls:     []string{"@risk-control:gxp-audit"},
		},
		{
			FilePath:         "auth.feature",
			TestName:         "Account lockout",
			Tags:             []string{"@gxp", "@critical-aspect", "@L2"},
			IsGxP:            true,
			IsCriticalAspect: true,
			IsManual:         true,
			RiskControls:     []string{"@risk-control:gxp-account-lockout"},
		},
		{
			FilePath: "general.feature",
			TestName: "Non-GxP test",
			Tags:     []string{"@L1"},
			IsGxP:    false,
		},
	}

	report := GenerateGxPReport(tests)

	assert.Contains(t, report, "# GxP Implementation Report")
	assert.Contains(t, report, "## Requirements Specifications (URS/FS)")
	assert.Contains(t, report, "Total GxP Features: 1")
	assert.Contains(t, report, "## Test Summary")
	assert.Contains(t, report, "- Total Tests: 3")
	assert.Contains(t, report, "- GxP Tests: 2")
	assert.Contains(t, report, "- Critical Aspects: 1")
	assert.Contains(t, report, "- Manual Tests: 1")
	assert.Contains(t, report, "## Risk Traceability Matrix")
	assert.Contains(t, report, "Login test")
	assert.Contains(t, report, "Account lockout")
}

func TestGenerateGxPReport_NoGxPTests(t *testing.T) {
	tests := []TestReference{
		{
			FilePath: "general.feature",
			TestName: "Regular test",
			Tags:     []string{"@L1"},
			IsGxP:    false,
		},
	}

	report := GenerateGxPReport(tests)

	assert.Contains(t, report, "# GxP Implementation Report")
	assert.Contains(t, report, "Total GxP Features: 0")
	assert.Contains(t, report, "- Total Tests: 1")
	assert.Contains(t, report, "- GxP Tests: 0")
	assert.Contains(t, report, "- Critical Aspects: 0")
	assert.Contains(t, report, "- Manual Tests: 0")
}

func TestGetGxPTests(t *testing.T) {
	tests := []TestReference{
		{TestName: "GxP test 1", IsGxP: true},
		{TestName: "Regular test", IsGxP: false},
		{TestName: "GxP test 2", IsGxP: true},
	}

	gxpTests := GetGxPTests(tests)

	assert.Len(t, gxpTests, 2)
	assert.Equal(t, "GxP test 1", gxpTests[0].TestName)
	assert.Equal(t, "GxP test 2", gxpTests[1].TestName)
}

func TestGetManualTests(t *testing.T) {
	tests := []TestReference{
		{TestName: "Manual test 1", IsManual: true},
		{TestName: "Automated test", IsManual: false},
		{TestName: "Manual test 2", IsManual: true},
	}

	manualTests := GetManualTests(tests)

	assert.Len(t, manualTests, 2)
	assert.Equal(t, "Manual test 1", manualTests[0].TestName)
	assert.Equal(t, "Manual test 2", manualTests[1].TestName)
}
