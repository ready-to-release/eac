module github.com/ready-to-release/eac/go/adapters/claude

go 1.25.0

require (
	github.com/anthropics/anthropic-sdk-go v1.27.0
	github.com/ready-to-release/eac/contracts/ai-provider/0.1.0 v0.0.0
	github.com/ready-to-release/eac/go/clibase v0.0.0
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/ready-to-release/eac/contracts/core/0.1.0 v0.0.0 // indirect
	github.com/ready-to-release/eac/go/core v0.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
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
