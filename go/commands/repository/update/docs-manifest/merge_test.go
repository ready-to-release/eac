package docsmanifest

import (
	"testing"
	"time"
)

func TestMergeManifest_PreservesDescriptions(t *testing.T) {
	// Setup: existing descriptions
	existingDescriptions := map[string]Description{
		"test/image.png": {
			Active:      true,
			Description: "A test image",
		},
	}

	// Existing cache with old usedIn
	existingCache := &CacheManifest{
		Assets: map[string]CacheAsset{
			"test/image.png": {
				UsedIn: []string{"docs/readme.md"},
			},
		},
	}

	// Discovered assets (same asset)
	discovered := map[string]DiscoveredAsset{
		"test/image.png": {
			RelPath:      "test/image.png",
			Category:     "test",
			Type:         "image",
			Size:         1024,
			LastModified: time.Now(),
		},
	}

	// Usage unchanged
	usage := map[string][]string{
		"test/image.png": {"docs/readme.md"},
	}

	// Merge
	mergeResult := MergeManifest(existingDescriptions, existingCache, discovered, usage)

	// Verify description was preserved
	if mergeResult.Descriptions["test/image.png"].Description != "A test image" {
		t.Errorf("Expected description to be preserved, got: %s", mergeResult.Descriptions["test/image.png"].Description)
	}

	// Verify active flag was preserved
	if !mergeResult.Descriptions["test/image.png"].Active {
		t.Error("Expected active flag to be preserved")
	}

	// Verify no changes detected
	if len(mergeResult.Result.Added) != 0 {
		t.Errorf("Expected 0 added, got %d", len(mergeResult.Result.Added))
	}
	if len(mergeResult.Result.Removed) != 0 {
		t.Errorf("Expected 0 removed, got %d", len(mergeResult.Result.Removed))
	}
}

func TestMergeManifest_AddsNewWithEmptyDescription(t *testing.T) {
	// Setup: empty existing descriptions
	existingDescriptions := map[string]Description{}

	// Discovered new asset
	discovered := map[string]DiscoveredAsset{
		"new/diagram.drawio.png": {
			RelPath:      "new/diagram.drawio.png",
			Category:     "new",
			Type:         "drawio",
			Size:         2048,
			LastModified: time.Now(),
		},
	}

	usage := map[string][]string{}

	// Merge
	mergeResult := MergeManifest(existingDescriptions, nil, discovered, usage)

	// Verify new asset added with empty description
	desc, exists := mergeResult.Descriptions["new/diagram.drawio.png"]
	if !exists {
		t.Fatal("Expected new asset to be added")
	}
	if desc.Description != "" {
		t.Errorf("Expected empty description for new asset, got: %s", desc.Description)
	}
	if !desc.Active {
		t.Error("Expected new asset to be active by default")
	}

	// Verify it's reported as added
	if len(mergeResult.Result.Added) != 1 || mergeResult.Result.Added[0] != "new/diagram.drawio.png" {
		t.Errorf("Expected 1 added asset, got: %v", mergeResult.Result.Added)
	}
}

func TestMergeManifest_RemovesDeletedAssets(t *testing.T) {
	// Setup: existing descriptions with asset
	existingDescriptions := map[string]Description{
		"old/deleted.png": {
			Active:      true,
			Description: "This will be deleted",
		},
	}

	// No discovered assets
	discovered := map[string]DiscoveredAsset{}
	usage := map[string][]string{}

	// Merge
	mergeResult := MergeManifest(existingDescriptions, nil, discovered, usage)

	// Verify asset is removed from descriptions
	if _, exists := mergeResult.Descriptions["old/deleted.png"]; exists {
		t.Error("Expected deleted asset to be removed from descriptions")
	}

	// Verify it's reported as removed
	if len(mergeResult.Result.Removed) != 1 || mergeResult.Result.Removed[0] != "old/deleted.png" {
		t.Errorf("Expected 1 removed asset, got: %v", mergeResult.Result.Removed)
	}
}

func TestMergeManifest_UpdatesUsedInReferences(t *testing.T) {
	// Setup: existing descriptions
	existingDescriptions := map[string]Description{
		"test/image.png": {
			Active:      true,
			Description: "Test image",
		},
	}

	// Existing cache with old usedIn
	existingCache := &CacheManifest{
		Assets: map[string]CacheAsset{
			"test/image.png": {
				UsedIn: []string{"docs/old.md"},
			},
		},
	}

	// Same asset discovered
	discovered := map[string]DiscoveredAsset{
		"test/image.png": {
			RelPath:      "test/image.png",
			Category:     "test",
			Type:         "image",
			Size:         1024,
			LastModified: time.Now(),
		},
	}

	// New usage
	usage := map[string][]string{
		"test/image.png": {"docs/another.md", "docs/new.md"},
	}

	// Merge
	mergeResult := MergeManifest(existingDescriptions, existingCache, discovered, usage)

	// Verify usedIn was updated in cache
	asset := mergeResult.Cache.Assets["test/image.png"]
	if len(asset.UsedIn) != 2 {
		t.Errorf("Expected 2 usedIn refs, got %d", len(asset.UsedIn))
	}

	// Verify it's reported as updated
	if len(mergeResult.Result.UpdatedUsedIn) != 1 {
		t.Errorf("Expected 1 updated usedIn, got %d", len(mergeResult.Result.UpdatedUsedIn))
	}

	// Verify description preserved
	if mergeResult.Descriptions["test/image.png"].Description != "Test image" {
		t.Errorf("Expected description preserved, got: %s", mergeResult.Descriptions["test/image.png"].Description)
	}
}

