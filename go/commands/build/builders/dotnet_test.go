package builders

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDotnetHandler_Name(t *testing.T) {
	h := &DotnetHandler{}
	if h.Name() != "dotnet" {
		t.Errorf("expected name 'dotnet', got '%s'", h.Name())
	}
}

func TestDotnetHandler_Capabilities(t *testing.T) {
	h := &DotnetHandler{}
	caps := h.Capabilities()

	expected := []string{"dotnet_module", "cross_compile"}
	if len(caps) != len(expected) {
		t.Fatalf("expected %d capabilities, got %d", len(expected), len(caps))
	}

	for i, cap := range expected {
		if caps[i] != cap {
			t.Errorf("expected capability[%d] = '%s', got '%s'", i, cap, caps[i])
		}
	}
}

func TestDotnetHandler_Requirements(t *testing.T) {
	h := &DotnetHandler{}
	reqs := h.Requirements()

	found := false
	for _, req := range reqs {
		if req == "dotnet" {
			found = true
			break
		}
	}
	if !found {
		t.Error("DotnetHandler requirements should include 'dotnet'")
	}
}

func TestDotnetHandler_IsContainer(t *testing.T) {
	h := &DotnetHandler{}
	if h.IsContainer() {
		t.Error("DotnetHandler should not be a container handler")
	}
}

func TestDotnetHandler_IsHostInstalled(t *testing.T) {
	h := &DotnetHandler{}
	if !h.IsHostInstalled() {
		t.Error("DotnetHandler should report as host-installed")
	}
}

// mockModuleForValidation implements core.ModuleContractPort for ValidateModule testing.
type mockModuleForValidation struct {
	componentRoot string
}

func (m *mockModuleForValidation) GetMoniker() string                              { return "test" }
func (m *mockModuleForValidation) GetName() string                                 { return "Test" }
func (m *mockModuleForValidation) GetDescription() string                          { return "" }
func (m *mockModuleForValidation) GetModuleGroup() string                          { return "" }
func (m *mockModuleForValidation) HasComponent(componentType string) bool          { return true }
func (m *mockModuleForValidation) GetComponentRoot(componentType string) string     { return m.componentRoot }
func (m *mockModuleForValidation) GetComponentRoots() map[string]string            { return nil }
func (m *mockModuleForValidation) GetComponentTypesDisplay() string                { return "" }
func (m *mockModuleForValidation) GetComponentAmp(string, string) float64          { return 1.0 }
func (m *mockModuleForValidation) GetComponentGroup(string) string                 { return "" }
func (m *mockModuleForValidation) GetDependsOn() []string                          { return nil }
func (m *mockModuleForValidation) GetVersioningScheme() string                     { return "" }
func (m *mockModuleForValidation) GetReleaseType() string                          { return "" }
func (m *mockModuleForValidation) GetChangelog() string                            { return "" }
func (m *mockModuleForValidation) HasVersioning() bool                             { return false }
func (m *mockModuleForValidation) GetMetadata() map[string]interface{}             { return nil }
func (m *mockModuleForValidation) GetContentHash() (string, error)                 { return "", nil }

func TestDotnetHandler_ValidateModule_WithCsproj(t *testing.T) {
	tmpDir := t.TempDir()
	compDir := filepath.Join(tmpDir, "src", "MyApp")
	os.MkdirAll(compDir, 0o755)
	os.WriteFile(filepath.Join(compDir, "MyApp.csproj"), []byte("<Project/>"), 0o644)

	h := &DotnetHandler{}
	module := &mockModuleForValidation{componentRoot: "src/MyApp"}

	err := h.ValidateModule(module, tmpDir, "dotnet")
	if err != nil {
		t.Errorf("expected no error for .csproj, got: %v", err)
	}
}

func TestDotnetHandler_ValidateModule_WithFsproj(t *testing.T) {
	tmpDir := t.TempDir()
	compDir := filepath.Join(tmpDir, "src", "MyApp")
	os.MkdirAll(compDir, 0o755)
	os.WriteFile(filepath.Join(compDir, "MyApp.fsproj"), []byte("<Project/>"), 0o644)

	h := &DotnetHandler{}
	module := &mockModuleForValidation{componentRoot: "src/MyApp"}

	err := h.ValidateModule(module, tmpDir, "dotnet")
	if err != nil {
		t.Errorf("expected no error for .fsproj, got: %v", err)
	}
}

