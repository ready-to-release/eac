package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// MCP Server for Structurizr architecture documentation

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

var workspaceRoot string

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0) // Disable timestamps in logs

	// Get workspace root from environment or use default
	workspaceRoot = os.Getenv("STRUCTURIZR_WORKSPACE_ROOT")
	if workspaceRoot == "" {
		workspaceRoot = "docs/reference/design"
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(workspaceRoot)
	if err != nil {
		log.Fatalf("Failed to resolve workspace root: %v", err)
	}
	workspaceRoot = absPath

	// Ensure workspace root exists
	if err := os.MkdirAll(workspaceRoot, 0755); err != nil {
		log.Fatalf("Failed to create workspace root: %v", err)
	}

	// Disabled to avoid stderr noise during MCP initialization
	// log.Printf("Structurizr MCP Server starting")
	// log.Printf("Workspace root: %s", workspaceRoot)

	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var request MCPRequest
		if err := json.Unmarshal(line, &request); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		response := handleRequest(request)

		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from stdin: %v", err)
	}
}

func handleRequest(request MCPRequest) MCPResponse {
	// log.Printf("Handling MCP request: %s", request.Method) // Disabled for clean MCP output

	switch request.Method {
	case "initialize":
		return handleInitialize(request.ID)
	case "tools/list":
		return handleToolsList(request.ID)
	case "tools/call":
		return handleToolsCall(request)
	default:
		return errorResponse(request.ID, -32601, fmt.Sprintf("Method not found: %s", request.Method))
	}
}

func handleInitialize(id interface{}) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]interface{}{
				"name":    "structurizr",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		},
	}
}

func handleToolsList(id interface{}) MCPResponse {
	tools := []Tool{
		{
			Name:        "create_workspace",
			Description: "Create a new Structurizr workspace for architecture documentation",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"module":      {Type: "string", Description: "Module name (e.g., 'cli', 'vscode', 'docs')"},
					"name":        {Type: "string", Description: "Workspace name (e.g., 'CLI Architecture')"},
					"description": {Type: "string", Description: "Workspace description"},
				},
				Required: []string{"module", "name", "description"},
			},
		},
		{
			Name:        "add_container",
			Description: "Add a container to a software system in the workspace",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"module":      {Type: "string", Description: "Module name"},
					"name":        {Type: "string", Description: "Container name"},
					"technology":  {Type: "string", Description: "Technology/platform (e.g., 'Go', 'React')"},
					"description": {Type: "string", Description: "Container's purpose and responsibilities"},
				},
				Required: []string{"module", "name", "technology", "description"},
			},
		},
		{
			Name:        "add_relationship",
			Description: "Define a relationship between two elements",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"module":      {Type: "string", Description: "Module name"},
					"source":      {Type: "string", Description: "Source element ID (snake_case)"},
					"destination": {Type: "string", Description: "Destination element ID (snake_case)"},
					"description": {Type: "string", Description: "Relationship description"},
					"technology":  {Type: "string", Description: "Technology/protocol (optional)"},
				},
				Required: []string{"module", "source", "destination", "description"},
			},
		},
		{
			Name:        "export_workspace",
			Description: "Export workspace DSL content to docs/reference/design/<module>/workspace.dsl",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"module": {Type: "string", Description: "Module name"},
				},
				Required: []string{"module"},
			},
		},
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func handleToolsCall(request MCPRequest) MCPResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(request.Params, &params); err != nil {
		return errorResponse(request.ID, -32602, fmt.Sprintf("Invalid params: %v", err))
	}

	// log.Printf("Executing tool: %s with args: %v", params.Name, params.Arguments) // Disabled for clean MCP output

	var result string
	var err error

	switch params.Name {
	case "create_workspace":
		result, err = toolCreateWorkspace(params.Arguments)
	case "add_container":
		result, err = toolAddContainer(params.Arguments)
	case "add_relationship":
		result, err = toolAddRelationship(params.Arguments)
	case "export_workspace":
		result, err = toolExportWorkspace(params.Arguments)
	default:
		return errorResponse(request.ID, -32601, fmt.Sprintf("Tool not found: %s", params.Name))
	}

	if err != nil {
		return errorResponse(request.ID, -32603, fmt.Sprintf("Tool execution failed: %v", err))
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		},
	}
}

func toolCreateWorkspace(args map[string]interface{}) (string, error) {
	module, _ := args["module"].(string)
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)

	if module == "" || name == "" || description == "" {
		return "", fmt.Errorf("module, name, and description are required")
	}

	modulePath := filepath.Join(workspaceRoot, module)
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create module directory: %w", err)
	}

	dslContent := generateBaseDSL(name, description)
	dslPath := filepath.Join(modulePath, "workspace.dsl")

	if err := os.WriteFile(dslPath, []byte(dslContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write DSL file: %w", err)
	}

	return fmt.Sprintf("Created workspace '%s' at %s", name, dslPath), nil
}

