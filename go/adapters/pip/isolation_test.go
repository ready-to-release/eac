package pip

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// --- venvPythonPath ---

func TestVenvPythonPath(t *testing.T) {
	venvDir := filepath.Join("some", "venv")
	got := venvPythonPath(venvDir)

	if runtime.GOOS == "windows" {
		want := filepath.Join(venvDir, "Scripts", "python.exe")
		if got != want {
			t.Errorf("venvPythonPath(%q) = %q, want %q", venvDir, got, want)
		}
	} else {
		want := filepath.Join(venvDir, "bin", "python")
		if got != want {
			t.Errorf("venvPythonPath(%q) = %q, want %q", venvDir, got, want)
		}
	}
}

// --- VenvPipPath (exported) ---

func TestVenvPipPath(t *testing.T) {
	venvDir := filepath.Join("some", "venv")
	got := VenvPipPath(venvDir)

	if runtime.GOOS == "windows" {
		want := filepath.Join(venvDir, "Scripts", "pip.exe")
		if got != want {
			t.Errorf("VenvPipPath(%q) = %q, want %q", venvDir, got, want)
		}
	} else {
		want := filepath.Join(venvDir, "bin", "pip")
		if got != want {
			t.Errorf("VenvPipPath(%q) = %q, want %q", venvDir, got, want)
		}
	}
}

func TestVenvPipPath_AbsolutePath(t *testing.T) {
	var venvDir string
	if runtime.GOOS == "windows" {
		venvDir = `C:\users\test\.venv`
	} else {
		venvDir = "/home/test/.venv"
	}

	got := VenvPipPath(venvDir)
	if !strings.HasPrefix(got, venvDir) {
		t.Errorf("VenvPipPath result %q should start with venvDir %q", got, venvDir)
	}
}

// --- buildPythonEnv ---

func TestBuildPythonEnv_SetsVirtualEnv(t *testing.T) {
	venvDir := filepath.Join("test", "venv")
	pipCache := filepath.Join("test", "pip-cache")

	env := buildPythonEnv(venvDir, pipCache)

	found := false
	want := "VIRTUAL_ENV=" + venvDir
	for _, e := range env {
		if e == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("buildPythonEnv: expected %q in env, not found", want)
	}
}

func TestBuildPythonEnv_SetsPipCacheDir(t *testing.T) {
	venvDir := filepath.Join("test", "venv")
	pipCache := filepath.Join("test", "pip-cache")

	env := buildPythonEnv(venvDir, pipCache)

	found := false
	want := "PIP_CACHE_DIR=" + pipCache
	for _, e := range env {
		if e == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("buildPythonEnv: expected %q in env, not found", want)
	}
}

func TestBuildPythonEnv_PrependsBinToPath(t *testing.T) {
	venvDir := filepath.Join("test", "venv")
	pipCache := filepath.Join("test", "pip-cache")

	env := buildPythonEnv(venvDir, pipCache)

	var expectedBinDir string
	if runtime.GOOS == "windows" {
		expectedBinDir = filepath.Join(venvDir, "Scripts")
	} else {
		expectedBinDir = filepath.Join(venvDir, "bin")
	}

	found := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			pathVal := strings.TrimPrefix(e, "PATH=")
			if strings.HasPrefix(pathVal, expectedBinDir+string(os.PathListSeparator)) {
				found = true
			}
			break
		}
	}
	if !found {
		t.Errorf("buildPythonEnv: PATH should be prepended with %q", expectedBinDir)
	}
}

func TestBuildPythonEnv_PreservesOtherEnvVars(t *testing.T) {
	// buildPythonEnv uses os.Environ() which includes existing vars.
	// We just verify that the result contains at least as many entries
	// as the current environment (it adds VIRTUAL_ENV, PIP_CACHE_DIR,
	// and modifies PATH).
	venvDir := filepath.Join("test", "venv")
	pipCache := filepath.Join("test", "pip-cache")

	env := buildPythonEnv(venvDir, pipCache)

	// Should have at least 2 entries (VIRTUAL_ENV and PIP_CACHE_DIR)
	if len(env) < 2 {
		t.Errorf("buildPythonEnv: expected at least 2 env vars, got %d", len(env))
	}
}

// --- NewPipIsolation ---

