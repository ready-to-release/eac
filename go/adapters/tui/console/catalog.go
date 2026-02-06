package console

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

// WidgetCatalog is a registry of all displayable widgets.
// Contains both singleton widgets (rendered once per frame) and template widgets
// (rendered once per instance, e.g., per UoW tab).
type WidgetCatalog struct {
	widgets map[string]*Widget
	order   []string // Insertion order for iteration

	// Template widgets (rendered per instance, not per frame)
	TabWidget *TabWidgetDef // The UoW tab template (nil until registered)
}

// TabWidgetDef is the template widget for UoW tabs.
// Unlike singleton widgets, this is instantiated once per UoW in the grid.
type TabWidgetDef struct {
	ID          string // Base zone prefix
	ElementName string // "Component" -- shown in Selected pane for any tab hover
	HelpText    string // Shown for any tab hover

	// Render produces the styled tab content for one UoW (without zone.Mark wrapping).
	Render func(instance TabInstance, sizing TabSizing) string

	// ColorMap is the status-to-color mapping, extracted from inline definition.
	ColorMap map[UoWStatus]TabBadgeColors
}

// TabBadgeColors holds the color palette for one UoW status.
type TabBadgeColors struct {
	Bg       lipgloss.Color // Badge background
	Text     lipgloss.Color // Badge text
	BgActive lipgloss.Color // Active/selected badge background (lighter)
	NameBg   lipgloss.Color // Subtle name area background
}

// NewWidgetCatalog creates an empty catalog.
func NewWidgetCatalog() *WidgetCatalog {
	return &WidgetCatalog{
		widgets: make(map[string]*Widget),
	}
}

// Register adds a singleton widget to the catalog. Panics on duplicate ID.
func (c *WidgetCatalog) Register(w *Widget) {
	if _, exists := c.widgets[w.ID]; exists {
		panic("duplicate widget ID: " + w.ID)
	}
	c.widgets[w.ID] = w
	c.order = append(c.order, w.ID)
}

// RegisterTabWidget registers the template widget for UoW tabs.
// Only one tab template exists. Panics if called twice.
func (c *WidgetCatalog) RegisterTabWidget(tw *TabWidgetDef) {
	if c.TabWidget != nil {
		panic("tab widget already registered")
	}
	c.TabWidget = tw
}

// Get returns a singleton widget by ID. Returns nil if not found.
func (c *WidgetCatalog) Get(id string) *Widget {
	return c.widgets[id]
}

// RenderWidget renders a single singleton widget with the given snapshot.
// Wraps output in zone.Mark if the widget has ZoneEnabled.
// Returns empty string if widget not found.
func (c *WidgetCatalog) RenderWidget(id string, snap WidgetSnapshot) string {
	w := c.widgets[id]
	if w == nil {
		return ""
	}
	content := w.Render(snap)
	if w.ZoneEnabled {
		return zone.Mark(w.ID, content)
	}
	return content
}

// RenderTab renders a single UoW tab using the registered tab template.
// Returns zone.Mark(instance.Moniker, content) since each tab has a unique zone.
func (c *WidgetCatalog) RenderTab(instance TabInstance, sizing TabSizing) string {
	if c.TabWidget == nil {
		return ""
	}
	content := c.TabWidget.Render(instance, sizing)
	return zone.Mark(instance.Moniker, content)
}

// HelpText returns the help text for a zone ID. Returns ("", false) if not found.
// Also checks if the zone ID matches a UoW tab moniker pattern (4+ colon-separated parts).
func (c *WidgetCatalog) HelpText(zoneID string) (string, bool) {
	// Singleton widgets first
	w := c.widgets[zoneID]
	if w != nil && w.HelpText != "" {
		return w.HelpText, true
	}
	// Tab widget: any moniker-like zone ID
	if c.TabWidget != nil && c.TabWidget.HelpText != "" {
		parts := strings.Split(zoneID, ":")
		if len(parts) >= 4 {
			return c.TabWidget.HelpText, true
		}
	}
	return "", false
}

// ElementName returns the human-readable element name for a zone ID.
func (c *WidgetCatalog) ElementName(zoneID string) string {
	w := c.widgets[zoneID]
	if w != nil {
		return w.ElementName
	}
	// Tab widget: any moniker-like zone ID
	if c.TabWidget != nil {
		parts := strings.Split(zoneID, ":")
		if len(parts) >= 4 {
			return c.TabWidget.ElementName
		}
	}
	return ""
}

// AllZoneIDs returns all singleton widget IDs that have zones enabled (for test validation).
// Tab widget is excluded since its instances have dynamic zone IDs (monikers).
func (c *WidgetCatalog) AllZoneIDs() []string {
	var ids []string
	for _, id := range c.order {
		if c.widgets[id].ZoneEnabled {
			ids = append(ids, id)
		}
	}
	return ids
}
