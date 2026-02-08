package ownership

import (
	"testing"
)

// TestResolve_SimpleRoot verifies basic root ownership.
func TestResolve_SimpleRoot(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "core",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/core"},
			},
		},
	})

	owner := r.Resolve("go/core/config/modules.go")
	if owner == nil {
		t.Fatal("expected owner, got nil")
	}
	if owner.Module != "core" || owner.Component != "go" {
		t.Errorf("expected core:go, got %s:%s", owner.Module, owner.Component)
	}
}

// TestResolve_SubOwnership verifies most-specific-root wins within a module.
func TestResolve_SubOwnership(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "eac-cli",
			Components: map[string]*ComponentOwnership{
				"go":        {Root: "go/cli/eac"},
				"testdata":  {Root: "go/cli/eac/testdata", ComponentType: "testdata"},
				"gherkin":   {Root: "go/cli/eac/specs"},
				"assets":    {Root: "go/cli/eac/specs/assets", ComponentType: "testdata"},
			},
		},
	})

	tests := []struct {
		path      string
		wantComp  string
	}{
		{"go/cli/eac/main.go", "go"},
		{"go/cli/eac/testdata/sample.txt", "testdata"},
		{"go/cli/eac/specs/godog_test.go", "gherkin"},
		{"go/cli/eac/specs/assets/test.json", "assets"},
	}

	for _, tc := range tests {
		owner := r.Resolve(tc.path)
		if owner == nil {
			t.Errorf("Resolve(%q): expected owner, got nil", tc.path)
			continue
		}
		if owner.Component != tc.wantComp {
			t.Errorf("Resolve(%q): expected component %q, got %q", tc.path, tc.wantComp, owner.Component)
		}
		if owner.Module != "eac-cli" {
			t.Errorf("Resolve(%q): expected module eac-cli, got %s", tc.path, owner.Module)
		}
	}
}

// TestResolve_FileOwnership verifies single-file ownership.
func TestResolve_FileOwnership(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "implicit-cli",
			Components: map[string]*ComponentOwnership{
				"pwsh":          {Root: "scripts/pwsh/go-invoker"},
				"pwsh-importer": {File: "importer.ps1", ComponentType: "pwsh"},
			},
		},
	})

	// File ownership matches exact file
	owner := r.Resolve("importer.ps1")
	if owner == nil {
		t.Fatal("expected owner for importer.ps1, got nil")
	}
	if owner.Component != "pwsh-importer" {
		t.Errorf("expected pwsh-importer, got %s", owner.Component)
	}

	// Root ownership still works
	owner = r.Resolve("scripts/pwsh/go-invoker/main.ps1")
	if owner == nil {
		t.Fatal("expected owner for main.ps1, got nil")
	}
	if owner.Component != "pwsh" {
		t.Errorf("expected pwsh, got %s", owner.Component)
	}

	// Unrelated file returns nil
	owner = r.Resolve("some/other/file.txt")
	if owner != nil {
		t.Errorf("expected nil for unrelated file, got %s:%s", owner.Module, owner.Component)
	}
}

// TestResolve_VirtualComponent verifies components with no root/file are skipped.
func TestResolve_VirtualComponent(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "docs",
			Components: map[string]*ComponentOwnership{
				"site":      {Root: "docs"},
				"tutorials": {ComponentType: "docs-pdf"}, // virtual: no root, no file
			},
		},
	})

	// Virtual component never matches
	owner := r.Resolve("docs/tutorials/intro.md")
	if owner == nil {
		t.Fatal("expected owner for docs file, got nil")
	}
	if owner.Component != "site" {
		t.Errorf("expected site, got %s", owner.Component)
	}
}

// TestResolve_CatchAllRoot verifies root: / acts as catch-all.
func TestResolve_CatchAllRoot(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "core",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/core"},
			},
		},
		{
			Moniker: "repository",
			Components: map[string]*ComponentOwnership{
				"repo-config": {Root: "/"},
			},
		},
	})

	// File under core module
	owner := r.Resolve("go/core/main.go")
	if owner == nil {
		t.Fatal("expected owner")
	}
	if owner.Module != "core" {
		t.Errorf("expected core, got %s", owner.Module)
	}

	// Root file falls to repository catch-all
	owner = r.Resolve("README.md")
	if owner == nil {
		t.Fatal("expected owner for README.md")
	}
	if owner.Module != "repository" || owner.Component != "repo-config" {
		t.Errorf("expected repository:repo-config, got %s:%s", owner.Module, owner.Component)
	}

	// .github file falls to catch-all
	owner = r.Resolve(".github/CODEOWNERS")
	if owner == nil {
		t.Fatal("expected owner for .github/CODEOWNERS")
	}
	if owner.Module != "repository" {
		t.Errorf("expected repository, got %s", owner.Module)
	}
}