func TestDotnetHandler_ValidateModule_WithSln(t *testing.T) {
	tmpDir := t.TempDir()
	compDir := filepath.Join(tmpDir, "src", "MyApp")
	os.MkdirAll(compDir, 0o755)
	os.WriteFile(filepath.Join(compDir, "MyApp.sln"), []byte(""), 0o644)

	h := &DotnetHandler{}
	module := &mockModuleForValidation{componentRoot: "src/MyApp"}

	err := h.ValidateModule(module, tmpDir, "dotnet")
	if err != nil {
		t.Errorf("expected no error for .sln, got: %v", err)
	}
}

func TestDotnetHandler_ValidateModule_SlnInParent(t *testing.T) {
	tmpDir := t.TempDir()
	compDir := filepath.Join(tmpDir, "src", "MyApp")
	os.MkdirAll(compDir, 0o755)
	// .sln at workspace root, component is nested
	os.WriteFile(filepath.Join(tmpDir, "Solution.sln"), []byte(""), 0o644)

	h := &DotnetHandler{}
	module := &mockModuleForValidation{componentRoot: "src/MyApp"}

	err := h.ValidateModule(module, tmpDir, "dotnet")
	if err != nil {
		t.Errorf("expected no error for .sln in parent dir, got: %v", err)
	}
}

func TestDotnetHandler_ValidateModule_NoProjectFile(t *testing.T) {
	tmpDir := t.TempDir()
	compDir := filepath.Join(tmpDir, "src", "MyApp")
	os.MkdirAll(compDir, 0o755)

	h := &DotnetHandler{}
	module := &mockModuleForValidation{componentRoot: "src/MyApp"}

	err := h.ValidateModule(module, tmpDir, "dotnet")
	if err == nil {
		t.Error("expected error when no project files exist, got nil")
	}
}

func TestDotnetRIDFromArtifactID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Direct .NET RIDs
		{"linux-x64", "linux-x64", "linux-x64"},
		{"linux-arm64", "linux-arm64", "linux-arm64"},
		{"win-x64", "win-x64", "win-x64"},
		{"win-arm64", "win-arm64", "win-arm64"},
		{"osx-x64", "osx-x64", "osx-x64"},
		{"osx-arm64", "osx-arm64", "osx-arm64"},
		// Go-style IDs mapped to .NET RIDs
		{"linux-amd64 to linux-x64", "linux-amd64", "linux-x64"},
		{"windows-amd64 to win-x64", "windows-amd64", "win-x64"},
		{"darwin-amd64 to osx-x64", "darwin-amd64", "osx-x64"},
		{"darwin-arm64 to osx-arm64", "darwin-arm64", "osx-arm64"},
		// Unknown with dash passed through
		{"unknown-with-dash", "freebsd-x64", "freebsd-x64"},
		// Unknown without dash returns empty
		{"no-dash", "unknown", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dotnetRIDFromArtifactID(tc.input)
			if got != tc.expected {
				t.Errorf("dotnetRIDFromArtifactID(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestDotnetHandler_BuildArgsForLibrary(t *testing.T) {
	args := []string{"build", "--no-restore", "--configuration", "Release"}

	expected := []string{"build", "--no-restore", "--configuration", "Release"}
	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d", len(expected), len(args))
	}
	for i, arg := range expected {
		if args[i] != arg {
			t.Errorf("arg[%d] = %q, want %q", i, args[i], arg)
		}
	}
}

func TestDotnetHandler_BuildArgsForPublish(t *testing.T) {
	tests := []struct {
		name     string
		rid      string
		outDir   string
		wantArgs []string
	}{
		{
			name:   "linux-x64 publish",
			rid:    "linux-x64",
			outDir: "/out/linux-x64",
			wantArgs: []string{
				"publish", "--no-restore", "--configuration", "Release",
				"--runtime", "linux-x64", "--self-contained", "true",
				"--output", "/out/linux-x64",
			},
		},
		{
			name:   "win-x64 publish",
			rid:    "win-x64",
			outDir: "/out/win-x64",
			wantArgs: []string{
				"publish", "--no-restore", "--configuration", "Release",
				"--runtime", "win-x64", "--self-contained", "true",
				"--output", "/out/win-x64",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{
				"publish", "--no-restore", "--configuration", "Release",
				"--runtime", tc.rid, "--self-contained", "true",
				"--output", tc.outDir,
			}

			if len(args) != len(tc.wantArgs) {
				t.Fatalf("got %d args, want %d", len(args), len(tc.wantArgs))
			}
			for i, arg := range tc.wantArgs {
				if args[i] != arg {
					t.Errorf("arg[%d] = %q, want %q", i, args[i], arg)
				}
			}
		})
	}
}
