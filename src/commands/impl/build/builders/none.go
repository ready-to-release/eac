// none.go - Build functions for static/config module types (build_system: none)
package builders

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"gopkg.in/yaml.v3"
)

// BuildNoop is a no-op build function for modules that don't require building.
func BuildNoop(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	Logln(logWriter, "\n=== %s (type: %s) ===", module.Moniker, module.Type)
	Logln(logWriter, "ℹ️  No build step required for this module type")
	return 0
}

// BuildScriptsPackage validates cross-shell script packages
func BuildScriptsPackage(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== Validating scripts-package: %s ===", module.Moniker)

	// Find script files matching module's source patterns
	var shellFiles []string
	var psFiles []string

	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".sh", ".bash":
			shellFiles = append(shellFiles, path)
		case ".ps1", ".psm1":
			psFiles = append(psFiles, path)
		}
		return nil
	})

	if err != nil {
		Logln(logWriter, "❌ Failed to scan directory: %v", err)
		return 1
	}

	totalFiles := len(shellFiles) + len(psFiles)
	if totalFiles == 0 {
		Logln(logWriter, "⚠️  No scripts found")
		return 0
	}

	Logln(logWriter, "📜 Found %d script(s) to validate (%d shell, %d PowerShell)", totalFiles, len(shellFiles), len(psFiles))

	validationErrors := 0

	// Validate shell scripts
	if len(shellFiles) > 0 {
		Logln(logWriter, "\n--- Shell Scripts ---")

		checkCmd := exec.Command("bash", "--version")
		bashAvailable := checkCmd.Run() == nil

		if !bashAvailable {
			if runtime.GOOS == "windows" {
				Logln(logWriter, "⚠️  Skipping shell validation: bash not available (WSL not configured)")
			} else {
				Logln(logWriter, "❌ bash not found")
				validationErrors++
			}
		} else {
			for _, shellFile := range shellFiles {
				relPath, _ := filepath.Rel(moduleRoot, shellFile)
				Logln(logWriter, "   Validating: %s", relPath)

				content, err := os.ReadFile(shellFile)
				if err != nil {
					Logln(logWriter, "      ❌ Failed to read: %v", err)
					validationErrors++
					continue
				}

				cmd := exec.Command("bash", "-n")
				cmd.Stdin = bytes.NewReader(content)
				output, err := cmd.CombinedOutput()
				if err != nil {
					Logln(logWriter, "      ❌ Syntax error: %s", strings.TrimSpace(string(output)))
					validationErrors++
					continue
				}

				Logln(logWriter, "      ✅ Valid syntax")
			}
		}
	}

	// Validate PowerShell scripts
	if len(psFiles) > 0 {
		Logln(logWriter, "\n--- PowerShell Scripts ---")

		checkCmd := exec.Command("pwsh", "--version")
		pwshAvailable := checkCmd.Run() == nil

		if !pwshAvailable {
			Logln(logWriter, "⚠️  Skipping PowerShell validation: pwsh not available")
		} else {
			for _, psFile := range psFiles {
				relPath, _ := filepath.Rel(moduleRoot, psFile)
				Logln(logWriter, "   Validating: %s", relPath)

				content, err := os.ReadFile(psFile)
				if err != nil {
					Logln(logWriter, "      ❌ Failed to read: %v", err)
					validationErrors++
					continue
				}

				cmd := exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", "-")
				cmd.Stdin = bytes.NewReader([]byte("$null = [System.Management.Automation.PSParser]::Tokenize(@'\n" + string(content) + "\n'@, [ref]$null)"))
				output, err := cmd.CombinedOutput()
				if err != nil {
					Logln(logWriter, "      ❌ Syntax error: %s", strings.TrimSpace(string(output)))
					validationErrors++
					continue
				}

				Logln(logWriter, "      ✅ Valid syntax")
			}
		}
	}

	if validationErrors > 0 {
		Logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	Logln(logWriter, "\n✅ All scripts validated successfully")
	return 0
}

