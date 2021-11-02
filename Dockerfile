FROM golang:1.17-alpine AS build_base

RUN apk add --no-cache git
WORKDIR /tmp/minimal-service

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test ./handlers/

RUN go build -o ./out/minimal-service .

FROM alpine:3.13

RUN apk add curl

COPY --from=build_base /tmp/minimal-service/out/minimal-service /app/minimal-service

ARG alt_port
ENV SERVICE_PORT=$alt_port

CMD ["/app/minimal-service"]
