package testview

import "strings"

// StripTagPrefixes removes @ prefix from tags for display.
func StripTagPrefixes(tags []string) []string {
	result := make([]string, len(tags))
	for i, tag := range tags {
		if tag != "" && tag[0] == '@' {
			result[i] = tag[1:]
		} else {
			result[i] = tag
		}
	}
	return result
}

// ExtractControlTags extracts control IDs from tags array.
// Tags like "@control:ai-2" return ["ai-2"].
func ExtractControlTags(tags []string) []string {
	var controls []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, "@control:") {
			control := strings.TrimPrefix(tag, "@control:")
			if control != "" {
				controls = append(controls, control)
			}
		}
	}
	return controls
}

// ExtractFeatureName extracts the feature name from a feature file path.
func ExtractFeatureName(filePath string) string {
	if filePath == "" {
		return ""
	}

	filePath = strings.ReplaceAll(filePath, "\\", "/")

	if !strings.Contains(filePath, "/specs/") && !strings.HasPrefix(filePath, "specs/") {
		return ""
	}

	parts := strings.Split(filePath, "/")

	specsIndex := -1
	for i, part := range parts {
		if part == "specs" {
			specsIndex = i
			break
		}
	}

	if specsIndex >= 0 && specsIndex+2 < len(parts) {
		return parts[specsIndex+2]
	}

	return ""
}
