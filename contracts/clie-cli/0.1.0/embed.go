// Package cliecli provides embedded contract files for the CLIE CLI.
// This includes the JSON schema for clie-cli.yml validation and the EBNF grammar
// for command parsing.
package cliecli

import "embed"

// FS is the embedded filesystem containing CLIE CLI contract files.
//
//go:embed schemas/clie-cli.schema.json
//go:embed schemas/command.ebnf
//go:embed schemas/clie-cli.yml.md
var FS embed.FS
