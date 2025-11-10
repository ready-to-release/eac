// Package main provides the r2r-cli command-line interface
//
// This file contains go:generate directives that copy contract files
// from the contracts directory to their appropriate locations before building.
//
//go:generate go run tools/copy.go ../../contracts/cli/0.1.0/command.ebnf internal/command-parser/command.ebnf
//go:generate go run tools/copy.go ../../contracts/cli/0.1.0/schema.json internal/validator/config/schema.json
package main
