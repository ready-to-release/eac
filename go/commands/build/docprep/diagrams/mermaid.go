package diagrams

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/commands/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// MermaidBlock is a type alias for DiagramBlock, preserving backward compatibility.
type MermaidBlock = DiagramBlock

// CacheStatus is a type alias for DiagramCacheStatus, preserving backward compatibility.
type CacheStatus = DiagramCacheStatus

// MermaidSizePresets maps size names to CSS width values.
var MermaidSizePresets = map[string]string{
	"small":  "33%",
	"medium": "50%",
	"large":  "66%",
	"full":   "100%",
}

// mermaidBlockPattern matches mermaid code blocks with optional size directive.
// Captures: (1) size directive value, (2) mermaid content.
var mermaidBlockPattern = regexp.MustCompile("(?s)```mermaid\\s*\n%%\\{(?:size|width):([^}]+)\\}%%\\s*\n(.*?)```")

// mermaidBlockPlain matches plain mermaid blocks without size directive.
var mermaidBlockPlain = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)```")

// MermaidConfig is the DiagramConfig for mermaid diagram processing.
var MermaidConfig = DiagramConfig{
	Name:            "mermaid",
	BlockPattern:    mermaidBlockPlain,
	FilePrefix:      "mermaid",
	HashFn:          HashContent,
	PreHashFn:       StripSizeDirective,
	BuildOutputPath: paths.MermaidBuildOutputPath,
	IndexFilename:   "mermaid-index.json",
	BuildImgTag:     mermaidImgTag,
	CacheSubdir:     "mermaid",
}

// mermaidImgTag builds an img tag with mermaid-specific size handling.
func mermaidImgTag(relPath, prefix string, _ DiagramBlock) string {
	imgWidth := "100%"
	if wrapperIdx := strings.LastIndex(prefix, "data-size=\""); wrapperIdx >= 0 {
		rest := prefix[wrapperIdx+11:]
		if endIdx := strings.Index(rest, "\""); endIdx > 0 {
			sizeVal := rest[:endIdx]
			if w, ok := MermaidSizePresets[strings.ToLower(sizeVal)]; ok {
				imgWidth = w
			} else if strings.HasSuffix(sizeVal, "%") || strings.HasSuffix(sizeVal, "px") {
				imgWidth = sizeVal
			}
		}
	}

	return fmt.Sprintf(
		"<img src=\"%s\" alt=\"Mermaid diagram\" style=\"display:block; width:%s; margin:0 auto;\">",
		relPath, imgWidth,
	)
}

// ProcessMermaidSizing wraps mermaid blocks with size directives in container divs.
func ProcessMermaidSizing(fileIndex *staging.FileIndex, logf func(string, ...any)) error {
	logf("    Processing mermaid diagram sizing...")

	wrapped := 0
	mdFiles := fileIndex.GetMarkdownFiles()

	for _, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := WrapMermaidBlocks(original)

		if modified != original {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			wrapped++
		}
	}

	logf("    Processed %d files, wrapped %d mermaid blocks with sizing", len(mdFiles), wrapped)
	return nil
}

// WrapMermaidBlocks finds mermaid blocks with size directives and wraps them.
func WrapMermaidBlocks(content string) string {
	result := mermaidBlockPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := mermaidBlockPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		sizeValue := strings.TrimSpace(submatches[1])
		mermaidContent := submatches[2]

		var wrapper strings.Builder
		wrapper.WriteString("<div class=\"mermaid-wrapper\" data-size=\"")
		wrapper.WriteString(strings.ToLower(sizeValue))
		wrapper.WriteString("\">\n\n")
		wrapper.WriteString("```mermaid\n")
		wrapper.WriteString(mermaidContent)
		wrapper.WriteString("```\n\n</div>")

		return wrapper.String()
	})

	return result
}

// ExtractMermaidBlocks scans a markdown file for mermaid code blocks.
// Returns all blocks with metadata for caching and rendering.
func ExtractMermaidBlocks(content, absSourcePath, stagingDir string) []MermaidBlock {
	return ExtractBlocks(MermaidConfig, content, absSourcePath, stagingDir)
}

// StripSizeDirective removes size directive lines from diagram content.
func StripSizeDirective(content string) string {
	lines := strings.Split(content, "\n")
	filtered := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%%{size:") || strings.HasPrefix(trimmed, "%%{width:") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// HashContent returns first 8 chars of SHA256 hash.
func HashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)[:8]
}

// CheckMermaidCache checks which diagrams have pre-rendered SVGs from the builder output.
func CheckMermaidCache(workspaceRoot string, blocks []MermaidBlock, debugf func(string, ...any)) ([]CacheStatus, error) {
	return CheckDiagramCache(MermaidConfig, workspaceRoot, blocks, debugf)
}

// ReplaceMermaidBlocksWithImages replaces mermaid code blocks with img references.
// extraPathPrefix is "" for PDF, "../" for site builds (from OutputMode.ExtraPathPrefix()).
func ReplaceMermaidBlocksWithImages(
	blocksByFile map[string][]MermaidBlock,
	statuses []CacheStatus,
	extraPathPrefix string,
	logf func(string, ...any),
) error {
	return ReplaceBlocksWithImages(MermaidConfig, blocksByFile, statuses, extraPathPrefix, logf)
}

// ScanForMermaidDiagrams scans all markdown files in staging directory.
// Returns all mermaid blocks found (grouped by file) and their cache statuses.
func ScanForMermaidDiagrams(
	fileIndex *staging.FileIndex,
	stagingDir, workspaceRoot string,
	logf func(string, ...any),
	debugf func(string, ...any),
) (map[string][]MermaidBlock, []CacheStatus, error) {
	return ScanForDiagrams(MermaidConfig, fileIndex, stagingDir, workspaceRoot, logf, debugf)
}
