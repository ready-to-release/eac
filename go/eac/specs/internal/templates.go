// Package internal provides shared helpers for godog BDD tests.
//
// This file contains template definitions for creating EAC configuration files
// in isolated test environments. Templates provide consistent, reusable
// configuration structures across all test scenarios.
//
// Design Principles:
// 1. Named templates provide complete, working configurations
// 2. Templates can be composed (base + overlay)
// 3. Template parameters use {{PLACEHOLDER}} syntax for substitution
// 4. All templates produce valid EAC configurations
package internal

import (
	"fmt"
	"strings"
)

// ============================================================================
// Template Types
// ============================================================================

// Template represents a named configuration template with content.
type Template struct {
	Name        string            // Unique template identifier
	Description string            // Human-readable description
	Files       map[string]string // Map of file path -> content
}

// TemplateParams holds parameters for template substitution.
type TemplateParams map[string]string

// ============================================================================
// Module Types Templates
// ============================================================================

// ModuleTypesGoLibrary is the module-types.yml for Go library modules.
const ModuleTypesGoLibrary = `types:
  - name: go-library
    description: Go library module
    capabilities:
      - go_module
`

// ModuleTypesGoCommand is the module-types.yml for Go command modules.
const ModuleTypesGoCommand = `types:
  - name: go-command
    description: Go CLI application
    capabilities:
      - go_module
      - executable
`

// ModuleTypesGo is the module-types.yml for both Go library and command modules.
const ModuleTypesGo = `types:
  - name: go-library
    description: Go library module
    capabilities:
      - go_module

  - name: go-command
    description: Go CLI application
    capabilities:
      - go_module
      - executable
`

// ModuleTypesDocker is the module-types.yml with Docker support.
const ModuleTypesDocker = `types:
  - name: go-library
    description: Go library module
    capabilities:
      - go_module

  - name: go-command
    description: Go CLI application
    capabilities:
      - go_module
      - executable

  - name: docker-image
    description: Docker image
    capabilities:
      - docker
`

// ModuleTypesComplete is the module-types.yml with all common module types.
const ModuleTypesComplete = `types:
  - name: go-library
    description: Go library module
    capabilities:
      - go_module

  - name: go-command
    description: Go CLI application
    capabilities:
      - go_module
      - executable

  - name: docker-image
    description: Docker image
    capabilities:
      - docker

  - name: mkdocs-book
    description: MkDocs documentation book
    capabilities:
      - mkdocs

  - name: gherkin-spec
    description: Gherkin specification module
    capabilities:
      - gherkin
`

// ============================================================================
// Modules Templates
// ============================================================================

// ModulesSingleGoLibrary is modules.yml with a single Go library module.
// Parameters: {{MODULE_NAME}}, {{MODULE_PATH}}
const ModulesSingleGoLibrary = `modules:
  - moniker: {{MODULE_NAME}}
    name: {{MODULE_NAME}} Module
    type: go-library
    files:
      root: {{MODULE_PATH}}
`

// ModulesTwoGoLibraries is modules.yml with two Go library modules.
// Parameters: {{MODULE1_NAME}}, {{MODULE1_PATH}}, {{MODULE2_NAME}}, {{MODULE2_PATH}}
const ModulesTwoGoLibraries = `modules:
  - moniker: {{MODULE1_NAME}}
    name: {{MODULE1_NAME}} Module
    type: go-library
    files:
      root: {{MODULE1_PATH}}

  - moniker: {{MODULE2_NAME}}
    name: {{MODULE2_NAME}} Module
    type: go-library
    files:
      root: {{MODULE2_PATH}}
`

// ============================================================================
// System Dependencies Templates
// ============================================================================

// SystemDependenciesMinimal is minimal system-dependencies.yml.
const SystemDependenciesMinimal = `dependencies:
  - name: go
    version: ">=1.21"
`

// SystemDependenciesWithDocker is system-dependencies.yml with Docker.
const SystemDependenciesWithDocker = `dependencies:
  - name: go
    version: ">=1.21"
  - name: docker
    version: ">=24.0"
`

// ============================================================================
// Named Template Registry
// ============================================================================

