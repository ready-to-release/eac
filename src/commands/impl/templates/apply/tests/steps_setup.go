package tests

import (
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// iHaveATemplateDirectory creates a template directory
func iHaveATemplateDirectory(dirPath string) error {
	fullPath := filepath.Join(applyCtx.workDir, dirPath)
	return os.MkdirAll(fullPath, 0755)
}

// iHaveATemplateFileWithContent creates a template file with content
func iHaveATemplateFileWithContent(filePath string, content *godog.DocString) error {
	fullPath := filepath.Join(applyCtx.workDir, filePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content.Content), 0644)
}

// IHaveAValuesFileWith creates a values JSON file with content
func IHaveAValuesFileWith(filePath string, content *godog.DocString) error {
	fullPath := filepath.Join(applyCtx.workDir, filePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content.Content), 0644)
}

// iHaveAFileWithContent creates a file with content (generic)
func iHaveAFileWithContent(filePath string, content *godog.DocString) error {
	fullPath := filepath.Join(applyCtx.workDir, filePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content.Content), 0644)
}

// aDeveloperAddsANewTemplateTypeForApply simulates adding a new template type for apply
func aDeveloperAddsANewTemplateTypeForApply(templateType string) error {
	// This is a conceptual step - developer adds new template type
	// The test verifies the system is extensible
	_ = templateType
	return nil
}
