package formats

import (
	"fmt"
	"strings"
)

// GetFormatter returns the appropriate formatter for the given format string.
func GetFormatter(format string) (ExportFormatter, error) {
	switch strings.ToLower(format) {
	case "json":
		return NewJSONFormatter(), nil
	case "csv":
		return NewCSVFormatter(), nil
	case "markdown", "md":
		return NewMarkdownFormatter(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s (supported: json, csv, markdown)", format)
	}
}