func TestMergeManifest_HandlesEmptyExistingManifest(t *testing.T) {
	// No existing descriptions
	existingDescriptions := map[string]Description{}

	discovered := map[string]DiscoveredAsset{
		"test/image.png": {
			RelPath:      "test/image.png",
			Category:     "test",
			Type:         "image",
			Size:         1024,
			LastModified: time.Now(),
		},
	}

	usage := map[string][]string{
		"test/image.png": {"docs/readme.md"},
	}

	// Merge
	mergeResult := MergeManifest(existingDescriptions, nil, discovered, usage)

	// Verify asset added
	if len(mergeResult.Descriptions) != 1 {
		t.Errorf("Expected 1 description, got %d", len(mergeResult.Descriptions))
	}

	// Verify all are marked as added
	if len(mergeResult.Result.Added) != 1 {
		t.Errorf("Expected 1 added, got %d", len(mergeResult.Result.Added))
	}

	// Verify NeedsDescriptionsWrite is true
	if !mergeResult.Result.NeedsDescriptionsWrite {
		t.Error("Expected NeedsDescriptionsWrite to be true for new manifest")
	}

	// Verify NeedsCacheWrite is true
	if !mergeResult.Result.NeedsCacheWrite {
		t.Error("Expected NeedsCacheWrite to be true for new manifest")
	}
}

func TestMergeManifest_HandlesNoChanges(t *testing.T) {
	// Setup: existing descriptions
	existingDescriptions := map[string]Description{
		"test/image.png": {
			Active:      true,
			Description: "Test",
		},
	}

	// Existing cache
	existingCache := &CacheManifest{
		Assets: map[string]CacheAsset{
			"test/image.png": {
				UsedIn: []string{"docs/readme.md"},
			},
		},
	}

	// Same asset discovered
	discovered := map[string]DiscoveredAsset{
		"test/image.png": {
			RelPath:      "test/image.png",
			Category:     "test",
			Type:         "image",
			Size:         1024,
			LastModified: time.Now(),
		},
	}

	// Same usage
	usage := map[string][]string{
		"test/image.png": {"docs/readme.md"},
	}

	// Merge
	mergeResult := MergeManifest(existingDescriptions, existingCache, discovered, usage)

	// Verify no changes
	if len(mergeResult.Result.Added) != 0 {
		t.Errorf("Expected 0 added, got %d", len(mergeResult.Result.Added))
	}
	if len(mergeResult.Result.Removed) != 0 {
		t.Errorf("Expected 0 removed, got %d", len(mergeResult.Result.Removed))
	}
	if len(mergeResult.Result.UpdatedUsedIn) != 0 {
		t.Errorf("Expected 0 updated, got %d", len(mergeResult.Result.UpdatedUsedIn))
	}
	if mergeResult.Result.NeedsDescriptionsWrite {
		t.Error("Expected NeedsDescriptionsWrite to be false when no changes")
	}
	if mergeResult.Result.NeedsCacheWrite {
		t.Error("Expected NeedsCacheWrite to be false when no changes")
	}
}

func TestCalculateStats(t *testing.T) {
	assets := map[string]CacheAsset{
		"cat1/used.png": {
			Category: "cat1",
			UsedIn:   []string{"docs/a.md"},
		},
		"cat1/unused.png": {
			Category: "cat1",
			UsedIn:   []string{},
		},
		"cat2/used.png": {
			Category: "cat2",
			UsedIn:   []string{"docs/b.md", "docs/c.md"},
		},
	}

	stats := calculateStats(assets)

	if stats.Total != 3 {
		t.Errorf("Expected total 3, got %d", stats.Total)
	}
	if stats.Used != 2 {
		t.Errorf("Expected used 2, got %d", stats.Used)
	}
	if stats.Unused != 1 {
		t.Errorf("Expected unused 1, got %d", stats.Unused)
	}

	cat1 := stats.ByCategory["cat1"]
	if cat1.Total != 2 || cat1.Used != 1 || cat1.Unused != 1 {
		t.Errorf("Cat1 stats wrong: %+v", cat1)
	}

	cat2 := stats.ByCategory["cat2"]
	if cat2.Total != 1 || cat2.Used != 1 || cat2.Unused != 0 {
		t.Errorf("Cat2 stats wrong: %+v", cat2)
	}
}
