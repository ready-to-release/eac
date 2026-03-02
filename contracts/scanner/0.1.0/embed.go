package scanner

import "embed"

// FS is the embedded filesystem containing scanner contract default files.
// Defaults are under schemas/defaults/.
//
//go:embed schemas/defaults/*.yml
var FS embed.FS

// DefaultPath returns the embedded path for a scanner default file.
// Example: DefaultPath("scanners.yml") returns "schemas/defaults/scanners.yml".
func DefaultPath(filename string) string {
	return "schemas/defaults/" + filename
}
