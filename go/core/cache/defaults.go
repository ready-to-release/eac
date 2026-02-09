package cache

// DefaultSkipSpecs returns the specs that --skip-cache (bare flag) expands to.
// Returns a fresh copy each call to prevent accidental mutation.
//
// Skipped:
//   - local:state - incremental state (triggers rebuild)
//   - local:work  - ephemeral work dirs (npm, preprocessing)
//
// Preserved:
//   - local:asset    - rendered assets (expensive to regenerate)
//   - local:registry - Docker images (slow to pull)
//   - local:layer    - BuildKit cache (very slow to rebuild)
//   - remote:*       - never touched by default
func DefaultSkipSpecs() []Spec {
	return []Spec{
		{Level: LevelLocal, Type: TypeState},
		{Level: LevelLocal, Type: TypeWork},
	}
}
