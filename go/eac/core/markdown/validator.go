package markdown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

// CodeBlock represents an extracted code block from markdown
type CodeBlock struct {
	Language string
	Content  string
	Line     int
}

// Section represents a markdown section with heading and content
type Section struct {
	Heading string
	Level   int
	Content string
	Line    int
}

// ValidationResult holds validation results for a markdown file
type ValidationResult struct {
	FilePath         string
	Valid            bool
	Errors           []ValidationError
	Warnings         []ValidationWarning
	CodeBlocks       []CodeBlock
	Sections         []Section
	LineCount        int
	ByteCount        int
}

// ValidationError represents a validation error
type ValidationError struct {
	Line    int
	Message string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Line    int
	Message string
}

// ValidatorOptions configures markdown validation
type ValidatorOptions struct {
	// ValidateCodeBlocks enables code block syntax validation
	ValidateCodeBlocks bool

	// RequiredSections lists section headings that must be present
	RequiredSections []string

	// CheckHeadingHierarchy ensures proper heading level progression
	CheckHeadingHierarchy bool

	// ExcludeDirs lists directories to skip during validation
	ExcludeDirs []string

	// AllowEmptyFiles allows empty markdown files
	AllowEmptyFiles bool
}

// DefaultValidatorOptions returns sensible defaults
func DefaultValidatorOptions() ValidatorOptions {
	return ValidatorOptions{
		ValidateCodeBlocks:    true,
		RequiredSections:      []string{},
		CheckHeadingHierarchy: true,
		ExcludeDirs:          []string{"node_modules", ".git", "out", ".vscode"},
		AllowEmptyFiles:      false,
	}
}

// Validator validates markdown files
type Validator struct {
	opts   ValidatorOptions
	md     goldmark.Markdown
	writer io.Writer
}

// NewValidator creates a new markdown validator
func NewValidator(opts ValidatorOptions, writer io.Writer) *Validator {
	return &Validator{
		opts: opts,
		md: goldmark.New(
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		),
		writer: writer,
	}
}

// ValidateFile validates a single markdown file
func (v *Validator) ValidateFile(filePath string) ValidationResult {
	result := ValidationResult{
		FilePath: filePath,
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Line:    0,
			Message: fmt.Sprintf("Failed to read file: %v", err),
		})
		return result
	}

	result.ByteCount = len(content)
	result.LineCount = bytes.Count(content, []byte("\n")) + 1

	// Check empty file
	if len(content) == 0 {
		if !v.opts.AllowEmptyFiles {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Line:    0,
				Message: "Empty file",
			})
		}
		return result
	}

	// Parse with goldmark
	doc := v.md.Parser().Parse(text.NewReader(content))

	// Extract code blocks and sections
	result.CodeBlocks = v.extractCodeBlocks(content, doc)
	result.Sections = v.extractSections(content, doc)

	// Validate code blocks
	if v.opts.ValidateCodeBlocks {
		for _, block := range result.CodeBlocks {
			if err := v.validateCodeBlock(block); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Line:    block.Line,
					Message: fmt.Sprintf("Invalid %s code: %v", block.Language, err),
				})
			}
		}
	}

	// Check required sections
	for _, required := range v.opts.RequiredSections {
		found := false
		for _, section := range result.Sections {
			if strings.EqualFold(section.Heading, required) {
				found = true
				break
			}
		}
		if !found {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Line:    0,
				Message: fmt.Sprintf("Missing required section: %s", required),
			})
		}
	}

	// Check heading hierarchy
	if v.opts.CheckHeadingHierarchy {
		v.checkHeadingHierarchy(&result)
	}

	return result
}

