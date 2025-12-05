# Risk Assessment Report (Detailed)

**Generated:** {{ .GeneratedAt }}

**Scope:** {{ .ScopeDescription }}

**Profile:** {{ .ProfileName }}

**Test Suite:** {{ .TestSuite }}

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Assessment Overview](#assessment-overview)
3. [Module Assessment Results](#module-assessment-results)
4. [Control Status by Module](#control-status-by-module)
5. [Risk Analysis](#risk-analysis)
6. [Evidence Summary](#evidence-summary)
7. [Detailed Findings](#detailed-findings)
8. [Recommendations](#recommendations)
9. [Report Files](#report-files)

---

## Executive Summary

| Metric | Value | Percentage |
|--------|-------|------------|
| **Total Controls Assessed** | {{ .Summary.TotalControls }} | 100% |
| **Satisfied** | {{ .Summary.Satisfied }} | {{ percentf .Summary.Satisfied .Summary.TotalControls 1 }}% |
| **Not Satisfied** | {{ .Summary.NotSatisfied }} | {{ percentf .Summary.NotSatisfied .Summary.TotalControls 1 }}% |
| **Modules Assessed** | {{ .Summary.ModulesAssessed }} | - |

### Compliance Status

{{ $compliance := mul .Summary.Satisfied 100 -}}
{{ $threshold90 := mul .Summary.TotalControls 90 -}}
{{ $threshold75 := mul .Summary.TotalControls 75 -}}
{{ $threshold50 := mul .Summary.TotalControls 50 -}}
{{ if gte $compliance $threshold90 -}}
🟢 **Excellent** - Compliance rate above 90%
{{ else if gte $compliance $threshold75 -}}
🟡 **Good** - Compliance rate above 75%, some improvements needed
{{ else if gte $compliance $threshold50 -}}
🟠 **Fair** - Compliance rate above 50%, significant improvements required
{{ else -}}
🔴 **Poor** - Compliance rate below 50%, immediate action required
{{ end }}

---

## Assessment Overview

This assessment evaluated **{{ .Summary.ModulesAssessed }} modules** against **{{ .Summary.TotalControls }} total control checks** defined in **{{ .ProfileName }}**.

### Test Evidence
- Test suite: **{{ .TestSuite }}**
- Evidence collected from test results and security scans

### Control Distribution
- **Satisfied controls:** {{ .Summary.Satisfied }} / {{ .Summary.TotalControls }}
- **Not satisfied controls:** {{ .Summary.NotSatisfied }} / {{ .Summary.TotalControls }}

---

## Module Assessment Results

| Module | Satisfied | Not Satisfied | Compliance % | Risk Score | Test Evidence | Security Evidence |
|--------|-----------|---------------|--------------|------------|---------------|-------------------|
{{ range .ModuleResults -}}
| {{ .Module }} | {{ .Satisfied }} | {{ .NotSatisfied }} | {{ percentf .Satisfied (add .Satisfied .NotSatisfied) 1 }}% | {{ .RiskScoreFormatted }} | {{ .TestEvidenceFormatted }} | {{ .SecurityEvidenceFormatted }} |
{{ end }}

---

## Control Status by Module

{{ range .ModuleResults -}}
### {{ .Module }}

**Compliance:** {{ percentf .Satisfied (add .Satisfied .NotSatisfied) 1 }}% ({{ .Satisfied }}/{{ add .Satisfied .NotSatisfied }} controls)

{{ if gt (len .SatisfiedControls) 0 -}}
#### ✓ Satisfied Controls ({{ len .SatisfiedControls }})

{{ join .SatisfiedControls ", " }}

{{ end -}}
{{ if gt (len .NotSatisfiedControls) 0 -}}
#### ✗ Not Satisfied Controls ({{ len .NotSatisfiedControls }})

{{ join .NotSatisfiedControls ", " }}

{{ end -}}
{{ if gt (len .NotSatisfiedFindings) 0 -}}
#### Findings

{{ range .NotSatisfiedFindings -}}
- **{{ .ControlID }}**: {{ .Title }}
{{ end }}

{{ end -}}
---

{{ end }}

## Risk Analysis

### Risk Distribution

{{$modulesWithRisk := 0 -}}
{{ range .ModuleResults -}}
{{ if .RiskScore -}}
{{ $modulesWithRisk = add $modulesWithRisk 1 -}}
{{ end -}}
{{ end -}}

**Modules with risk scores:** {{ $modulesWithRisk }} / {{ .Summary.ModulesAssessed }}

### Detailed Risk Scores

| Module | Risk Level | Score | Likelihood | Impact | Reasoning |
|--------|------------|-------|------------|--------|-----------|
{{ range .ModuleResults -}}
{{ if .RiskScore -}}
| {{ .Module }} | {{ .RiskScoreFormatted }} | {{ .RiskScore.Score }} | {{ .RiskScore.Likelihood }}/5 | {{ .RiskScore.Impact }}/5 | {{ truncate .RiskScore.Reasoning 80 }} |
{{ end -}}
{{ end }}

---

## Evidence Summary

### Test Coverage

{{ $totalTests := 0 -}}
{{ $passedTests := 0 -}}
{{ $modulesWithTests := 0 -}}
{{ range .ModuleResults -}}
{{ if ne .TestEvidenceFormatted "N/A" -}}
{{ $modulesWithTests = add $modulesWithTests 1 -}}
{{ end -}}
{{ end }}

- **Modules with test evidence:** {{ $modulesWithTests }} / {{ .Summary.ModulesAssessed }}
- **Test suite used:** {{ .TestSuite }}

### Security Scans

{{ $modulesWithScans := 0 -}}
{{ range .ModuleResults -}}
{{ if ne .SecurityEvidenceFormatted "N/A" -}}
{{ $modulesWithScans = add $modulesWithScans 1 -}}
{{ end -}}
{{ end }}

- **Modules with security scans:** {{ $modulesWithScans }} / {{ .Summary.ModulesAssessed }}
- **Vulnerability scanning:** Performed
- **SBOM generation:** Performed

---

## Detailed Findings

{{ if gt .Summary.NotSatisfied 0 -}}
### Controls Requiring Attention

{{ range .ModuleResults -}}
{{ if gt .NotSatisfied 0 -}}
#### {{ .Module }} ({{ .NotSatisfied }} controls not satisfied)

**Risk Level:** {{ .RiskScoreFormatted }}

**Test Evidence:** {{ .TestEvidenceFormatted }}

**Security Evidence:** {{ .SecurityEvidenceFormatted }}

**Controls not satisfied:**
{{ range .NotSatisfiedFindings -}}
- **{{ .ControlID }}**: {{ .Title }}
{{ end }}

{{ end -}}
{{ end -}}
{{ else -}}
### No Issues Found

✅ All assessed controls are satisfied across all modules.

{{ end -}}

---

## Recommendations

{{ if gt .Summary.NotSatisfied 0 -}}
### High Priority Actions

The following modules require immediate attention:

{{ range .ModuleResults -}}
{{ if gt .NotSatisfied 5 -}}
**{{ .Module }}** - {{ .NotSatisfied }} controls not satisfied

Priority: {{ .RiskScoreFormatted }}

Action items:
{{ $count := 0 -}}
{{ range .NotSatisfiedFindings -}}
{{ if lt $count 5 -}}
- Address {{ .ControlID }}: {{ .Title }}
{{ $count = add $count 1 -}}
{{ end -}}
{{ end }}

{{ end -}}
{{ end -}}

### General Recommendations

1. **Continuous Monitoring**: Regularly run risk assessments to track compliance
2. **Evidence Collection**: Ensure all modules have test and security evidence
3. **Control Implementation**: Prioritize implementing controls for high-risk modules
4. **Documentation**: Maintain up-to-date control implementation documentation

{{ else -}}
### Maintain Current Practices

✅ All assessed controls are satisfied. Recommendations:

1. **Continue monitoring**: Run regular assessments to maintain compliance
2. **Stay current**: Keep security scans and test suites up to date
3. **Document changes**: Record any changes to control implementations
4. **Review periodically**: Assess against updated profiles as they become available

{{ end -}}

---

## Report Files

### Aggregated Report
- `{{ .AggregatedReportPath }}`

### Individual Module Reports
{{ range .ModuleResults -}}
- **{{ .Module }}**: `{{ .ReportPath }}`
{{ end -}}

---

*Generated by EAC Risk Assessment Tool*

*Report Type: Detailed Analysis*
