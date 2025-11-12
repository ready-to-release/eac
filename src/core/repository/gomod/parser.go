package gomod

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseGoMod parses a go.mod file and extracts module information
func ParseGoMod(filePath, rootPath string) (*GoModInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer file.Close()

	info := &GoModInfo{
		FilePath: filePath,
		Requires: []Require{},
		Replaces: []Replace{},
	}

	// Extract module path
	modulePath, err := extractModulePath(file)
	if err != nil {
		return nil, err
	}
	info.ModulePath = modulePath

	// Reset file pointer
	file.Seek(0, 0)

	// Extract requires
	requires, err := extractRequires(file)
	if err != nil {
		return nil, err
	}
	info.Requires = requires

	// Reset file pointer
	file.Seek(0, 0)

	// Extract replaces
	replaces, err := extractReplaces(file)
	if err != nil {
		return nil, err
	}
	info.Replaces = replaces

	// Extract module directory relative to root
	moduleDir, err := ExtractModuleDir(filePath, rootPath)
	if err != nil {
		return nil, err
	}
	info.ModuleDir = moduleDir

	return info, nil
}

// extractModulePath extracts the module declaration from go.mod
// Example: "module github.com/ready-to-release/eac/src/cli"
func extractModulePath(file *os.File) (string, error) {
	scanner := bufio.NewScanner(file)
	moduleRegex := regexp.MustCompile(`^module\s+(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := moduleRegex.FindStringSubmatch(line); matches != nil {
			return strings.TrimSpace(matches[1]), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning file: %w", err)
	}

	return "", fmt.Errorf("no module declaration found")
}

// extractRequires extracts require statements from go.mod
// Only includes direct dependencies (not marked with // indirect)
func extractRequires(file *os.File) ([]Require, error) {
	scanner := bufio.NewScanner(file)
	var requires []Require
	inRequireBlock := false

	// Regex patterns
	requireBlockStart := regexp.MustCompile(`^require\s*\($`)
	requireBlockEnd := regexp.MustCompile(`^\)`)
	singleRequire := regexp.MustCompile(`^require\s+(\S+)\s+(\S+)`)
	blockRequire := regexp.MustCompile(`^\s*(\S+)\s+(\S+)`)
	indirectComment := regexp.MustCompile(`//\s*indirect`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Check for require block start
		if requireBlockStart.MatchString(trimmed) {
			inRequireBlock = true
			continue
		}

		// Check for require block end
		if inRequireBlock && requireBlockEnd.MatchString(trimmed) {
			inRequireBlock = false
			continue
		}

		// Parse single-line require
		if matches := singleRequire.FindStringSubmatch(trimmed); matches != nil {
			indirect := indirectComment.MatchString(line)
			requires = append(requires, Require{
				Path:     matches[1],
				Version:  matches[2],
				Indirect: indirect,
			})
			continue
		}

		// Parse require within block
		if inRequireBlock {
			if matches := blockRequire.FindStringSubmatch(trimmed); matches != nil {
				indirect := indirectComment.MatchString(line)
				requires = append(requires, Require{
					Path:     matches[1],
					Version:  matches[2],
					Indirect: indirect,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning requires: %w", err)
	}

	return requires, nil
}

// extractReplaces extracts replace directives from go.mod
// Example: "replace github.com/ready-to-release/eac/src/core => ../internal"
func extractReplaces(file *os.File) ([]Replace, error) {
	scanner := bufio.NewScanner(file)
	var replaces []Replace

	// Regex patterns
	replaceRegex := regexp.MustCompile(`^replace\s+(\S+)(?:\s+\S+)?\s+=>\s+(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Parse replace directive
		if matches := replaceRegex.FindStringSubmatch(line); matches != nil {
			replaces = append(replaces, Replace{
				OldPath: strings.TrimSpace(matches[1]),
				NewPath: strings.TrimSpace(matches[2]),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning replaces: %w", err)
	}

	return replaces, nil
}

// FilterInternalDependencies filters requires to only include internal modules
// Internal modules are those that start with the base module path
func FilterInternalDependencies(requires []Require, baseModulePath string) []string {
	var internal []string

	for _, req := range requires {
		// Skip indirect dependencies
		if req.Indirect {
			continue
		}

		// Check if it's an internal module
		if IsInternalModule(req.Path, baseModulePath) {
			internal = append(internal, req.Path)
		}
	}

	return internal
}

// ParseAllGoMods finds and parses all go.mod files in the repository
func ParseAllGoMods(rootPath string, excludeDirs []string) ([]*GoModInfo, error) {
	// Find all go.mod files
	goModFiles, err := FindGoModFiles(rootPath, excludeDirs)
	if err != nil {
		return nil, fmt.Errorf("failed to find go.mod files: %w", err)
	}

	// Parse each go.mod file
	var infos []*GoModInfo
	for _, goModPath := range goModFiles {
		info, err := ParseGoMod(goModPath, rootPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", goModPath, err)
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// GetModuleNameFromPath extracts a simple module name from the module directory
// Example: "src/cli" -> "cli"
// Example: "src/mcp/commands" -> "mcp-commands"
func GetModuleNameFromPath(moduleDir string) string {
	// Convert to forward slashes for consistency
	moduleDir = filepath.ToSlash(moduleDir)

	// Remove "src/" prefix if present
	moduleDir = strings.TrimPrefix(moduleDir, "src/")

	// Replace slashes with hyphens
	name := strings.ReplaceAll(moduleDir, "/", "-")

	return name
}
