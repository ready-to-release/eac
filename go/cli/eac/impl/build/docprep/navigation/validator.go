package navigation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"gopkg.in/yaml.v3"
)

// ValidateNavEntry validates a single nav entry recursively.
// Returns (validatedEntry, referencedFiles) or (nil, empty) if invalid.
func ValidateNavEntry(entry any, dirPath string, actualFiles, actualDirs map[string]bool) (any, map[string]bool) {
	referenced := make(map[string]bool)

	switch v := entry.(type) {
	case string:
		if strings.HasSuffix(v, "/") {
			dirName := strings.TrimSuffix(v, "/")
			if actualDirs[dirName] {
				referenced[dirName+"/"] = true
				return v, referenced
			}
			if strings.Contains(dirName, "/") {
				fullPath := filepath.Join(dirPath, dirName)
				if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
					referenced[dirName+"/"] = true
					return v, referenced
				}
			}
			return nil, referenced
		} else if strings.HasSuffix(v, ".md") {
			if actualFiles[v] {
				referenced[v] = true
				return v, referenced
			}
			if strings.Contains(v, "/") {
				fullPath := filepath.Join(dirPath, v)
				if _, err := os.Stat(fullPath); err == nil {
					referenced[v] = true
					return v, referenced
				}
			}
			return nil, referenced
		} else {
			if actualDirs[v] {
				referenced[v+"/"] = true
				return v, referenced
			}
			if strings.Contains(v, "/") {
				fullPath := filepath.Join(dirPath, v)
				if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
					referenced[v+"/"] = true
					return v, referenced
				}
			}
			return nil, referenced
		}

	case map[string]any:
		validatedMap := make(map[string]any)

		for title, content := range v {
			switch c := content.(type) {
			case string:
				if strings.HasSuffix(c, ".md") {
					if actualFiles[c] {
						validatedMap[title] = c
						referenced[c] = true
					} else if strings.Contains(c, "/") {
						fullPath := filepath.Join(dirPath, c)
						if _, err := os.Stat(fullPath); err == nil {
							validatedMap[title] = c
							referenced[c] = true
						}
					}
				} else if strings.HasSuffix(c, "/") {
					dirName := strings.TrimSuffix(c, "/")
					if actualDirs[dirName] {
						validatedMap[title] = c
						referenced[dirName+"/"] = true
					} else if strings.Contains(dirName, "/") {
						fullPath := filepath.Join(dirPath, dirName)
						if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
							validatedMap[title] = c
							referenced[dirName+"/"] = true
						}
					}
				} else {
					if actualDirs[c] {
						validatedMap[title] = c
						referenced[c+"/"] = true
					} else if strings.Contains(c, "/") {
						fullPath := filepath.Join(dirPath, c)
						if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
							validatedMap[title] = c
							referenced[c+"/"] = true
						}
					}
				}

			case []any:
				validatedItems := []any{}
				for _, item := range c {
					validated, refs := ValidateNavEntry(item, dirPath, actualFiles, actualDirs)
					if validated != nil {
						validatedItems = append(validatedItems, validated)
						for ref := range refs {
							referenced[ref] = true
						}
					}
				}

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

	return entry, referenced
}

// ValidateAndCleanNav validates existing .nav.yml, removes broken refs, adds new files.
func ValidateAndCleanNav(
	book *config.Book,
	stagingDir, navPath, dirPath string,
	logf func(string, ...any),
) error {
	navFile, err := ParseNavFile(navPath)
	if err != nil {
		relPath, relErr := filepath.Rel(stagingDir, navPath)
		if relErr != nil {
			relPath = navPath
		}
		logf("    Failed to parse %s, regenerating", relPath)
		return GenerateNavForDir(book, stagingDir, dirPath, logf)
	}

	actualFiles := ScanMarkdownFiles(dirPath)
	actualDirs := ScanSubdirectories(dirPath)

	validatedNav := make([]any, 0)
	referencedFiles := make(map[string]bool)

	for _, entry := range navFile.Nav {
		validated, referenced := ValidateNavEntry(entry, dirPath, actualFiles, actualDirs)
		if validated != nil {
			validatedNav = append(validatedNav, validated)
			for file := range referenced {
				referencedFiles[file] = true
			}
		}
	}

	newFiles := []string{}
	for file := range actualFiles {
		if !referencedFiles[file] {
			newFiles = append(newFiles, file)
		}
	}

	if len(newFiles) > 0 {
		sortedNew := SortNewFiles(book, stagingDir, newFiles, dirPath)
		for _, file := range sortedNew {
			validatedNav = append(validatedNav, file)
		}

		relDir, relErr := filepath.Rel(stagingDir, dirPath)
		if relErr != nil || relDir == "." {
			relDir = "(root)"
		}
		logf("    Added %d new files to %s/.nav.yml", len(newFiles), relDir)
	}

	navFile.Nav = validatedNav
	data, err := yaml.Marshal(navFile)
	if err != nil {
		return fmt.Errorf("marshaling nav file: %w", err)
	}

	return os.WriteFile(navPath, data, 0o644)
}