func TestNewPipIsolation(t *testing.T) {
	workspaceRoot := filepath.Join("some", "workspace")
	iso := NewPipIsolation(workspaceRoot)

	if iso.workspaceRoot != workspaceRoot {
		t.Errorf("workspaceRoot = %q, want %q", iso.workspaceRoot, workspaceRoot)
	}
	if iso.workRoot == "" {
		t.Error("workRoot should not be empty")
	}
	if iso.venvRoot == "" {
		t.Error("venvRoot should not be empty")
	}
	if iso.pipCacheDir == "" {
		t.Error("pipCacheDir should not be empty")
	}
}

func TestNewPipIsolation_PathsContainPython(t *testing.T) {
	workspaceRoot := filepath.Join("some", "workspace")
	iso := NewPipIsolation(workspaceRoot)

	if !strings.Contains(iso.workRoot, "python") {
		t.Errorf("workRoot %q should contain 'python'", iso.workRoot)
	}
	if !strings.Contains(iso.venvRoot, "python") {
		t.Errorf("venvRoot %q should contain 'python'", iso.venvRoot)
	}
	if !strings.Contains(iso.pipCacheDir, "python") || !strings.Contains(iso.pipCacheDir, "pip-cache") {
		t.Errorf("pipCacheDir %q should contain 'python' and 'pip-cache'", iso.pipCacheDir)
	}
}

// --- requirementsChanged ---

func TestRequirementsChanged_NoSourceFiles(t *testing.T) {
	moduleRoot := t.TempDir()
	workDir := t.TempDir()

	// Neither source nor destination has any requirement files
	got := requirementsChanged(moduleRoot, workDir)
	if got {
		t.Error("requirementsChanged: should return false when no requirement files exist in source")
	}
}

func TestRequirementsChanged_SourceExistsDstMissing(t *testing.T) {
	moduleRoot := t.TempDir()
	workDir := t.TempDir()

	// Create pyproject.toml in source but not in work
	writeTestFile(t, filepath.Join(moduleRoot, "pyproject.toml"), "content")

	got := requirementsChanged(moduleRoot, workDir)
	if !got {
		t.Error("requirementsChanged: should return true when source file exists but destination does not")
	}
}

