package render

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// RenderStructAsMarkdown converts a Go struct to markdown using YAML serialization
//
// Example:
//
//	type Person struct {
//	    Name  string `yaml:"name"`
//	    Age   int    `yaml:"age"`
//	    Email string `yaml:"email"`
//	}
//	person := Person{Name: "Alice", Age: 30, Email: "alice@example.com"}
//	markdown := RenderStructAsMarkdown(person)
//
// Output:
//
//	```yaml
//	name: Alice
//	age: 30
//	email: alice@example.com
//	```
func RenderStructAsMarkdown(v interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return fmt.Sprintf("```yaml\n%s```", string(yamlBytes)), nil
}

// RenderStructAsMarkdownOrPanic is a convenience wrapper that panics on error
// Useful for cases where marshaling is guaranteed to succeed
func RenderStructAsMarkdownOrPanic(v interface{}) string {
	result, err := RenderStructAsMarkdown(v)
	if err != nil {
		panic(err)
	}
	return result
}

// RenderStructWithTitle renders a struct with a markdown title
//
// Example:
//
//	markdown := RenderStructWithTitle("Configuration", config)
//
// Output:
//
//	## Configuration
//
//	```yaml
//	...
//	```
func RenderStructWithTitle(title string, v interface{}) (string, error) {
	rendered, err := RenderStructAsMarkdown(v)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("## %s\n\n%s", title, rendered), nil
}

// RenderMultipleStructs renders multiple structs with their titles
//
// Example:
//
//	sections := map[string]interface{}{
//	    "Database Config": dbConfig,
//	    "Server Config": serverConfig,
//	}
//	markdown := RenderMultipleStructs(sections)
func RenderMultipleStructs(sections map[string]interface{}) (string, error) {
	var result string

	for title, v := range sections {
		section, err := RenderStructWithTitle(title, v)
		if err != nil {
			return "", fmt.Errorf("failed to render section %s: %w", title, err)
		}
		result += section + "\n\n"
	}

	return result, nil
}

// RenderStructSliceAsMarkdown renders a slice of structs as YAML
//
// Example:
//
//	people := []Person{
//	    {Name: "Alice", Age: 30},
//	    {Name: "Bob", Age: 25},
//	}
//	markdown := RenderStructSliceAsMarkdown(people)
//
// Output:
//
//	```yaml
//	- name: Alice
//	  age: 30
//	- name: Bob
//	  age: 25
//	```
func RenderStructSliceAsMarkdown(slice interface{}) (string, error) {
	return RenderStructAsMarkdown(slice)
}
