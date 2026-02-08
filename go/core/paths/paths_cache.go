// Package paths provides centralized path constants and utilities for the EAC repository.
package paths

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CacheRootPath returns the root EAC cache directory (.cache/eac).
// All transient build caches should be stored under this directory.
func CacheRootPath(repoRoot string) string {
	return filepath.Join(repoRoot, EACCacheRoot)
}

// BuildCachePath returns the path to the build acceleration cache directory.
// Used for file hashes and build state tracking.
// Path: .cache/eac/build/
func BuildCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "build")
}

// FileHashCachePath returns the path to a book's file hash cache.
// Used for incremental preprocessing to track which files have changed.
// Path: .cache/eac/build/hashes/{bookName}.json
func FileHashCachePath(repoRoot, bookName string) string {
	return filepath.Join(BuildCachePath(repoRoot), "hashes", bookName+".json")
}

// InputHashCachePath returns the path to the module input hash cache.
// Used for caching computed source file hashes with mtime metadata.
// Path: .cache/eac/build/input-hashes.json
func InputHashCachePath(repoRoot string) string {
	return filepath.Join(BuildCachePath(repoRoot), "input-hashes.json")
}

// BuildStateCachePath returns the path to the build state cache directory.
// Used for tracking incremental build state (PDF/site content hashes).
// Path: .cache/eac/build/state/
func BuildStateCachePath(repoRoot string) string {
	return filepath.Join(BuildCachePath(repoRoot), "state")
}

// IncrementalCachePath returns the root path for UoW incremental state.
// State files (state.json) for build/test/lint/scan units live here,
// separate from output artifacts (logs, results) which remain in out/.
// Path: .cache/eac/incremental/
func IncrementalCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "incremental")
}

// SemaphoreCachePath returns the path to the semaphore coordination directory.
// Cross-process capacity semaphore files live here.
// Path: .cache/eac/semaphores/
func SemaphoreCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "semaphores")
}

// PreprocessCachePath returns the path to preprocessing state cache.
// Path: .cache/eac/preprocess/
func PreprocessCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "preprocess")
}

// NpmWorkCachePath returns the path to NPM isolation work directories.
// Path: .cache/eac/npm/work/
func NpmWorkCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "npm", "work")
}

// NpmDownloadCachePath returns the path for NPM_CONFIG_CACHE.
// Path: .cache/eac/npm/cache/
func NpmDownloadCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "npm", "cache")
}

// StagingCachePath returns the path to the staging cache directory.
// Used for persistent staging areas that survive across builds.
// Path: .cache/eac/staging/
func StagingCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "staging")
}

// BookStagingCachePath returns the path to a book's staging directory.
// The staging directory persists across builds for incremental preprocessing.
// Path: .cache/eac/staging/{moniker}/{bookName}
func BookStagingCachePath(repoRoot, moniker, bookName string) string {
	return filepath.Join(StagingCachePath(repoRoot), moniker, bookName)
}

// StructurizrCachePath returns the path to the Structurizr cache directory in docs/assets/cache
// This is the git-tracked cache location for CI optimization (pre-rendered Structurizr diagrams).
func StructurizrCachePath(repoRoot string) string {
	return filepath.Join(DocsCachePath(repoRoot), "structurizr")
}

// StructurizrDocsCachePath returns the path to a cached Structurizr SVG in docs/assets/cache
// Filename format: {module}_{viewKey}_{dslHash}.svg
// The dslHash is the first 8 characters of the SHA256 hash of the workspace.dsl content
// This ensures all views from the same DSL file share the same hash for cache invalidation.
func StructurizrDocsCachePath(repoRoot, module, viewKey, dslHash string) string {
	filename := fmt.Sprintf("%s_%s_%s.svg", module, viewKey, dslHash)
	return filepath.Join(StructurizrCachePath(repoRoot), filename)
}

// StructurizrModuleBuildOutputPath returns the path to a module's structurizr build output directory.
// Each module with a structurizr component gets its own build output.
// Path: out/build/{module}/structurizr-structurizr-render/structurizr/
func StructurizrModuleBuildOutputPath(repoRoot, moduleName string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, moduleName, "structurizr-structurizr-render", "structurizr")
}

// StructurizrAccelCachePath returns the path to the structurizr acceleration cache.
// Used for incremental builds to avoid re-rendering unchanged diagrams.
// Path: .cache/eac/structurizr/
func StructurizrAccelCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "structurizr")
}

// MarkdownCommandsBuildOutputPath returns the markdown-commands fragment output dir.
// Path: out/build/<module>/markdown-commands-markdown-commands/markdown-commands/
func MarkdownCommandsBuildOutputPath(repoRoot, moduleName string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, moduleName,
		"markdown-commands-markdown-commands", "markdown-commands")
}

// DrawioBuildOutputPath returns the path to the drawio build output directory.
// This is where rendered drawio PNGs are written by the drawio builder.
// Path: out/build/docs/drawio-drawio-render/drawio/
func DrawioBuildOutputPath(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, "docs", "drawio-drawio-render", "drawio")
}

// MermaidBuildOutputPath returns the path to the mermaid build output directory.
// This is where rendered mermaid SVGs and the index manifest are written.
// Path: out/build/docs/mermaid-mermaid-render/mermaid/
func MermaidBuildOutputPath(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, "docs", "mermaid-mermaid-render", "mermaid")
}

// DrawioAccelCachePath returns the path to the drawio acceleration cache.
// Used for incremental builds to avoid re-rendering unchanged diagrams.
// Path: .cache/eac/drawio/
func DrawioAccelCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "drawio")
}

