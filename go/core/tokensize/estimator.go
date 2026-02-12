// Package tokensize provides low-cost heuristics for estimating token counts in source files.
// This helps identify files that may exceed Claude's token limits.
package tokensize

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

// EstimationMethod identifies the algorithm used for token estimation.
type EstimationMethod string

const (
	// MethodCharDiv4 estimates tokens as characters / 4 (default heuristic).
	MethodCharDiv4 EstimationMethod = "char/4"
)

// Estimate holds the token estimation result for a file.
type Estimate struct {
	FilePath   string           `json:"file_path"`
	Tokens     int              `json:"tokens"`
	Characters int              `json:"characters"`
	Bytes      int64            `json:"bytes"`
	Lines      int              `json:"lines"`
	Method     EstimationMethod `json:"method"`
}

// EstimateFile reads a file and estimates its token count.
func EstimateFile(filePath string) (*Estimate, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return EstimateContent(filePath, content), nil
}

// EstimateContent estimates token count from raw bytes.
// This is useful for testing or when file content is already in memory.
//
// Heuristic limitations (char/4):
//   - Assumes an average of 4 characters per token, which is a rough
//     approximation derived from English prose. Actual tokenisation depends
//     on the model's BPE vocabulary and varies by language and content type.
//   - Code-heavy files (especially languages with long identifiers like Java
//     or Go) tend to have fewer tokens per character, so the estimate may
//     overcount.
//   - JSON, YAML, and other structured data formats are more token-dense;
//     the estimate may undercount for these.
//   - Non-ASCII content (CJK characters, emoji) can consume multiple tokens
//     per character, making the heuristic less reliable.
//   - The heuristic does not account for whitespace collapsing or special
//     token boundaries; leading/trailing whitespace is counted equally.
//   - For precise token counts, use the model's actual tokeniser; this
//     heuristic is intended only as a fast pre-filter to flag files that are
//     likely to exceed context limits.
func EstimateContent(filePath string, content []byte) *Estimate {
	chars := len(content)
	lines := bytes.Count(content, []byte{'\n'})
	if len(content) > 0 && content[len(content)-1] != '\n' {
		lines++ // Count last line if not newline-terminated
	}

	return &Estimate{
		FilePath:   filePath,
		Tokens:     chars / 4, // char/4 heuristic -- see doc comment for limitations
		Characters: chars,
		Bytes:      int64(len(content)),
		Lines:      lines,
		Method:     MethodCharDiv4,
	}
}

// ExpandGlobPatterns expands glob patterns against the filesystem.
// Returns absolute file paths. Directories are excluded.
func ExpandGlobPatterns(workspaceRoot string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		absPattern := pattern
		if !filepath.IsAbs(pattern) {
			absPattern = filepath.Join(workspaceRoot, pattern)
		}

		matches, err := doublestar.FilepathGlob(absPattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}

			if !seen[match] {
				seen[match] = true
				result = append(result, match)
			}
		}
	}

	return result, nil
}
