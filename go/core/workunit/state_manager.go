package workunit

// State management has been split into focused files:
//   - uow_state.go:    StateManager struct, core CRUD operations, unit-level change detection
//   - module_state.go: Module-granularity change detection and state persistence
//   - test_state.go:   Test-set-aware change detection with dependency propagation
