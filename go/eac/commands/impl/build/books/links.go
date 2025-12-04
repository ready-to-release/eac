package books

// cleanupLinksForPDF processes all markdown files in staging for PDF compatibility (Step 6)
//
// Currently disabled - mkdocs-with-pdf handles anchor generation internally.
// This is a placeholder for future link normalization if needed.
//
// The mkdocs-with-pdf plugin generates anchor IDs in the format:
//   - Page anchor: path/to/page/:
//   - Heading anchor: path/to/page/:heading-slug
func (p *Preprocessor) cleanupLinksForPDF() error {
	// No-op for now - links are handled by mkdocs-with-pdf
	p.log("    Link normalization: skipped (handled by mkdocs-with-pdf)")
	return nil
}
