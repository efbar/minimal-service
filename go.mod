module github.com/efbar/minimal-service

go 1.15

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/hashicorp/consul v1.10.1
	github.com/hashicorp/consul/api v1.10.1
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
)

replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.1+incompatible
