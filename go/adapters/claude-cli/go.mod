module github.com/ready-to-release/eac/go/adapters/claude-cli

go 1.24.4

require (
	github.com/ready-to-release/eac/contracts/ai-provider/0.1.0 v0.0.0
	github.com/ready-to-release/eac/go/clibase v0.0.0
)

require (
	github.com/ready-to-release/eac/contracts/core/0.1.0 v0.0.0 // indirect
	github.com/ready-to-release/eac/go/core v0.0.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/ready-to-release/eac/contracts/ai-provider/0.1.0 => ../../../contracts/ai-provider/0.1.0
	github.com/ready-to-release/eac/go/clibase => ../../clibase
	github.com/ready-to-release/eac/go/core => ../../core
)

replace github.com/ready-to-release/eac/contracts/core/0.1.0 => ../../../contracts/core/0.1.0

replace github.com/ready-to-release/eac/contracts/container-runtime/0.1.0 => ../../../contracts/container-runtime/0.1.0

replace github.com/ready-to-release/eac/contracts/tui/0.1.0 => ../../../contracts/tui/0.1.0

replace github.com/ready-to-release/eac/contracts/scanner/0.1.0 => ../../../contracts/scanner/0.1.0

replace github.com/ready-to-release/eac/contracts/runner/0.1.0 => ../../../contracts/runner/0.1.0
