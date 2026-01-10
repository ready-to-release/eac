// structurizr.go provides Structurizr diagram processing for book builds
package books

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	design "github.com/ready-to-release/eac/go/eac/commands/impl/design/helper"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// StructurizrMarker represents a Structurizr diagram marker found in markdown
type StructurizrMarker struct {
	Module    string // Module name (e.g., "eac-commands")
	ViewKey   string // View key (e.g., "SystemContext")
	StartPos  int    // Start position of the marker in the file
	EndPos    int    // End position of the marker in the file
	FullMatch string // The full marker text
}

// structurizrMarkerPattern matches <!-- structurizr:module:viewKey --> markers
// Examples:
//   - <!-- structurizr:eac-commands:SystemContext -->
//   - <!-- structurizr:r2r-cli:Containers -->
var structurizrMarkerPattern = regexp.MustCompile(`<!--\s*structurizr:([^:]+):([^>\s]+)\s*-->`)

// processStructurizrDiagrams scans staging markdown for Structurizr markers
// and replaces them with img tags pointing to cached SVGs
func (p *Preprocessor) processStructurizrDiagrams() error {
	// Build cache of DSL hashes for each module
	dslHashes := make(map[string]string)

	// Scan for markers in staging directory
	markersByFile := make(map[string][]StructurizrMarker)
	modulesUsed := make(map[string]bool)

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		markers := extractStructurizrMarkers(string(content))
		if len(markers) > 0 {
			markersByFile[path] = markers
			for _, m := range markers {
				modulesUsed[m.Module] = true
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("scanning for structurizr markers: %w", err)
	}

	if len(markersByFile) == 0 {
		p.log("    No Structurizr markers found")
		return nil
	}

	// Get DSL hashes for all used modules
	for module := range modulesUsed {
		hash, err := design.GetModuleDSLHash(module)
		if err != nil {
			p.log("    Warning: could not get DSL hash for %s: %v", module, err)
			continue
		}
		dslHashes[module] = hash
	}

	// Count total markers and replace them
	totalMarkers := 0
	for _, markers := range markersByFile {
		totalMarkers += len(markers)
	}
	p.log("    Found %d Structurizr marker(s) in %d file(s)", totalMarkers, len(markersByFile))

	// Replace markers with img tags
	replaced := 0
	missing := 0

	for filePath, markers := range markersByFile {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", filePath, err)
		}

		modified := string(content)

		// Replace in reverse order to preserve positions
		for i := len(markers) - 1; i >= 0; i-- {
			marker := markers[i]

			// Get DSL hash for this module
			dslHash, ok := dslHashes[marker.Module]
			if !ok {
				p.log("    Warning: no DSL hash for module %s (marker: %s:%s)",
					marker.Module, marker.Module, marker.ViewKey)
				missing++
				continue
			}

			// Build expected cache path (source location)
			sourceCachePath := paths.StructurizrDocsCachePath(p.workspaceRoot, marker.Module, marker.ViewKey, dslHash)

			// Check if cached SVG exists in source
			if _, err := os.Stat(sourceCachePath); os.IsNotExist(err) {
				p.log("    Warning: cached SVG not found for %s:%s (expected: %s)",
					marker.Module, marker.ViewKey, filepath.Base(sourceCachePath))
				missing++
				continue
			}

			// Calculate staging cache path
			// Source: docs/assets/cache/structurizr/file.svg
			// Staging: [stagingDir]/assets/cache/structurizr/file.svg
			svgFilename := filepath.Base(sourceCachePath)
			stagingCachePath := filepath.Join(p.stagingDir, "assets", "cache", "structurizr", svgFilename)

			// Verify the SVG was copied to staging
			if _, err := os.Stat(stagingCachePath); os.IsNotExist(err) {
				p.log("    Warning: SVG not found in staging for %s:%s (expected: %s)",
					marker.Module, marker.ViewKey, filepath.Base(stagingCachePath))
				missing++
				continue
			}

			// Calculate relative path from markdown file to staging SVG
			relPath, err := p.linkTranslator.CalculateRelativePath(filePath, stagingCachePath)
			if err != nil {
				return fmt.Errorf("calculating relative path for %s:%s: %w",
					marker.Module, marker.ViewKey, err)
			}

			// For site builds (non-PDF), MkDocs converts file.md to file/index.html,
			// adding an extra directory level. Prepend ../ to account for this.
			if !p.pdfMode {
				relPath = "../" + relPath
			}

			// Build img tag
			altText := fmt.Sprintf("%s %s diagram", marker.Module, marker.ViewKey)
			imgTag := fmt.Sprintf(
				"<img src=\"%s\" alt=\"%s\" style=\"display:block; width:100%%; margin:1em auto;\">",
				relPath, altText,
			)

			// Replace marker with img tag
			modified = modified[:marker.StartPos] + imgTag + modified[marker.EndPos:]
			replaced++
		}

		// Write back modified content
		if err := os.WriteFile(filePath, []byte(modified), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", filePath, err)
		}
	}

	p.log("    Replaced %d marker(s), %d missing", replaced, missing)

	if missing > 0 {
		p.log("    Run 'r2r eac update structurizr' to generate missing diagrams")
	}

	return nil
}

// extractStructurizrMarkers finds all Structurizr markers in content
func extractStructurizrMarkers(content string) []StructurizrMarker {
	var markers []StructurizrMarker

	matches := structurizrMarkerPattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		if len(match) >= 6 {
			markers = append(markers, StructurizrMarker{
				Module:    content[match[2]:match[3]],
				ViewKey:   content[match[4]:match[5]],
				StartPos:  match[0],
				EndPos:    match[1],
				FullMatch: content[match[0]:match[1]],
			})
		}
	}

	return markers
}
