ifneq (,$(wildcard ./.env))
    include .env
    export
endif

PROJECT_NAME=minimal-service
VERSION=1.0.0
GOFILES=$(wildcard *./.go)

## build
build:
	@-go build -o $(GOPATH)/bin/$(PROJECT_NAME) $(GOFILES)

## run
run: build
	@-go run main.go

## install
install: build
	@-go install

## test
test: build
	@-go test -race -v ./handlers

## clean
clean:
	@-go clean
	@-rm -f $(GOPATH)/bin/$(PROJECT_NAME)

## build-linux
build-linux: test
	@-CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(GOPATH)/bin/$(PROJECT_NAME) $(GOFILES)

## build-mac
build-mac: test
	@-CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(GOPATH)/bin/$(PROJECT_NAME) $(GOFILES)

## build-windows
build-windows: test
	@-CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(GOPATH)/bin/$(PROJECT_NAME) $(GOFILES)

## build-docker
docker-build:
	@-docker build . -t $(PROJECT_NAME):v$(VERSION)

.PHONY: help test clean
help: Makefile
	@echo
	@echo " Choose a command from list below"
	@echo " usage: make <command>"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo