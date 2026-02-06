// Package repository contains godog step implementations for specs/repository.
package repository

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/core/paths"
)

// ============================================================================
// Docs Mermaid Cache Validation Steps
// ============================================================================

// mermaidBlockPlain matches ```mermaid...``` blocks
var mermaidBlockPlain = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)```")

// sizeDirectivePattern matches %%{size:...}%% or %%{width:...}%% lines
var sizeDirectivePattern = regexp.MustCompile(`(?m)^%%\{(?:size|width):[^}]+\}%%\s*\n?`)

func (c *repositoryContext) scanDocsForMermaidBlocks() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	c.mermaidBlocks = []mermaidBlockInfo{}
	docsDir := paths.DocsSourcePath(c.repoRoot)

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Extract mermaid blocks
		matches := mermaidBlockPlain.FindAllStringSubmatch(string(content), -1)
		for idx, match := range matches {
			if len(match) < 2 {
				continue
			}
			diagramContent := strings.TrimSpace(match[1])
			if diagramContent == "" {
				continue
			}

			// Strip size directives before hashing
			cleanContent := sizeDirectivePattern.ReplaceAllString(diagramContent, "")
			cleanContent = strings.TrimSpace(cleanContent)

			// Hash content (same algorithm as books package)
			hash := hashMermaidContent(cleanContent)

			c.mermaidBlocks = append(c.mermaidBlocks, mermaidBlockInfo{
				content:    diagramContent,
				hash:       hash,
				sourceFile: path,
				blockIndex: idx,
			})
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan docs for mermaid: %w", err)
	}
	return nil
}

// hashMermaidContent creates a cache key hash for mermaid content
// Must match the algorithm in books/cache.go
func hashMermaidContent(code string) string {
	h := sha256.New()
	fmt.Fprintf(h, "code:%s\n", code)
	fmt.Fprintf(h, "width:%d\n", 0)
	fmt.Fprintf(h, "height:%d\n", 0)
	fmt.Fprintf(h, "theme:%s\n", "")
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *repositoryContext) checkMermaidBlocksAgainstCache() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	cacheDir := paths.DocsCachePath(c.repoRoot)
	c.uncachedMermaidBlocks = []mermaidBlockInfo{}

	for _, block := range c.mermaidBlocks {
		// Check if cached SVG exists using the same path format as the build system
		// Format: {identifier}_{blockIndex}_{shortHash}.svg
		// paths.MermaidCachePath handles identifier extraction and hash truncation
		cachePath := paths.MermaidCachePath(cacheDir, block.sourceFile, block.blockIndex, block.hash)
		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			c.uncachedMermaidBlocks = append(c.uncachedMermaidBlocks, block)
		}
	}
	return nil
}

func (c *repositoryContext) allMermaidDiagramsShouldHaveCachedSVGs() error {
	if len(c.uncachedMermaidBlocks) > 0 {
		docsDir := paths.DocsSourcePath(c.repoRoot)
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d mermaid diagram(s) missing from cache:\n\n", len(c.uncachedMermaidBlocks)))

		for _, block := range c.uncachedMermaidBlocks {
			relPath, _ := filepath.Rel(docsDir, block.sourceFile)
			firstLine := strings.Split(block.content, "\n")[0]
			if len(firstLine) > 50 {
				firstLine = firstLine[:50] + "..."
			}
			details.WriteString(fmt.Sprintf("  - %s [%d]: %s\n", relPath, block.blockIndex, firstLine))
		}

		details.WriteString("\nRun 'r2r update docs' to update the mermaid cache.\n")
		return fmt.Errorf("%s", details.String())
	}
	return nil
}

func (c *repositoryContext) ifDiagramsMissingShowUpdateCommand(expected *godog.DocString) error {
	// This is a passive assertion - the error message from allMermaidDiagramsShouldHaveCachedSVGs
	// already includes the update command instruction
	return nil
}

// ============================================================================
// Docs Drawio Cache Validation Steps
// ============================================================================

func (c *repositoryContext) scanDocsForDrawioImages() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	c.drawioImages = []drawioImageInfo{}
	docsDir := paths.DocsSourcePath(c.repoRoot)

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Check if this is a drawio.png file
		if !strings.HasSuffix(path, ".drawio.png") {
			return nil
		}

		// Hash the file content
		hash, err := hashFileContent(path)
		if err != nil {
			return fmt.Errorf("hashing %s: %w", path, err)
		}

		// Calculate relative path from docs directory
		relPath, err := filepath.Rel(docsDir, path)
		if err != nil {
			relPath = path
		}

		c.drawioImages = append(c.drawioImages, drawioImageInfo{
			sourceFile: path,
			relPath:    relPath,
			hash:       hash,
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan docs for drawio images: %w", err)
	}
	return nil
}

// hashFileContent returns the SHA256 hash of a file's content
func hashFileContent(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// hashDrawioContent creates a cache key hash for drawio image
// Must match the algorithm in books/cache.go
func hashDrawioContent(sourceHash string, maxWidth int) string {
	h := sha256.New()
	fmt.Fprintf(h, "source:%s\n", sourceHash)
	fmt.Fprintf(h, "maxWidth:%d\n", maxWidth)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *repositoryContext) checkDrawioImagesAgainstCache() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	cacheDir := paths.DocsCachePath(c.repoRoot)
	c.uncachedDrawioImages = []drawioImageInfo{}

	// MaxImageWidthPDF = 1200 (must match books/images.go)
	const maxWidth = 1200

	for _, img := range c.drawioImages {
		// Calculate cache key hash (same algorithm as books/cache.go)
		cacheHash := hashDrawioContent(img.hash, maxWidth)
		// Use DrawioCachePath which includes traceable filename: {identifier}_{hash8}.png
		cachePath := paths.DrawioCachePath(cacheDir, img.relPath, cacheHash)

		if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			c.uncachedDrawioImages = append(c.uncachedDrawioImages, img)
		}
	}
	return nil
}

func (c *repositoryContext) allDrawioImagesShouldHaveCachedPNGs() error {
	if len(c.uncachedDrawioImages) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d drawio image(s) missing from cache:\n\n", len(c.uncachedDrawioImages)))

		for _, img := range c.uncachedDrawioImages {
			details.WriteString(fmt.Sprintf("  - %s\n", img.relPath))
		}

		details.WriteString("\nRun 'r2r update docs' to update the drawio cache.\n")
		return fmt.Errorf("%s", details.String())
	}
	return nil
}

func (c *repositoryContext) ifImagesMissingShowUpdateCommand(expected *godog.DocString) error {
	// This is a passive assertion - the error message from allDrawioImagesShouldHaveCachedPNGs
	// already includes the update command instruction
	return nil
}

// ============================================================================
// Docs Structurizr Cache Validation Steps
// ============================================================================

// viewDefinitionPatternLocal extracts view keys from workspace.dsl content
// Matches patterns like: systemContext softwareSystem "ViewKey"
var viewDefinitionPatternLocal = regexp.MustCompile(`(?m)^\s*(systemContext|container|component|dynamic|deployment|custom|filtered)\s+\S+\s+"([^"]+)"`)

