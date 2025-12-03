// Package books provides book preprocessing for MkDocs sites.
// It aggregates static content with dynamically-generated content from EAC commands.
package books

import (
	"fmt"
	"io"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// Preprocessor handles book preprocessing before MkDocs build
type Preprocessor struct {
	book          *config.Book
	workspaceRoot string
	stagingDir    string
	logWriter     io.Writer
}

// NewPreprocessor creates a new book preprocessor
func NewPreprocessor(book *config.Book, workspaceRoot, stagingDir string, logWriter io.Writer) *Preprocessor {
	return &Preprocessor{
		book:          book,
		workspaceRoot: workspaceRoot,
		stagingDir:    stagingDir,
		logWriter:     logWriter,
	}
}

// Preprocess runs the 5-step preprocessing pipeline
func (p *Preprocessor) Preprocess() error {
	p.log("📚 Book preprocessing: %s", p.book.Name)

	// Step 1: Copy static files to staging
	p.log("  Step 1: Copying static files...")
	if err := p.copyStaticFiles(); err != nil {
		return fmt.Errorf("step 1 (copy): %w", err)
	}

	// Step 2: Execute commands (capture outputs)
	p.log("  Step 2: Executing commands...")
	commandOutputs, err := p.executeCommands()
	if err != nil {
		return fmt.Errorf("step 2 (commands): %w", err)
	}

	// Step 3: Generate .nav.yml for generated sections
	p.log("  Step 3: Generating navigation...")
	if err := p.generateNavigation(); err != nil {
		return fmt.Errorf("step 3 (nav): %w", err)
	}

	// Step 4: Insert generated sections into parent navs
	p.log("  Step 4: Inserting nav sections...")
	if err := p.insertNavSections(); err != nil {
		return fmt.Errorf("step 4 (insert nav): %w", err)
	}

	// Step 5: Insert inline command outputs at markers
	p.log("  Step 5: Inserting inline content...")
	if err := p.insertInlineContent(commandOutputs); err != nil {
		return fmt.Errorf("step 5 (inline): %w", err)
	}

	p.log("✅ Book preprocessing complete: %s", p.book.Name)
	return nil
}

// log writes a formatted message to the log writer
func (p *Preprocessor) log(format string, args ...any) {
	fmt.Fprintf(p.logWriter, format+"\n", args...)
}
