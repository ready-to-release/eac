package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
)

// createValidWorkspace creates a minimal valid workspace for testing.
func createValidWorkspace(t *testing.T, dir string) {
	t.Helper()
	// Create .git directory
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
}

// createEacWorkspace creates a workspace with only .eac/repository.yml (no .git).
func createEacWorkspace(t *testing.T, dir string) {
	t.Helper()
	eacDir := filepath.Join(dir, paths.EACConfigRelPath)
	if err := os.MkdirAll(eacDir, 0o755); err != nil {
		t.Fatalf("failed to create .eac: %v", err)
	}
	repoFile := filepath.Join(eacDir, "repository.yml")
	if err := os.WriteFile(repoFile, []byte("name: test\n"), 0o644); err != nil {
		t.Fatalf("failed to create repository.yml: %v", err)
	}
}

// saveAndClearEnv saves env vars and clears them for clean test state.
func saveAndClearEnv(t *testing.T) func() {
	t.Helper()
	saved := map[string]string{
		environments.EnvR2RRepoRoot:      os.Getenv(environments.EnvR2RRepoRoot),
		environments.EnvR2RContainerRoot: os.Getenv(environments.EnvR2RContainerRoot),
		environments.EnvR2RPWD:           os.Getenv(environments.EnvR2RPWD),
		environments.EnvR2RDockerMode:    os.Getenv(environments.EnvR2RDockerMode),
	}

	// Clear all
	for k := range saved {
		os.Unsetenv(k)
	}

	return func() {
		for k, v := range saved {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}
}

func TestDetect_EnvOverride(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)

	cleanup := ForTesting(t, tempDir)
	defer cleanup()

	ws, err := Detect()
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Root = %q, want %q", ws.Root, tempDir)
	}
	if ws.Source != "env:R2R_REPO_ROOT" {
		t.Errorf("Source = %q, want %q", ws.Source, "env:R2R_REPO_ROOT")
	}
}

func TestDetect_EnvOverride_TakesPrecedenceOverGit(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Create a temp directory that is a valid workspace
	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)

	// Set env to override
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)
	defer os.Unsetenv(environments.EnvR2RRepoRoot)

	ws, err := Detect()
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}

	// Even though we're inside a git repo (the eac repo), the env override takes precedence
	if ws.Source != "env:R2R_REPO_ROOT" {
		t.Errorf("Source = %q, want %q", ws.Source, "env:R2R_REPO_ROOT")
	}
}

func TestDetect_GitFallback(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Don't set any env vars, let it fall through to git detection
	// We're running in the eac repo, so git detection should succeed
	ws, err := Detect()
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}

	if ws.Source != "git" {
		t.Errorf("Source = %q, want %q", ws.Source, "git")
	}

	// Verify .git exists at detected root
	gitPath := filepath.Join(ws.Root, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		t.Errorf(".git not found at detected root %q", ws.Root)
	}
}

func TestDetect_ModeExplicit_FailsWithoutEnv(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	_, err := DetectWithOptions(Options{
		Mode: ModeExplicit,
	})

	if err == nil {
		t.Fatal("DetectWithOptions() expected error but got none")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error should wrap ErrNotFound, got: %v", err)
	}

	var detErr *DetectionError
	if !errors.As(err, &detErr) {
		t.Errorf("error should be *DetectionError, got: %T", err)
	} else {
		if detErr.Source != "env:R2R_REPO_ROOT" {
			t.Errorf("Source = %q, want %q", detErr.Source, "env:R2R_REPO_ROOT")
		}
	}
}

func TestDetect_ModeExplicit_SucceedsWithEnv(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)

	ws, err := DetectWithOptions(Options{
		Mode: ModeExplicit,
	})
	if err != nil {
		t.Fatalf("DetectWithOptions() unexpected error: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Root = %q, want %q", ws.Root, tempDir)
	}
}