func toolAddContainer(args map[string]interface{}) (string, error) {
	module, _ := args["module"].(string)
	name, _ := args["name"].(string)
	technology, _ := args["technology"].(string)
	description, _ := args["description"].(string)

	if module == "" || name == "" {
		return "", fmt.Errorf("module and name are required")
	}

	dslPath := filepath.Join(workspaceRoot, module, "workspace.dsl")
	if _, err := os.Stat(dslPath); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace not found for module %s", module)
	}

	dsl, err := os.ReadFile(dslPath)
	if err != nil {
		return "", fmt.Errorf("failed to read DSL file: %w", err)
	}

	containerID := sanitizeID(name)
	containerDef := fmt.Sprintf("\n            %s = container \"%s\" \"%s\" \"%s\"\n", containerID, name, description, technology)

	systemIdx := strings.Index(string(dsl), "system = softwareSystem")
	if systemIdx == -1 {
		return "", fmt.Errorf("system not found in workspace")
	}

	insertIdx := strings.Index(string(dsl)[systemIdx:], "# Containers will be added here")
	if insertIdx == -1 {
		insertIdx = strings.Index(string(dsl)[systemIdx:], "}")
	}
	insertIdx += systemIdx

	newDSL := string(dsl[:insertIdx]) + containerDef + string(dsl[insertIdx:])

	if err := os.WriteFile(dslPath, []byte(newDSL), 0644); err != nil {
		return "", fmt.Errorf("failed to write DSL file: %w", err)
	}

	return fmt.Sprintf("Added container '%s' to %s module", name, module), nil
}

func toolAddRelationship(args map[string]interface{}) (string, error) {
	module, _ := args["module"].(string)
	source, _ := args["source"].(string)
	destination, _ := args["destination"].(string)
	description, _ := args["description"].(string)
	technology, _ := args["technology"].(string)

	if module == "" || source == "" || destination == "" {
		return "", fmt.Errorf("module, source, and destination are required")
	}

	dslPath := filepath.Join(workspaceRoot, module, "workspace.dsl")
	if _, err := os.Stat(dslPath); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace not found for module %s", module)
	}

	dsl, err := os.ReadFile(dslPath)
	if err != nil {
		return "", fmt.Errorf("failed to read DSL file: %w", err)
	}

	var relationshipDef string
	if technology != "" {
		relationshipDef = fmt.Sprintf("\n        %s -> %s \"%s\" \"%s\"\n", source, destination, description, technology)
	} else {
		relationshipDef = fmt.Sprintf("\n        %s -> %s \"%s\"\n", source, destination, description)
	}

	insertMarker := "# Define relationships here"
	insertIdx := strings.Index(string(dsl), insertMarker)
	if insertIdx == -1 {
		insertIdx = strings.Index(string(dsl), "    }\n\n    views {")
		if insertIdx == -1 {
			return "", fmt.Errorf("could not find insertion point for relationship")
		}
	} else {
		insertIdx += len(insertMarker)
	}

	newDSL := string(dsl[:insertIdx]) + relationshipDef + string(dsl[insertIdx:])

	if err := os.WriteFile(dslPath, []byte(newDSL), 0644); err != nil {
		return "", fmt.Errorf("failed to write DSL file: %w", err)
	}

	return fmt.Sprintf("Added relationship: %s -> %s in %s module", source, destination, module), nil
}

func toolExportWorkspace(args map[string]interface{}) (string, error) {
	module, _ := args["module"].(string)
	if module == "" {
		return "", fmt.Errorf("module is required")
	}

	dslPath := filepath.Join(workspaceRoot, module, "workspace.dsl")
	if _, err := os.Stat(dslPath); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace not found for module %s", module)
	}

	content, err := os.ReadFile(dslPath)
	if err != nil {
		return "", fmt.Errorf("failed to read DSL file: %w", err)
	}

	return fmt.Sprintf("Exported DSL to %s (%d bytes)", dslPath, len(content)), nil
}

func generateBaseDSL(name, description string) string {
	return fmt.Sprintf(`workspace "%s" "%s" {

    model {
        # Define your software systems, containers, and components here

        system = softwareSystem "System" "Main software system" {
            # Containers will be added here
        }

        # Define relationships here
    }

    views {
        systemContext system "SystemContext" {
            include *
            autoLayout
        }

        container system "Containers" {
            include *
            autoLayout
        }

        styles {
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
            }
        }
    }

}
`, name, description)
}

func sanitizeID(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "_")
	id = strings.ReplaceAll(id, "-", "_")

	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func errorResponse(id interface{}, code int, message string) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
}
