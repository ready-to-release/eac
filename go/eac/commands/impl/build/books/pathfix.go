package books

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

// pathMapping tracks how a file was remapped from source to staging.
type pathMapping struct {
	stagedPath   string // Path in staging
	sourcePrefix string // Prefix stripped from source (e.g., "docs/explanation/")
	depthChange  int    // Number of directory levels removed
	fileDepth    int    // Depth of file within content tree (for threshold calculation)
}

// fixRelativePaths corrects relative links in markdown files after copy
// This runs after step 1 (copy) to adjust links based on path remapping.
func (p *Preprocessor) fixRelativePaths() error {
	p.log("    Fixing relative paths based on source mappings...")

	// Build mappings for all copied markdown files
	mappings, err := p.buildPathMappings()
	if err != nil {
		return err
	}

	if len(mappings) == 0 {
		p.log("    No path mappings to adjust")
		return nil
	}

	// Process each mapped file
	fixed := 0
	for _, m := range mappings {
		if m.depthChange == 0 {
			continue // No adjustment needed
		}

		changed, err := fixFileRelativePaths(m.stagedPath, m.depthChange, m.fileDepth)
		if err != nil {
			return err
		}
		if changed {
			fixed++
		}
	}

	p.log("    Fixed relative paths in %d files", fixed)
	return nil
}

// buildPathMappings analyzes copy sources to build path remapping info.
func (p *Preprocessor) buildPathMappings() ([]pathMapping, error) {
	var mappings []pathMapping

	copySources := p.book.GetCopySources()
	p.log("    Found %d copy sources", len(copySources))

	for _, src := range copySources {
		// Only process markdown files
		if !strings.Contains(src.From, ".md") {
			p.log("    Skipping non-markdown source: %s", src.From)
			continue
		}

		p.log("    Processing markdown source: %s -> %s", src.From, src.To)
		srcMappings, err := p.buildMappingsForSource(src)
		if err != nil {
			return nil, err
		}
		p.log("    Built %d mappings", len(srcMappings))
		mappings = append(mappings, srcMappings...)
	}

	return mappings, nil
}

// buildMappingsForSource builds mappings for a single copy source.
func (p *Preprocessor) buildMappingsForSource(src config.Source) ([]pathMapping, error) {
	var mappings []pathMapping

	// Calculate the prefix that gets stripped
	// e.g., "docs/explanation/**/*.md" -> "docs/explanation/"
	sourcePrefix := extractSourcePrefix(src.From)
	depthChange := countPathDepth(sourcePrefix)

	// Adjust for shared parent: if prefix starts with "docs/", assets also come from "docs/"
	// so the effective depth change is reduced by 1 (the shared "docs/" level)
	if strings.HasPrefix(sourcePrefix, "docs/") && depthChange > 1 {
		depthChange--
	}

	// Find all files matching this source
	pattern := filepath.Join(p.workspaceRoot, src.From)
	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil || info.IsDir() {
			continue
		}

		// Skip excluded files
		if isExcluded(match, p.workspaceRoot, src.Exclude) {
			continue
		}

		// Calculate the staged path
		relPath, err := calculateRelativePath(match, src.From, p.workspaceRoot)
		if err != nil {
			continue
		}
		stagedPath := filepath.Join(p.stagingDir, src.To, relPath)

		// Calculate file depth (directory levels, not counting the file itself)
		fileDepth := countPathDepth(filepath.Dir(relPath))

		mappings = append(mappings, pathMapping{
			stagedPath:   stagedPath,
			sourcePrefix: sourcePrefix,
			depthChange:  depthChange,
			fileDepth:    fileDepth,
		})
	}

	return mappings, nil
}

// extractSourcePrefix extracts the fixed directory prefix from a glob pattern
// e.g., "docs/explanation/**/*.md" -> "docs/explanation/".
func extractSourcePrefix(pattern string) string {
	// Normalize to forward slashes
	pattern = filepath.ToSlash(pattern)

	// Find where the glob starts (first *, ?, [, or {)
	globStart := strings.IndexAny(pattern, "*?[{")
	if globStart == -1 {
		// No glob, return directory part
		return filepath.ToSlash(filepath.Dir(pattern)) + "/"
	}

	// Get the part before the glob
	prefix := pattern[:globStart]

	// Find the last directory separator before the glob
	lastSlash := strings.LastIndex(prefix, "/")
	if lastSlash == -1 {
		return ""
	}

	return prefix[:lastSlash+1]
}

// countPathDepth counts the number of directory components in a path.
func countPathDepth(path string) int {
	// Normalize path separators: convert both OS-native separators and backslashes to forward slashes
	// This ensures consistent behavior across platforms (Windows paths work on Linux and vice versa)
	path = filepath.ToSlash(path)
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.Trim(path, "/")
	// Handle "." which represents root directory (depth 0)
	if path == "" || path == "." {
		return 0
	}
	return strings.Count(path, "/") + 1
}

// linkPattern matches markdown links and image references.
var linkPattern = regexp.MustCompile(`(\[.+?\]\()(\.\./)+([^)]+\))|(!\[.*?\]\()(\.\./)+([^)]+\))`)

// relativePathPattern matches sequences of "../".
var relativePathPattern = regexp.MustCompile(`^((?:\.\./)+)(.*)$`)

// fixFileRelativePaths adjusts relative paths in a single file
// fileDepth is the depth of the file within the content tree (for threshold).
func fixFileRelativePaths(filePath string, depthChange, fileDepth int) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	original := string(content)
	modified := adjustRelativePaths(original, depthChange, fileDepth)

	if modified == original {
		return false, nil
	}

	err = os.WriteFile(filePath, []byte(modified), 0o644)
	return err == nil, err
}

// adjustRelativePaths adjusts all relative paths in content
// Only adjusts links that go OUTSIDE the content tree (more ../ than fileDepth).
func adjustRelativePaths(content string, depthChange, fileDepth int) string {
	// Match markdown links: [text](../path) and ![alt](../path)
	// Also match links with attributes: [text](../path){attrs} or [text](../path) { attrs }
	// Note: attr_list can have optional whitespace before/after braces
	linkRe := regexp.MustCompile(`(!?\[.+?\]\()(\.\./)+(.*?)(\)\s*(?:\{[^}]*\})?)`)

	return linkRe.ReplaceAllStringFunc(content, func(match string) string {
		// Parse the match
		submatches := linkRe.FindStringSubmatch(match)
		if len(submatches) < 5 {
			return match
		}

		prefix := submatches[1] // "[text](" or "![alt]("
		_ = submatches[4]       // suffix ")" or "){attrs}" - used implicitly in remaining

		// Count ../ sequences from the match
		afterPrefix := match[len(prefix):]
		upCount := 0
		for strings.HasPrefix(afterPrefix, "../") {
			upCount++
			afterPrefix = afterPrefix[3:]
		}

		// Only adjust links that go OUTSIDE the content tree
		// Links with upCount <= fileDepth stay within the tree and shouldn't be adjusted
		if upCount <= fileDepth {
			return match
		}

		// The remaining part is the path + suffix
		remaining := afterPrefix

		// Adjust: reduce the number of ../ by depthChange
		newUpCount := upCount - depthChange
		if newUpCount < 0 {
			newUpCount = 0
		}

		// Rebuild the link
		var result strings.Builder
		result.WriteString(prefix)
		for i := 0; i < newUpCount; i++ {
			result.WriteString("../")
		}
		result.WriteString(remaining)

		return result.String()
	})
}
