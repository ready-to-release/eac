package nuget

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewNuGetIsolation(t *testing.T) {
	iso := NewNuGetIsolation("/workspace")
	if iso.workspaceRoot != "/workspace" {
		t.Errorf("workspaceRoot = %q, want /workspace", iso.workspaceRoot)
	}
	if !strings.Contains(iso.workRoot, "nuget") {
		t.Errorf("workRoot = %q, should contain 'nuget'", iso.workRoot)
	}
	if !strings.Contains(iso.nugetCache, "nuget") {
		t.Errorf("nugetCache = %q, should contain 'nuget'", iso.nugetCache)
	}
}

func TestPrepareIsolatedEnv_CreatesWorkDir(t *testing.T) {
	tmpDir := t.TempDir()
	moduleRoot := filepath.Join(tmpDir, "module")
	outputDir := filepath.Join(tmpDir, "output", "test-key")

	// Create module root with a .csproj file
	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moduleRoot, "Test.csproj"), []byte("<Project/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	iso := NewNuGetIsolation(tmpDir)
	env, err := iso.PrepareIsolatedEnv(moduleRoot, outputDir)
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv failed: %v", err)
	}

	if env.WorkDir == "" {
		t.Fatal("WorkDir is empty")
	}
	if _, err := os.Stat(env.WorkDir); os.IsNotExist(err) {
		t.Errorf("WorkDir %q does not exist", env.WorkDir)
	}
	if env.SourceRoot != moduleRoot {
		t.Errorf("SourceRoot = %q, want %q", env.SourceRoot, moduleRoot)
	}
}

func TestPrepareIsolatedEnv_CopiesProjectFiles(t *testing.T) {
	tmpDir := t.TempDir()
	moduleRoot := filepath.Join(tmpDir, "module")
	outputDir := filepath.Join(tmpDir, "output", "copy-test")

	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create various project files
	files := map[string]string{
		"App.csproj":            "<Project/>",
		"App.sln":               "Solution",
		"Directory.Build.props": "<Project/>",
		"global.json":           "{}",
		"README.md":             "Not a project file",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(moduleRoot, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	iso := NewNuGetIsolation(tmpDir)
	env, err := iso.PrepareIsolatedEnv(moduleRoot, outputDir)
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv failed: %v", err)
	}

	// Project files should be copied
	for _, name := range []string{"App.csproj", "App.sln", "Directory.Build.props", "global.json"} {
		if _, err := os.Stat(filepath.Join(env.WorkDir, name)); os.IsNotExist(err) {
			t.Errorf("Expected %s to be copied to work dir", name)
		}
	}

	// README.md should NOT be copied
	if _, err := os.Stat(filepath.Join(env.WorkDir, "README.md")); !os.IsNotExist(err) {
		t.Error("README.md should not be copied to work dir")
	}
}

func TestPrepareIsolatedEnv_SetsNugetPackagesEnv(t *testing.T) {
	tmpDir := t.TempDir()
	moduleRoot := filepath.Join(tmpDir, "module")
	outputDir := filepath.Join(tmpDir, "output", "env-test")

	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moduleRoot, "Test.csproj"), []byte("<Project/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	iso := NewNuGetIsolation(tmpDir)
	env, err := iso.PrepareIsolatedEnv(moduleRoot, outputDir)
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv failed: %v", err)
	}

	found := false
	for _, e := range env.Env {
		if strings.HasPrefix(e, "NUGET_PACKAGES=") {
			found = true
			if !strings.Contains(e, "nuget") {
				t.Errorf("NUGET_PACKAGES = %q, should contain 'nuget'", e)
			}
			break
		}
	}
	if !found {
		t.Error("NUGET_PACKAGES not found in env")
	}
}

func TestPrepareIsolatedEnv_DiscoversSlnFile(t *testing.T) {
	tmpDir := t.TempDir()
	moduleRoot := filepath.Join(tmpDir, "module")
	outputDir := filepath.Join(tmpDir, "output", "sln-test")

	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moduleRoot, "MyApp.sln"), []byte("Solution"), 0o644); err != nil {
		t.Fatal(err)
	}

	iso := NewNuGetIsolation(tmpDir)
	env, err := iso.PrepareIsolatedEnv(moduleRoot, outputDir)
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv failed: %v", err)
	}

	if env.SlnFile != "MyApp.sln" {
		t.Errorf("SlnFile = %q, want MyApp.sln", env.SlnFile)
	}
}

func TestPrepareIsolatedEnv_FallbackKey(t *testing.T) {
	tmpDir := t.TempDir()
	moduleRoot := filepath.Join(tmpDir, "my-module")

	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moduleRoot, "Test.csproj"), []byte("<Project/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	iso := NewNuGetIsolation(tmpDir)
	env, err := iso.PrepareIsolatedEnv(moduleRoot, "")
	if err != nil {
		t.Fatalf("PrepareIsolatedEnv failed: %v", err)
	}

	// Should use module root basename as key
	if !strings.HasSuffix(env.WorkDir, "my-module") {
		t.Errorf("WorkDir = %q, should end with 'my-module'", env.WorkDir)
	}
}

func TestProjectFilesChanged_NoWorkDir(t *testing.T) {
	tmpDir := t.TempDir()
	moduleRoot := filepath.Join(tmpDir, "module")
	workDir := filepath.Join(tmpDir, "nonexistent")

	if err := os.MkdirAll(moduleRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moduleRoot, "App.csproj"), []byte("<Project/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should return true since work dir doesn't exist
	if !projectFilesChanged(moduleRoot, workDir) {
		t.Error("Expected projectFilesChanged to return true when work dir doesn't exist")
	}
}

func TestDiscoverSlnFile_NoSln(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "App.csproj"), []byte("<Project/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := discoverSlnFile(tmpDir)
	if result != "" {
		t.Errorf("discoverSlnFile = %q, want empty", result)
	}
}

func TestSyncDirectory_CopiesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	// Create source files
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file1.cs"), []byte("class A{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "sub", "file2.cs"), []byte("class B{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := syncDirectory(srcDir, dstDir); err != nil {
		t.Fatalf("syncDirectory failed: %v", err)
	}

	// Verify files copied
	if _, err := os.Stat(filepath.Join(dstDir, "file1.cs")); os.IsNotExist(err) {
		t.Error("file1.cs not copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "sub", "file2.cs")); os.IsNotExist(err) {
		t.Error("sub/file2.cs not copied")
	}
}