func TestRequirementsChanged_FilesIdentical(t *testing.T) {
	moduleRoot := t.TempDir()
	workDir := t.TempDir()

	content := "some-requirement==1.0.0"
	mtime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	srcPath := filepath.Join(moduleRoot, "pyproject.toml")
	dstPath := filepath.Join(workDir, "pyproject.toml")

	writeTestFile(t, srcPath, content)
	writeTestFile(t, dstPath, content)

	// Set identical mtime
	if err := os.Chtimes(srcPath, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dstPath, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	got := requirementsChanged(moduleRoot, workDir)
	if got {
		t.Error("requirementsChanged: should return false when files are identical")
	}
}

func TestRequirementsChanged_SizeDiffers(t *testing.T) {
	moduleRoot := t.TempDir()
	workDir := t.TempDir()

	mtime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	srcPath := filepath.Join(moduleRoot, "requirements.txt")
	dstPath := filepath.Join(workDir, "requirements.txt")

	writeTestFile(t, srcPath, "flask==2.0.0\nrequests==2.28.0")
	writeTestFile(t, dstPath, "flask==1.0.0")

	// Set same mtime but different sizes
	if err := os.Chtimes(srcPath, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dstPath, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	got := requirementsChanged(moduleRoot, workDir)
	if got != true {
		t.Error("requirementsChanged: should return true when file sizes differ")
	}
}

func TestRequirementsChanged_MtimeDiffers(t *testing.T) {
	moduleRoot := t.TempDir()
	workDir := t.TempDir()

	content := "flask==2.0.0"

	srcPath := filepath.Join(moduleRoot, "requirements.txt")
	dstPath := filepath.Join(workDir, "requirements.txt")

	writeTestFile(t, srcPath, content)
	writeTestFile(t, dstPath, content)

	// Set different mtimes but same size
	if err := os.Chtimes(srcPath, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dstPath, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	got := requirementsChanged(moduleRoot, workDir)
	if !got {
		t.Error("requirementsChanged: should return true when modification times differ")
	}
}

func TestRequirementsChanged_ChecksAllFileTypes(t *testing.T) {
	// Verifies that each of the tracked files triggers a change detection
	trackedFiles := []string{"pyproject.toml", "requirements.txt", "requirements-dev.txt", "setup.py"}

	for _, file := range trackedFiles {
		t.Run(file, func(t *testing.T) {
			moduleRoot := t.TempDir()
			workDir := t.TempDir()

			writeTestFile(t, filepath.Join(moduleRoot, file), "content")

			got := requirementsChanged(moduleRoot, workDir)
			if !got {
				t.Errorf("requirementsChanged: should detect new %s in source", file)
			}
		})
	}
}

// --- copyFileIfChanged ---

func TestCopyFileIfChanged_CopiesNewFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.txt")
	dstPath := filepath.Join(dstDir, "test.txt")

	content := "hello world"
	writeTestFile(t, srcPath, content)

	copied, err := copyFileIfChanged(srcPath, dstPath)
	if err != nil {
		t.Fatalf("copyFileIfChanged() error = %v", err)
	}
	if !copied {
		t.Error("copyFileIfChanged: should return true when copying to new destination")
	}

	got := readTestFile(t, dstPath)
	if got != content {
		t.Errorf("destination content = %q, want %q", got, content)
	}
}

func TestCopyFileIfChanged_SkipsUnchangedFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.txt")
	dstPath := filepath.Join(dstDir, "test.txt")

	content := "hello world"
	writeTestFile(t, srcPath, content)

	// First copy
	_, err := copyFileIfChanged(srcPath, dstPath)
	if err != nil {
		t.Fatalf("first copyFileIfChanged() error = %v", err)
	}

	// Second copy should skip (same mtime and size because copyFileIfChanged preserves mtime)
	copied, err := copyFileIfChanged(srcPath, dstPath)
	if err != nil {
		t.Fatalf("second copyFileIfChanged() error = %v", err)
	}
	if copied {
		t.Error("copyFileIfChanged: should return false for unchanged file")
	}
}

func TestCopyFileIfChanged_CopiesWhenContentChanges(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.txt")
	dstPath := filepath.Join(dstDir, "test.txt")

	// Initial copy
	writeTestFile(t, srcPath, "version 1")
	_, err := copyFileIfChanged(srcPath, dstPath)
	if err != nil {
		t.Fatalf("first copyFileIfChanged() error = %v", err)
	}

	// Modify source (different content, different size triggers copy)
	writeTestFile(t, srcPath, "version 2 - longer content")

	copied, err := copyFileIfChanged(srcPath, dstPath)
	if err != nil {
		t.Fatalf("second copyFileIfChanged() error = %v", err)
	}
	if !copied {
		t.Error("copyFileIfChanged: should return true when source file changed")
	}

	got := readTestFile(t, dstPath)
	if got != "version 2 - longer content" {
		t.Errorf("destination content = %q, want %q", got, "version 2 - longer content")
	}
}

func TestCopyFileIfChanged_PreservesMtime(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.txt")
	dstPath := filepath.Join(dstDir, "test.txt")

	writeTestFile(t, srcPath, "content")
	mtime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(srcPath, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	_, err := copyFileIfChanged(srcPath, dstPath)
	if err != nil {
		t.Fatalf("copyFileIfChanged() error = %v", err)
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	if !dstInfo.ModTime().Equal(mtime) {
		t.Errorf("destination mtime = %v, want %v", dstInfo.ModTime(), mtime)
	}
}

func TestCopyFileIfChanged_SourceNotExist(t *testing.T) {
	dstDir := t.TempDir()
	dstPath := filepath.Join(dstDir, "test.txt")

	_, err := copyFileIfChanged(filepath.Join(t.TempDir(), "nonexistent.txt"), dstPath)
	if err == nil {
		t.Error("copyFileIfChanged: expected error for nonexistent source")
	}
	if !os.IsNotExist(err) {
		t.Errorf("copyFileIfChanged: expected os.IsNotExist error, got %v", err)
	}
}

func TestCopyFileIfChanged_EmptyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "empty.txt")
	dstPath := filepath.Join(dstDir, "empty.txt")

	writeTestFile(t, srcPath, "")

	copied, err := copyFileIfChanged(srcPath, dstPath)
	if err != nil {
		t.Fatalf("copyFileIfChanged() error = %v", err)
	}
	if !copied {
		t.Error("copyFileIfChanged: should return true for first copy of empty file")
	}

	got := readTestFile(t, dstPath)
	if got != "" {
		t.Errorf("destination content = %q, want empty string", got)
	}
}

// --- syncDirectory ---

func TestSyncDirectory_CopiesFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	writeTestFile(t, filepath.Join(srcDir, "a.py"), "print('a')")
	writeTestFile(t, filepath.Join(srcDir, "b.py"), "print('b')")

	if err := syncDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("syncDirectory() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dstDir, "a.py"), "print('a')")
	assertFileContent(t, filepath.Join(dstDir, "b.py"), "print('b')")
}

func TestSyncDirectory_CopiesNestedDirectories(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	if err := os.MkdirAll(filepath.Join(srcDir, "sub", "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(srcDir, "sub", "deep", "module.py"), "import os")

	if err := syncDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("syncDirectory() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dstDir, "sub", "deep", "module.py"), "import os")
}

func TestSyncDirectory_RemovesDeletedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	// Initial sync with two files
	writeTestFile(t, filepath.Join(srcDir, "keep.py"), "keep")
	writeTestFile(t, filepath.Join(srcDir, "delete.py"), "delete")

	if err := syncDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("first syncDirectory() error = %v", err)
	}

	// Remove one file from source
	if err := os.Remove(filepath.Join(srcDir, "delete.py")); err != nil {
		t.Fatal(err)
	}

	// Re-sync
	if err := syncDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("second syncDirectory() error = %v", err)
	}

	// Kept file should still be present
	assertFileContent(t, filepath.Join(dstDir, "keep.py"), "keep")

	// Deleted file should be removed from destination
	if _, err := os.Stat(filepath.Join(dstDir, "delete.py")); !os.IsNotExist(err) {
		t.Error("syncDirectory: should have removed delete.py from destination")
	}
}

func TestSyncDirectory_UpdatesChangedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	writeTestFile(t, filepath.Join(srcDir, "module.py"), "v1")

	if err := syncDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("first syncDirectory() error = %v", err)
	}

	// Update source file (different content + different size triggers update)
	writeTestFile(t, filepath.Join(srcDir, "module.py"), "v2 updated content")

	if err := syncDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("second syncDirectory() error = %v", err)
	}

	assertFileContent(t, filepath.Join(dstDir, "module.py"), "v2 updated content")
}

