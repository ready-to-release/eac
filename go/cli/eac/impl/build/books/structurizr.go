// structurizr.go provides Structurizr diagram processing for book builds.
// Reads pre-rendered SVGs from the structurizr builder output (structurizr-index.json)
// and replaces <!-- structurizr:module:viewKey --> markers with img tags.
package books

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ready-to-release/eac/go/core/paths"
)

// StructurizrMarker represents a Structurizr diagram marker found in markdown.
type StructurizrMarker struct {
	Module    string // Module name (e.g., "eac-cli")
	ViewKey   string // View key (e.g., "SystemContext")
	StartPos  int    // Start position of the marker in the file
	EndPos    int    // End position of the marker in the file
	FullMatch string // The full marker text
}

// structurizrMarkerPattern matches <!-- structurizr:module:viewKey --> markers
// Examples:
//   - <!-- structurizr:eac-cli:SystemContext -->
//   - <!-- structurizr:r2r-cli:Containers -->
var structurizrMarkerPattern = regexp.MustCompile(`<!--\s*structurizr:([^:]+):([^>\s]+)\s*-->`)

// structurizrIndexEntry mirrors the builder's index entry for JSON unmarshaling.
type structurizrIndexEntry struct {
	Module      string `json:"module"`
	ViewKey     string `json:"view_key"`
	DSLHash     string `json:"dsl_hash"`
	SVGFilename string `json:"svg_filename"`
}

// structurizrIndex mirrors the builder's index manifest for JSON unmarshaling.
type structurizrIndex struct {
	Entries []structurizrIndexEntry `json:"entries"`
}

// processStructurizrDiagrams scans staging markdown for Structurizr markers
// and replaces them with img tags pointing to builder output SVGs.
func (p *Preprocessor) processStructurizrDiagrams() error {
	// Scan for markers in staging directory using file index
	markersByFile := make(map[string][]StructurizrMarker)

	for _, path := range p.fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		markers := extractStructurizrMarkers(string(content))
		if len(markers) > 0 {
			markersByFile[path] = markers
		}
	}

	if len(markersByFile) == 0 {
		p.log("    No Structurizr markers found")
		return nil
	}

	// Count total markers
	totalMarkers := 0
	for _, markers := range markersByFile {
		totalMarkers += len(markers)
	}
	p.log("    Found %d Structurizr marker(s) in %d file(s)", totalMarkers, len(markersByFile))

	// Collect unique module names from markers
	moduleSet := make(map[string]struct{})
	for _, markers := range markersByFile {
		for _, m := range markers {
			moduleSet[m.Module] = struct{}{}
		}
	}

	// Load per-module structurizr-index.json files and build combined lookup
	svgLookup := make(map[string]string)
	// Also track which builder output directory each module's SVGs live in
	moduleBuildDirs := make(map[string]string)
	for moduleName := range moduleSet {
		builderOutputDir := paths.StructurizrModuleBuildOutputPath(p.workspaceRoot, moduleName)
		indexPath := filepath.Join(builderOutputDir, "structurizr-index.json")

		indexData, err := os.ReadFile(indexPath)
		if err != nil {
			// Index not found — markers referencing this module will be reported as missing below
			p.warn("structurizr builder output not found for module %s (build structurizr for that module first)", moduleName)
			continue
		}

		var idx structurizrIndex
		if err := json.Unmarshal(indexData, &idx); err != nil {
			p.warn("corrupt structurizr-index.json for module %s: %v", moduleName, err)
			continue
		}

		moduleBuildDirs[moduleName] = builderOutputDir
		for _, entry := range idx.Entries {
			key := entry.Module + ":" + entry.ViewKey
			svgLookup[key] = entry.SVGFilename
		}
	}

	log.Debugf("structurizr: loaded builder indexes for %d module(s) with %d total entries", len(moduleSet), len(svgLookup))

	// Copy builder SVGs to staging rendered directory
	stagingDir := paths.RenderedAssetsPath(p.stagingDir, "structurizr")
	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		return fmt.Errorf("creating staging directory: %w", err)
	}

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

			// Look up SVG from builder index
			lookupKey := marker.Module + ":" + marker.ViewKey
			svgFilename, found := svgLookup[lookupKey]
			if !found {
				p.warn("structurizr SVG not found in builder output for %s:%s",
					marker.Module, marker.ViewKey)
				missing++
				continue
			}

			// Copy SVG from per-module builder output to staging
			srcPath := filepath.Join(moduleBuildDirs[marker.Module], svgFilename)
			stagingPath := filepath.Join(stagingDir, svgFilename)

			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				p.warn("structurizr SVG file missing: %s", srcPath)
				missing++
				continue
			}

			if err := copyFile(srcPath, stagingPath); err != nil {
				return fmt.Errorf("copying SVG to staging: %w", err)
			}

			// Calculate relative path from markdown file to staging SVG
			relPath, err := p.linkTranslator.CalculateRelativePath(filePath, stagingPath)
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
		if err := os.WriteFile(filePath, []byte(modified), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", filePath, err)
		}
	}

	p.log("    Replaced %d marker(s), %d missing", replaced, missing)

	if missing > 0 {
		return fmt.Errorf("%d structurizr diagram(s) not found in builder output (run structurizr build first)", missing)
	}

	return nil
}

// extractStructurizrMarkers finds all Structurizr markers in content.
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
