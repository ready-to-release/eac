package core

import "embed"

// FS is the embedded filesystem containing all core contract files.
// Schemas are under schemas/ and defaults are under schemas/defaults/.
//
//go:embed schemas/*.schema.json
//go:embed schemas/*.json
//go:embed schemas/defaults/*.yml
var FS embed.FS

// SchemaPath returns the embedded path for a schema file.
// Example: SchemaPath("repository.schema.json") returns "schemas/repository.schema.json".
func SchemaPath(filename string) string {
	return "schemas/" + filename
}

// DefaultPath returns the embedded path for a default file.
// Example: DefaultPath("repository.yml") returns "schemas/defaults/repository.yml".
func DefaultPath(filename string) string {
	return "schemas/defaults/" + filename
}
