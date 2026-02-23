module github.com/ready-to-release/eac/go/adapters/pytest

go 1.24.4

require (
	github.com/ready-to-release/eac/go/adapters/pip v0.0.0
	github.com/ready-to-release/eac/go/clibase v0.0.0
	github.com/ready-to-release/eac/go/core v0.0.0
)

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/defenseunicorns/go-oscal v0.7.0 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/huandu/go-clone v1.7.3 // indirect
	github.com/huandu/go-clone/generic v1.7.3 // indirect
	github.com/ready-to-release/eac/contracts/container-runtime/0.1.0 v0.0.0 // indirect
	github.com/ready-to-release/eac/contracts/core/0.1.0 v0.0.0 // indirect
	github.com/ready-to-release/eac/contracts/runner/0.1.0 v0.0.0 // indirect
	github.com/ready-to-release/eac/contracts/scanner/0.1.0 v0.0.0 // indirect
	github.com/ready-to-release/eac/contracts/tui/0.1.0 v0.0.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/ready-to-release/eac/contracts/core/0.1.0 => ../../../contracts/core/0.1.0
	github.com/ready-to-release/eac/go/adapters/pip => ../pip
	github.com/ready-to-release/eac/go/clibase => ../../clibase
	github.com/ready-to-release/eac/go/core => ../../core
)

replace github.com/ready-to-release/eac/contracts/container-runtime/0.1.0 => ../../../contracts/container-runtime/0.1.0

replace github.com/ready-to-release/eac/contracts/tui/0.1.0 => ../../../contracts/tui/0.1.0

replace github.com/ready-to-release/eac/contracts/scanner/0.1.0 => ../../../contracts/scanner/0.1.0

replace github.com/ready-to-release/eac/contracts/runner/0.1.0 => ../../../contracts/runner/0.1.0
