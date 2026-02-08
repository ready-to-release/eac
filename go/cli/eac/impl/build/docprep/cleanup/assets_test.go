package cleanup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupUnreferencedAssets_ReferencedKept(t *testing.T) {
	tmpDir := t.TempDir()

	// Create markdown referencing an image
	mdContent := "# Title\n\n![image](assets/logo.png)\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create referenced asset
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "logo.png"), []byte("PNG"), 0o644); err != nil {
		t.Fatal(err)
	}

	noop := func(string, ...any) {}
	if err := CleanupUnreferencedAssets(tmpDir, noop, noop); err != nil {
		t.Fatal(err)
	}

	// Referenced asset should still exist
	if _, err := os.Stat(filepath.Join(assetsDir, "logo.png")); os.IsNotExist(err) {
		t.Errorf("Referenced asset was deleted")
	}
}

func TestCleanupUnreferencedAssets_UnreferencedDeleted(t *testing.T) {
	tmpDir := t.TempDir()

	// Create markdown that does NOT reference anything
	mdContent := "# Title\n\nNo images here.\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create unreferenced asset
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "unused.png"), []byte("PNG"), 0o644); err != nil {
		t.Fatal(err)
	}

	noop := func(string, ...any) {}
	if err := CleanupUnreferencedAssets(tmpDir, noop, noop); err != nil {
		t.Fatal(err)
	}

	// Unreferenced asset should be gone
	if _, err := os.Stat(filepath.Join(assetsDir, "unused.png")); !os.IsNotExist(err) {
		t.Errorf("Unreferenced asset was not deleted")
	}
}

func TestCleanupUnreferencedAssets_EmptyDirPruned(t *testing.T) {
	tmpDir := t.TempDir()

	// Create markdown
	mdContent := "# Title\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create nested empty dir with an unreferenced asset
	nestedDir := filepath.Join(tmpDir, "assets", "images", "old")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "stale.jpg"), []byte("JPEG"), 0o644); err != nil {
		t.Fatal(err)
	}

	noop := func(string, ...any) {}
	if err := CleanupUnreferencedAssets(tmpDir, noop, noop); err != nil {
		t.Fatal(err)
	}

	// Asset should be deleted and empty dirs pruned
	if _, err := os.Stat(filepath.Join(nestedDir, "stale.jpg")); !os.IsNotExist(err) {
		t.Errorf("Unreferenced asset was not deleted")
	}

	// The empty directory chain should be pruned
	if _, err := os.Stat(nestedDir); !os.IsNotExist(err) {
		t.Errorf("Empty directory was not pruned")
	}
}

func TestCleanupUnreferencedAssets_ExternalLinksIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	// Markdown references external image and a local one
	mdContent := `# Title

![external](https://example.com/image.png)
![local](assets/local.png)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatal(err)
	}

	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "local.png"), []byte("PNG"), 0o644); err != nil {
		t.Fatal(err)
	}

	noop := func(string, ...any) {}
	if err := CleanupUnreferencedAssets(tmpDir, noop, noop); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(assetsDir, "local.png")); os.IsNotExist(err) {
		t.Errorf("Local referenced asset was deleted")
	}
}

func TestCleanupUnreferencedAssets_NoAssets(t *testing.T) {
	tmpDir := t.TempDir()

	mdContent := "# Empty\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatal(err)
	}

	noop := func(string, ...any) {}
	// Should not error with no assets
	if err := CleanupUnreferencedAssets(tmpDir, noop, noop); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupUnreferencedAssets_NonAssetExtensionsKept(t *testing.T) {
	tmpDir := t.TempDir()

	mdContent := "# Title\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a .yml file that should NOT be cleaned up (not an asset extension)
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), []byte("key: value"), 0o644); err != nil {
		t.Fatal(err)
	}

	noop := func(string, ...any) {}
	if err := CleanupUnreferencedAssets(tmpDir, noop, noop); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "config.yml")); os.IsNotExist(err) {
		t.Errorf("Non-asset file was incorrectly deleted")
	}
}

func TestCleanupUnreferencedAssets_HTMLSrcReference(t *testing.T) {
	tmpDir := t.TempDir()

	// Reference via HTML img tag
	mdContent := `# Title

<img src="assets/diagram.svg" alt="diagram">
`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.md"), []byte(mdContent), 0o644); err != nil {
		t.Fatal(err)
	}

	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "diagram.svg"), []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	noop := func(string, ...any) {}
	if err := CleanupUnreferencedAssets(tmpDir, noop, noop); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(assetsDir, "diagram.svg")); os.IsNotExist(err) {
		t.Errorf("HTML-referenced asset was deleted")
	}
}
