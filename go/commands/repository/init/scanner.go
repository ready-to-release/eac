// File: go/cli/eac/impl/init/scanner.go
package init

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ScanResult represents the detected structure of a repository
type ScanResult struct {
	Modules []ModuleInfo `json:"modules"`
}

// ModuleInfo describes a detected module
type ModuleInfo struct {
	Name      string   `json:"name"`       // Inferred module name (from directory or manifest)
	Root      string   `json:"root"`       // Relative path to module root
	Language  string   `json:"language"`   // Primary language (go, python, rust, typescript, dotnet, java)
	BuildTool string   `json:"build_tool"` // Build tool (go, cargo, npm, pip, dotnet, maven, gradle)
	Files     []string `json:"files"`      // Key files found (go.mod, Cargo.toml, etc.)
}

// packageManagerFiles maps file patterns to language and build tool information
var packageManagerFiles = map[string]struct {
	language  string
	buildTool string
	detector  func(root string, file string) (*ModuleInfo, error)
}{
	"go.mod": {
		language:  "go",
		buildTool: "go",
		detector:  detectGoModule,
	},
	"Cargo.toml": {
		language:  "rust",
		buildTool: "cargo",
		detector:  detectRustModule,
	},
	"package.json": {
		language:  "javascript", // May be upgraded to typescript
		buildTool: "npm",
		detector:  detectNodeModule,
	},
	"pyproject.toml": {
		language:  "python",
		buildTool: "pip",
		detector:  detectPythonModule,
	},
	"setup.py": {
		language:  "python",
		buildTool: "pip",
		detector:  detectPythonSetupModule,
	},
	"*.csproj": {
		language:  "dotnet",
		buildTool: "dotnet",
		detector:  detectDotNetModule,
	},
	"*.fsproj": {
		language:  "dotnet",
		buildTool: "dotnet",
		detector:  detectDotNetModule,
	},
	"pom.xml": {
		language:  "java",
		buildTool: "maven",
		detector:  detectMavenModule,
	},
	"build.gradle": {
		language:  "java",
		buildTool: "gradle",
		detector:  detectGradleModule,
	},
}

// vendorDirectories are directories that should be filtered out
var vendorDirectories = map[string]bool{
	"vendor":        true,
	"node_modules":  true,
	"target":        true,
	".git":          true,
	".svn":          true,
	".hg":           true,
	"dist":          true,
	"build":         true,
	"out":           true,
	"bin":           true,
	"obj":           true,
	".venv":         true,
	"venv":          true,
	".pytest_cache": true,
	"__pycache__":   true,
	".idea":         true,
	".vscode":       true,
}

// ScanRepository scans a repository and returns detected modules
func ScanRepository(repoRoot string) (*ScanResult, error) {
	if repoRoot == "" {
		return nil, fmt.Errorf("repository root cannot be empty")
	}

	// Verify directory exists
	info, err := os.Stat(repoRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s", repoRoot)
		}
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", repoRoot)
	}

	result := &ScanResult{
		Modules: []ModuleInfo{},
	}

	// Track module roots to avoid duplicates
	moduleRoots := make(map[string]bool)

	// Walk the directory tree
	err = filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor directories
		if info.IsDir() && shouldSkipDirectory(info.Name()) {
			return filepath.SkipDir
		}

		// Check if this is a package manager file
		if !info.IsDir() {
			moduleInfo := detectModule(repoRoot, path, info.Name())
			if moduleInfo != nil {
				// Get the directory containing this file
				moduleRoot := filepath.Dir(path)

				// Skip if we've already found a module at this root
				if moduleRoots[moduleRoot] {
					return nil
				}

				// Mark this root as found
				moduleRoots[moduleRoot] = true

				// Convert absolute path to relative
				relRoot, err := filepath.Rel(repoRoot, moduleRoot)
				if err != nil {
					relRoot = moduleRoot
				}
				moduleInfo.Root = relRoot

				result.Modules = append(result.Modules, *moduleInfo)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	return result, nil
}

// shouldSkipDirectory returns true if the directory should be skipped during scanning
func shouldSkipDirectory(name string) bool {
	return vendorDirectories[name]
}

// detectModule detects if a file is a package manager file and returns module info
func detectModule(repoRoot, fullPath, fileName string) *ModuleInfo {
	// Direct file name matches
	if pm, ok := packageManagerFiles[fileName]; ok {
		moduleInfo, err := pm.detector(repoRoot, fullPath)
		if err == nil && moduleInfo != nil {
			return moduleInfo
		}
	}

	// Pattern matches (e.g., *.csproj)
	for pattern, pm := range packageManagerFiles {
		if strings.Contains(pattern, "*") {
			ext := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(fileName, ext) {
				moduleInfo, err := pm.detector(repoRoot, fullPath)
				if err == nil && moduleInfo != nil {
					return moduleInfo
				}
			}
		}
	}

	return nil
}

// detectGoModule detects a Go module from go.mod
func detectGoModule(repoRoot, goModPath string) (*ModuleInfo, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}

	// Parse module name from go.mod
	moduleName := ""
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Get last component of module path
				fullModuleName := parts[1]
				pathParts := strings.Split(fullModuleName, "/")
				moduleName = pathParts[len(pathParts)-1]
				break
			}
		}
	}

	// Fallback to directory name if we couldn't parse module name
	if moduleName == "" {
		moduleName = filepath.Base(filepath.Dir(goModPath))
	}

	return &ModuleInfo{
		Name:      moduleName,
		Language:  "go",
		BuildTool: "go",
		Files:     []string{"go.mod"},
	}, nil
}

