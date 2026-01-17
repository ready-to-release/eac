package oscal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var catalogLog = logging.C()

// LoadCatalog loads an OSCAL catalog from a URL or local file path.
// It uses go-oscal types for parsing and validation.
func LoadCatalog(catalogPath string) (*oscalTypes.Catalog, error) {
	var data []byte
	var err error

	// Check if it's a URL or local file
	if strings.HasPrefix(catalogPath, "http://") || strings.HasPrefix(catalogPath, "https://") {
		data, err = fetchCatalogFromURL(catalogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch catalog from URL: %w", err)
		}
	} else {
		data, err = os.ReadFile(catalogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read catalog file: %w", err)
		}
	}

	// Parse OSCAL document (may be wrapped in OscalModels or direct Catalog)
	// Try parsing as OscalModels first (NIST format)
	var oscalDoc oscalTypes.OscalModels
	if err := json.Unmarshal(data, &oscalDoc); err != nil {
		return nil, fmt.Errorf("failed to parse OSCAL JSON: %w", err)
	}

	// Extract catalog from wrapper
	if oscalDoc.Catalog == nil {
		return nil, fmt.Errorf("document does not contain a catalog")
	}

	catalog := oscalDoc.Catalog

	// Debug: Check what we actually parsed
	catalogLog.Debugf("Parsed catalog - UUID: '%s', Metadata.Title: '%s'", catalog.UUID, catalog.Metadata.Title)
	if catalog.Groups != nil {
		catalogLog.Debugf("Catalog has %d groups", len(*catalog.Groups))
	}
	if catalog.Controls != nil {
		catalogLog.Debugf("Catalog has %d controls", len(*catalog.Controls))
	}

	// Basic validation
	if catalog.UUID == "" {
		// Try to extract UUID from raw JSON to debug the issue
		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err == nil {
			if catalogData, ok := rawData["catalog"].(map[string]interface{}); ok {
				if rawUUID, ok := catalogData["uuid"].(string); ok {
					return nil, fmt.Errorf("catalog UUID exists in JSON (%s) but not parsed into struct - possible go-oscal version mismatch", rawUUID)
				}
			}
		}
		return nil, fmt.Errorf("catalog missing required UUID field")
	}

	// Validate that catalog has controls
	if catalog.Groups == nil && catalog.Controls == nil {
		return nil, fmt.Errorf("catalog contains no groups or controls")
	}

	return catalog, nil
}

