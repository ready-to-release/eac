package content

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
)

// imagePatterns consolidates all image-related regex patterns.
// Previously these were three separate overlapping patterns; now a single base
// pattern captures all markdown images with an optional attr_list group, and
// specialised helpers filter by drawio extension or attr presence.
var imagePatterns = struct {
	// base matches any markdown image with optional attr_list: ![alt](path){attrs}
	base *regexp.Regexp
	// attrList matches key-value pairs inside an attr_list block
	attrList *regexp.Regexp
}{
	base:     regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)(\s*\{:?\s*([^}]*)\})?`),
	attrList: regexp.MustCompile(`(\w+)\s*=\s*"?([^"\s}]+)"?`),
}

// isDrawioImage returns true if the image path ends with .drawio.
func isDrawioImage(path string) bool {
	return strings.HasSuffix(path, ".drawio")
}

// hasAttrList returns true if the optional attrs capture group is non-empty.
func hasAttrList(match []string) bool {
	return len(match) > 4 && strings.TrimSpace(match[4]) != ""
}

// ConvertAttrListImages converts ![alt](path){attrs} to <img> tags.
func ConvertAttrListImages(content string, adjustPaths bool) (string, int) {
	count := 0

	result := imagePatterns.base.ReplaceAllStringFunc(content, func(match string) string {
		parts := imagePatterns.base.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		alt := parts[1]
		src := parts[2]

		// Skip non-attr-list images and drawio images
		if !hasAttrList(parts) || isDrawioImage(src) {
			return match
		}

		attrs := parts[4]
		htmlAttrs := ParseAttrListToHTML(attrs)
		if htmlAttrs == "" {
			return match
		}

		adjustedSrc := src
		if adjustPaths && !strings.HasPrefix(src, "/") && !strings.HasPrefix(src, "http") {
			adjustedSrc = "../" + src
		}

		count++
		return fmt.Sprintf(`<img src="%s" %s alt="%s">`, adjustedSrc, htmlAttrs, alt)
	})

	return result, count
}

// ParseAttrListToHTML converts attr_list attributes to HTML attributes.
func ParseAttrListToHTML(attrs string) string {
	var htmlParts []string

	matches := imagePatterns.attrList.FindAllStringSubmatch(attrs, -1)

	for _, m := range matches {
		if len(m) >= 3 {
			key := strings.ToLower(m[1])
			value := m[2]

			switch key {
			case "width", "height", "style", "class", "id":
				htmlParts = append(htmlParts, fmt.Sprintf(`%s="%s"`, key, value))
			}
		}
	}

	return strings.Join(htmlParts, " ")
}

// AddImageWidthConstraints adds width="100%" to images that don't have explicit dimensions.
func AddImageWidthConstraints(content string) (string, int) {
	count := 0

	result := imagePatterns.base.ReplaceAllStringFunc(content, func(match string) string {
		if strings.Contains(match, "width=") || strings.Contains(match, "style=") {
			return match
		}

		parts := imagePatterns.base.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		alt := parts[1]
		path := parts[2]
		existingAttrs := ""
		if len(parts) > 3 {
			existingAttrs = parts[3]
		}

		if isDrawioImage(path) {
			return match
		}

		lowerPath := strings.ToLower(path)
		if strings.Contains(lowerPath, "icon") ||
			strings.Contains(lowerPath, "badge") ||
			strings.Contains(lowerPath, "emoji") ||
			strings.Contains(lowerPath, "favicon") {
			return match
		}

		count++
		if existingAttrs != "" {
			newAttrs := strings.TrimSuffix(strings.TrimSpace(existingAttrs), "}")
			return "![" + alt + "](" + path + ")" + newAttrs + ` width="100%" }`
		}

		return "![" + alt + "](" + path + `){ width="100%" }`
	})

	return result, count
}

// ConvertDrawioImages converts ![alt](path.drawio) to [View interactive diagram](url).
func ConvertDrawioImages(content, relFileDir, siteURL string) (string, int) {
	count := 0

	result := imagePatterns.base.ReplaceAllStringFunc(content, func(match string) string {
		parts := imagePatterns.base.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		imgPath := parts[2]

		// Only process drawio images
		if !isDrawioImage(imgPath) {
			return match
		}

		var urlPath string
		if strings.HasPrefix(imgPath, "/") {
			urlPath = imgPath[1:]
		} else {
			urlPath = filepath.ToSlash(filepath.Join(relFileDir, imgPath))
		}
		urlPath = filepath.ToSlash(filepath.Clean(urlPath))
		urlPath = strings.TrimPrefix(urlPath, "../")

		fullURL, err := url.JoinPath(siteURL, urlPath)
		if err != nil {
			return match
		}

		count++
		return fmt.Sprintf("[View interactive diagram](%s)", fullURL)
	})

	return result, count
}

// ProcessAttrListImages converts attr_list images in all markdown files.
func ProcessAttrListImages(fileIndex *staging.FileIndex, logf func(string, ...any)) error {
	logf("    Converting attr_list images to HTML...")

	converted := 0
	mdFiles := fileIndex.GetMarkdownFiles()

	for _, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified, count := ConvertAttrListImages(original, false)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			converted += count
		}
	}

	logf("    Processed %d files, converted %d images to HTML", len(mdFiles), converted)
	return nil
}

// ProcessImageWidthConstraints adds width constraints to all markdown files.
func ProcessImageWidthConstraints(fileIndex *staging.FileIndex, logf func(string, ...any)) error {
	logf("    Processing images for width constraints...")

	imagesFixed := 0
	mdFiles := fileIndex.GetMarkdownFiles()

	for _, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified, count := AddImageWidthConstraints(original)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			imagesFixed += count
		}
	}

	logf("    Processed %d files, added width constraints to %d images", len(mdFiles), imagesFixed)
	return nil
}

// ProcessDrawioToLinks converts .drawio image references to links in all markdown files.
func ProcessDrawioToLinks(fileIndex *staging.FileIndex, stagingDir, siteURL string, logf func(string, ...any)) error {
	if siteURL == "" {
		logf("    Skipping drawio conversion (no site_url configured)")
		return nil
	}

	logf("    Converting .drawio images to links...")

	if !strings.HasSuffix(siteURL, "/") {
		siteURL += "/"
	}

	converted := 0
	mdFiles := fileIndex.GetMarkdownFiles()

	for _, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fileDir := filepath.Dir(path)
		relFileDir, relErr := filepath.Rel(stagingDir, fileDir)
		if relErr != nil {
			relFileDir = fileDir
		}

		original := string(content)
		modified, count := ConvertDrawioImages(original, relFileDir, siteURL)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			converted += count
		}
	}

	logf("    Processed %d files, converted %d .drawio images to links", len(mdFiles), converted)
	return nil
}
