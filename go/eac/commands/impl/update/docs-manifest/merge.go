package docsmanifest

import (
	"slices"
	"sort"
	"time"
)

// MergeResult contains both the updated descriptions and cache after merging.
type MergeResult struct {
	Descriptions map[string]Description
	Cache        *CacheManifest
	Result       *UpdateResult
}

// MergeManifest merges discovered assets with existing descriptions,
// creating updated descriptions and a new cache manifest.
func MergeManifest(
	existingDescriptions map[string]Description,
	existingCache *CacheManifest,
	discovered map[string]DiscoveredAsset,
	usage map[string][]string,
) *MergeResult {
	result := &UpdateResult{}

	// Build the new descriptions and cache maps
	newDescriptions := make(map[string]Description)
	newCacheAssets := make(map[string]CacheAsset)

	// Process discovered assets
	for path, disc := range discovered {
		// Get usedIn list for this asset
		usedIn := usage[path]
		if usedIn == nil {
			usedIn = []string{}
		}
		sort.Strings(usedIn)

		// Check if this asset existed before in descriptions
		var desc Description
		if existingDesc, exists := existingDescriptions[path]; exists {
			// Preserve the description and active status
			desc = existingDesc
		} else {
			// New asset - default to active: true, empty description
			desc = Description{
				Active:      true,
				Description: "",
			}
			result.Added = append(result.Added, path)
		}

		// Check if usedIn changed (for cache comparison)
		if existingCache != nil {
			if oldAsset, exists := existingCache.Assets[path]; exists {
				if !stringSlicesEqual(oldAsset.UsedIn, usedIn) {
					result.UpdatedUsedIn = append(result.UpdatedUsedIn, path)
				}
			}
		}

		newDescriptions[path] = desc

		newCacheAssets[path] = CacheAsset{
			UsedIn:       usedIn,
			Type:         disc.Type,
			Category:     disc.Category,
			LastModified: disc.LastModified.Format(time.RFC3339),
			Size:         disc.Size,
			Hash:         disc.Hash,
		}
	}

	// Find removed assets (in existing descriptions but not in discovered)
	for path := range existingDescriptions {
		if _, exists := discovered[path]; !exists {
			result.Removed = append(result.Removed, path)
		}
	}

	// Sort result slices for consistent output
	sort.Strings(result.Added)
	sort.Strings(result.Removed)
	sort.Strings(result.UpdatedUsedIn)

	// Calculate stats
	stats := calculateStats(newCacheAssets)
	result.TotalAssets = stats.Total
	result.UsedAssets = stats.Used
	result.UnusedAssets = stats.Unused

	// Determine if we need to write descriptions (only when new assets discovered)
	result.NeedsDescriptionsWrite = len(result.Added) > 0

	// Cache always needs write if anything changed or no existing cache
	result.NeedsCacheWrite = len(result.Added) > 0 ||
		len(result.Removed) > 0 ||
		len(result.UpdatedUsedIn) > 0 ||
		existingCache == nil

	// Create new cache manifest
	cache := &CacheManifest{
		Generated: time.Now().Format(time.RFC3339),
		Assets:    newCacheAssets,
		Stats:     stats,
	}

	return &MergeResult{
		Descriptions: newDescriptions,
		Cache:        cache,
		Result:       result,
	}
}

// calculateStats computes statistics for the cache manifest.
func calculateStats(assets map[string]CacheAsset) *ManifestStats {
	stats := &ManifestStats{
		ByCategory: make(map[string]CategoryStats),
	}

	for _, asset := range assets {
		stats.Total++
		if len(asset.UsedIn) > 0 {
			stats.Used++
		} else {
			stats.Unused++
		}

		// Update category stats
		if asset.Category != "" {
			cat := stats.ByCategory[asset.Category]
			cat.Total++
			if len(asset.UsedIn) > 0 {
				cat.Used++
			} else {
				cat.Unused++
			}
			stats.ByCategory[asset.Category] = cat
		}
	}

	return stats
}

// stringSlicesEqual checks if two string slices are equal.
func stringSlicesEqual(a, b []string) bool {
	return slices.Equal(a, b)
}
