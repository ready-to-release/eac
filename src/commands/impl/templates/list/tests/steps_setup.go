package tests

import (
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// iHaveATemplateDirectory creates a template directory
func iHaveATemplateDirectory(dirPath string) error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}
	fullPath := filepath.Join(ctx.WorkDir(), dirPath)
	return os.MkdirAll(fullPath, 0755)
}

// iHaveATemplateFileWithContent creates a template file with content
func iHaveATemplateFileWithContent(filePath string, content *godog.DocString) error {
	ctx, err := getActiveContext()
	if err != nil {
		return err
	}
	fullPath := filepath.Join(ctx.WorkDir(), filePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content.Content), 0644)
}
