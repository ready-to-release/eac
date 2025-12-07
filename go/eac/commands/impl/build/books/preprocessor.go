// Package books provides book preprocessing for MkDocs sites.
// It aggregates static content with dynamically-generated content from EAC commands.
package books

import (
	"fmt"
	"io"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// Preprocessor handles book preprocessing before MkDocs build
type Preprocessor struct {
	book           *config.Book
	workspaceRoot  string
	stagingDir     string
	logWriter      io.Writer
	pdfMode        bool
	linkTranslator *LinkTranslator // Handles source → staging path translations
	assetCache     *AssetCache     // Persistent cache for expensive operations (mermaid, etc.)
}

// NewPreprocessor creates a new book preprocessor
// pdfMode enables PDF-specific processing like link normalization
func NewPreprocessor(book *config.Book, workspaceRoot, stagingDir string, logWriter io.Writer, pdfMode bool) *Preprocessor {
	return &Preprocessor{
		book:           book,
		workspaceRoot:  workspaceRoot,
		stagingDir:     stagingDir,
		logWriter:      logWriter,
		pdfMode:        pdfMode,
		linkTranslator: NewLinkTranslator(workspaceRoot, stagingDir, logWriter, pdfMode),
		assetCache:     NewAssetCache(workspaceRoot),
	}
}

// Preprocess runs the preprocessing pipeline
func (p *Preprocessor) Preprocess() error {
	p.log("📚 Book preprocessing: %s", p.book.Name)

	startTime := time.Now()

	// Step 1: Copy static files to staging
	p.log("  Step 1: Copying static files...")
	if err := p.copyStaticFiles(); err != nil {
		return fmt.Errorf("step 1 (copy): %w", err)
	}

	// Step 1b: Convert attr_list images to HTML (before link translation)
	// Converts: ![alt](img.png){width=100} -> <img src="img.png" width="100" alt="alt">
	// Path adjustments are handled by the link translator in Step 2
	p.log("  Step 1b: Converting attr_list images to HTML...")
	if err := p.convertAttrListImagesToHTML(); err != nil {
		return fmt.Errorf("step 1b (attr_list images): %w", err)
	}

	// Step 2: Build and apply link translations
	// Handles ALL link processing: path depth adjustment, external URLs, etc.
	// Analyzes source markdown to extract all relative links, calculates new paths
	// for staging directory structure, and applies translations to fix all paths
	// NOTE: This processes both markdown links and HTML img src attributes
	p.log("  Step 2: Building link translations...")
	if err := p.linkTranslator.BuildTranslations(p.book.SiteURL); err != nil {
		return fmt.Errorf("step 2 (build translations): %w", err)
	}

	p.log("  Step 2: Applying link translations...")
	if err := p.linkTranslator.ApplyAllTranslations(); err != nil {
		return fmt.Errorf("step 2 (apply translations): %w", err)
	}

	// Step 3: Execute commands (capture outputs)
	p.log("  Step 3: Executing commands...")
	commandOutputs, err := p.executeCommands()
	if err != nil {
		return fmt.Errorf("step 3 (commands): %w", err)
	}

	// Step 4: Ensure root index.md exists
	// If no index.md, generate one with book metadata and TOC
	// If index.md exists from copy, create toc.md for separate TOC
	p.log("  Step 4: Ensuring root index...")
	if err := p.ensureRootIndex(); err != nil {
		return fmt.Errorf("step 4 (root index): %w", err)
	}

	// Step 5: Ensure .nav.yml exists in all directories
	// Scans staging and creates navigation for any directory missing .nav.yml
	p.log("  Step 5: Ensuring navigation structure...")
	if err := p.ensureNavigationStructure(); err != nil {
		return fmt.Errorf("step 5 (navigation): %w", err)
	}

	// Step 6: Insert inline command outputs at markers
	p.log("  Step 6: Inserting inline content...")
	if err := p.insertInlineContent(commandOutputs); err != nil {
		return fmt.Errorf("step 6 (inline): %w", err)
	}

	// PDF-specific processing steps (only in PDF mode)
	if p.pdfMode {
		// Step 8: Strip nav titles (awesome-nav warns about top-level titles)
		p.log("  Step 8: Stripping nav titles...")
		if err := p.stripNavTitles(); err != nil {
			return fmt.Errorf("step 8 (strip nav titles): %w", err)
		}

		// Step 9: Process mermaid diagram sizing
		// Wraps mermaid blocks with size directives in container divs
		p.log("  Step 9: Processing mermaid sizing...")
		if err := p.processMermaidSizing(); err != nil {
			return fmt.Errorf("step 9 (mermaid sizing): %w", err)
		}

		// Step 9b: Process mermaid diagrams with caching
		// Scans for diagrams, renders cache misses, replaces blocks with img tags
		// Only modifies staging markdown (source stays pure)
		p.log("  Step 9b: Processing mermaid diagrams...")
		blocksByFile, err := p.scanForMermaidDiagrams()
		if err != nil {
			return fmt.Errorf("step 9b (mermaid scan): %w", err)
		}

		// Replace mermaid blocks with img tags in staging
		if err := p.replaceMermaidBlocksWithImages(blocksByFile); err != nil {
			return fmt.Errorf("step 9b (mermaid replace): %w", err)
		}

		// Step 10: Convert .drawio images to links
		// Interactive diagrams can't display in PDFs, so convert to GitHub Pages links
		p.log("  Step 10: Converting .drawio images to links...")
		if err := p.convertDrawioToLinks(); err != nil {
			return fmt.Errorf("step 10 (drawio to links): %w", err)
		}

		// Step 11: Add image width constraints for PDF
		// Ensures large diagrams fit within PDF page boundaries
		p.log("  Step 11: Adding image width constraints...")
		if err := p.cleanupLinksForPDF(); err != nil {
			return fmt.Errorf("step 11 (image constraints): %w", err)
		}

		// Step 12: Optimize drawio images for PDF
		// Resizes large drawio.png files to reduce PDF size and improve compatibility
		p.log("  Step 12: Optimizing drawio images...")
		if err := p.optimizeDrawioImages(); err != nil {
			return fmt.Errorf("step 12 (drawio optimization): %w", err)
		}
	}

	elapsed := time.Since(startTime)
	p.log("✅ Book preprocessing complete: %s (took %v)", p.book.Name, elapsed)

	// Log cache statistics
	stats := p.assetCache.Stats()
	if stats.MermaidHits+stats.MermaidMisses > 0 {
		hitRate := float64(stats.MermaidHits) / float64(stats.MermaidHits+stats.MermaidMisses) * 100
		p.log("   📊 Persistent cache: %d mermaid hits, %d misses (%.1f%% hit rate)",
			stats.MermaidHits, stats.MermaidMisses, hitRate)
	}

	return nil
}

// log writes a formatted message to the log writer
func (p *Preprocessor) log(format string, args ...any) {
	fmt.Fprintf(p.logWriter, format+"\n", args...)
}
