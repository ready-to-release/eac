package design

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// ModuleInfo contains information about a module with architecture documentation
type ModuleInfo struct {
	Name         string
	Path         string
	HasWorkspace bool
	HasDocs      bool
	HasDecisions bool
	Description  string
	ViewCount    int
	DocCount     int
	DecisionCount int
}

// getRepoRoot returns the repository root directory
// Walks up the directory tree to find the repository root
func getRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree to find the repository root
	dir := cwd
	for {
		// Check if we're at a directory that has "src" as a subdirectory
		srcPath := filepath.Join(dir, "src")
		if stat, err := os.Stat(srcPath); err == nil && stat.IsDir() {
			// Found the root - this directory has a src subdirectory
			return dir, nil
		}

		// Check if we're in a src subdirectory structure
		base := filepath.Base(dir)
		parent := filepath.Dir(dir)

		// If we're in src/commands/design, src/commands, or src/cli
		if base == "design" || base == "commands" || base == "cli" {
			grandparent := filepath.Dir(parent)
			if filepath.Base(parent) == "src" {
				return grandparent, nil
			}
		}

		// If we're in src, go up one level
		if base == "src" {
			return parent, nil
		}

		// Move up one directory
		nextDir := filepath.Dir(dir)
		if nextDir == dir {
			// Reached the root of the filesystem, assume current dir
			return cwd, nil
		}
		dir = nextDir
	}
}

// ListAvailableModules scans for modules with architecture documentation
func ListAvailableModules() ([]ModuleInfo, error) {
	root, err := getRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to determine repository root: %w", err)
	}

	// Load module contracts
	registry, err := modules.LoadFromWorkspaceLatest(root)
	if err != nil {
		return nil, fmt.Errorf("failed to load module contracts: %w", err)
	}

	var moduleList []ModuleInfo

	// Iterate through all module contracts
	for _, module := range registry.All() {
		moniker := module.Moniker

		// Check if module has a design/ subdirectory in specs
		modulePath := GetModulePath(moniker)

		// Check if design directory exists
		if _, err := os.Stat(modulePath); os.IsNotExist(err) {
			// Skip modules without design folder
			continue
		}

		// Create module info
		moduleInfo := ModuleInfo{
			Name: moniker,
			Path: modulePath,
		}

		// Check for workspace.dsl
		workspacePath := filepath.Join(modulePath, "workspace.dsl")
		if _, err := os.Stat(workspacePath); err == nil {
			moduleInfo.HasWorkspace = true
		}

		// Check for docs folder
		docsPath := filepath.Join(modulePath, "docs")
		if stat, err := os.Stat(docsPath); err == nil && stat.IsDir() {
			moduleInfo.HasDocs = true
			// Count markdown files in docs
			docFiles, _ := filepath.Glob(filepath.Join(docsPath, "*.md"))
			moduleInfo.DocCount = len(docFiles)
		}

		// Check for decisions folder
		decisionsPath := filepath.Join(modulePath, "decisions")
		if stat, err := os.Stat(decisionsPath); err == nil && stat.IsDir() {
			moduleInfo.HasDecisions = true
			// Count markdown files in decisions
			decisionFiles, _ := filepath.Glob(filepath.Join(decisionsPath, "*.md"))
			moduleInfo.DecisionCount = len(decisionFiles)
		}

		// Count views in workspace.dsl (rough estimate)
		if moduleInfo.HasWorkspace {
			moduleInfo.ViewCount = countViewsInWorkspace(workspacePath)
		}

		// Only include if has workspace.dsl
		if moduleInfo.HasWorkspace {
			moduleList = append(moduleList, moduleInfo)
		}
	}

	return moduleList, nil
}

// ValidateModule checks if a module exists and has required files
func ValidateModule(module string) error {
	modulePath := GetModulePath(module)

	// Check if module directory exists
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return fmt.Errorf("module '%s' not found", module)
	}

	// Check if workspace.dsl exists
	workspacePath := filepath.Join(modulePath, "workspace.dsl")
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("module '%s' has no workspace.dsl file", module)
	}

	return nil
}

// GetModulePath returns the path to a module's design directory
// Module parameter should be the module moniker (e.g., "src-cli", "src-commands")
func GetModulePath(moniker string) string {
	root, err := getRepoRoot()
	if err != nil {
		// Fallback to relative path
		return filepath.Join("specs", moniker, "design")
	}
	return filepath.Join(root, "specs", moniker, "design")
}

// ModuleExists checks if a module exists (has a directory)
func ModuleExists(module string) bool {
	modulePath := GetModulePath(module)
	_, err := os.Stat(modulePath)
	return err == nil
}

// GetModuleStatus returns a human-readable status for a module
func (m *ModuleInfo) GetStatus() string {
	if !m.HasWorkspace {
		return "❌ Missing"
	}
	if m.HasDocs && m.HasDecisions {
		return "✅ Ready"
	}
	return "⚠️  Partial"
}

// countViewsInWorkspace counts the number of views defined in workspace.dsl
func countViewsInWorkspace(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	// Count view definitions (rough estimate)
	// Look for systemContext, container, component, deployment, dynamic, filtered
	viewCount := 0
	lines := string(content)

	// Simple counting of view keywords
	viewKeywords := []string{
		"systemContext",
		"container ",
		"component ",
		"deployment",
		"dynamic",
		"filtered",
	}

	for _, keyword := range viewKeywords {
		start := 0
		for {
			idx := indexOf(lines[start:], keyword)
			if idx == -1 {
				break
			}
			viewCount++
			start += idx + len(keyword)
		}
	}

	return viewCount
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
