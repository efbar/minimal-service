FROM golang:1.15-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/yags

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Unit tests
RUN CGO_ENABLED=0 go test ./handlers/

# Build the Go app
RUN go build -o ./out/yags .

# Start fresh from a smaller image
FROM alpine:3.12.3
RUN apk add curl

COPY --from=build_base /tmp/yags/out/yags /app/yags

# This container exposes port 9090 to the outside world
EXPOSE 9090
ARG alt_port
ENV SERVICE_PORT=$alt_port

# Run the binary program produced by `go install`
CMD ["/app/yags"]