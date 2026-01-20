package formats

import (
	"encoding/json"
	"io"
)

// JSONFormatter exports scenarios as JSON.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Write(export *ManualTestExport, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(export)
}

func (f *JSONFormatter) FileExtension() string {
	return "json"
}
