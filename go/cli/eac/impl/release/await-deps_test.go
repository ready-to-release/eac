package release

import (
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"hello", 5, "hello"},
		{"", 5, ""},
		{"abcdefgh", 6, "abc..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
		{"a very long string that needs truncation", 20, "a very long strin..."},
		{"short", 100, "short"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestTruncateString_LengthConstraint(t *testing.T) {
	// Verify truncated strings don't exceed maxLen
	inputs := []string{
		"this is a long commit message that should be truncated",
		"fix: resolve issue with authentication flow",
		"chore(deps): update all dependencies to latest versions",
	}

	for _, input := range inputs {
		for maxLen := 10; maxLen <= 50; maxLen += 10 {
			result := truncateString(input, maxLen)
			if len(result) > maxLen {
				t.Errorf("truncateString(%q, %d) returned string of length %d, exceeds maxLen",
					input, maxLen, len(result))
			}
		}
	}
}

func TestDepCIStatus_Struct(t *testing.T) {
	status := DepCIStatus{
		Moniker:       "core",
		LastCommit:    "abc123def456789",
		LastCommitMsg: "fix: something important",
		CIWorkflow:    "ci-core.yaml",
		Status:        "success",
		RunID:         12345,
		RunURL:        "https://github.com/org/repo/actions/runs/12345",
	}

	if status.Moniker != "core" {
		t.Errorf("Moniker = %q, want %q", status.Moniker, "core")
	}
	if status.Status != "success" {
		t.Errorf("Status = %q, want %q", status.Status, "success")
	}
	if status.RunID != 12345 {
		t.Errorf("RunID = %d, want %d", status.RunID, 12345)
	}
	if status.CIWorkflow != "ci-core.yaml" {
		t.Errorf("CIWorkflow = %q, want %q", status.CIWorkflow, "ci-core.yaml")
	}
}

func TestDepCIStatus_SkippedStatus(t *testing.T) {
	status := DepCIStatus{
		Moniker:    "static-module",
		Status:     "skipped",
		SkipReason: "no CI workflow",
	}

	if status.Status != "skipped" {
		t.Errorf("Status = %q, want %q", status.Status, "skipped")
	}
	if status.SkipReason != "no CI workflow" {
		t.Errorf("SkipReason = %q, want %q", status.SkipReason, "no CI workflow")
	}
}

func TestDepCIStatus_FailedStatus(t *testing.T) {
	status := DepCIStatus{
		Moniker:       "failing-module",
		LastCommit:    "def456",
		LastCommitMsg: "chore: update deps",
		CIWorkflow:    "ci-failing-module.yaml",
		Status:        "failed",
		RunID:         67890,
		RunURL:        "https://github.com/org/repo/actions/runs/67890",
	}

	if status.Status != "failed" {
		t.Errorf("Status = %q, want %q", status.Status, "failed")
	}
	if status.RunURL == "" {
		t.Error("RunURL should be set for failed status")
	}
}

func TestDepCIStatus_NotFoundStatus(t *testing.T) {
	status := DepCIStatus{
		Moniker:    "new-module",
		LastCommit: "aaa111",
		Status:     "not_found",
	}

	if status.Status != "not_found" {
		t.Errorf("Status = %q, want %q", status.Status, "not_found")
	}
	// For not_found, RunID and RunURL should be zero/empty
	if status.RunID != 0 {
		t.Errorf("RunID = %d, want 0 for not_found", status.RunID)
	}
	if status.RunURL != "" {
		t.Errorf("RunURL = %q, want empty for not_found", status.RunURL)
	}
}

func TestDepCIStatus_RunningStatus(t *testing.T) {
	status := DepCIStatus{
		Moniker: "building-module",
		Status:  "running",
		RunID:   11111,
		RunURL:  "https://github.com/org/repo/actions/runs/11111",
	}

	if status.Status != "running" {
		t.Errorf("Status = %q, want %q", status.Status, "running")
	}
}

func TestDepCIStatus_TimeoutStatus(t *testing.T) {
	status := DepCIStatus{
		Moniker: "slow-module",
		Status:  "timeout",
		RunID:   22222,
		RunURL:  "https://github.com/org/repo/actions/runs/22222",
	}

	if status.Status != "timeout" {
		t.Errorf("Status = %q, want %q", status.Status, "timeout")
	}
}

func TestDepCIStatus_AllValidStatuses(t *testing.T) {
	validStatuses := []string{
		"success",
		"failed",
		"running",
		"not_found",
		"skipped",
		"timeout",
	}

	for _, s := range validStatuses {
		status := DepCIStatus{Status: s}
		if status.Status != s {
			t.Errorf("Status assignment failed for %q", s)
		}
	}
}

func TestDepCIStatus_CommitSHAHandling(t *testing.T) {
	// Test that we can handle both full and short SHAs
	fullSHA := "abc123def456789012345678901234567890abcd"
	status := DepCIStatus{
		Moniker:    "test-module",
		LastCommit: fullSHA,
		Status:     "success",
	}

	if len(status.LastCommit) != 40 {
		t.Errorf("LastCommit length = %d, want 40 for full SHA", len(status.LastCommit))
	}

	// Code should handle short SHAs by slicing [:7]
	shortSHA := status.LastCommit[:7]
	if shortSHA != "abc123d" {
		t.Errorf("Short SHA = %q, want %q", shortSHA, "abc123d")
	}
}

func TestDepCIStatus_EmptyValues(t *testing.T) {
	// Test zero-value struct
	var status DepCIStatus

	if status.Moniker != "" {
		t.Errorf("Moniker should be empty by default")
	}
	if status.Status != "" {
		t.Errorf("Status should be empty by default")
	}
	if status.RunID != 0 {
		t.Errorf("RunID should be 0 by default")
	}
}
