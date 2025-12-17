package github

import (
	"testing"
	"time"
)

func TestMockAPI_GetTreeFiles(t *testing.T) {
	mock := NewMockAPI()
	mock.AddTreeFiles("abc123", []string{"go/main.go", "README.md", "docs/index.md"})

	files, err := mock.GetTreeFiles("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}

	// Verify call tracking
	if mock.CallCount("GetTreeFiles") != 1 {
		t.Errorf("expected 1 call to GetTreeFiles, got %d", mock.CallCount("GetTreeFiles"))
	}
}

func TestMockAPI_GetTreeFiles_NotFound(t *testing.T) {
	mock := NewMockAPI()

	files, err := mock.GetTreeFiles("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files for unknown SHA, got %d", len(files))
	}
}

func TestMockAPI_GetTreeFiles_Error(t *testing.T) {
	mock := NewMockAPI()
	mock.TreeError = errTest

	_, err := mock.GetTreeFiles("abc123")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestMockAPI_FindRunBySHA(t *testing.T) {
	mock := NewMockAPI()
	mock.AddSuccessRun("ci-test.yaml", "abc123")
	mock.AddSuccessRun("ci-test.yaml", "def456")

	run, err := mock.FindRunBySHA("ci-test.yaml", "abc123", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if run == nil {
		t.Fatal("expected run, got nil")
	}

	if run.HeadSHA != "abc123" {
		t.Errorf("expected SHA abc123, got %s", run.HeadSHA)
	}
}

func TestMockAPI_FindRunBySHA_NotFound(t *testing.T) {
	mock := NewMockAPI()
	mock.AddSuccessRun("ci-test.yaml", "abc123")

	run, err := mock.FindRunBySHA("ci-test.yaml", "nonexistent", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if run != nil {
		t.Error("expected nil for non-matching SHA")
	}
}

func TestMockAPI_HasRecentSuccess(t *testing.T) {
	mock := NewMockAPI()
	// Add a recent successful run
	mock.AddRun("ci-test.yaml", WorkflowRun{
		ID:         1,
		HeadSHA:    "abc123",
		Status:     "completed",
		Conclusion: "success",
		CreatedAt:  time.Now().Add(-30 * time.Minute),
	})

	hasRecent, err := mock.HasRecentSuccess("ci-test.yaml", "abc123", 2*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hasRecent {
		t.Error("expected hasRecent=true for recent run")
	}
}

func TestMockAPI_HasRecentSuccess_TooOld(t *testing.T) {
	mock := NewMockAPI()
	// Add an old successful run
	mock.AddRun("ci-test.yaml", WorkflowRun{
		ID:         1,
		HeadSHA:    "abc123",
		Status:     "completed",
		Conclusion: "success",
		CreatedAt:  time.Now().Add(-3 * time.Hour),
	})

	hasRecent, err := mock.HasRecentSuccess("ci-test.yaml", "abc123", 2*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hasRecent {
		t.Error("expected hasRecent=false for old run")
	}
}

func TestMockAPI_HasRecentSuccess_WrongSHA(t *testing.T) {
	mock := NewMockAPI()
	mock.AddRun("ci-test.yaml", WorkflowRun{
		ID:         1,
		HeadSHA:    "different",
		Status:     "completed",
		Conclusion: "success",
		CreatedAt:  time.Now(),
	})

	hasRecent, err := mock.HasRecentSuccess("ci-test.yaml", "abc123", 2*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hasRecent {
		t.Error("expected hasRecent=false for different SHA")
	}
}

func TestMockAPI_ListRuns_StatusFilter(t *testing.T) {
	mock := NewMockAPI()
	mock.AddSuccessRun("ci-test.yaml", "sha1")
	mock.AddRun("ci-test.yaml", WorkflowRun{
		ID:         2,
		HeadSHA:    "sha2",
		Status:     "completed",
		Conclusion: "failure",
	})

	runs, err := mock.ListRuns("ci-test.yaml", ListRunsOpts{Status: "success"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(runs) != 1 {
		t.Errorf("expected 1 successful run, got %d", len(runs))
	}

	if runs[0].HeadSHA != "sha1" {
		t.Errorf("expected sha1, got %s", runs[0].HeadSHA)
	}
}

func TestMockAPI_ListRuns_Limit(t *testing.T) {
	mock := NewMockAPI()
	for i := 0; i < 10; i++ {
		mock.AddSuccessRun("ci-test.yaml", "sha"+string(rune('0'+i)))
	}

	runs, err := mock.ListRuns("ci-test.yaml", ListRunsOpts{Limit: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(runs) != 3 {
		t.Errorf("expected 3 runs with limit, got %d", len(runs))
	}
}

func TestMockAPI_ReleaseExists(t *testing.T) {
	mock := NewMockAPI()
	mock.AddRelease("v1.0.0", "Release 1.0.0")
	mock.AddRelease("v1.1.0", "Release 1.1.0")

	exists, err := mock.ReleaseExists("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !exists {
		t.Error("expected release to exist")
	}

	exists, err = mock.ReleaseExists("v2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exists {
		t.Error("expected release to not exist")
	}
}

func TestMockAPI_ListReleases(t *testing.T) {
	mock := NewMockAPI()
	mock.AddRelease("v1.0.0", "Release 1.0.0")
	mock.AddRelease("v1.1.0", "Release 1.1.0")
	mock.AddRelease("v1.2.0", "Release 1.2.0")

	releases, err := mock.ListReleases(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 2 {
		t.Errorf("expected 2 releases with limit, got %d", len(releases))
	}
}

func TestMockAPI_CallTracking(t *testing.T) {
	mock := NewMockAPI()

	mock.GetTreeFiles("sha1")
	mock.GetTreeFiles("sha2")
	mock.FindRunBySHA("workflow", "sha", 10)

	if mock.CallCount("GetTreeFiles") != 2 {
		t.Errorf("expected 2 calls to GetTreeFiles")
	}

	if mock.CallCount("FindRunBySHA") != 1 {
		t.Errorf("expected 1 call to FindRunBySHA")
	}

	last := mock.LastCall("GetTreeFiles")
	if last == nil {
		t.Fatal("expected last call, got nil")
	}

	if last.Args[0] != "sha2" {
		t.Errorf("expected last call with sha2, got %v", last.Args[0])
	}
}

func TestMockAPI_Reset(t *testing.T) {
	mock := NewMockAPI()
	mock.GetTreeFiles("sha1")
	mock.Reset()

	if mock.CallCount("GetTreeFiles") != 0 {
		t.Error("expected calls to be reset")
	}
}

// errTest is a test error
var errTest = &testError{"test error"}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
