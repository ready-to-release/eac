package commandparser_test

import (
	"fmt"
	
	parser "github.com/ready-to-release/eac/src/cli/internal/command-parser"
	"github.com/ready-to-release/eac/src/cli/internal/validator"
)

// Example demonstrates parsing and validating a command
func Example() {
	// Create parser and validator
	p := parser.NewParser()
	v := validator.NewCommandValidator()
	
	// Example 1: Simple command
	args1 := []string{"r2r", "version", "--json"}
	parsed1 := p.Parse(args1)
	fmt.Printf("Command: %v\n", args1)
	fmt.Printf("  Viper args: %v\n", parsed1.ViperArgs)
	fmt.Printf("  Container args: %v\n\n", parsed1.ContainerArgs)
	
	// Example 2: Run command with container arguments
	args2 := []string{"r2r", "--r2r-debug", "run", "pwsh", "Write-Host", "Hello", "World"}
	parsed2 := p.Parse(args2)
	fmt.Printf("Command: %v\n", args2)
	fmt.Printf("  Viper args: %v\n", parsed2.ViperArgs)
	fmt.Printf("  Container args: %v\n", parsed2.ContainerArgs)
	fmt.Printf("  Boundary at index: %d\n\n", parsed2.ArgumentBoundary)
	
	// Example 3: Validation
	args3 := []string{"r2r", "run", "python", "script.py", "--verbose"}
	result := v.ValidateCommand(args3)
	fmt.Printf("Command: %v\n", args3)
	fmt.Printf("  Valid: %v\n", result.Valid)
	fmt.Printf("  Summary: %s\n", result.Summary())
	
	// Output:
	// Command: [r2r version --json]
	//   Viper args: [r2r version --json]
	//   Container args: []
	//
	// Command: [r2r --r2r-debug run pwsh Write-Host Hello World]
	//   Viper args: [r2r --r2r-debug run pwsh]
	//   Container args: [Write-Host Hello World]
	//   Boundary at index: 4
	//
	// Command: [r2r run python script.py --verbose]
	//   Valid: true
	//   Summary: Command valid
}

// ExampleParser_SplitArguments demonstrates splitting arguments
func ExampleParser_SplitArguments() {
	p := parser.NewParser()
	
	args := []string{"r2r", "run", "python", "app.py", "--port", "8080"}
	viper, container := p.SplitArguments(args)
	
	fmt.Printf("Full command: %v\n", args)
	fmt.Printf("Viper processes: %v\n", viper)
	fmt.Printf("Container receives: %v\n", container)
	
	// Output:
	// Full command: [r2r run python app.py --port 8080]
	// Viper processes: [r2r run python]
	// Container receives: [app.py --port 8080]
}

// ExampleCommandValidator_ValidateCommand demonstrates validation
func ExampleCommandValidator_ValidateCommand() {
	v := validator.NewCommandValidator()
	
	// Valid command
	valid := []string{"r2r", "run", "pwsh"}
	result1 := v.ValidateCommand(valid)
	fmt.Printf("Command: %v -> %s\n", valid, result1.Summary())
	
	// Invalid command
	invalid := []string{"r2r", "run"}
	result2 := v.ValidateCommand(invalid)
	fmt.Printf("Command: %v -> %s\n", invalid, result2.Summary())
	if !result2.Valid {
		fmt.Printf("  Error: %s\n", result2.Errors[0])
	}
	
	// Output:
	// Command: [r2r run pwsh] -> Command valid
	// Command: [r2r run] -> Command invalid: 1 error(s), 0 warning(s)
	//   Error: run command requires an extension name
}