// fetchCatalogFromURL fetches catalog JSON from a URL with timeout.
func fetchCatalogFromURL(url string) ([]byte, error) {
	// Create context with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// ExtractControlIDs extracts all control IDs from a catalog.
// It recursively walks through groups and controls using go-oscal types.
func ExtractControlIDs(catalog *oscalTypes.Catalog) ([]string, error) {
	if catalog == nil {
		return nil, fmt.Errorf("catalog is nil")
	}

	controlIDs := make([]string, 0)
	seen := make(map[string]bool)

	// Process top-level controls
	if catalog.Controls != nil {
		for _, control := range *catalog.Controls {
			if control.ID != "" && !seen[control.ID] {
				controlIDs = append(controlIDs, control.ID)
				seen[control.ID] = true
			}
		}
	}

	// Process groups recursively
	if catalog.Groups != nil {
		for _, group := range *catalog.Groups {
			extractControlIDsFromGroup(&group, &controlIDs, seen)
		}
	}

	if len(controlIDs) == 0 {
		return nil, fmt.Errorf("catalog contains no controls")
	}

	return controlIDs, nil
}

// extractControlIDsFromGroup recursively extracts control IDs from a group.
func extractControlIDsFromGroup(group *oscalTypes.Group, controlIDs *[]string, seen map[string]bool) {
	if group == nil {
		return
	}

	// Process controls in this group
	if group.Controls != nil {
		for _, control := range *group.Controls {
			if control.ID != "" && !seen[control.ID] {
				*controlIDs = append(*controlIDs, control.ID)
				seen[control.ID] = true
			}
		}
	}

	// Recursively process nested groups
	if group.Groups != nil {
		for _, nestedGroup := range *group.Groups {
			extractControlIDsFromGroup(&nestedGroup, controlIDs, seen)
		}
	}
}

// ValidateControlIDsAgainstCatalog checks if all control IDs exist in the catalog.
// Returns an error with details about missing control IDs.
func ValidateControlIDsAgainstCatalog(controlIDs []string, catalog *oscalTypes.Catalog) error {
	if catalog == nil {
		return fmt.Errorf("catalog is nil")
	}

	// Extract all valid control IDs from catalog
	catalogControlIDs, err := ExtractControlIDs(catalog)
	if err != nil {
		return fmt.Errorf("failed to extract control IDs from catalog: %w", err)
	}

	// Build a set for efficient lookup
	validControls := make(map[string]bool)
	for _, id := range catalogControlIDs {
		validControls[strings.ToLower(id)] = true
	}

	// Check each control ID
	var missingIDs []string
	for _, id := range controlIDs {
		normalizedID := strings.ToLower(id)
		if !validControls[normalizedID] {
			missingIDs = append(missingIDs, id)
		}
	}

	if len(missingIDs) > 0 {
		return fmt.Errorf("control IDs not found in catalog: %s", strings.Join(missingIDs, ", "))
	}

	return nil
}

// ControlInfo holds information about a control for profile generation.
type ControlInfo struct {
	ID          string
	Title       string
	Description string
}

// GetControlInfo retrieves title and description for specific control IDs from a catalog.
func GetControlInfo(catalog *oscalTypes.Catalog, controlIDs []string) ([]ControlInfo, error) {
	if catalog == nil {
		return nil, fmt.Errorf("catalog is nil")
	}

	// Build a map of control ID to control for efficient lookup
	controlMap := make(map[string]*oscalTypes.Control)

	// Process top-level controls
	if catalog.Controls != nil {
		for i := range *catalog.Controls {
			control := &(*catalog.Controls)[i]
			if control.ID != "" {
				controlMap[strings.ToLower(control.ID)] = control
			}
		}
	}

	// Process groups recursively
	if catalog.Groups != nil {
		for i := range *catalog.Groups {
			addControlsFromGroup(&(*catalog.Groups)[i], controlMap)
		}
	}

	// Extract info for requested control IDs
	var controlInfo []ControlInfo
	for _, id := range controlIDs {
		normalizedID := strings.ToLower(id)
		if control, found := controlMap[normalizedID]; found {
			info := ControlInfo{
				ID:    id,
				Title: control.Title,
			}

			// Extract description from parts (if available)
			if control.Parts != nil {
				for _, part := range *control.Parts {
					if part.Name == "statement" && part.Prose != "" {
						// Truncate long descriptions
						desc := part.Prose
						if len(desc) > 200 {
							desc = desc[:197] + "..."
						}
						info.Description = desc
						break
					}
				}
			}

			controlInfo = append(controlInfo, info)
		} else {
			// Control not found - add minimal info
			controlInfo = append(controlInfo, ControlInfo{
				ID:    id,
				Title: "Control not found in catalog",
			})
		}
	}

	return controlInfo, nil
}

// addControlsFromGroup recursively adds controls from a group to the control map.
func addControlsFromGroup(group *oscalTypes.Group, controlMap map[string]*oscalTypes.Control) {
	if group == nil {
		return
	}

	// Add controls from this group
	if group.Controls != nil {
		for i := range *group.Controls {
			control := &(*group.Controls)[i]
			if control.ID != "" {
				controlMap[strings.ToLower(control.ID)] = control
			}
		}
	}

	// Recursively process nested groups
	if group.Groups != nil {
		for i := range *group.Groups {
			addControlsFromGroup(&(*group.Groups)[i], controlMap)
		}
	}
}

// GetControl retrieves a single control by ID from the catalog.
func GetControl(catalog *oscalTypes.Catalog, controlID string) *oscalTypes.Control {
	if catalog == nil {
		return nil
	}

	normalizedID := strings.ToLower(controlID)

	// Build control map
	controlMap := make(map[string]*oscalTypes.Control)

	// Add top-level controls
	if catalog.Controls != nil {
		for i := range *catalog.Controls {
			control := &(*catalog.Controls)[i]
			if control.ID != "" {
				controlMap[strings.ToLower(control.ID)] = control
			}
		}
	}

	// Add controls from groups
	if catalog.Groups != nil {
		for i := range *catalog.Groups {
			addControlsFromGroup(&(*catalog.Groups)[i], controlMap)
		}
	}

	return controlMap[normalizedID]
}

// GetControlStatement extracts the statement prose from a control.
func GetControlStatement(control *oscalTypes.Control) string {
	if control == nil || control.Parts == nil {
		return ""
	}

	for _, part := range *control.Parts {
		if part.Name == "statement" && part.Prose != "" {
			return part.Prose
		}
	}

	return ""
}

// CatalogHasControl checks if a control ID exists in the catalog.
func CatalogHasControl(catalog *oscalTypes.Catalog, controlID string) bool {
	return GetControl(catalog, controlID) != nil
}
