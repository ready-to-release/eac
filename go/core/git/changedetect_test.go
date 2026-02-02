package git

import (
	"testing"
)

// Tests for change detection related methods (HeadCommit, UncommittedFiles)

func TestMockRepository_HeadCommit(t *testing.T) {
	fullSHA := "abc123def456789012345678901234567890dead"
	mock := NewMockRepository("/test/repo").WithHeadSHA(fullSHA)

	commit, err := mock.HeadCommit()
	if err != nil {
		t.Fatalf("HeadCommit failed: %v", err)
	}

	if commit != fullSHA {
		t.Errorf("expected %q, got %q", fullSHA, commit)
	}
}

func TestMockRepository_HeadCommit_NoCommit(t *testing.T) {
	mock := NewMockRepository("/test/repo")

	_, err := mock.HeadCommit()
	if err == nil {
		t.Error("expected error for repo with no HEAD commit")
	}
}

func TestMockRepository_HeadCommit_Error(t *testing.T) {
	mock := NewMockRepository("/test/repo").
		WithHeadSHA("abc123").
		WithError("HeadCommit", errMock)

	_, err := mock.HeadCommit()
	if err != errMock {
		t.Errorf("expected injected error, got %v", err)
	}
}

func TestMockRepository_UncommittedFiles(t *testing.T) {
	files := []string{"modified.go", "new.txt", "renamed.md"}
	mock := NewMockRepository("/test/repo").WithUncommittedFiles(files)

	result, err := mock.UncommittedFiles()
	if err != nil {
		t.Fatalf("UncommittedFiles failed: %v", err)
	}

	if len(result) != len(files) {
		t.Errorf("expected %d files, got %d", len(files), len(result))
	}

	for i, f := range files {
		if result[i] != f {
			t.Errorf("file[%d]: expected %q, got %q", i, f, result[i])
		}
	}
}

func TestMockRepository_UncommittedFiles_Empty(t *testing.T) {
	mock := NewMockRepository("/test/repo")

	result, err := mock.UncommittedFiles()
	if err != nil {
		t.Fatalf("UncommittedFiles failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 files for clean repo, got %d", len(result))
	}
}

func TestMockRepository_UncommittedFiles_Error(t *testing.T) {
	mock := NewMockRepository("/test/repo").
		WithUncommittedFiles([]string{"file.go"}).
		WithError("UncommittedFiles", errMock)

	_, err := mock.UncommittedFiles()
	if err != errMock {
		t.Errorf("expected injected error, got %v", err)
	}
}

func TestParseStatusPorcelain_Empty(t *testing.T) {
	result := parseStatusPorcelain("")
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestParseStatusPorcelain_ModifiedFile(t *testing.T) {
	// Format: XY filename (2 status chars + space + filename)
	output := " M file.go" // Modified in working tree
	result := parseStatusPorcelain(output)

	if len(result) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result))
	}
	// Note: line[3:] skips "XY " to get the filename
	if result[0] != "file.go" {
		t.Errorf("expected 'file.go', got %q", result[0])
	}
}

func TestParseStatusPorcelain_MultipleFiles(t *testing.T) {
	output := ` M modified.go
A  added.txt
?? untracked.md
 D deleted.go`

	result := parseStatusPorcelain(output)

	expected := []string{"modified.go", "added.txt", "untracked.md", "deleted.go"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d files, got %d: %v", len(expected), len(result), result)
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("file[%d]: expected %q, got %q", i, exp, result[i])
		}
	}
}

func TestParseStatusPorcelain_QuotedPath(t *testing.T) {
	output := ` M "path with spaces/file.go"`
	result := parseStatusPorcelain(output)

	if len(result) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result))
	}
	if result[0] != "path with spaces/file.go" {
		t.Errorf("expected 'path with spaces/file.go', got %q", result[0])
	}
}

func TestParseStatusPorcelain_SkipsShortLines(t *testing.T) {
	output := "AB" // Too short to have a valid file path
	result := parseStatusPorcelain(output)

	if len(result) != 0 {
		t.Errorf("expected 0 files for too-short line, got %d: %v", len(result), result)
	}
}

// Helper error for testing
var errMock = errMockError{}

type errMockError struct{}

func (e errMockError) Error() string { return "mock error" }
