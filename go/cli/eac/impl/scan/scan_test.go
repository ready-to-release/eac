package scan

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/evidence"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseScannerList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []evidence.ScannerType
		wantErr bool
		errMsg  string
	}{
		{
			name:  "single valid scanner sbom",
			input: "sbom",
			want:  []evidence.ScannerType{evidence.ScannerType(tool.ScannerToolIDForCategory("sbom"))},
		},
		{
			name:  "single valid scanner vuln",
			input: "vuln",
			want:  []evidence.ScannerType{evidence.ScannerType(tool.ScannerToolIDForCategory("vuln"))},
		},
		{
			name:  "single valid scanner secrets",
			input: "secrets",
			want:  []evidence.ScannerType{evidence.ScannerType(tool.ScannerToolIDForCategory("secrets"))},
		},
		{
			name:  "single valid scanner sast",
			input: "sast",
			want:  []evidence.ScannerType{evidence.ScannerType(tool.ScannerToolIDForCategory("sast"))},
		},
		{
			name:  "single valid scanner zap",
			input: "zap",
			want:  []evidence.ScannerType{evidence.ScannerType(tool.ScannerToolIDForCategory("zap"))},
		},
		{
			name:  "multiple valid scanners",
			input: "sbom,vuln,secrets",
			want: []evidence.ScannerType{
				evidence.ScannerType(tool.ScannerToolIDForCategory("sbom")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("vuln")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("secrets")),
			},
		},
		{
			name:  "all valid scanner types",
			input: "sbom,vuln,secrets,iac,compliance,sast,zap",
			want: []evidence.ScannerType{
				evidence.ScannerType(tool.ScannerToolIDForCategory("sbom")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("vuln")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("secrets")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("iac")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("compliance")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("sast")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("zap")),
			},
		},
		{
			name:    "invalid scanner name",
			input:   "invalid",
			wantErr: true,
			errMsg:  "invalid scanner type: invalid",
		},
		{
			name:    "valid and invalid scanner mixed",
			input:   "sbom,badscanner",
			wantErr: true,
			errMsg:  "invalid scanner type: badscanner",
		},
		{
			name:  "whitespace handling around entries",
			input: " sbom , vuln ",
			want: []evidence.ScannerType{
				evidence.ScannerType(tool.ScannerToolIDForCategory("sbom")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("vuln")),
			},
		},
		{
			name:  "case insensitivity uppercase",
			input: "SBOM",
			want:  []evidence.ScannerType{evidence.ScannerType(tool.ScannerToolIDForCategory("sbom"))},
		},
		{
			name:  "case insensitivity mixed case",
			input: "SbOm,VULN,Secrets",
			want: []evidence.ScannerType{
				evidence.ScannerType(tool.ScannerToolIDForCategory("sbom")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("vuln")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("secrets")),
			},
		},
		{
			name:  "empty strings between commas are skipped",
			input: "sbom,,vuln",
			want: []evidence.ScannerType{
				evidence.ScannerType(tool.ScannerToolIDForCategory("sbom")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("vuln")),
			},
		},
		{
			name:  "trailing comma produces empty segment that is skipped",
			input: "sbom,",
			want:  []evidence.ScannerType{evidence.ScannerType(tool.ScannerToolIDForCategory("sbom"))},
		},
		{
			name:  "leading comma produces empty segment that is skipped",
			input: ",vuln",
			want:  []evidence.ScannerType{evidence.ScannerType(tool.ScannerToolIDForCategory("vuln"))},
		},
		{
			name:    "completely empty input",
			input:   "",
			wantErr: true,
			errMsg:  "no valid scanner types specified",
		},
		{
			name:    "only whitespace",
			input:   "   ",
			wantErr: true,
			errMsg:  "no valid scanner types specified",
		},
		{
			name:    "only commas",
			input:   ",,,",
			wantErr: true,
			errMsg:  "no valid scanner types specified",
		},
		{
			name:  "whitespace and case combined",
			input: "  SBOM  ,  vuln  ,  IAC  ",
			want: []evidence.ScannerType{
				evidence.ScannerType(tool.ScannerToolIDForCategory("sbom")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("vuln")),
				evidence.ScannerType(tool.ScannerToolIDForCategory("iac")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScannerList(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseSeverityList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []evidence.Severity
		wantErr bool
		errMsg  string
	}{
		{
			name:  "single valid severity LOW",
			input: "LOW",
			want:  []evidence.Severity{evidence.SeverityLow},
		},
		{
			name:  "single valid severity MEDIUM",
			input: "MEDIUM",
			want:  []evidence.Severity{evidence.SeverityMedium},
		},
		{
			name:  "single valid severity HIGH",
			input: "HIGH",
			want:  []evidence.Severity{evidence.SeverityHigh},
		},
		{
			name:  "single valid severity CRITICAL",
			input: "CRITICAL",
			want:  []evidence.Severity{evidence.SeverityCritical},
		},
		{
			name:  "multiple valid severities",
			input: "LOW,HIGH,CRITICAL",
			want: []evidence.Severity{
				evidence.SeverityLow,
				evidence.SeverityHigh,
				evidence.SeverityCritical,
			},
		},
		{
			name:  "all valid severity levels",
			input: "LOW,MEDIUM,HIGH,CRITICAL",
			want: []evidence.Severity{
				evidence.SeverityLow,
				evidence.SeverityMedium,
				evidence.SeverityHigh,
				evidence.SeverityCritical,
			},
		},
		{
			name:    "invalid severity",
			input:   "UNKNOWN",
			wantErr: true,
			errMsg:  "invalid severity: UNKNOWN",
		},
		{
			name:    "valid and invalid severity mixed",
			input:   "HIGH,BOGUS",
			wantErr: true,
			errMsg:  "invalid severity: BOGUS",
		},
		{
			name:  "whitespace handling around entries",
			input: " HIGH , LOW ",
			want: []evidence.Severity{
				evidence.SeverityHigh,
				evidence.SeverityLow,
			},
		},
		{
			name:  "case insensitivity lowercase input",
			input: "low",
			want:  []evidence.Severity{evidence.SeverityLow},
		},
		{
			name:  "case insensitivity mixed case input",
			input: "High,medium,CrItIcAl",
			want: []evidence.Severity{
				evidence.SeverityHigh,
				evidence.SeverityMedium,
				evidence.SeverityCritical,
			},
		},
		{
			name:  "empty input returns nil without error",
			input: "",
			want:  nil,
		},
		{
			name:  "only whitespace returns nil without error",
			input: "   ",
			want:  nil,
		},
		{
			name:  "only commas returns nil without error",
			input: ",,,",
			want:  nil,
		},
		{
			name:  "empty strings between commas are skipped",
			input: "HIGH,,CRITICAL",
			want: []evidence.Severity{
				evidence.SeverityHigh,
				evidence.SeverityCritical,
			},
		},
		{
			name:  "trailing comma",
			input: "HIGH,",
			want:  []evidence.Severity{evidence.SeverityHigh},
		},
		{
			name:  "leading comma",
			input: ",LOW",
			want:  []evidence.Severity{evidence.SeverityLow},
		},
		{
			name:  "whitespace and case combined",
			input: "  low  ,  MEDIUM  ,  high  ",
			want: []evidence.Severity{
				evidence.SeverityLow,
				evidence.SeverityMedium,
				evidence.SeverityHigh,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSeverityList(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
