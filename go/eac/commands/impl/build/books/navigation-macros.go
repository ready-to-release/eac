package books

import (
	"os"
	"path/filepath" // Used for filepath.Rel
	"regexp"
	"strings"
)

// diátaxisSections are the four documentation sections that should have navigation macros.
var diataxisSections = map[string]bool{
	"tutorials":     true,
	"how-to-guides": true,
	"explanation":   true,
	"reference":     true,
}

// stripMacros removes Jinja2 macro calls from markdown files (PDF only)
// The macros plugin is only enabled for site builds, not PDF builds
// Common macros: {{ diataxis_footer() }}, {{ page_breadcrumb() }}.
// macroPattern matches {{ macro_name() }} or {{ macro_name(args) }}.
var macroPattern = regexp.MustCompile(`\{\{\s*\w+\([^)]*\)\s*\}\}`)

func (p *Preprocessor) stripMacros() error {
	p.log("    Stripping macros from markdown files...")

	stripped := 0
	filesModified := 0

	// Use file index for efficient iteration
	for _, path := range p.fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := macroPattern.ReplaceAllString(original, "")

		if modified != original {
			// Count how many macros were stripped
			matches := macroPattern.FindAllString(original, -1)
			stripped += len(matches)
			filesModified++

			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
		}
	}

	p.log("    Stripped %d macros from %d files", stripped, filesModified)
	return nil
}

// navTitlePattern matches title line at start of .nav.yml files.
var navTitlePattern = regexp.MustCompile(`(?m)^title:\s*.*\n?`)

// stripNavTitles removes the 'title' field from .nav.yml files
// The awesome-nav plugin warns that title has no effect at top level.
func (p *Preprocessor) stripNavTitles() error {
	p.log("    Stripping titles from .nav.yml files...")

	stripped := 0

	// Use file index for efficient iteration
	for _, path := range p.fileIndex.GetAllFiles() {
		if !strings.HasSuffix(path, ".nav.yml") {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		modified := navTitlePattern.ReplaceAllString(string(content), "")

		if modified != string(content) {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			stripped++
		}
	}

	p.log("    Stripped titles from %d .nav.yml files", stripped)
	return nil
}

// Patterns for macro injection.
var (
	h1TitlePattern    = regexp.MustCompile(`(?m)^#\s+.+$`)            // First H1 title line
	breadcrumbPattern = regexp.MustCompile(`\{\{\s*page_breadcrumb\(\)\s*\}\}`)
	footerPattern     = regexp.MustCompile(`\{\{\s*diataxis_footer\(\)\s*\}\}`)
)

// injectMacros adds Jinja2 navigation macros to markdown files (site builds only)
// Injects {{ page_breadcrumb() }} after the title and {{ diataxis_footer() }} at the end
// Only processes files in Diátaxis sections (tutorials, how-to-guides, explanation, reference).
func (p *Preprocessor) injectMacros() error {
	p.log("    Injecting navigation macros into markdown files...")

	injected := 0
	filesModified := 0

	// Use file index for efficient iteration
	for _, path := range p.fileIndex.GetMarkdownFiles() {
		// Get relative path from staging root to determine section
		relPath, err := filepath.Rel(p.stagingDir, path)
		if err != nil {
			return err
		}

		// Normalize path separators and get first directory component
		relPath = filepath.ToSlash(relPath)
		parts := strings.Split(relPath, "/")
		if len(parts) == 0 {
			continue
		}

		// Check if file is in a Diátaxis section
		section := parts[0]
		if !diataxisSections[section] {
			continue // Skip files not in a Diátaxis section
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := original
		macrosAdded := 0

		// Check if breadcrumb already exists
		if !breadcrumbPattern.MatchString(modified) {
			// Find first H1 title and insert breadcrumb after it
			loc := h1TitlePattern.FindStringIndex(modified)
			if loc != nil {
				titleEnd := loc[1]
				// Insert breadcrumb macro after the title line
				modified = modified[:titleEnd] + "\n\n{{ page_breadcrumb() }}" + modified[titleEnd:]
				macrosAdded++
			}
		}

		// Check if footer already exists
		if !footerPattern.MatchString(modified) {
			// Append footer at end of file
			modified = strings.TrimRight(modified, "\n\r\t ") + "\n\n{{ diataxis_footer() }}\n"
			macrosAdded++
		}

		if modified != original {
			injected += macrosAdded
			filesModified++

			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
		}
	}

	p.log("    Injected %d macros into %d files", injected, filesModified)
	return nil
}