func TestDetect_ModeGitOnly_IgnoresEnvOverride(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)

	ws, err := DetectWithOptions(Options{
		Mode: ModeGitOnly,
	})
	if err != nil {
		t.Fatalf("DetectWithOptions() unexpected error: %v", err)
	}

	// Should use git detection, not env override
	if ws.Source != "git" {
		t.Errorf("Source = %q, want %q", ws.Source, "git")
	}

	// Root should NOT be tempDir (unless tempDir happens to be inside a git repo)
	// In practice, it will be the actual eac repo root
	if ws.Root == tempDir {
		t.Errorf("Root should not be tempDir in ModeGitOnly when tempDir is not a real git repo")
	}
}

func TestDetect_ValidationFailsForNonexistent(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	nonexistent := filepath.Join(t.TempDir(), "does-not-exist")
	os.Setenv(environments.EnvR2RRepoRoot, nonexistent)

	_, err := Detect()

	if err == nil {
		t.Fatal("Detect() expected error but got none")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("error should wrap ErrInvalidPath, got: %v", err)
	}

	var detErr *DetectionError
	if errors.As(err, &detErr) {
		if detErr.Op != "validate" {
			t.Errorf("Op = %q, want %q", detErr.Op, "validate")
		}
	}
}

func TestDetect_ValidationFailsForFile(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Create a file, not a directory
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "not-a-dir")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	os.Setenv(environments.EnvR2RRepoRoot, filePath)

	_, err := Detect()

	if err == nil {
		t.Fatal("Detect() expected error but got none")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("error should wrap ErrInvalidPath, got: %v", err)
	}
}

func TestDetect_ValidationFailsForMissingMarkers(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Create empty directory (no .git, no .eac/repository.yml)
	tempDir := t.TempDir()
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)

	_, err := Detect()

	if err == nil {
		t.Fatal("Detect() expected error but got none")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("error should wrap ErrInvalidPath, got: %v", err)
	}
}

func TestDetect_ValidationDisabled(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Create empty directory (no .git, no .eac/repository.yml)
	tempDir := t.TempDir()
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)

	// With validation disabled, should succeed
	ws, err := DetectWithOptions(Options{
		Mode:     ModeAuto,
		Validate: false,
	})

	if err != nil {
		t.Fatalf("DetectWithOptions() unexpected error: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Root = %q, want %q", ws.Root, tempDir)
	}
}

func TestDetect_EacWorkspaceValid(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Create workspace with only .eac/repository.yml (no .git)
	tempDir := t.TempDir()
	createEacWorkspace(t, tempDir)
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)

	ws, err := Detect()
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Root = %q, want %q", ws.Root, tempDir)
	}
}

func TestDetect_RequireGit(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Create workspace with only .eac/repository.yml (no .git)
	tempDir := t.TempDir()
	createEacWorkspace(t, tempDir)
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)

	_, err := DetectWithOptions(Options{
		Mode:       ModeAuto,
		Validate:   true,
		RequireGit: true,
	})

	if err == nil {
		t.Fatal("DetectWithOptions() expected error but got none")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("error should wrap ErrInvalidPath, got: %v", err)
	}
}

func TestDetect_DockerMode(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	os.Setenv(environments.EnvR2RDockerMode, "true")

	ws, err := DetectWithOptions(Options{
		Mode:     ModeAuto,
		Validate: false, // Skip validation since /var/task doesn't exist locally
	})

	if err != nil {
		t.Fatalf("DetectWithOptions() unexpected error: %v", err)
	}

	if ws.Source != "env:environments.EnvR2RDockerMode" {
		t.Errorf("Source = %q, want %q", ws.Source, "env:environments.EnvR2RDockerMode")
	}

	if ws.Root != paths.ContainerRepoRoot {
		t.Errorf("Root = %q, want %q", ws.Root, paths.ContainerRepoRoot)
	}

	if !ws.IsContainer {
		t.Error("IsContainer should be true")
	}
}