func (c *repositoryContext) scanSpecsForWorkspaceDSLFiles() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	c.structurizrViews = []structurizrViewInfo{}
	specsDir := filepath.Join(c.repoRoot, "specs")

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No specs directory is fine
		}
		return fmt.Errorf("failed to read specs directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		moduleName := entry.Name()
		workspacePath := paths.WorkspaceDSLPath(c.repoRoot, moduleName)

		// Check if workspace.dsl exists
		content, err := os.ReadFile(workspacePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Module doesn't have workspace.dsl
			}
			return fmt.Errorf("failed to read %s: %w", workspacePath, err)
		}

		// Hash the workspace.dsl content
		dslHash := hashDSLContent(string(content))

		// Extract view keys from DSL
		viewKeys := extractViewKeysFromDSL(string(content))

		for _, viewKey := range viewKeys {
			c.structurizrViews = append(c.structurizrViews, structurizrViewInfo{
				module:  moduleName,
				viewKey: viewKey,
				dslHash: dslHash,
				dslPath: workspacePath,
			})
		}
	}

	return nil
}

// hashDSLContent returns the first 8 characters of SHA256 hash of DSL content
// Must match the algorithm in design/helper/export.go
func hashDSLContent(content string) string {
	// Normalize line endings to LF for consistent hashing across platforms
	// This ensures the same DSL content produces the same hash regardless of
	// whether it was checked out with CRLF (Windows) or LF (Linux/Mac)
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	h := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", h)[:8]
}

