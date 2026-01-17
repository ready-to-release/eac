## Test Execution Results

**Generated:** {{ .GeneratedAt }}  
**Last Run:** {{ .LastRun }}  
**Modules Tested:** {{ .ModulesTested }}  
**Total Tests:** {{ .TotalTests }} ({{ .TotalPassed }} passed, {{ .TotalFailed }} failed)  

---

### Module Overview

| Module | Tests | Passed | Failed | Skipped | Duration | Controls |
|--------|-------|--------|--------|---------|----------|----------|
{{ range .ModuleStats -}}
| {{ .Module }} | {{ .Total }} | {{ .Passed }} | {{ .Failed }} | {{ .Skipped }} | {{ formatDurationSec .DurationSeconds }} | {{ join .Controls ", " }} |
{{ end }}

---

{{ if gt (len .SpecCoverage) 0 -}}

### Specification Coverage

| Feature | Scenarios | Status | Controls |
|---------|-----------|--------|----------|
{{ range .SpecCoverage -}}
| {{ .FeatureName }} | {{ .ScenarioCount }} | {{ formatStatus .PassedCount .FailedCount .SkippedCount }} | {{ join .Controls ", " }} |
{{ end }}

**Summary:**

- Features: {{ len .SpecCoverage }}

---
{{ end }}

### Test Results by Module

{{ range .ModuleStats -}}

#### {{ .Module }}

**Total:** {{ .Total }} tests ({{ .Passed }} passed, {{ .Failed }} failed{{ if gt .Skipped 0 }}, {{ .Skipped }} skipped{{ end }}) - {{ formatDurationSec .DurationSeconds }}

| # | Type | Name | Suite | Status | Tags |
|---|------|------|-------|--------|------|
{{ range $idx, $test := .Tests -}}
| {{ add $idx 1 }} | {{ $test.Type }} | {{ truncate $test.Name 60 }} | {{ $test.Suite }} | {{ statusIcon $test.Status }} | {{ join $test.Tags " " }} |
{{ end }}
{{ end }}

---

### Summary

#### By Type

{{ range .SummaryByType -}}

- **{{ .Type }}**: {{ .Count }} tests ({{ .Passed }} passed, {{ .Failed }} failed)
{{ end }}

#### By Suite

{{ range .SummaryBySuite -}}

- **{{ .Suite }}**: {{ .Count }} tests ({{ .Passed }} passed, {{ .Failed }} failed)
{{ end }}

{{ if gt (len .ControlSummary) 0 -}}

#### By Control

{{ range .ControlSummary -}}

- **{{ .ControlID }}**: {{ .TestCount }} tests across {{ .ModuleCount }} modules ({{ .PassedCount }} passed, {{ .FailedCount }} failed)
{{ end }}

---

**Unique controls tested:** {{ len .ControlSummary }}
{{ end }}