// detectPythonModule detects a Python module from pyproject.toml
func detectPythonModule(repoRoot, tomlPath string) (*ModuleInfo, error) {
	content, err := os.ReadFile(tomlPath)
	if err != nil {
		return nil, err
	}

	// Parse TOML to get project name
	var config struct {
		Project struct {
			Name string `toml:"name"`
		} `toml:"project"`
	}

	if err := toml.Unmarshal(content, &config); err != nil {
		// If parsing fails, use directory name
		return &ModuleInfo{
			Name:      filepath.Base(filepath.Dir(tomlPath)),
			Language:  "python",
			BuildTool: "pip",
			Files:     []string{"pyproject.toml"},
			}, nil
	}

	moduleName := config.Project.Name
	if moduleName == "" {
		moduleName = filepath.Base(filepath.Dir(tomlPath))
	}

	return &ModuleInfo{
		Name:      moduleName,
		Language:  "python",
		BuildTool: "pip",
		Files:     []string{"pyproject.toml"},
	}, nil
}

// detectPythonSetupModule detects a Python module from setup.py
func detectPythonSetupModule(repoRoot, setupPath string) (*ModuleInfo, error) {
	// For setup.py, we'll use the directory name as the module name
	// since parsing setup.py reliably would require running Python code
	moduleName := filepath.Base(filepath.Dir(setupPath))

	return &ModuleInfo{
		Name:      moduleName,
		Language:  "python",
		BuildTool: "pip",
		Files:     []string{"setup.py"},
	}, nil
}

// detectRustModule detects a Rust module from Cargo.toml
func detectRustModule(repoRoot, cargoPath string) (*ModuleInfo, error) {
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return nil, err
	}

	// Parse TOML to get package name
	var config struct {
		Package struct {
			Name string `toml:"name"`
		} `toml:"package"`
	}

	if err := toml.Unmarshal(content, &config); err != nil {
		// If parsing fails, use directory name
		return &ModuleInfo{
			Name:      filepath.Base(filepath.Dir(cargoPath)),
			Language:  "rust",
			BuildTool: "cargo",
			Files:     []string{"Cargo.toml"},
			}, nil
	}

	moduleName := config.Package.Name
	if moduleName == "" {
		moduleName = filepath.Base(filepath.Dir(cargoPath))
	}

	return &ModuleInfo{
		Name:      moduleName,
		Language:  "rust",
		BuildTool: "cargo",
		Files:     []string{"Cargo.toml"},
	}, nil
}

// detectNodeModule detects a Node.js/TypeScript module from package.json
func detectNodeModule(repoRoot, packagePath string) (*ModuleInfo, error) {
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return nil, err
	}

	// Parse package.json
	var pkg struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		// If parsing fails, use directory name
		return &ModuleInfo{
			Name:      filepath.Base(filepath.Dir(packagePath)),
			Language:  "javascript",
			BuildTool: "npm",
			Files:     []string{"package.json"},
			}, nil
	}

	moduleName := pkg.Name
	if moduleName == "" {
		moduleName = filepath.Base(filepath.Dir(packagePath))
	}

	// Check if this is a TypeScript project
	language := "javascript"
	files := []string{"package.json"}

	moduleDir := filepath.Dir(packagePath)
	tsconfigPath := filepath.Join(moduleDir, "tsconfig.json")
	if _, err := os.Stat(tsconfigPath); err == nil {
		language = "typescript"
		files = append(files, "tsconfig.json")
	}

	return &ModuleInfo{
		Name:      moduleName,
		Language:  language,
		BuildTool: "npm",
		Files:     files,
	}, nil
}

// detectDotNetModule detects a .NET module from .csproj or .fsproj
func detectDotNetModule(repoRoot, projPath string) (*ModuleInfo, error) {
	// Use the project file name (without extension) as the module name
	fileName := filepath.Base(projPath)
	moduleName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	return &ModuleInfo{
		Name:      moduleName,
		Language:  "dotnet",
		BuildTool: "dotnet",
		Files:     []string{fileName},
	}, nil
}

// detectMavenModule detects a Java module from pom.xml
func detectMavenModule(repoRoot, pomPath string) (*ModuleInfo, error) {
	// Use directory name as module name for Maven projects
	// Parsing pom.xml would require XML parsing which is more complex
	moduleName := filepath.Base(filepath.Dir(pomPath))

	return &ModuleInfo{
		Name:      moduleName,
		Language:  "java",
		BuildTool: "maven",
		Files:     []string{"pom.xml"},
	}, nil
}

// detectGradleModule detects a Java module from build.gradle
func detectGradleModule(repoRoot, gradlePath string) (*ModuleInfo, error) {
	// Use directory name as module name for Gradle projects
	moduleName := filepath.Base(filepath.Dir(gradlePath))

	return &ModuleInfo{
		Name:      moduleName,
		Language:  "java",
		BuildTool: "gradle",
		Files:     []string{"build.gradle"},
	}, nil
}