// namedTemplates maps template names to their definitions.
var namedTemplates = map[string]*Template{
	"minimal": {
		Name:        "minimal",
		Description: "Minimal EAC configuration with a single Go library module",
		Files: map[string]string{
			".r2r/eac/modules.yml":      ModulesSingleGoLibrary,
			".r2r/eac/module-types.yml": ModuleTypesGoLibrary,
		},
	},
	"minimal-go": {
		Name:        "minimal-go",
		Description: "Minimal EAC config for Go development",
		Files: map[string]string{
			".r2r/eac/modules.yml":            ModulesSingleGoLibrary,
			".r2r/eac/module-types.yml":       ModuleTypesGo,
			".r2r/eac/system-dependencies.yml": SystemDependenciesMinimal,
		},
	},
	"minimal-with-docker": {
		Name:        "minimal-with-docker",
		Description: "Minimal EAC config with Docker support",
		Files: map[string]string{
			".r2r/eac/modules.yml":            ModulesSingleGoLibrary,
			".r2r/eac/module-types.yml":       ModuleTypesDocker,
			".r2r/eac/system-dependencies.yml": SystemDependenciesWithDocker,
		},
	},
	"multi-module": {
		Name:        "multi-module",
		Description: "EAC config with two Go library modules",
		Files: map[string]string{
			".r2r/eac/modules.yml":            ModulesTwoGoLibraries,
			".r2r/eac/module-types.yml":       ModuleTypesGo,
			".r2r/eac/system-dependencies.yml": SystemDependenciesMinimal,
		},
	},
	"complete": {
		Name:        "complete",
		Description: "Complete EAC config with all module types",
		Files: map[string]string{
			".r2r/eac/modules.yml":            ModulesSingleGoLibrary,
			".r2r/eac/module-types.yml":       ModuleTypesComplete,
			".r2r/eac/system-dependencies.yml": SystemDependenciesWithDocker,
		},
	},
}

// ============================================================================
// Template Functions
// ============================================================================

// GetTemplate returns a named template by name.
// Returns nil if the template doesn't exist.
func GetTemplate(name string) *Template {
	return namedTemplates[name]
}

// ListTemplates returns all available template names.
func ListTemplates() []string {
	names := make([]string, 0, len(namedTemplates))
	for name := range namedTemplates {
		names = append(names, name)
	}
	return names
}

// ApplyTemplate applies a named template to the test context with parameter substitution.
// Parameters are substituted using {{KEY}} syntax in template content.
//
// Example:
//
//	ApplyTemplate(ctx, "minimal", TemplateParams{
//	    "MODULE_NAME": "test-module",
//	    "MODULE_PATH": "go/test-module",
//	})
func ApplyTemplate(ctx *TestContext, templateName string, params TemplateParams) error {
	tmpl := GetTemplate(templateName)
	if tmpl == nil {
		return fmt.Errorf("template %q not found. Available: %v", templateName, ListTemplates())
	}

	for path, content := range tmpl.Files {
		// Apply parameter substitution
		resolvedContent := substituteParams(content, params)
		if err := CreateFile(ctx, path, resolvedContent); err != nil {
			return fmt.Errorf("failed to create %s from template %s: %w", path, templateName, err)
		}
	}

	return nil
}

// ApplyTemplateFile applies a single file from a template with parameter substitution.
// Useful when you only need part of a template.
func ApplyTemplateFile(ctx *TestContext, templateName, filePath string, params TemplateParams) error {
	tmpl := GetTemplate(templateName)
	if tmpl == nil {
		return fmt.Errorf("template %q not found", templateName)
	}

	content, ok := tmpl.Files[filePath]
	if !ok {
		return fmt.Errorf("file %q not found in template %q", filePath, templateName)
	}

	resolvedContent := substituteParams(content, params)
	return CreateFile(ctx, filePath, resolvedContent)
}

// substituteParams replaces {{KEY}} placeholders with values from params.
func substituteParams(content string, params TemplateParams) string {
	result := content
	for key, value := range params {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// ============================================================================
// Template Composition Functions
// ============================================================================

// MergeTemplates combines multiple templates into one.
// Later templates override files from earlier templates.
func MergeTemplates(templates ...*Template) *Template {
	merged := &Template{
		Name:        "merged",
		Description: "Merged from multiple templates",
		Files:       make(map[string]string),
	}

	for _, t := range templates {
		if t == nil {
			continue
		}
		for path, content := range t.Files {
			merged.Files[path] = content
		}
	}

	return merged
}

// TemplateWithOverrides creates a new template with file overrides.
// Useful for modifying specific files while keeping the rest of a template.
func TemplateWithOverrides(base *Template, overrides map[string]string) *Template {
	if base == nil {
		return &Template{
			Name:  "custom",
			Files: overrides,
		}
	}

	result := &Template{
		Name:        base.Name + "-custom",
		Description: base.Description + " (with overrides)",
		Files:       make(map[string]string),
	}

	// Copy base files
	for path, content := range base.Files {
		result.Files[path] = content
	}

	// Apply overrides
	for path, content := range overrides {
		result.Files[path] = content
	}

	return result
}

// ============================================================================
// Convenience Constants for Direct Use
// ============================================================================

// DefaultModuleParams returns default parameters for a single module template.
func DefaultModuleParams(moduleName string) TemplateParams {
	return TemplateParams{
		"MODULE_NAME": moduleName,
		"MODULE_PATH": "go/" + moduleName,
	}
}

// TwoModuleParams returns parameters for a two-module template.
func TwoModuleParams(module1, module2 string) TemplateParams {
	return TemplateParams{
		"MODULE1_NAME": module1,
		"MODULE1_PATH": "go/" + module1,
		"MODULE2_NAME": module2,
		"MODULE2_PATH": "go/" + module2,
	}
}