// TestResolve_CrossDirectoryOwnership verifies modules with components in different directories.
func TestResolve_CrossDirectoryOwnership(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "eac-cli",
			Components: map[string]*ComponentOwnership{
				"go":         {Root: "go/cli/eac"},
				"dockerfile": {Root: "containers/drawio-oci"},
			},
		},
	})

	owner := r.Resolve("go/cli/eac/main.go")
	if owner == nil || owner.Component != "go" {
		t.Errorf("expected go component")
	}

	owner = r.Resolve("containers/drawio-oci/Dockerfile")
	if owner == nil || owner.Component != "dockerfile" {
		t.Errorf("expected dockerfile component")
	}
}

// TestResolve_SameRootExtensionTieBreaking verifies extension-based tie-breaking
// when multiple components share the same root.
func TestResolve_SameRootExtensionTieBreaking(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "repository",
			Components: map[string]*ComponentOwnership{
				"repo-config": {Root: "/", ComponentType: "repository"},
				"assets":      {Root: "/", ComponentType: "assets", Extensions: []string{".md", ".markdown", ".yml", ".yaml", ".json"}},
				"workflows":   {Root: ".github/workflows"},
			},
		},
	})

	tests := []struct {
		path     string
		wantComp string
	}{
		// Workflow files go to sub-owner (more specific root)
		{".github/workflows/ci.yaml", "workflows"},
		// Markdown files go to assets component (extension match)
		{"README.md", "assets"},
		// YAML files go to assets component (extension match)
		{".golangci.yml", "assets"},
		{"config.yaml", "assets"},
		// Unknown extension goes to catch-all (repo-config has no extensions)
		{"LICENSE", "repo-config"},
		{".gitignore", "repo-config"},
		{"Makefile", "repo-config"},
	}

	for _, tc := range tests {
		owner := r.Resolve(tc.path)
		if owner == nil {
			t.Errorf("Resolve(%q): expected owner, got nil", tc.path)
			continue
		}
		if owner.Component != tc.wantComp {
			t.Errorf("Resolve(%q): expected component %q, got %q", tc.path, tc.wantComp, owner.Component)
		}
	}
}

// TestResolve_FileOwnershipBeatsRoot verifies file ownership takes precedence over root.
func TestResolve_FileOwnershipBeatsRoot(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "repository",
			Components: map[string]*ComponentOwnership{
				"repo-config": {Root: "/"},
			},
		},
		{
			Moniker: "implicit-cli",
			Components: map[string]*ComponentOwnership{
				"importer": {File: "importer.sh"},
			},
		},
	})

	owner := r.Resolve("importer.sh")
	if owner == nil {
		t.Fatal("expected owner")
	}
	if owner.Module != "implicit-cli" || owner.Component != "importer" {
		t.Errorf("expected implicit-cli:importer, got %s:%s", owner.Module, owner.Component)
	}
}

// TestResolve_NoMatch verifies nil returned for unowned files.
func TestResolve_NoMatch(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "core",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/core"},
			},
		},
	})

	owner := r.Resolve("completely/unrelated/file.txt")
	if owner != nil {
		t.Errorf("expected nil, got %s:%s", owner.Module, owner.Component)
	}
}

// TestResolve_WindowsPathNormalization verifies backslash paths work.
func TestResolve_WindowsPathNormalization(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "core",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/core"},
			},
		},
	})

	owner := r.Resolve("go\\core\\config\\modules.go")
	if owner == nil {
		t.Fatal("expected owner for Windows path")
	}
	if owner.Component != "go" {
		t.Errorf("expected go, got %s", owner.Component)
	}
}

// TestResolveModule returns files owned by a specific module.
func TestResolveModule(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "core",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/core"},
			},
		},
		{
			Moniker: "eac-cli",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/cli/eac"},
			},
		},
	})

	allFiles := []string{
		"go/core/main.go",
		"go/core/config/types.go",
		"go/cli/eac/cmd.go",
		"unowned/file.txt",
	}

	coreFiles := r.ResolveModule("core", allFiles)
	if len(coreFiles) != 2 {
		t.Errorf("expected 2 core files, got %d", len(coreFiles))
	}

	eacFiles := r.ResolveModule("eac-cli", allFiles)
	if len(eacFiles) != 1 {
		t.Errorf("expected 1 eac-cli file, got %d", len(eacFiles))
	}
}

// TestValidate_AllOwned verifies no errors when all files are owned.
func TestValidate_AllOwned(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "repository",
			Components: map[string]*ComponentOwnership{
				"all": {Root: "/"},
			},
		},
	})

	errs := r.Validate([]string{"README.md", "go/main.go"})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