func TestDetect_EnvOverrideBeforeDockerMode(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)

	os.Setenv(environments.EnvR2RRepoRoot, tempDir)
	os.Setenv(environments.EnvR2RDockerMode, "true")

	ws, err := Detect()
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}

	// R2R_REPO_ROOT should take precedence over environments.EnvR2RDockerMode
	if ws.Source != "env:R2R_REPO_ROOT" {
		t.Errorf("Source = %q, want %q", ws.Source, "env:R2R_REPO_ROOT")
	}

	if ws.Root != tempDir {
		t.Errorf("Root = %q, want %q", ws.Root, tempDir)
	}
}

func TestDetect_DistRoot(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)

	containerRoot := filepath.Join(tempDir, "container")
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)
	os.Setenv(environments.EnvR2RContainerRoot, containerRoot)

	ws, err := Detect()
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}

	if ws.DistRoot != containerRoot {
		t.Errorf("DistRoot = %q, want %q", ws.DistRoot, containerRoot)
	}
}

func TestDetect_StartPath(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Create a nested directory structure
	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)
	nestedDir := filepath.Join(tempDir, "a", "b", "c")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	// Detect from nested directory
	ws, err := DetectWithOptions(Options{
		Mode:      ModeGitOnly,
		StartPath: nestedDir,
	})
	if err != nil {
		t.Fatalf("DetectWithOptions() unexpected error: %v", err)
	}

	if ws.Root != tempDir {
		t.Errorf("Root = %q, want %q", ws.Root, tempDir)
	}
}

func TestForTesting(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)

	// Before ForTesting, env var should be empty
	if got := os.Getenv(environments.EnvR2RRepoRoot); got != "" {
		t.Errorf("EnvR2RRepoRoot before ForTesting = %q, want empty", got)
	}

	cleanup := ForTesting(t, tempDir)

	// During ForTesting, env var should be set
	if got := os.Getenv(environments.EnvR2RRepoRoot); got != tempDir {
		t.Errorf("EnvR2RRepoRoot during ForTesting = %q, want %q", got, tempDir)
	}

	cleanup()

	// After cleanup, env var should be empty again
	if got := os.Getenv(environments.EnvR2RRepoRoot); got != "" {
		t.Errorf("EnvR2RRepoRoot after cleanup = %q, want empty", got)
	}
}

func TestForTesting_PreservesExistingValue(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	original := "/original/path"
	os.Setenv(environments.EnvR2RRepoRoot, original)

	tempDir := t.TempDir()
	cleanup := ForTesting(t, tempDir)

	// During ForTesting, should be the new value
	if got := os.Getenv(environments.EnvR2RRepoRoot); got != tempDir {
		t.Errorf("EnvR2RRepoRoot during ForTesting = %q, want %q", got, tempDir)
	}

	cleanup()

	// After cleanup, should restore original
	if got := os.Getenv(environments.EnvR2RRepoRoot); got != original {
		t.Errorf("EnvR2RRepoRoot after cleanup = %q, want %q", got, original)
	}
}

func TestRequireIsolation_Fails(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Use a fake test that we can check if it fails
	var fatalCalled bool
	fakeT := &fakeTestingT{
		fatalf: func(format string, args ...any) {
			fatalCalled = true
		},
	}

	RequireIsolation(fakeT)

	if !fatalCalled {
		t.Error("RequireIsolation should have called Fatalf when R2R_REPO_ROOT not set")
	}
}

func TestRequireIsolation_Succeeds(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	os.Setenv(environments.EnvR2RRepoRoot, "/some/path")

	var fatalCalled bool
	fakeT := &fakeTestingT{
		fatalf: func(format string, args ...any) {
			fatalCalled = true
		},
	}

	RequireIsolation(fakeT)

	if fatalCalled {
		t.Error("RequireIsolation should not have called Fatalf when R2R_REPO_ROOT is set")
	}
}

func TestRoot(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)

	root, err := Root()
	if err != nil {
		t.Fatalf("Root() unexpected error: %v", err)
	}

	if root != tempDir {
		t.Errorf("Root() = %q, want %q", root, tempDir)
	}
}

