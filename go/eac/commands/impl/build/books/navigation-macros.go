package books

import (
	"os"
	"path/filepath"
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
func (p *Preprocessor) stripMacros() error {
	p.log("    Stripping macros from markdown files...")

	// Pattern matches {{ macro_name() }} or {{ macro_name(args) }}
	macroPattern := regexp.MustCompile(`\{\{\s*\w+\([^)]*\)\s*\}\}`)

	stripped := 0
	filesModified := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

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

		return nil
	})
	if err != nil {
		return err
	}

	p.log("    Stripped %d macros from %d files", stripped, filesModified)
	return nil
}

// stripNavTitles removes the 'title' field from .nav.yml files
// The awesome-nav plugin warns that title has no effect at top level.
func (p *Preprocessor) stripNavTitles() error {
	p.log("    Stripping titles from .nav.yml files...")

	stripped := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".nav.yml") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Simple regex to remove title line at start of file
		titlePattern := regexp.MustCompile(`(?m)^title:\s*.*\n?`)
		modified := titlePattern.ReplaceAllString(string(content), "")

		if modified != string(content) {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			stripped++
		}

		return nil
	})
	if err != nil {
		return err
	}

	p.log("    Stripped titles from %d .nav.yml files", stripped)
	return nil
}

// injectMacros adds Jinja2 navigation macros to markdown files (site builds only)
// Injects {{ page_breadcrumb() }} after the title and {{ diataxis_footer() }} at the end
// Only processes files in Diátaxis sections (tutorials, how-to-guides, explanation, reference).
func (p *Preprocessor) injectMacros() error {
	p.log("    Injecting navigation macros into markdown files...")

	// Pattern to find the first H1 title line
	titlePattern := regexp.MustCompile(`(?m)^#\s+.+$`)

	// Pattern to check if macros already exist (for idempotency)
	breadcrumbPattern := regexp.MustCompile(`\{\{\s*page_breadcrumb\(\)\s*\}\}`)
	footerPattern := regexp.MustCompile(`\{\{\s*diataxis_footer\(\)\s*\}\}`)

	injected := 0
	filesModified := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Get relative path from staging root to determine section
		relPath, err := filepath.Rel(p.stagingDir, path)
		if err != nil {
			return err
		}

		// Normalize path separators and get first directory component
		relPath = filepath.ToSlash(relPath)
		parts := strings.Split(relPath, "/")
		if len(parts) == 0 {
			return nil
		}

		// Check if file is in a Diátaxis section
		section := parts[0]
		if !diataxisSections[section] {
			return nil // Skip files not in a Diátaxis section
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
			loc := titlePattern.FindStringIndex(modified)
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

		return nil
	})
	if err != nil {
		return err
	}

	p.log("    Injected %d macros into %d files", injected, filesModified)
	return nil
}
