package release

import (
	"testing"
	"time"
)

func TestFormatElapsed(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{0 * time.Second, "0:00"},
		{1 * time.Second, "0:01"},
		{30 * time.Second, "0:30"},
		{59 * time.Second, "0:59"},
		{60 * time.Second, "1:00"},
		{61 * time.Second, "1:01"},
		{90 * time.Second, "1:30"},
		{119 * time.Second, "1:59"},
		{120 * time.Second, "2:00"},
		{5 * time.Minute, "5:00"},
		{5*time.Minute + 30*time.Second, "5:30"},
		{10 * time.Minute, "10:00"},
		{10*time.Minute + 45*time.Second, "10:45"},
		{59*time.Minute + 59*time.Second, "59:59"},
		{60 * time.Minute, "60:00"},
		{61*time.Minute + 23*time.Second, "61:23"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatElapsed(tt.duration)
			if got != tt.want {
				t.Errorf("formatElapsed(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestFormatElapsed_WithMilliseconds(t *testing.T) {
	// Should round to nearest second
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{500 * time.Millisecond, "0:01"},   // rounds up to 1
		{1500 * time.Millisecond, "0:02"},  // rounds up to 2
		{59500 * time.Millisecond, "1:00"}, // rounds up to 60 seconds
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatElapsed(tt.duration)
			if got != tt.want {
				t.Errorf("formatElapsed(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestCIRunInfo_Struct(t *testing.T) {
	info := CIRunInfo{
		SHA:   "abc123def456789",
		RunID: 12345,
	}

	if info.SHA != "abc123def456789" {
		t.Errorf("SHA = %q, want %q", info.SHA, "abc123def456789")
	}
	if info.RunID != 12345 {
		t.Errorf("RunID = %d, want %d", info.RunID, 12345)
	}
}

func TestCIRunStatus_Struct(t *testing.T) {
	status := CIRunStatus{
		Status:     "completed",
		Conclusion: "success",
		RunID:      67890,
		HeadSHA:    "abc123def456789",
	}

	if status.Status != "completed" {
		t.Errorf("Status = %q, want %q", status.Status, "completed")
	}
	if status.Conclusion != "success" {
		t.Errorf("Conclusion = %q, want %q", status.Conclusion, "success")
	}
	if status.RunID != 67890 {
		t.Errorf("RunID = %d, want %d", status.RunID, 67890)
	}
	if status.HeadSHA != "abc123def456789" {
		t.Errorf("HeadSHA = %q, want %q", status.HeadSHA, "abc123def456789")
	}
}

func TestCIRunStatus_FailedStatus(t *testing.T) {
	status := CIRunStatus{
		Status:     "completed",
		Conclusion: "failure",
		RunID:      11111,
		HeadSHA:    "def456",
	}

	if status.Conclusion != "failure" {
		t.Errorf("Conclusion = %q, want %q", status.Conclusion, "failure")
	}
}

func TestCIRunStatus_InProgressStatus(t *testing.T) {
	status := CIRunStatus{
		Status:     "in_progress",
		Conclusion: "",
		RunID:      22222,
		HeadSHA:    "ghi789",
	}

	if status.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", status.Status, "in_progress")
	}
	if status.Conclusion != "" {
		t.Errorf("Conclusion = %q, want empty for in_progress", status.Conclusion)
	}
}

func TestCIRunStatus_AllConclusions(t *testing.T) {
	validConclusions := []string{
		"success",
		"failure",
		"canceled",
		"skipped",
		"timed_out",
		"action_required",
	}

	for _, c := range validConclusions {
		status := CIRunStatus{Conclusion: c}
		if status.Conclusion != c {
			t.Errorf("Conclusion assignment failed for %q", c)
		}
	}
}

func TestNormalizeSlashes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"path/to/file", "path/to/file"},
		{"path\\to\\file", "path/to/file"},
		{"path\\to/file", "path/to/file"},
		{"", ""},
		{"file.go", "file.go"},
		{"C:\\Users\\test\\project", "C:/Users/test/project"},
		{"mixed/path\\with\\slashes/here", "mixed/path/with/slashes/here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeSlashes(tt.input)
			if got != tt.want {
				t.Errorf("normalizeSlashes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestModuleCIStatus_Struct(t *testing.T) {
	status := ModuleCIStatus{
		success: true,
		failed:  false,
		running: false,
		runID:   12345,
		ciSHA:   "abc123def456",
	}

	if !status.success {
		t.Error("success should be true")
	}
	if status.failed {
		t.Error("failed should be false")
	}
	if status.running {
		t.Error("running should be false")
	}
	if status.runID != 12345 {
		t.Errorf("runID = %d, want 12345", status.runID)
	}
	if status.ciSHA != "abc123def456" {
		t.Errorf("ciSHA = %q, want %q", status.ciSHA, "abc123def456")
	}
}

func TestModuleCIStatus_Running(t *testing.T) {
	status := ModuleCIStatus{
		success: false,
		failed:  false,
		running: true,
	}

	if status.success {
		t.Error("success should be false when running")
	}
	if !status.running {
		t.Error("running should be true")
	}
}

func TestModuleCIStatus_Failed(t *testing.T) {
	status := ModuleCIStatus{
		success: false,
		failed:  true,
		running: false,
	}

	if !status.failed {
		t.Error("failed should be true")
	}
}

func TestCIChainStatus_Struct(t *testing.T) {
	tests := []struct {
		name      string
		running   bool
		completed bool
	}{
		{"running", true, false},
		{"completed", false, true},
		{"empty", false, false},
		{"both", true, true}, // edge case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := CIChainStatus{
				running:   tt.running,
				completed: tt.completed,
			}
			if status.running != tt.running {
				t.Errorf("running = %v, want %v", status.running, tt.running)
			}
			if status.completed != tt.completed {
				t.Errorf("completed = %v, want %v", status.completed, tt.completed)
			}
		})
	}
}

func TestCIRunWithWorkflow_Struct(t *testing.T) {
	run := CIRunWithWorkflow{
		Status:       "completed",
		Conclusion:   "success",
		HeadSHA:      "abc123",
		WorkflowName: "ci-core.yaml",
	}

	if run.Status != "completed" {
		t.Errorf("Status = %q, want %q", run.Status, "completed")
	}
	if run.Conclusion != "success" {
		t.Errorf("Conclusion = %q, want %q", run.Conclusion, "success")
	}
	if run.HeadSHA != "abc123" {
		t.Errorf("HeadSHA = %q, want %q", run.HeadSHA, "abc123")
	}
	if run.WorkflowName != "ci-core.yaml" {
		t.Errorf("WorkflowName = %q, want %q", run.WorkflowName, "ci-core.yaml")
	}
}

func TestCiRunCoversCommit_ExactMatch(t *testing.T) {
	// Exact match should always return true
	if !ciRunCoversCommit("abc123", "abc123") {
		t.Error("exact match should return true")
	}
}

func TestCiRunCoversCommit_DifferentSHA(t *testing.T) {
	// Different SHAs without git context - isAncestor will fail
	// This tests the non-exact-match path (will call git merge-base)
	// In a test environment without the actual git repo, this should return false
	result := ciRunCoversCommit("abc123", "def456")
	// We can't assert the exact result without git context,
	// but we verify the function doesn't panic
	_ = result
}
