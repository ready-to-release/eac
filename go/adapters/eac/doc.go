// Package eac provides an adapter for invoking EAC commands through
// the tool abstraction layer. This allows EAC to be treated as a
// first-class tool in the registry, with native and container modes.
//
// The central type is [EACPort], a minimal interface with a single
// Execute method. Production code uses [New] to obtain an adapter
// that routes through the tool executor. Tests use [NewMock] for
// a simple in-memory mock that records calls.
package eac