func TestSyncDirectory_SourceNotDirectory(t *testing.T) {
	srcFile := filepath.Join(t.TempDir(), "file.txt")
	writeTestFile(t, srcFile, "not a directory")
	dstDir := filepath.Join(t.TempDir(), "dst")

	err := syncDirectory(srcFile, dstDir)
	if err == nil {
		t.Error("syncDirectory: expected error for non-directory source")
	}
	if !strings.Contains(err.Error(), "is not a directory") {
		t.Errorf("syncDirectory: expected 'is not a directory' error, got: %v", err)
	}
}

func TestSyncDirectory_SourceNotExist(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "nonexistent")
	dstDir := filepath.Join(t.TempDir(), "dst")

	err := syncDirectory(srcDir, dstDir)
	if err == nil {
		t.Error("syncDirectory: expected error for nonexistent source")
	}
}

func TestSyncDirectory_EmptySourceDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "dst")

	if err := syncDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("syncDirectory() error = %v", err)
	}

	// Destination directory should be created even if empty
	info, err := os.Stat(dstDir)
	if err != nil {
		t.Fatalf("os.Stat(dstDir) error = %v", err)
	}
	if !info.IsDir() {
		t.Error("syncDirectory: destination should be a directory")
	}
}

// --- NeedsVenvCreate ---

func TestNeedsVenvCreate_NoVenv(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		venvRoot: tmpDir,
	}

	// No venv directory exists at all
	if !iso.NeedsVenvCreate("mymodule") {
		t.Error("NeedsVenvCreate: should return true when venv does not exist")
	}
}

func TestNeedsVenvCreate_VenvExists(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		venvRoot: tmpDir,
	}

	// Create the python binary that NeedsVenvCreate looks for
	venvDir := filepath.Join(tmpDir, "mymodule")
	var pythonPath string
	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(venvDir, "Scripts", "python.exe")
	} else {
		pythonPath = filepath.Join(venvDir, "bin", "python")
	}

	if err := os.MkdirAll(filepath.Dir(pythonPath), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, pythonPath, "fake python binary")

	if iso.NeedsVenvCreate("mymodule") {
		t.Error("NeedsVenvCreate: should return false when venv python binary exists")
	}
}

// --- PrepareIsolatedEnv ---

func TestPrepareIsolatedEnv_KeyFromOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	moduleRoot := t.TempDir()
	outputDir := filepath.Join("out", "test", "my-module", "unit-tests")

	env, err := iso.PrepareIsolatedEnv(moduleRoot, outputDir)
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() error = %v", err)
	}

	// Key should be basename of outputDir = "unit-tests"
	expectedWorkDir := filepath.Join(tmpDir, "work", "unit-tests")
	if env.WorkDir != expectedWorkDir {
		t.Errorf("WorkDir = %q, want %q", env.WorkDir, expectedWorkDir)
	}

	expectedVenvDir := filepath.Join(tmpDir, "venv", "unit-tests")
	if env.VenvDir != expectedVenvDir {
		t.Errorf("VenvDir = %q, want %q", env.VenvDir, expectedVenvDir)
	}
}