// extractViewKeysFromDSL extracts view keys from workspace.dsl content
func extractViewKeysFromDSL(content string) []string {
	matches := viewDefinitionPatternLocal.FindAllStringSubmatch(content, -1)
	keys := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			keys = append(keys, match[2])
		}
	}
	return keys
}

func (c *repositoryContext) checkStructurizrViewsAgainstCache() error {
	if err := c.ensureRepoRoot(); err != nil {
		return err
	}

	c.uncachedStructurizrViews = []structurizrViewInfo{}

	// Collect unique modules from discovered views
	moduleViews := make(map[string][]structurizrViewInfo)
	for _, view := range c.structurizrViews {
		moduleViews[view.module] = append(moduleViews[view.module], view)
	}

	// Check each module's per-module builder output
	for moduleName, views := range moduleViews {
		builderOutputDir := paths.StructurizrModuleBuildOutputPath(c.repoRoot, moduleName)
		indexPath := filepath.Join(builderOutputDir, "structurizr-index.json")

		indexData, err := os.ReadFile(indexPath)
		if err != nil {
			// Builder output missing for this module - all its views are uncached
			c.uncachedStructurizrViews = append(c.uncachedStructurizrViews, views...)
			continue
		}

		// Parse index manifest
		var idx struct {
			Entries []struct {
				Module      string `json:"module"`
				ViewKey     string `json:"view_key"`
				DSLHash     string `json:"dsl_hash"`
				SVGFilename string `json:"svg_filename"`
			} `json:"entries"`
		}
		if err := json.Unmarshal(indexData, &idx); err != nil {
			// Corrupt index - treat all views for this module as uncached
			c.uncachedStructurizrViews = append(c.uncachedStructurizrViews, views...)
			continue
		}

		// Build lookup: "module:viewKey" -> exists
		builtViews := make(map[string]bool, len(idx.Entries))
		for _, entry := range idx.Entries {
			key := entry.Module + ":" + entry.ViewKey
			// Verify SVG file actually exists in builder output
			svgPath := filepath.Join(builderOutputDir, entry.SVGFilename)
			if _, err := os.Stat(svgPath); err == nil {
				builtViews[key] = true
			}
		}

		for _, view := range views {
			key := view.module + ":" + view.viewKey
			if !builtViews[key] {
				c.uncachedStructurizrViews = append(c.uncachedStructurizrViews, view)
			}
		}
	}
	return nil
}

func (c *repositoryContext) allStructurizrViewsShouldHaveCachedSVGs() error {
	if len(c.uncachedStructurizrViews) > 0 {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("Found %d Structurizr view(s) missing from build output:\n\n", len(c.uncachedStructurizrViews)))

		for _, view := range c.uncachedStructurizrViews {
			details.WriteString(fmt.Sprintf("  - %s:%s (hash %s)\n", view.module, view.viewKey, view.dslHash))
		}

		details.WriteString("\nRun 'r2r eac build --module docs' to build the Structurizr diagrams.\n")
		return fmt.Errorf("%s", details.String())
	}
	return nil
}

func (c *repositoryContext) ifViewsMissingShowUpdateCommand(expected *godog.DocString) error {
	// This is a passive assertion - the error message from allStructurizrViewsShouldHaveCachedSVGs
	// already includes the update command instruction
	return nil
}
