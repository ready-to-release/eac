// Package render provides stateless rendering functions for the console TUI.
//
// All render functions take explicit data parameters (not Model) to enable
// isolated testing and loose coupling with the console state.
//
// This package is internal to the console implementation and does not affect
// any interface boundaries (Console interface in ports.go).
package render
