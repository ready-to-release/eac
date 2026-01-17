package books

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// validateNavEntry validates a single nav entry recursively
// Returns (validatedEntry, referencedFiles) or (nil, empty) if invalid.
func (p *Preprocessor) validateNavEntry(entry any, dirPath string, actualFiles, actualDirs map[string]bool) (any, map[string]bool) {
	referenced := make(map[string]bool)

	switch v := entry.(type) {
	case string:
		// Simple file or directory reference: "modules.md" or "subdir/" or "subdir/file.md"
		if strings.HasSuffix(v, "/") {
			// Directory reference
			dirName := strings.TrimSuffix(v, "/")
			if actualDirs[dirName] {
				referenced[dirName+"/"] = true
				return v, referenced // Valid directory
			}
			// For relative paths (containing /), check filesystem
			if strings.Contains(dirName, "/") {
				fullPath := filepath.Join(dirPath, dirName)
				if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
					referenced[dirName+"/"] = true
					return v, referenced // Valid directory
				}
			}
			// Directory missing, skip
			return nil, referenced
		} else if strings.HasSuffix(v, ".md") {
			// File reference
			if actualFiles[v] {
				referenced[v] = true
				return v, referenced // Valid file
			}
			// For relative paths (containing /), check filesystem
			if strings.Contains(v, "/") {
				fullPath := filepath.Join(dirPath, v)
				navLog.Debugf("[VALIDATE] Checking relative file: %s -> %s", v, fullPath)
				if _, err := os.Stat(fullPath); err == nil {
					navLog.Debugf("[VALIDATE] ✓ File exists: %s", v)
					referenced[v] = true
					return v, referenced // Valid file
				} else {
					navLog.Debugf("[VALIDATE] ✗ File missing: %s (error: %v)", v, err)
				}
			}
			// File missing, skip
			return nil, referenced
		}
		// Unknown format, keep as-is
		return v, referenced

	case map[string]any:
		// Titled section: {"Section Name": [...]} or {"Title": "file.md"}
		validatedMap := make(map[string]any)

		for title, content := range v {
			switch c := content.(type) {
			case string:
				// {"Title": "file.md"}
				if strings.HasSuffix(c, ".md") {
					if actualFiles[c] {
						validatedMap[title] = c
						referenced[c] = true
					} else if strings.Contains(c, "/") {
						// For relative paths, check filesystem
						fullPath := filepath.Join(dirPath, c)
						if _, err := os.Stat(fullPath); err == nil {
							validatedMap[title] = c
							referenced[c] = true
						}
					}
				} else if strings.HasSuffix(c, "/") {
					// {"Title": "subdir/"}
					dirName := strings.TrimSuffix(c, "/")
					if actualDirs[dirName] {
						validatedMap[title] = c
						referenced[dirName+"/"] = true
					} else if strings.Contains(dirName, "/") {
						// For relative paths, check filesystem
						fullPath := filepath.Join(dirPath, dirName)
						if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
							validatedMap[title] = c
							referenced[dirName+"/"] = true
						}
					}
				} else {
					// {"Title": "subdir"} - directory without trailing slash
					// awesome-nav allows both formats
					if actualDirs[c] {
						validatedMap[title] = c
						referenced[c+"/"] = true
					} else if strings.Contains(c, "/") {
						// For relative paths, check filesystem
						fullPath := filepath.Join(dirPath, c)
						if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
							validatedMap[title] = c
							referenced[c+"/"] = true
						}
					}
				}
				// If file/dir missing, omit this entry

			case []any:
				// {"Section": [...items...]}
				validatedItems := []any{}
				for _, item := range c {
					validated, refs := p.validateNavEntry(item, dirPath, actualFiles, actualDirs)
					if validated != nil {
						validatedItems = append(validatedItems, validated)
						for ref := range refs {
							referenced[ref] = true
						}
					}
				}

				// Only include section if it has valid items
				if len(validatedItems) > 0 {
					validatedMap[title] = validatedItems
				}
			}
		}

		if len(validatedMap) > 0 {
			return validatedMap, referenced
		}
		return nil, referenced
	}

	return entry, referenced // Unknown type, keep as-is
}

// validateAndCleanNav validates existing .nav.yml, removes broken refs, adds new files.
func (p *Preprocessor) validateAndCleanNav(navPath, dirPath string) error {
	// 1. Parse existing .nav.yml
	navFile, err := p.parseNavFile(navPath)
	if err != nil {
		// If parse fails, regenerate from scratch
		relPath, relErr := filepath.Rel(p.stagingDir, navPath)
		if relErr != nil {
			relPath = navPath
		}
		p.log("    ⚠️  Failed to parse %s, regenerating", relPath)
		return p.generateNavForDir(dirPath)
	}

	// 2. Scan directory for actual files
	actualFiles := p.scanMarkdownFiles(dirPath)
	actualDirs := p.scanSubdirectories(dirPath)

	// 3. Process nav entries
	validatedNav := make([]any, 0)
	referencedFiles := make(map[string]bool)

	for _, entry := range navFile.Nav {
		validated, referenced := p.validateNavEntry(entry, dirPath, actualFiles, actualDirs)
		if validated != nil {
			validatedNav = append(validatedNav, validated)
			// Track what files are referenced
			for file := range referenced {
				referencedFiles[file] = true
			}
		}
	}

	// 4. Find unreferenced files (new from command sources or other additions)
	newFiles := []string{}
	for file := range actualFiles {
		if !referencedFiles[file] {
			newFiles = append(newFiles, file)
		}
	}

	// 5. Add new files with intelligent ordering
	if len(newFiles) > 0 {
		sortedNew := p.sortNewFiles(newFiles, dirPath)
		for _, file := range sortedNew {
			validatedNav = append(validatedNav, file)
		}

		relDir, relErr := filepath.Rel(p.stagingDir, dirPath)
		if relErr != nil || relDir == "." {
			relDir = "(root)"
		}
		p.log("    Added %d new files to %s/.nav.yml", len(newFiles), relDir)
	}

	// 6. Write updated .nav.yml
	// Preserve title if set - awesome-nav uses it for navigation display
	navFile.Nav = validatedNav
	data, err := yaml.Marshal(navFile)
	if err != nil {
		return fmt.Errorf("marshaling nav file: %w", err)
	}

	return os.WriteFile(navPath, data, 0o644)
}
