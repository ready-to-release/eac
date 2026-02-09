package diagrams

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/linking"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// StructurizrMarker represents a Structurizr diagram marker found in markdown.
type StructurizrMarker struct {
	Module    string // Module name (e.g., "eac")
	ViewKey   string // View key (e.g., "SystemContext")
	StartPos  int    // Start position of the marker in the file
	EndPos    int    // End position of the marker in the file
	FullMatch string // The full marker text
}

// structurizrMarkerPattern matches <!-- structurizr:module:viewKey --> markers.
var structurizrMarkerPattern = regexp.MustCompile(`<!--\s*structurizr:([^:]+):([^>\s]+)\s*-->`)

// structurizrIndexEntry mirrors the builder's index entry.
type structurizrIndexEntry struct {
	Module      string `json:"module"`
	ViewKey     string `json:"view_key"`
	DSLHash     string `json:"dsl_hash"`
	SVGFilename string `json:"svg_filename"`
}

// structurizrIndex mirrors the builder's index manifest.
type structurizrIndex struct {
	Entries []structurizrIndexEntry `json:"entries"`
}

// ExtractStructurizrMarkers finds all Structurizr markers in content.
func ExtractStructurizrMarkers(content string) []StructurizrMarker {
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

// ProcessStructurizrDiagrams scans staging markdown for Structurizr markers
// and replaces them with img tags pointing to builder output SVGs.
func ProcessStructurizrDiagrams(
	fileIndex *staging.FileIndex,
	stagingDir, workspaceRoot string,
	extraPathPrefix string,
	logf func(string, ...any),
	warnf func(string, ...any),
	debugf func(string, ...any),
) error {
	if debugf == nil {
		debugf = func(string, ...any) {}
	}

	markersByFile := make(map[string][]StructurizrMarker)

	for _, path := range fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		markers := ExtractStructurizrMarkers(string(content))
		if len(markers) > 0 {
			markersByFile[path] = markers
		}
	}

	if len(markersByFile) == 0 {
		logf("    No Structurizr markers found")
		return nil
	}

	totalMarkers := 0
	for _, markers := range markersByFile {
		totalMarkers += len(markers)
	}
	logf("    Found %d Structurizr marker(s) in %d file(s)", totalMarkers, len(markersByFile))

	moduleSet := make(map[string]struct{})
	for _, markers := range markersByFile {
		for _, m := range markers {
			moduleSet[m.Module] = struct{}{}
		}
	}

	svgLookup := make(map[string]string)
	moduleBuildDirs := make(map[string]string)
	for moduleName := range moduleSet {
		builderOutputDir := paths.StructurizrModuleBuildOutputPath(workspaceRoot, moduleName)
		indexPath := filepath.Join(builderOutputDir, "structurizr-index.json")

		indexData, err := os.ReadFile(indexPath)
		if err != nil {
			warnf("structurizr builder output not found for module %s (build structurizr for that module first)", moduleName)
			continue
		}

		var idx structurizrIndex
		if err := json.Unmarshal(indexData, &idx); err != nil {
			warnf("corrupt structurizr-index.json for module %s: %v", moduleName, err)
			continue
		}

		moduleBuildDirs[moduleName] = builderOutputDir
		for _, entry := range idx.Entries {
			key := entry.Module + ":" + entry.ViewKey
			svgLookup[key] = entry.SVGFilename
		}
	}

	debugf("structurizr: loaded builder indexes for %d module(s) with %d total entries", len(moduleSet), len(svgLookup))

	cacheDir := filepath.Join(stagingDir, "assets", "cache", "structurizr")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("creating staging directory: %w", err)
	}

	replaced := 0
	missing := 0

	for filePath, markers := range markersByFile {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", filePath, err)
		}

		modified := string(content)

		for i := len(markers) - 1; i >= 0; i-- {
			marker := markers[i]

			lookupKey := marker.Module + ":" + marker.ViewKey
			svgFilename, found := svgLookup[lookupKey]
			if !found {
				warnf("structurizr SVG not found in builder output for %s:%s",
					marker.Module, marker.ViewKey)
				missing++
				continue
			}

			srcPath := filepath.Join(moduleBuildDirs[marker.Module], svgFilename)
			stagingPath := filepath.Join(cacheDir, svgFilename)

			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				warnf("structurizr SVG file missing: %s", srcPath)
				missing++
				continue
			}

			if err := staging.CopyFile(srcPath, stagingPath); err != nil {
				return fmt.Errorf("copying SVG to staging: %w", err)
			}

			relPath, err := linking.CalculateRelativePath(filePath, stagingPath)
			if err != nil {
				return fmt.Errorf("calculating relative path for %s:%s: %w",
					marker.Module, marker.ViewKey, err)
			}

			relPath = extraPathPrefix + relPath

			altText := fmt.Sprintf("%s %s diagram", marker.Module, marker.ViewKey)
			imgTag := fmt.Sprintf(
				"<img src=\"%s\" alt=\"%s\" style=\"display:block; width:100%%; margin:1em auto;\">",
				relPath, altText,
			)

			modified = modified[:marker.StartPos] + imgTag + modified[marker.EndPos:]
			replaced++
		}

		if err := os.WriteFile(filePath, []byte(modified), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", filePath, err)
		}
	}

	logf("    Replaced %d marker(s), %d missing", replaced, missing)

	if missing > 0 {
		return fmt.Errorf("%d structurizr diagram(s) not found in builder output (run structurizr build first)", missing)
	}

	return nil
}
