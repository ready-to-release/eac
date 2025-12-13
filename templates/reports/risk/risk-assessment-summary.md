# Risk Assessment Summary

**Generated:** {{ .GeneratedAt }}
**Scope:** {{ .ScopeDescription }}
**Profile:** {{ .ProfileName }}

## Overview

This assessment evaluated risk compliance controls (tagged with `@control:<id>`) against test evidence. Controls address risks as defined in control catalogs. Risk assessments are documented separately and map risks to controls.

## Executive Summary

| Metric | Value |
|--------|-------|
| **Modules Assessed** | {{ .Summary.ModulesAssessed }} |
| **Total Controls** | {{ .Summary.TotalControls }} |
| **Satisfied** | {{ .Summary.Satisfied }} ({{ percentf .Summary.Satisfied .Summary.TotalControls 1 }}%) |
| **Not Satisfied** | {{ .Summary.NotSatisfied }} ({{ percentf .Summary.NotSatisfied .Summary.TotalControls 1 }}%) |
| **Overall Compliance** | {{ riskBadge "Low" }} |

## Top Risk Modules

{{ range .ModuleResults -}}
{{ if gt .NotSatisfied 10 -}}

- **{{ .Module }}**: {{ .NotSatisfied }} controls not satisfied
{{ end -}}
{{ end }}

{{ if eq .Summary.NotSatisfied 0 -}}

## Status

✅ **All assessed controls are satisfied**

Continue monitoring and maintain current security practices.
{{ else -}}

## Priority Actions Required

**{{ .Summary.NotSatisfied }} controls require attention across {{ .Summary.ModulesAssessed }} modules.**

Top 5 modules by unsatisfied controls:
{{ $count := 0 -}}
{{ range .ModuleResults -}}
{{ if lt $count 5 -}}
{{ if gt .NotSatisfied 0 -}}

- **{{ .Module }}**: {{ .NotSatisfied }} controls
{{ $count = add $count 1 -}}
{{ end -}}
{{ end -}}
{{ end }}

{{ end -}}