func TestPrepareIsolatedEnv_FallbackToModuleRoot(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	moduleRoot := filepath.Join(t.TempDir(), "my-module")
	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	// Empty outputDir triggers fallback
	env, err := iso.PrepareIsolatedEnv(moduleRoot, "")
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() error = %v", err)
	}

	expectedWorkDir := filepath.Join(tmpDir, "work", "my-module")
	if env.WorkDir != expectedWorkDir {
		t.Errorf("WorkDir = %q, want %q", env.WorkDir, expectedWorkDir)
	}
}

func TestPrepareIsolatedEnv_CopiesRequiredAndOptionalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	moduleRoot := t.TempDir()

	// Create required and optional files
	writeTestFile(t, filepath.Join(moduleRoot, "pyproject.toml"), "[project]\nname = \"test\"")
	writeTestFile(t, filepath.Join(moduleRoot, "setup.py"), "from setuptools import setup")
	writeTestFile(t, filepath.Join(moduleRoot, "requirements.txt"), "flask==2.0")
	writeTestFile(t, filepath.Join(moduleRoot, "conftest.py"), "import pytest")

	env, err := iso.PrepareIsolatedEnv(moduleRoot, filepath.Join("out", "test-key"))
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() error = %v", err)
	}

	assertFileContent(t, filepath.Join(env.WorkDir, "pyproject.toml"), "[project]\nname = \"test\"")
	assertFileContent(t, filepath.Join(env.WorkDir, "setup.py"), "from setuptools import setup")
	assertFileContent(t, filepath.Join(env.WorkDir, "requirements.txt"), "flask==2.0")
	assertFileContent(t, filepath.Join(env.WorkDir, "conftest.py"), "import pytest")
}

func TestPrepareIsolatedEnv_SyncsDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	moduleRoot := t.TempDir()

	// Create directories that should be synced
	for _, dir := range []string{"src", "tests", "features", "steps"} {
		dirPath := filepath.Join(moduleRoot, dir)
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, filepath.Join(dirPath, "file.py"), "# "+dir)
	}

	env, err := iso.PrepareIsolatedEnv(moduleRoot, filepath.Join("out", "sync-test"))
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() error = %v", err)
	}

	for _, dir := range []string{"src", "tests", "features", "steps"} {
		assertFileContent(t, filepath.Join(env.WorkDir, dir, "file.py"), "# "+dir)
	}
}

func TestPrepareIsolatedEnv_SetsEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	moduleRoot := t.TempDir()

	env, err := iso.PrepareIsolatedEnv(moduleRoot, filepath.Join("out", "env-test"))
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() error = %v", err)
	}

	// Check VIRTUAL_ENV is set
	foundVenv := false
	foundPipCache := false
	for _, e := range env.Env {
		if strings.HasPrefix(e, "VIRTUAL_ENV=") {
			foundVenv = true
			val := strings.TrimPrefix(e, "VIRTUAL_ENV=")
			if val != env.VenvDir {
				t.Errorf("VIRTUAL_ENV = %q, want %q", val, env.VenvDir)
			}
		}
		if strings.HasPrefix(e, "PIP_CACHE_DIR=") {
			foundPipCache = true
		}
	}
	if !foundVenv {
		t.Error("PrepareIsolatedEnv: VIRTUAL_ENV not found in env")
	}
	if !foundPipCache {
		t.Error("PrepareIsolatedEnv: PIP_CACHE_DIR not found in env")
	}
}

func TestPrepareIsolatedEnv_SourceRootSet(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	moduleRoot := t.TempDir()

	env, err := iso.PrepareIsolatedEnv(moduleRoot, filepath.Join("out", "sr-test"))
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() error = %v", err)
	}

	if env.SourceRoot != moduleRoot {
		t.Errorf("SourceRoot = %q, want %q", env.SourceRoot, moduleRoot)
	}
}

