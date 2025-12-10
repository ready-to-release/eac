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

	// Step 8: Strip nav titles and macros (PDF only)
	// awesome-nav warns about top-level titles which conflict with PDF navigation
	// Macros like {{ diataxis_footer() }} are only processed by site builds
	if p.pdfMode {
		p.log("  Step 8a: Stripping nav titles...")
		if err := p.stripNavTitles(); err != nil {
			return fmt.Errorf("step 8a (strip nav titles): %w", err)
		}
		p.log("  Step 8b: Stripping macros...")
		if err := p.stripMacros(); err != nil {
			return fmt.Errorf("step 8b (strip macros): %w", err)
		}
	}

	// Step 9: Process mermaid diagram sizing (both PDF and site)
	// Wraps mermaid blocks with size directives in container divs
	p.log("  Step 9: Processing mermaid sizing...")
	if err := p.processMermaidSizing(); err != nil {
		return fmt.Errorf("step 9 (mermaid sizing): %w", err)
	}

	// Step 9b: Process mermaid diagrams with caching (both PDF and site)
	// Scans for diagrams, uses cached SVGs, replaces blocks with img tags
	// Only modifies staging markdown (source stays pure)
	p.log("  Step 9b: Processing mermaid diagrams...")
	blocksByFile, statuses, err := p.scanForMermaidDiagrams()
	if err != nil {
		return fmt.Errorf("step 9b (mermaid scan): %w", err)
	}

	// Replace mermaid blocks with img tags in staging
	if err := p.replaceMermaidBlocksWithImages(blocksByFile, statuses); err != nil {
		return fmt.Errorf("step 9b (mermaid replace): %w", err)
	}

	// Step 10: Convert .drawio to cached images (PDF only)
	// Interactive .drawio diagrams can't display in PDFs, so render to static images
	// For HTML site, .drawio files are rendered by JavaScript viewer
	if p.pdfMode {
		p.log("  Step 10: Converting .drawio to cached images...")
		if err := p.convertDrawioToLinks(); err != nil {
			return fmt.Errorf("step 10 (drawio to images): %w", err)
		}
	}

	// Step 11: Add image width constraints (both PDF and site)
	// Ensures large diagrams fit within page/container boundaries
	p.log("  Step 11: Adding image width constraints...")
	if err := p.cleanupLinksForPDF(); err != nil {
		return fmt.Errorf("step 11 (image constraints): %w", err)
	}

	// Step 12: Optimize drawio images (both PDF and site)
	// Uses cached optimized versions for faster builds and smaller output
	p.log("  Step 12: Optimizing drawio images...")
	if err := p.optimizeDrawioImages(); err != nil {
		return fmt.Errorf("step 12 (drawio optimization): %w", err)
	}

	// Step 13: Clean up unreferenced assets (runs for both PDF and HTML)
	// Removes any files in staging that are not referenced by markdown
	// This catches orphaned images, stale assets, and intermediate files
	p.log("  Step 13: Cleaning up unreferenced assets...")
	if err := p.cleanupUnreferencedAssets(); err != nil {
		return fmt.Errorf("step 13 (cleanup unreferenced): %w", err)
	}

	elapsed := time.Since(startTime)
	p.log("✅ Book preprocessing complete: %s (took %v)", p.book.Name, elapsed)

	// Log cache statistics
	stats := p.assetCache.Stats()
	if stats.MermaidHits+stats.MermaidMisses > 0 {
		hitRate := float64(stats.MermaidHits) / float64(stats.MermaidHits+stats.MermaidMisses) * 100
		p.log("   📊 Mermaid cache: %d hits, %d misses (%.1f%% hit rate)",
			stats.MermaidHits, stats.MermaidMisses, hitRate)
	}
	if stats.DrawioHits+stats.DrawioMisses > 0 {
		hitRate := float64(stats.DrawioHits) / float64(stats.DrawioHits+stats.DrawioMisses) * 100
		p.log("   📊 Drawio cache: %d hits, %d misses (%.1f%% hit rate)",
			stats.DrawioHits, stats.DrawioMisses, hitRate)
	}

	return nil
}

// log writes a formatted message to the log writer
func (p *Preprocessor) log(format string, args ...any) {
	fmt.Fprintf(p.logWriter, format+"\n", args...)
}