// TestValidate_UnownedFiles verifies detection of unowned files.
func TestValidate_UnownedFiles(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "core",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/core"},
			},
		},
	})

	errs := r.Validate([]string{"go/core/main.go", "orphan.txt"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Issue != "unowned" {
		t.Errorf("expected 'unowned', got %q", errs[0].Issue)
	}
	if errs[0].FilePath != "orphan.txt" {
		t.Errorf("expected orphan.txt, got %q", errs[0].FilePath)
	}
}

// TestValidate_MultiOwned verifies detection of files claimed by multiple modules.
func TestValidate_MultiOwned(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "mod-a",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "shared/lib"},
			},
		},
		{
			Moniker: "mod-b",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "shared/lib"},
			},
		},
	})

	errs := r.Validate([]string{"shared/lib/main.go"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Issue != "multi-owned" {
		t.Errorf("expected 'multi-owned', got %q", errs[0].Issue)
	}
	if len(errs[0].Owners) != 2 {
		t.Errorf("expected 2 owners, got %d", len(errs[0].Owners))
	}
}

// TestResolve_EmptyRoot verifies empty root is treated as catch-all (same as /).
func TestResolve_EmptyRoot(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "all",
			Components: map[string]*ComponentOwnership{
				"catch-all": {Root: ""},
			},
		},
	})

	// Empty root should not match anything (it's a virtual/unresolved component)
	owner := r.Resolve("some/file.go")
	if owner != nil {
		t.Errorf("expected nil for empty root, got %s:%s", owner.Module, owner.Component)
	}
}

// TestResolve_RootExactMatch verifies files at the root directory itself match.
func TestResolve_RootExactMatch(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "core",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/core"},
			},
		},
	})

	// A file named exactly "go/core" shouldn't match (that would be a directory)
	// Files directly under the root should match
	owner := r.Resolve("go/core/go.mod")
	if owner == nil {
		t.Fatal("expected owner for go/core/go.mod")
	}
}

// TestResolve_PartialRootMismatch ensures partial root prefixes don't match.
func TestResolve_PartialRootMismatch(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "core",
			Components: map[string]*ComponentOwnership{
				"go": {Root: "go/core"},
			},
		},
	})

	// "go/corefoo/bar.go" should NOT match "go/core" root
	owner := r.Resolve("go/corefoo/bar.go")
	if owner != nil {
		t.Errorf("expected nil for partial root match, got %s:%s", owner.Module, owner.Component)
	}
}

// TestResolve_CrossModuleSubOwnership verifies that sub-ownership only applies
// within the same module. Cross-module overlap is a validation error, but resolver
// picks the most-specific-root regardless of module.
func TestResolve_CrossModuleMostSpecificWins(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "broad",
			Components: map[string]*ComponentOwnership{
				"all": {Root: "shared"},
			},
		},
		{
			Moniker: "specific",
			Components: map[string]*ComponentOwnership{
				"lib": {Root: "shared/lib"},
			},
		},
	})

	// Most specific root wins even across modules
	owner := r.Resolve("shared/lib/code.go")
	if owner == nil {
		t.Fatal("expected owner")
	}
	if owner.Module != "specific" {
		t.Errorf("expected specific, got %s", owner.Module)
	}

	// Parent still owns non-overlapping files
	owner = r.Resolve("shared/other.go")
	if owner == nil {
		t.Fatal("expected owner")
	}
	if owner.Module != "broad" {
		t.Errorf("expected broad, got %s", owner.Module)
	}
}

// TestResolveComponent returns files for a specific module:component pair.
func TestResolveComponent(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker: "eac-cli",
			Components: map[string]*ComponentOwnership{
				"go":       {Root: "go/cli/eac"},
				"testdata": {Root: "go/cli/eac/testdata", ComponentType: "testdata"},
			},
		},
	})

	allFiles := []string{
		"go/cli/eac/main.go",
		"go/cli/eac/cmd/root.go",
		"go/cli/eac/testdata/sample.txt",
	}

	goFiles := r.ResolveComponent("eac-cli", "go", allFiles)
	if len(goFiles) != 2 {
		t.Errorf("expected 2 go files, got %d: %v", len(goFiles), goFiles)
	}

	tdFiles := r.ResolveComponent("eac-cli", "testdata", allFiles)
	if len(tdFiles) != 1 {
		t.Errorf("expected 1 testdata file, got %d: %v", len(tdFiles), tdFiles)
	}
}

// TestNewResolver_NilComponents handles nil component entries gracefully.
func TestNewResolver_NilComponents(t *testing.T) {
	r := NewResolver([]ModuleDefinition{
		{
			Moniker:    "empty",
			Components: nil,
		},
	})

	owner := r.Resolve("any/file.go")
	if owner != nil {
		t.Errorf("expected nil, got %s:%s", owner.Module, owner.Component)
	}
}

// TestNewResolver_EmptyModules handles empty module list.
func TestNewResolver_EmptyModules(t *testing.T) {
	r := NewResolver(nil)
	owner := r.Resolve("any/file.go")
	if owner != nil {
		t.Errorf("expected nil, got %s:%s", owner.Module, owner.Component)
	}
}