func TestWorkingDir_EnvOverride(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	expected := "/custom/working/dir"
	os.Setenv(environments.EnvR2RPWD, expected)

	got, err := WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() unexpected error: %v", err)
	}

	// filepath.Clean normalizes the path
	if got != filepath.Clean(expected) {
		t.Errorf("WorkingDir() = %q, want %q", got, filepath.Clean(expected))
	}
}

func TestWorkingDir_Fallback(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	got, err := WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() unexpected error: %v", err)
	}

	cwd, _ := os.Getwd()
	if got != cwd {
		t.Errorf("WorkingDir() = %q, want %q", got, cwd)
	}
}

func TestIsInContainer(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	if IsInContainer() {
		t.Error("IsInContainer() should be false when environments.EnvR2RDockerMode not set")
	}

	os.Setenv(environments.EnvR2RDockerMode, "true")
	if !IsInContainer() {
		t.Error("IsInContainer() should be true when environments.EnvR2RDockerMode=true")
	}

	os.Setenv(environments.EnvR2RDockerMode, "false")
	if IsInContainer() {
		t.Error("IsInContainer() should be false when environments.EnvR2RDockerMode=false")
	}
}

func TestDistRoot(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	tempDir := t.TempDir()
	createValidWorkspace(t, tempDir)
	os.Setenv(environments.EnvR2RRepoRoot, tempDir)

	// Without container root, should return workspace root
	got := DistRoot()
	if got != tempDir {
		t.Errorf("DistRoot() = %q, want %q", got, tempDir)
	}

	// With container root, should return container root
	containerRoot := "/container/root"
	os.Setenv(environments.EnvR2RContainerRoot, containerRoot)
	got = DistRoot()
	if got != containerRoot {
		t.Errorf("DistRoot() = %q, want %q", got, containerRoot)
	}
}

func TestDetectionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DetectionError
		contains []string
	}{
		{
			name: "with path",
			err: &DetectionError{
				Op:      "detect",
				Path:    "/some/path",
				Source:  "git",
				Message: "not a git repository",
			},
			contains: []string{"detect", "/some/path", "git", "not a git repository"},
		},
		{
			name: "without path",
			err: &DetectionError{
				Op:      "validate",
				Source:  "env:R2R_REPO_ROOT",
				Message: "not set",
			},
			contains: []string{"validate", "env:R2R_REPO_ROOT", "not set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			for _, substr := range tt.contains {
				if !containsStr(msg, substr) {
					t.Errorf("Error() = %q, should contain %q", msg, substr)
				}
			}
		})
	}
}

func TestDetectionError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &DetectionError{
		Op:      "test",
		Source:  "test",
		Message: "test",
		Err:     inner,
	}

	if !errors.Is(err, inner) {
		t.Error("Unwrap should allow errors.Is to find inner error")
	}
}

func TestMustDetect_Panics(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Set to nonexistent path
	os.Setenv(environments.EnvR2RRepoRoot, "/nonexistent/path")

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustDetect() should panic on error")
		}
	}()

	MustDetect()
}

func TestRootOrPanic_Panics(t *testing.T) {
	restore := saveAndClearEnv(t)
	defer restore()

	// Set to nonexistent path
	os.Setenv(environments.EnvR2RRepoRoot, "/nonexistent/path")

	defer func() {
		if r := recover(); r == nil {
			t.Error("RootOrPanic() should panic on error")
		}
	}()

	RootOrPanic()
}

// fakeTestingT implements the interface needed for RequireIsolation testing.
type fakeTestingT struct {
	fatalf func(format string, args ...any)
}

func (f *fakeTestingT) Helper() {}

func (f *fakeTestingT) Fatalf(format string, args ...any) {
	if f.fatalf != nil {
		f.fatalf(format, args...)
	}
}

// containsStr checks if s contains substr.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
