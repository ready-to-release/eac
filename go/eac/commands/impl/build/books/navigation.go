package books

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"gopkg.in/yaml.v3"
)

// NavFile represents a .nav.yml file structure
type NavFile struct {
	Title string `yaml:"title,omitempty"`
	Nav   []any  `yaml:"nav"`
}

// generateNavigation creates .nav.yml for generated sections (Step 3)
func (p *Preprocessor) generateNavigation() error {
	if len(p.book.GeneratedNav) == 0 {
		p.log("    No generated nav sections defined")
		return nil
	}

	for _, navConfig := range p.book.GeneratedNav {
		sectionDir := filepath.Join(p.stagingDir, navConfig.Section)

		// Check if section directory exists (created by command sources)
		if _, err := os.Stat(sectionDir); os.IsNotExist(err) {
			p.log("    Skipping nav for %s (directory not found)", navConfig.Section)
			continue
		}

		// Get ordered command sources for this section
		sources := p.getSourcesForSection(navConfig.Section)
		if len(sources) == 0 {
			p.log("    Skipping nav for %s (no sources)", navConfig.Section)
			continue
		}

		// Build nav entries from sources
		var navEntries []any
		for _, src := range sources {
			filename := filepath.Base(src.Target)
			navEntries = append(navEntries, filename)
		}

		// Create .nav.yml
		nav := NavFile{
			Title: navConfig.Title,
			Nav:   navEntries,
		}

		navPath := filepath.Join(sectionDir, ".nav.yml")
		data, err := yaml.Marshal(nav)
		if err != nil {
			return err
		}

		if err := os.WriteFile(navPath, data, 0644); err != nil {
			return err
		}

		p.log("    Generated: %s/.nav.yml", navConfig.Section)
	}

	return nil
}

// insertNavSections inserts generated sections into parent navs (Step 4)
func (p *Preprocessor) insertNavSections() error {
	if len(p.book.GeneratedNav) == 0 {
		return nil
	}

	for _, navConfig := range p.book.GeneratedNav {
		parentNavPath := filepath.Join(p.stagingDir, navConfig.InsertInto, ".nav.yml")

		// Read existing nav
		data, err := os.ReadFile(parentNavPath)
		if err != nil {
			if os.IsNotExist(err) {
				p.log("    Parent nav not found: %s/.nav.yml", navConfig.InsertInto)
				continue
			}
			return err
		}

		var nav NavFile
		if err := yaml.Unmarshal(data, &nav); err != nil {
			return err
		}

		// Insert section at position
		sectionName := filepath.Base(navConfig.Section)

		// Check if section already exists
		if containsNavEntry(nav.Nav, sectionName) {
			p.log("    Section '%s' already in %s/.nav.yml", sectionName, navConfig.InsertInto)
			continue
		}

		nav.Nav = insertAtPosition(nav.Nav, sectionName, navConfig.Position)

		// Write updated nav
		data, err = yaml.Marshal(nav)
		if err != nil {
			return err
		}

		if err := os.WriteFile(parentNavPath, data, 0644); err != nil {
			return err
		}

		p.log("    Updated: %s/.nav.yml (inserted '%s')", navConfig.InsertInto, sectionName)
	}

	return nil
}

// getSourcesForSection returns command sources whose target is in the given section
func (p *Preprocessor) getSourcesForSection(section string) []config.Source {
	var sources []config.Source

	// Normalize section path (convert to forward slashes for comparison)
	normalizedSection := filepath.ToSlash(section)

	for _, src := range p.book.GetCommandSources() {
		// Check if target is in this section
		// Normalize target dir to forward slashes for cross-platform comparison
		targetDir := filepath.ToSlash(filepath.Dir(src.Target))
		if targetDir == normalizedSection || strings.HasPrefix(targetDir, normalizedSection+"/") {
			sources = append(sources, src)
		}
	}

	// Sort by order
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Order < sources[j].Order
	})

	return sources
}

// insertAtPosition inserts item at specified position in nav
func insertAtPosition(nav []any, item string, position string) []any {
	switch {
	case position == "first":
		return append([]any{item}, nav...)

	case position == "last" || position == "":
		return append(nav, item)

	case strings.HasPrefix(position, "after:"):
		afterItem := strings.TrimPrefix(position, "after:")
		for i, entry := range nav {
			if str, ok := entry.(string); ok && str == afterItem {
				// Insert after this item
				result := make([]any, 0, len(nav)+1)
				result = append(result, nav[:i+1]...)
				result = append(result, item)
				result = append(result, nav[i+1:]...)
				return result
			}
		}
		// Fallback to last if not found
		return append(nav, item)

	default:
		return append(nav, item)
	}
}

// containsNavEntry checks if a nav entry already exists
func containsNavEntry(nav []any, item string) bool {
	for _, entry := range nav {
		if str, ok := entry.(string); ok && str == item {
			return true
		}
	}
	return false
}
