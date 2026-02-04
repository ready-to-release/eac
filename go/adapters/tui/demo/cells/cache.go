package cells

// CacheCell displays cache hit/miss status for the selected unit.
type CacheCell struct {
	hit    bool
	reason string
}

// NewCacheCell creates a new cache cell.
func NewCacheCell() *CacheCell {
	return &CacheCell{}
}

// SetCacheInfo sets the cache status.
func (c *CacheCell) SetCacheInfo(hit bool, reason string) {
	c.hit = hit
	c.reason = reason
}

// Render returns the cache status display.
func (c *CacheCell) Render(width, height int) string {
	if c.hit {
		return "cache: hit"
	}

	if c.reason == "" {
		return "cache: miss"
	}

	return "cache: miss (" + c.reason + ")"
}

// ZoneID returns empty string (not an interactive zone).
func (c *CacheCell) ZoneID() string {
	return ""
}
