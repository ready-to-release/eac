package docsmanifest

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// imageRefPattern matches markdown image references and HTML img tags.
// Matches patterns like:
//   - ![alt](path/to/image.png)
//   - <img src="path/to/image.png">
var imageRefPattern = regexp.MustCompile(`(?:!\[[^\]]*\]\(([^)]+)\)|<img[^>]+src=["']([^"']+)["'])`)

// ScanUsage scans all markdown files in docsDir and returns a map of
// asset paths to the list of markdown files that reference them.
// Asset paths are relative to assetsDir with forward slashes.
func ScanUsage(docsDir, assetsDir string) (map[string][]string, error) {
	usage := make(map[string][]string)

	// Get the relative path from docsDir to assetsDir for resolving references
	assetsRelToDocsDir, err := filepath.Rel(docsDir, assetsDir)
	if err != nil {
		return nil, err
	}
	assetsRelToDocsDir = filepath.ToSlash(assetsRelToDocsDir)

	err = filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// Get the directory containing this markdown file (for relative path resolution)
		mdDir := filepath.Dir(path)

		// Get path of markdown file relative to repo (docs/...)
		mdRelPath, err := filepath.Rel(filepath.Dir(docsDir), path)
		if err != nil {
			mdRelPath = path
		}
		mdRelPath = filepath.ToSlash(mdRelPath)

		// Scan file for image references
		refs, err := extractImageRefs(path)
		if err != nil {
			// Non-fatal: skip files we can't read
			return nil
		}

		for _, ref := range refs {
			// Resolve the reference to an absolute path
			var absRefPath string
			if filepath.IsAbs(ref) {
				absRefPath = ref
			} else {
				// Relative reference - resolve from markdown file's directory
				absRefPath = filepath.Join(mdDir, ref)
			}
			absRefPath = filepath.Clean(absRefPath)

			// Check if this reference points to something in assetsDir
			relToAssets, err := filepath.Rel(assetsDir, absRefPath)
			if err != nil {
				continue
			}

			// Skip if it's outside assetsDir (starts with ..)
			if strings.HasPrefix(relToAssets, "..") {
				continue
			}

			// Normalize to forward slashes
			assetPath := filepath.ToSlash(relToAssets)

			// Add this markdown file to the usage list
			if _, exists := usage[assetPath]; !exists {
				usage[assetPath] = []string{}
			}
			// Avoid duplicates
			if !slices.Contains(usage[assetPath], mdRelPath) {
				usage[assetPath] = append(usage[assetPath], mdRelPath)
			}
		}

		return nil
	})

	return usage, err
}

// extractImageRefs extracts all image references from a markdown file.
func extractImageRefs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var refs []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		// Find all image references in this line
		matches := imageRefPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			// match[1] is from markdown syntax, match[2] is from HTML img tag
			ref := match[1]
			if ref == "" {
				ref = match[2]
			}
			if ref != "" {
				// Clean up the reference (remove query strings, fragments)
				ref = cleanImageRef(ref)
				if ref != "" {
					refs = append(refs, ref)
				}
			}
		}
	}

	return refs, scanner.Err()
}

// cleanImageRef removes query strings and fragments from an image reference.
func cleanImageRef(ref string) string {
	// Remove query string
	if idx := strings.Index(ref, "?"); idx > 0 {
		ref = ref[:idx]
	}
	// Remove fragment
	if idx := strings.Index(ref, "#"); idx > 0 {
		ref = ref[:idx]
	}
	// Skip external URLs
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ""
	}
	return strings.TrimSpace(ref)
}

