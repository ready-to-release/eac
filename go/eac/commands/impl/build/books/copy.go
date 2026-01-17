package books

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

// copyStaticFiles copies all copy-type sources to staging (Step 1).
func (p *Preprocessor) copyStaticFiles() error {
	copySources := p.book.GetCopySources()
	if len(copySources) == 0 {
		p.log("    No copy sources defined")
		return nil
	}

	for _, src := range copySources {
		count, err := p.copySingleSource(src)
		if err != nil {
			return err
		}
		p.log("    Copied %d files from %s", count, src.From)
	}

	return nil
}

// copySingleSource copies files matching a single copy source.
func (p *Preprocessor) copySingleSource(src config.Source) (int, error) {
	from := src.From
	to := src.To
	exclude := src.Exclude

	// Build the glob pattern (relative to workspace root)
	pattern := filepath.Join(p.workspaceRoot, from)

	// Use doublestar for glob matching with ** support
	basePath, _ := doublestar.SplitPattern(pattern)
	if basePath == "" {
		basePath = p.workspaceRoot
	}

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return 0, err
	}

	var copied int
	for _, match := range matches {
		// Check if it's a file (skip directories)
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}

		// Check exclusions
		if isExcluded(match, p.workspaceRoot, exclude) {
			continue
		}

		// Skip .drawio files in PDF mode (they're interactive diagrams for web only)
		if p.pdfMode && strings.HasSuffix(strings.ToLower(match), ".drawio") {
			continue
		}

		// Calculate relative path from the base of the glob pattern
		relPath, err := calculateRelativePath(match, from, p.workspaceRoot)
		if err != nil {
			return copied, err
		}

		// Destination in staging
		destPath := filepath.Join(p.stagingDir, to, relPath)

		// DEBUG: Log markdown file copies
		if strings.HasSuffix(strings.ToLower(match), ".md") {
			log.Debugf("[COPY] MD file: %s -> %s", match, destPath)
		}

		// Copy file
		if err := copyFile(match, destPath); err != nil {
			return copied, err
		}

		// Track source → staging mapping for link translation
		// Track all files so we can generate correct relative paths for images, etc.
		p.linkTranslator.AddFileMapping(destPath, match)

		copied++
	}

	return copied, nil
}

// calculateRelativePath calculates the relative path for a matched file.
func calculateRelativePath(matchPath, pattern, workspaceRoot string) (string, error) {
	// Build full pattern path with forward slashes for doublestar
	fullPattern := filepath.ToSlash(filepath.Join(workspaceRoot, pattern))

	// Use doublestar.SplitPattern to find the base directory
	// SplitPattern works with forward slashes
	baseDir, _ := doublestar.SplitPattern(fullPattern)
	if baseDir == "" || baseDir == "." {
		baseDir = workspaceRoot
	} else {
		// Convert back to OS-specific path separators
		baseDir = filepath.FromSlash(baseDir)
	}

	// Get relative path from base directory
	rel, err := filepath.Rel(baseDir, matchPath)
	if err != nil {
		return "", err
	}

	return rel, nil
}

// isExcluded checks if a file matches any exclusion pattern.
func isExcluded(path, workspaceRoot string, excludePatterns []string) bool {
	// Normalize path to forward slashes for cross-platform matching
	normalizedPath := filepath.ToSlash(path)

	// Always exclude files in any 'lfs' directory (large file storage)
	// These are typically large binary files not suitable for PDF embedding
	if strings.Contains(normalizedPath, "/lfs/") {
		return true
	}

	for _, pattern := range excludePatterns {
		// Make pattern relative to workspace root for matching
		fullPattern := filepath.ToSlash(filepath.Join(workspaceRoot, pattern))

		// Use doublestar for glob matching
		if matched, matchErr := doublestar.Match(fullPattern, normalizedPath); matchErr == nil && matched {
			return true
		}

		// Also try matching just the relative path
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

// copyFile copies a file from src to dst, creating directories as needed.
func copyFile(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy contents
	_, err = io.Copy(dstFile, srcFile)
	return err
}
