package releasenotes

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
)

// ReleaseNotes represents the structure of a RELEASE-NOTES.md file.
type ReleaseNotes struct {
	Versions []ReleaseNotesVersion
}

// ReleaseNotesVersion represents a single version entry in RELEASE-NOTES.md.
type ReleaseNotesVersion struct {
	Number   string
	Date     time.Time
	Sections []ReleaseNotesSection
}

// ReleaseNotesSection represents a section within a version entry.
type ReleaseNotesSection struct {
	Header  string
	Content string
}

// Parse reads and parses a RELEASE-NOTES.md file.
func Parse(path string) (*ReleaseNotes, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseContent(string(content))
}

// ParseContent parses RELEASE-NOTES.md content.
func ParseContent(content string) (*ReleaseNotes, error) {
	rn := &ReleaseNotes{
		Versions: []ReleaseNotesVersion{},
	}

	// Regex to match version headers: ## [0.0.7] - 2025-12-11
	versionHeaderRegex := regexp.MustCompile(`(?m)^##\s+\[([^\]]+)\]\s+-\s+(\d{4}-\d{2}-\d{2})`)
	// Regex to match section headers: ### Section Name
	sectionHeaderRegex := regexp.MustCompile(`(?m)^###\s+(.+)`)

	lines := strings.Split(content, "\n")

	var currentVersion *ReleaseNotesVersion
	var currentSection *ReleaseNotesSection

	for _, line := range lines {
		// Check for version header
		if match := versionHeaderRegex.FindStringSubmatch(line); match != nil {
			// Save previous version if exists
			if currentVersion != nil {
				// Save last section if exists
				if currentSection != nil {
					currentSection.Content = strings.TrimSpace(currentSection.Content)
					currentVersion.Sections = append(currentVersion.Sections, *currentSection)
					currentSection = nil
				}
				rn.Versions = append(rn.Versions, *currentVersion)
			}

			// Parse date
			date, err := time.Parse("2006-01-02", match[2])
			if err != nil {
				date = time.Time{} // Use zero time if parse fails
			}

			// Start new version
			currentVersion = &ReleaseNotesVersion{
				Number:   match[1],
				Date:     date,
				Sections: []ReleaseNotesSection{},
			}
			continue
		}

		// Check for section header (only within a version)
		if currentVersion != nil {
			if match := sectionHeaderRegex.FindStringSubmatch(line); match != nil {
				// Save previous section if exists
				if currentSection != nil {
					currentSection.Content = strings.TrimSpace(currentSection.Content)
					currentVersion.Sections = append(currentVersion.Sections, *currentSection)
				}

				// Start new section
				currentSection = &ReleaseNotesSection{
					Header:  strings.TrimSpace(match[1]),
					Content: "",
				}
				continue
			}

			// Accumulate content for current section
			if currentSection != nil {
				// Skip the main header (# Release release_notes)
				if !strings.HasPrefix(strings.TrimSpace(line), "#") {
					currentSection.Content += line + "\n"
				}
			}
		}
	}

	// Save last version and section
	if currentVersion != nil {
		if currentSection != nil {
			currentSection.Content = strings.TrimSpace(currentSection.Content)
			currentVersion.Sections = append(currentVersion.Sections, *currentSection)
		}
		rn.Versions = append(rn.Versions, *currentVersion)
	}

	return rn, nil
}

// ValidateVersion checks if a version exists in the release notes and has content.
func (r *ReleaseNotes) ValidateVersion(version string) error {
	for _, v := range r.Versions {
		if v.Number == version {
			// Check that version has at least one section with content
			if len(v.Sections) == 0 {
				return fmt.Errorf("version [%s] exists but has no sections", version)
			}

			hasContent := false
			for _, section := range v.Sections {
				if strings.TrimSpace(section.Content) != "" {
					hasContent = true
					break
				}
			}

			if !hasContent {
				return fmt.Errorf("version [%s] exists but all sections are empty", version)
			}

			return nil
		}
	}

	return fmt.Errorf("version [%s] not found", version)
}

// GetVersion returns the ReleaseNotesVersion for a specific version number.
func (r *ReleaseNotes) GetVersion(version string) (*ReleaseNotesVersion, error) {
	for _, v := range r.Versions {
		if v.Number == version {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("version [%s] not found", version)
}

// GenerateTemplate creates a RELEASE-NOTES.md file from a template.
// Uses the repository configuration system to locate the template file.
func GenerateTemplate(workspaceRoot string, cfg *config.RepositoryConfig, path, module, version string, date time.Time) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // G301: Release notes directory should be world-readable
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get template path from config
	// For isolated tests, use R2R_CONTAINER_ROOT to find templates (they're in the real repo, not the temp test dir)
	templateRoot := workspaceRoot
	if containerRoot := os.Getenv(environments.EnvR2RContainerRoot); containerRoot != "" {
		templateRoot = containerRoot
	}
	templatePath := cfg.TemplatePathAbs(templateRoot, "reports", "release", "release-notes-template.md")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Parse template
	tmpl, err := template.New("release-notes").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template with data
	data := struct {
		Version string
		Date    string
	}{
		Version: version,
		Date:    date.Format("2006-01-02"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write generated content
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