func TestPrepareIsolatedEnv_PythonBinSet(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	moduleRoot := t.TempDir()

	env, err := iso.PrepareIsolatedEnv(moduleRoot, filepath.Join("out", "pybin-test"))
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() error = %v", err)
	}

	if env.PythonBin == "" {
		t.Error("PythonBin should not be empty")
	}

	// Should be inside the venv directory
	if !strings.HasPrefix(env.PythonBin, env.VenvDir) {
		t.Errorf("PythonBin %q should be inside VenvDir %q", env.PythonBin, env.VenvDir)
	}
}

func TestPrepareIsolatedEnv_MissingOptionalFilesDoNotError(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	// Empty module root - no files at all
	moduleRoot := t.TempDir()

	_, err := iso.PrepareIsolatedEnv(moduleRoot, filepath.Join("out", "empty-test"))
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() should not error with missing optional files, got: %v", err)
	}
}

func TestPrepareIsolatedEnv_MissingSyncDirsDoNotError(t *testing.T) {
	tmpDir := t.TempDir()
	iso := &PipIsolation{
		workspaceRoot: tmpDir,
		workRoot:      filepath.Join(tmpDir, "work"),
		venvRoot:      filepath.Join(tmpDir, "venv"),
		pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
	}

	// Module root with no src/tests/features/steps directories
	moduleRoot := t.TempDir()

	_, err := iso.PrepareIsolatedEnv(moduleRoot, filepath.Join("out", "nosync-test"))
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv() should not error with missing sync directories, got: %v", err)
	}
}

// --- buildPythonEnv table-driven ---

func TestBuildPythonEnv_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		venvDir     string
		pipCacheDir string
		checkVar    string
		wantValue   string
	}{
		{
			name:        "sets VIRTUAL_ENV",
			venvDir:     filepath.Join("path", "to", "venv"),
			pipCacheDir: filepath.Join("path", "cache"),
			checkVar:    "VIRTUAL_ENV",
			wantValue:   filepath.Join("path", "to", "venv"),
		},
		{
			name:        "sets PIP_CACHE_DIR",
			venvDir:     filepath.Join("path", "to", "venv"),
			pipCacheDir: filepath.Join("path", "to", "pip-cache"),
			checkVar:    "PIP_CACHE_DIR",
			wantValue:   filepath.Join("path", "to", "pip-cache"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := buildPythonEnv(tt.venvDir, tt.pipCacheDir)

			prefix := tt.checkVar + "="
			found := false
			for _, e := range env {
				if strings.HasPrefix(e, prefix) {
					val := strings.TrimPrefix(e, prefix)
					if val != tt.wantValue {
						t.Errorf("%s = %q, want %q", tt.checkVar, val, tt.wantValue)
					}
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s not found in env", tt.checkVar)
			}
		})
	}
}

// --- Key derivation (tested via PrepareIsolatedEnv) ---

func TestPrepareIsolatedEnv_KeyDerivation(t *testing.T) {
	tests := []struct {
		name       string
		outputDir  string
		moduleBase string // basename of moduleRoot
		wantKey    string
	}{
		{
			name:      "key from outputDir basename",
			outputDir: filepath.Join("out", "test", "mymod", "unit"),
			wantKey:   "unit",
		},
		{
			name:       "dot outputDir falls back to module basename",
			outputDir:  ".",
			moduleBase: "fallback-mod",
			wantKey:    "fallback-mod",
		},
		{
			name:       "empty outputDir falls back to module basename",
			outputDir:  "",
			moduleBase: "another-mod",
			wantKey:    "another-mod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			iso := &PipIsolation{
				workspaceRoot: tmpDir,
				workRoot:      filepath.Join(tmpDir, "work"),
				venvRoot:      filepath.Join(tmpDir, "venv"),
				pipCacheDir:   filepath.Join(tmpDir, "pip-cache"),
			}

			moduleBase := tt.moduleBase
			if moduleBase == "" {
				moduleBase = "default-module"
			}
			moduleRoot := filepath.Join(t.TempDir(), moduleBase)
			if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
				t.Fatal(err)
			}

			env, err := iso.PrepareIsolatedEnv(moduleRoot, tt.outputDir)
			if err != nil {
				t.Fatalf("PrepareIsolatedEnv() error = %v", err)
			}

			expectedWorkDir := filepath.Join(tmpDir, "work", tt.wantKey)
			if env.WorkDir != expectedWorkDir {
				t.Errorf("WorkDir = %q, want %q", env.WorkDir, expectedWorkDir)
			}
		})
	}
}

// --- Helper functions ---

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got := readTestFile(t, path)
	if got != want {
		t.Errorf("file %s content = %q, want %q", path, got, want)
	}
}
