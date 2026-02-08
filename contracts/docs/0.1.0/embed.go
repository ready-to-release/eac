// Package docs provides embedded contract files for documentation.
// This includes the JSON schema for docs manifest validation.
package docs

import "embed"

// FS is the embedded filesystem containing docs contract files.
//
//go:embed schemas/manifest.schema.json
var FS embed.FS
