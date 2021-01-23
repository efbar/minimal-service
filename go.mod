module github.com/efbar/minimal-service

go 1.15

require (
	github.com/hashicorp/consul v1.9.1
	github.com/hashicorp/consul/api v1.8.1
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
)
