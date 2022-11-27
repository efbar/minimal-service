module github.com/efbar/minimal-service

go 1.16

require (
	github.com/hashicorp/consul v1.14.1
	github.com/hashicorp/consul/api v1.17.0
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
)

replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.1+incompatible
