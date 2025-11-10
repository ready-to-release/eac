package gomod

import (
	"os"
	"path/filepath"
	"strings"
)

// FindGoModFiles finds all go.mod files in the repository
// Excludes: vendor/, .git/, and specified exclude directories
func FindGoModFiles(rootPath string, excludeDirs []string) ([]string, error) {
	var goModFiles []string

	// Default exclude directories
	defaultExcludes := []string{
		".git",
		"vendor",
		"node_modules",
	}
	excludeDirs = append(excludeDirs, defaultExcludes...)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() {
			for _, exclude := range excludeDirs {
				if info.Name() == exclude {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if it's a go.mod file
		if info.Name() == "go.mod" {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			goModFiles = append(goModFiles, absPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return goModFiles, nil
}

// ExtractModuleDir extracts the relative directory from a go.mod file path
// Example: /c/projects/eac/src/cli/go.mod -> src/cli
func ExtractModuleDir(goModPath, rootPath string) (string, error) {
	absGoMod, err := filepath.Abs(goModPath)
	if err != nil {
		return "", err
	}

	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return "", err
	}

	// Get directory containing go.mod
	modDir := filepath.Dir(absGoMod)

	// Make it relative to root
	relPath, err := filepath.Rel(absRoot, modDir)
	if err != nil {
		return "", err
	}

	// Convert to forward slashes for consistency
	relPath = filepath.ToSlash(relPath)

	return relPath, nil
}

// IsInternalModule checks if a module path belongs to this repository
func IsInternalModule(modulePath, baseModulePath string) bool {
	return strings.HasPrefix(modulePath, baseModulePath)
}
