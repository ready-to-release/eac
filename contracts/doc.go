// Package contracts provides embedded contract files (schemas, EBNF grammars, defaults)
// for use by other modules in the EAC ecosystem.
//
// This package serves as the single source of truth for all contract definitions.
// Instead of copying contract files at build time, consuming modules import this
// package and access files via the embedded filesystem.
//
// # Usage
//
// To read a schema file:
//
//	schemaBytes, err := contracts.FS.ReadFile(contracts.EACCorePath("repository.schema.json"))
//	if err != nil {
//	    // handle error
//	}
//
// To read an R2R CLI file:
//
//	ebnfBytes, err := contracts.FS.ReadFile(contracts.R2RCLIPath("command.ebnf"))
//	if err != nil {
//	    // handle error
//	}
//
// # Versioning
//
// Contract versions are exposed as constants (EACCoreVersion, R2RCLIVersion, etc.)
// and used by the helper functions to construct paths. This ensures consuming code
// automatically uses the correct version without hardcoding paths.
package contracts