// MermaidAccelCachePath returns the path to the mermaid acceleration cache.
// Used for incremental builds to avoid re-rendering unchanged diagrams.
// Path: .cache/eac/mermaid/
func MermaidAccelCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "mermaid")
}

// PlantUMLAccelCachePath returns the path to the plantuml acceleration cache.
// Used for incremental builds to avoid re-rendering unchanged diagrams.
// Path: .cache/eac/plantuml/
func PlantUMLAccelCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "plantuml")
}

// PlantUMLBuildOutputPath returns the path to the plantuml build output directory.
// This is where rendered plantuml SVGs and the index manifest are written.
// Path: out/build/docs/plantuml-plantuml-render/plantuml/
func PlantUMLBuildOutputPath(repoRoot string) string {
	return filepath.Join(repoRoot, OutDir, BuildDir, "docs", "plantuml-plantuml-render", "plantuml")
}

// PDFScreenshotsCachePath returns the path to the PDF screenshots cache directory.
// Located in .cache/eac/pdf-screenshots/ (not git-tracked).
func PDFScreenshotsCachePath(repoRoot string) string {
	return filepath.Join(CacheRootPath(repoRoot), "pdf-screenshots")
}

// PDFScreenshotsDirPath returns the path to a specific PDF's screenshot directory
// Each PDF gets its own directory named by the file's SHA256 hash (first 12 chars).
func PDFScreenshotsDirPath(repoRoot, hash string) string {
	return filepath.Join(PDFScreenshotsCachePath(repoRoot), hash)
}

// ============================================================================
// Traceable Cache Path Helpers (V2 - human-readable filenames)
// ============================================================================

// SanitizeForCacheName converts a source path to a cache-friendly identifier.
// The result contains the last 2 meaningful path components for traceability.
//
// Examples:
//
//	"docs/assets/architecture/modules-overview.drawio.png" -> "architecture_modules-overview"
//	"docs/explanation/cd-model/overview.md" -> "cd-model_overview"
//	"docs/reference/eac/architecture/index.md" -> "eac_architecture_index"
//
// Rules:
//  1. Strip common prefixes (docs/assets/, docs/, assets/)
//  2. Remove file extensions (.drawio.png, .md, .png)
//  3. Take last 2 path components (parent_basename)
//  4. Convert path separators to underscores
//  5. Lowercase for consistency
//  6. Limit total length to 64 characters
func SanitizeForCacheName(sourcePath string) string {
	// Normalize path separators to forward slashes
	path := strings.ReplaceAll(sourcePath, "\\", "/")

	// Strip common prefixes
	for _, prefix := range []string{"docs/assets/", "docs/", "assets/"} {
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
			break // Only strip one prefix
		}
	}

	// Remove extensions (order matters - check longest first)
	for _, ext := range []string{".drawio.png", ".drawio", ".md", ".png"} {
		if strings.HasSuffix(path, ext) {
			path = strings.TrimSuffix(path, ext)
			break // Only strip one extension
		}
	}

	// Split into path components
	parts := strings.Split(path, "/")

	// Take last 2 components (parent_basename)
	var result string
	if len(parts) >= 2 {
		result = parts[len(parts)-2] + "_" + parts[len(parts)-1]
	} else if len(parts) == 1 {
		// Single component - prefix with parent directory name from original path
		// If original was "docs/index.md", we want "docs_index"
		originalParts := strings.Split(strings.ReplaceAll(sourcePath, "\\", "/"), "/")
		if len(originalParts) >= 2 {
			parentIdx := len(originalParts) - 2
			parent := originalParts[parentIdx]
			// Remove extension from parent if it has one (shouldn't happen but be safe)
			result = parent + "_" + parts[0]
		} else {
			result = parts[0]
		}
	}

	// Lowercase for consistency
	result = strings.ToLower(result)

	// Limit length to 64 characters (truncate from the left to keep the end which is more specific)
	if len(result) > 64 {
		result = result[len(result)-64:]
	}

	return result
}

// CacheHashLength is the number of characters from a hash to include in cache filenames.
// 8 hex chars = 32 bits, enough to avoid collisions within a repository.
const CacheHashLength = 8

// truncateHash returns first n characters of hash, or full hash if shorter.
func truncateHash(hash string, n int) string {
	if len(hash) > n {
		return hash[:n]
	}
	return hash
}

// DrawioCachePath returns path with traceable filename for drawio cache.
// Format: {cacheRoot}/drawio/{identifier}_{hash8}.png
//
// Example:
//
//	DrawioCachePath("/cache", "docs/assets/arch/overview.drawio.png", "abc123def456")
//	-> "/cache/drawio/arch_overview_abc123de.png"
func DrawioCachePath(cacheRoot, sourcePath, contentHash string) string {
	identifier := SanitizeForCacheName(sourcePath)
	shortHash := truncateHash(contentHash, CacheHashLength)
	filename := fmt.Sprintf("%s_%s.png", identifier, shortHash)
	return filepath.Join(cacheRoot, "drawio", filename)
}

// MermaidCachePath returns path with traceable filename for mermaid cache.
// Format: {cacheRoot}/mermaid/{identifier}_{blockIndex}_{hash8}.svg
//
// Example:
//
//	MermaidCachePath("/cache", "docs/cd-model/overview.md", 0, "abc123def456")
//	-> "/cache/mermaid/cd-model_overview_0_abc123de.svg"
func MermaidCachePath(cacheRoot, sourcePath string, blockIndex int, contentHash string) string {
	identifier := SanitizeForCacheName(sourcePath)
	shortHash := truncateHash(contentHash, CacheHashLength)
	filename := fmt.Sprintf("%s_%d_%s.svg", identifier, blockIndex, shortHash)
	return filepath.Join(cacheRoot, "mermaid", filename)
}
