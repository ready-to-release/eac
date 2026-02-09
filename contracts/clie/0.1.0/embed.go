// Package clie provides embedded contract files for the CLIE CLI.
// This includes the JSON schema for clie.yml validation and the EBNF grammar
// for command parsing.
package clie

import "embed"

// FS is the embedded filesystem containing CLIE CLI contract files.
//
//go:embed schemas/clie.schema.json
//go:embed schemas/command.ebnf
//go:embed schemas/clie.yml.md
var FS embed.FS
