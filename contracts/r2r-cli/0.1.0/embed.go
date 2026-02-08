// Package r2rcli provides embedded contract files for the R2R CLI.
// This includes the JSON schema for r2r-cli.yml validation and the EBNF grammar
// for command parsing.
package r2rcli

import "embed"

// FS is the embedded filesystem containing R2R CLI contract files.
//
//go:embed schemas/r2r-cli.schema.json
//go:embed schemas/command.ebnf
//go:embed schemas/r2r-cli.yml.md
var FS embed.FS
