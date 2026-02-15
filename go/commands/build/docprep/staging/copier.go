package staging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/core/config"
)

// CopyStats tracks copy operation statistics.
type CopyStats struct {
	Copied  int // Files copied (new or changed)
	Skipped int // Files skipped (unchanged - same mtime/size)
	Lazy    int // Files skipped (lazy asset copy - unreferenced)
}

// FileMapper tracks source-to-staging file mappings.
type FileMapper interface {
	AddFileMapping(stagingPath, sourcePath string)
	GetSourcePath(stagingPath string) (string, bool)
}

// SimpleFileMap is a basic FileMapper implementation.
type SimpleFileMap struct {
	mappings map[string]string // staging -> source
}

// NewSimpleFileMap creates a new SimpleFileMap.
func NewSimpleFileMap() *SimpleFileMap {
	return &SimpleFileMap{mappings: make(map[string]string)}
}

// AddFileMapping records a staging-to-source mapping.
func (m *SimpleFileMap) AddFileMapping(stagingPath, sourcePath string) {
	m.mappings[stagingPath] = sourcePath
}

// GetSourcePath returns the source path for a given staging path.
func (m *SimpleFileMap) GetSourcePath(stagingPath string) (string, bool) {
	src, ok := m.mappings[stagingPath]
	return src, ok
}

// AllMappings returns all staging->source mappings as [2]string pairs.
func (m *SimpleFileMap) AllMappings() [][2]string {
	result := make([][2]string, 0, len(m.mappings))
	for staging, source := range m.mappings {
		result = append(result, [2]string{staging, source})
	}
	return result
}

// CopyConfig holds the parameters for a copy operation.
type CopyConfig struct {
	Book             *config.Book
	WorkspaceRoot    string
	StagingDir       string
	IsPDF            bool
	SkipDrawioFiles  bool
	ReferencedAssets map[string]bool
	FileMapper       FileMapper
	Logf             func(string, ...any)
	Warnf            func(string, ...any)
}

// CopyAllSources copies all copy-type sources to staging.
func CopyAllSources(cfg *CopyConfig) error {
	copySources := cfg.Book.GetCopySources()
	if len(copySources) == 0 {
		cfg.Logf("    No copy sources defined")
		return nil
	}

	var totalStats CopyStats
	for i := range copySources {
		src := &copySources[i]
		stats, err := copySingleSource(cfg, *src)
		if err != nil {
			return err
		}
		totalStats.Copied += stats.Copied
		totalStats.Skipped += stats.Skipped
		totalStats.Lazy += stats.Lazy
	}

	cfg.Logf("    Copied %d files (%d unchanged, %d lazy-skipped)",
		totalStats.Copied, totalStats.Skipped, totalStats.Lazy)

	// Clean up orphaned files
	if err := CleanOrphanedStagedFiles(cfg.StagingDir, cfg.FileMapper, cfg.Logf); err != nil {
		return err
	}

	return nil
}

// copySingleSource copies files matching a single copy source.
func copySingleSource(cfg *CopyConfig, src config.Source) (CopyStats, error) {
	var stats CopyStats
	from := src.From
	to := src.To
	exclude := src.Exclude

	isAssetCopy := strings.Contains(from, "assets/") || strings.HasPrefix(from, "assets")

	pattern := filepath.Join(cfg.WorkspaceRoot, from)

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return stats, err
	}

	for _, match := range matches {
		srcInfo, err := os.Stat(match)
		if err != nil {
			continue
		}
		if srcInfo.IsDir() {
			continue
		}

		if IsExcluded(match, cfg.WorkspaceRoot, exclude) {
			continue
		}

		// Skip .drawio files when mode says so
		if cfg.SkipDrawioFiles && strings.HasSuffix(strings.ToLower(match), ".drawio") {
			continue
		}

		// Lazy asset copying
		if isAssetCopy && cfg.ReferencedAssets != nil && len(cfg.ReferencedAssets) > 0 {
			if !IsAssetNeeded(match, cfg.ReferencedAssets) {
				stats.Lazy++
				continue
			}
		}

		relPath, err := CalculateRelativePath(match, from, cfg.WorkspaceRoot)
		if err != nil {
			return stats, err
		}

		destPath := filepath.Join(cfg.StagingDir, to, relPath)

		// Track mapping (always, even if skipped)
		cfg.FileMapper.AddFileMapping(destPath, match)

		if !NeedsCopy(srcInfo, destPath) {
			stats.Skipped++
			continue
		}

		if err := CopyFilePreserveMtime(match, destPath, srcInfo); err != nil {
			return stats, err
		}

		stats.Copied++
	}

	return stats, nil
}

// IsAssetNeeded checks if an asset file should be copied based on references.
func IsAssetNeeded(assetPath string, referencedAssets map[string]bool) bool {
	normalized := filepath.ToSlash(assetPath)

	essentialDirs := []string{"/css/", "/js/", "/logo/", "/templates/"}
	for _, dir := range essentialDirs {
		if strings.Contains(normalized, dir) {
			return true
		}
	}

	if strings.HasSuffix(normalized, "manifest.json") {
		return true
	}

	return IsAssetReferenced(assetPath, referencedAssets)
}

// CalculateRelativePath calculates the relative path for a matched file.
func CalculateRelativePath(matchPath, pattern, workspaceRoot string) (string, error) {
	fullPattern := filepath.ToSlash(filepath.Join(workspaceRoot, pattern))

	baseDir, _ := doublestar.SplitPattern(fullPattern)
	if baseDir == "" || baseDir == "." {
		baseDir = workspaceRoot
	} else {
		baseDir = filepath.FromSlash(baseDir)
	}

	rel, err := filepath.Rel(baseDir, matchPath)
	if err != nil {
		return "", fmt.Errorf("calculating relative path: %w", err)
	}

	return rel, nil
}

// IsExcluded checks if a file matches any exclusion pattern.
func IsExcluded(path, workspaceRoot string, excludePatterns []string) bool {
	normalizedPath := filepath.ToSlash(path)

	// Always exclude files in any 'lfs' directory
	if strings.Contains(normalizedPath, "/lfs/") {
		return true
	}

	for _, pattern := range excludePatterns {
		fullPattern := filepath.ToSlash(filepath.Join(workspaceRoot, pattern))

		if matched, matchErr := doublestar.Match(fullPattern, normalizedPath); matchErr == nil && matched {
			return true
		}

		relPath, relErr := filepath.Rel(workspaceRoot, path)
		if relErr == nil {
			normalizedRelPath := filepath.ToSlash(relPath)
			if matched, matchErr := doublestar.Match(pattern, normalizedRelPath); matchErr == nil && matched {
				return true
			}
		}
	}
	return false
}
