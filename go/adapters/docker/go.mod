module github.com/ready-to-release/eac/go/adapters/docker

go 1.24.4

require (
	github.com/docker/docker v28.5.2+incompatible
	github.com/docker/go-connections v0.6.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces v0.0.0
	github.com/ready-to-release/eac/go/core v0.0.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/defenseunicorns/go-oscal v0.7.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/ready-to-release/eac/contracts v0.0.0-20260129143239-c3ff637b8ca9 // indirect
	github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces v0.0.0-00010101000000-000000000000 // indirect
	github.com/ready-to-release/eac/contracts/testing/0.1.0/interfaces v0.0.0-00010101000000-000000000000 // indirect
	github.com/ready-to-release/eac/contracts/tui-adapter/0.1.0/interfaces v0.0.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.40.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.40.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.40.0 // indirect
	go.opentelemetry.io/otel/sdk v1.40.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.40.0 // indirect
	go.opentelemetry.io/otel/trace v1.40.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)

replace (
	github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces => ../../../contracts/core/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces => ../../../contracts/docker-adapter/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/security/0.1.0/interfaces => ../../../contracts/security/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/testing/0.1.0/interfaces => ../../../contracts/testing/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/tools/0.1.0/interfaces => ../../../contracts/tools/0.1.0/interfaces
	github.com/ready-to-release/eac/contracts/tui-adapter/0.1.0/interfaces => ../../../contracts/tui-adapter/0.1.0/interfaces
	github.com/ready-to-release/eac/go/core => ../../core
	github.com/ready-to-release/eac/go/godog => ../../godog
)