// ValidateDirectory validates all markdown files in a directory
func (v *Validator) ValidateDirectory(rootDir string) ([]ValidationResult, error) {
	var results []ValidationResult
	var markdownFiles []string

	// Find all markdown files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() {
			for _, excluded := range v.opts.ExcludeDirs {
				if info.Name() == excluded {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check for markdown files
		ext := filepath.Ext(path)
		if ext == ".md" || ext == ".markdown" {
			markdownFiles = append(markdownFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Validate each file
	for _, mdFile := range markdownFiles {
		result := v.ValidateFile(mdFile)
		results = append(results, result)
	}

	return results, nil
}

// extractCodeBlocks extracts all code blocks from the AST
func (v *Validator) extractCodeBlocks(source []byte, doc ast.Node) []CodeBlock {
	var blocks []CodeBlock

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if codeBlock, ok := n.(*ast.FencedCodeBlock); ok {
			lang := string(codeBlock.Language(source))
			content := extractNodeContent(source, codeBlock)
			line := 0
			if codeBlock.Lines().Len() > 0 {
				line = codeBlock.Lines().At(0).Start
			}

			blocks = append(blocks, CodeBlock{
				Language: lang,
				Content:  content,
				Line:     line,
			})
		}

		return ast.WalkContinue, nil
	})

	return blocks
}

// extractSections extracts all sections with their headings
func (v *Validator) extractSections(source []byte, doc ast.Node) []Section {
	var sections []Section

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			headingText := extractHeadingText(source, heading)
			line := 0
			if heading.Lines().Len() > 0 {
				line = heading.Lines().At(0).Start
			}

			sections = append(sections, Section{
				Heading: headingText,
				Level:   heading.Level,
				Line:    line,
			})
		}

		return ast.WalkContinue, nil
	})

	return sections
}

// validateCodeBlock validates a code block based on its language
func (v *Validator) validateCodeBlock(block CodeBlock) error {
	switch strings.ToLower(block.Language) {
	case "json":
		var data interface{}
		return json.Unmarshal([]byte(block.Content), &data)

	case "yaml", "yml":
		var data interface{}
		return yaml.Unmarshal([]byte(block.Content), &data)

	// Add more language validators as needed
	// case "go":
	//     // Could use go/parser to validate Go syntax

	default:
		// Unknown language - skip validation
		return nil
	}
}

// checkHeadingHierarchy validates proper heading level progression
func (v *Validator) checkHeadingHierarchy(result *ValidationResult) {
	if len(result.Sections) == 0 {
		return
	}

	prevLevel := 0
	for _, section := range result.Sections {
		if prevLevel > 0 && section.Level > prevLevel+1 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Line:    section.Line,
				Message: fmt.Sprintf("Heading level jump: H%d to H%d (expected H%d)", prevLevel, section.Level, prevLevel+1),
			})
		}
		prevLevel = section.Level
	}
}

// PrintResults prints validation results to the configured writer
func (v *Validator) PrintResults(results []ValidationResult, moduleRoot string) int {
	totalErrors := 0
	totalWarnings := 0
	validFiles := 0

	fmt.Fprintf(v.writer, "\n📝 Validated %d markdown file(s)\n", len(results))

	for _, result := range results {
		relPath, _ := filepath.Rel(moduleRoot, result.FilePath)
		fmt.Fprintf(v.writer, "\n   %s\n", relPath)

		// Print errors
		for _, err := range result.Errors {
			fmt.Fprintf(v.writer, "      ❌ Line %d: %s\n", err.Line, err.Message)
			totalErrors++
		}

		// Print warnings
		for _, warn := range result.Warnings {
			fmt.Fprintf(v.writer, "      ⚠️  Line %d: %s\n", warn.Line, warn.Message)
			totalWarnings++
		}

		// Print summary if valid
		if result.Valid && len(result.Warnings) == 0 {
			codeBlockInfo := ""
			if len(result.CodeBlocks) > 0 {
				codeBlockInfo = fmt.Sprintf(", %d code block(s)", len(result.CodeBlocks))
			}
			fmt.Fprintf(v.writer, "      ✅ Valid (%d lines%s)\n", result.LineCount, codeBlockInfo)
			validFiles++
		}
	}

	// Summary
	fmt.Fprintf(v.writer, "\n")
	if totalErrors > 0 {
		fmt.Fprintf(v.writer, "❌ Validation failed: %d error(s), %d warning(s)\n", totalErrors, totalWarnings)
		return 1
	} else if totalWarnings > 0 {
		fmt.Fprintf(v.writer, "⚠️  Validation passed with %d warning(s)\n", totalWarnings)
		fmt.Fprintf(v.writer, "✅ %d/%d files valid\n", validFiles, len(results))
		return 0
	} else {
		fmt.Fprintf(v.writer, "✅ All markdown files validated successfully\n")
		return 0
	}
}

// Helper functions

func extractNodeContent(source []byte, n ast.Node) string {
	var buf bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		buf.Write(line.Value(source))
	}
	return strings.TrimSpace(buf.String())
}

func extractHeadingText(source []byte, heading *ast.Heading) string {
	var buf bytes.Buffer
	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			buf.Write(textNode.Segment.Value(source))
		}
	}
	return strings.TrimSpace(buf.String())
}
