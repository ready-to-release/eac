module github.com/ready-to-release/eac/src/commands

go 1.24.4

require (
	github.com/cucumber/godog v0.15.1
	github.com/docker/docker v28.0.0+incompatible
	github.com/docker/go-connections v0.6.0
	github.com/jedib0t/go-pretty/v6 v6.6.9
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/ready-to-release/eac/src/internal v0.0.0
	github.com/spf13/cobra v1.10.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Microsoft/go-winio v0.4.21 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cucumber/gherkin/go/v26 v26.2.0 // indirect
	github.com/cucumber/messages/go/v21 v21.0.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gofrs/uuid v4.3.1+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)

replace github.com/ready-to-release/eac/src/internal => ../internal