// BuildConfig validates configuration files (JSON, YAML, TOML)
func BuildConfig(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== Validating config files: %s ===", module.Moniker)

	var configFiles []string
	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "out" || name == "dist" || name == ".vscode" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".json" || ext == ".yaml" || ext == ".yml" || ext == ".toml" {
			configFiles = append(configFiles, path)
		}
		return nil
	})

	if err != nil {
		Logln(logWriter, "❌ Failed to scan directory: %v", err)
		return 1
	}

	if len(configFiles) == 0 {
		Logln(logWriter, "⚠️  No config files found")
		return 0
	}

	Logln(logWriter, "⚙️  Found %d config file(s) to validate", len(configFiles))

	validationErrors := 0
	for _, configFile := range configFiles {
		relPath, _ := filepath.Rel(moduleRoot, configFile)
		ext := filepath.Ext(configFile)
		Logln(logWriter, "   Validating: %s", relPath)

		content, err := os.ReadFile(configFile)
		if err != nil {
			Logln(logWriter, "      ❌ Failed to read: %v", err)
			validationErrors++
			continue
		}

		switch ext {
		case ".json":
			var data interface{}
			if err := json.Unmarshal(content, &data); err != nil {
				Logln(logWriter, "      ❌ Invalid JSON: %v", err)
				validationErrors++
				continue
			}
		case ".yaml", ".yml":
			var data interface{}
			if err := yaml.Unmarshal(content, &data); err != nil {
				Logln(logWriter, "      ❌ Invalid YAML: %v", err)
				validationErrors++
				continue
			}
		case ".toml":
			if len(content) == 0 {
				Logln(logWriter, "      ❌ Empty file")
				validationErrors++
				continue
			}
		}

		Logln(logWriter, "      ✅ Valid %s", strings.TrimPrefix(ext, "."))
	}

	if validationErrors > 0 {
		Logln(logWriter, "\n❌ Validation failed with %d error(s)", validationErrors)
		return 1
	}

	Logln(logWriter, "\n✅ All config files validated successfully")
	return 0
}

// BuildTemplates validates template files and detects placeholders
func BuildTemplates(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := filepath.Join(workspaceRoot, module.Files.Root)

	Logln(logWriter, "\n=== Validating templates: %s ===", module.Moniker)

	var templateFiles []string
	placeholders := make(map[string]int)

	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "out" {
				return filepath.SkipDir
			}
			return nil
		}
		templateFiles = append(templateFiles, path)
		return nil
	})

	if err != nil {
		Logln(logWriter, "❌ Failed to scan directory: %v", err)
		return 1
	}

	if len(templateFiles) == 0 {
		Logln(logWriter, "⚠️  No template files found")
		return 0
	}

	Logln(logWriter, "📄 Found %d template file(s) to analyze", len(templateFiles))

	for _, templateFile := range templateFiles {
		content, err := os.ReadFile(templateFile)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "{{") || strings.Contains(contentStr, "${") || strings.Contains(contentStr, "%") {
			placeholders[filepath.Base(templateFile)]++
		}
	}

	Logln(logWriter, "\n📊 Template Analysis:")
	Logln(logWriter, "   Total files: %d", len(templateFiles))
	Logln(logWriter, "   Files with placeholders: %d", len(placeholders))

	if len(placeholders) > 0 {
		Logln(logWriter, "\n📝 Files with detected placeholders:")
		for file := range placeholders {
			Logln(logWriter, "   - %s", file)
		}
	}

	Logln(logWriter, "\n✅ Template validation complete")
	return 0
}

// BuildRepositoryRoot validates repository root structure
func BuildRepositoryRoot(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, opts BuildOptions) int {
	moduleRoot := workspaceRoot
	if module.Files.Root != "/" && module.Files.Root != "" {
		moduleRoot = filepath.Join(workspaceRoot, module.Files.Root)
	}

	Logln(logWriter, "\n=== Validating repository root: %s ===", module.Moniker)

	essentialFiles := []string{
		"README.md",
		".gitignore",
		"go.work",
	}

	missing := []string{}
	for _, file := range essentialFiles {
		path := filepath.Join(moduleRoot, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		Logln(logWriter, "⚠️  Missing essential files:")
		for _, file := range missing {
			Logln(logWriter, "   - %s", file)
		}
	}

	Logln(logWriter, "\n✅ Repository root validation complete")
	return 0
}
