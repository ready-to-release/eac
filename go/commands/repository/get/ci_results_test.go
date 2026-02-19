package get

import (
	"testing"
	"time"
)

func TestClassifyInput(t *testing.T) {
	tests := []struct {
		name       string
		positional []string
		wantType   inputType
		wantValue  string
	}{
		{
			name:       "empty input -> auto",
			positional: []string{},
			wantType:   inputAuto,
			wantValue:  "",
		},
		{
			name:       "numeric run ID",
			positional: []string{"12345678"},
			wantType:   inputRunID,
			wantValue:  "12345678",
		},
		{
			name:       "short SHA (7 chars)",
			positional: []string{"abc1234"},
			wantType:   inputSHA,
			wantValue:  "abc1234",
		},
		{
			name:       "full SHA (40 chars)",
			positional: []string{"abcdef1234567890abcdef1234567890abcdef12"},
			wantType:   inputSHA,
			wantValue:  "abcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name:       "module name -> auto",
			positional: []string{"core"},
			wantType:   inputAuto,
			wantValue:  "",
		},
		{
			name:       "module name with hyphen -> auto",
			positional: []string{"eac-ext"},
			wantType:   inputAuto,
			wantValue:  "",
		},
		{
			name:       "hex string too short (6 chars) -> auto (module name)",
			positional: []string{"abc123"},
			wantType:   inputAuto,
			wantValue:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotValue := classifyInput(tt.positional)
			if gotType != tt.wantType {
				t.Errorf("classifyInput(%v) type = %v, want %v", tt.positional, gotType, tt.wantType)
			}
			if gotValue != tt.wantValue {
				t.Errorf("classifyInput(%v) value = %q, want %q", tt.positional, gotValue, tt.wantValue)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{"sub-second", 500, "500ms"},
		{"seconds", 45000, "45s"},
		{"minutes", 123000, "2m3s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := (time.Duration(tt.ms) * time.Millisecond)
			got := formatDuration(d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", d, got, tt.want)
			}
		})
	}
}

func TestBuildRunLinks(t *testing.T) {
	links := buildRunLinks(12345, "owner/repo")

	if links.WebURL != "https://github.com/owner/repo/actions/runs/12345" {
		t.Errorf("WebURL = %q", links.WebURL)
	}
	if links.ViewLogs != "gh run view 12345 --repo owner/repo --log" {
		t.Errorf("ViewLogs = %q", links.ViewLogs)
	}
	if links.ViewFailedLogs != "gh run view 12345 --repo owner/repo --log-failed" {
		t.Errorf("ViewFailedLogs = %q", links.ViewFailedLogs)
	}
	if links.DownloadAll != "gh run download 12345 --repo owner/repo" {
		t.Errorf("DownloadAll = %q", links.DownloadAll)
	}
}
