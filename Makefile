SHELL := /bin/bash

# The name of the executable (default is current directory name)
TARGET := $(shell echo $${PWD\#\#*/})
.DEFAULT_GOAL: $(TARGET)

# These will be provided to the target
VERSION := $(shell git describe --abbrev=0 --tags | sed 's/v//g')
BUILD := $(shell git rev-parse --short HEAD)

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build clean pb fmt lint test test-all serve server agent

all: clean build

install-agent: agent
	@cp ./bin/pmd-agent ${GOPATH}/bin/

serve: server
	@./bin/pmd-server

pb:
	@protoc -I/usr/local/include -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:. \
		rpc/service.proto
	@protoc -I/usr/local/include -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:. \
		rpc/service.proto
	@protoc -I/usr/local/include -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--swagger_out=logtostderr=true:. \
		rpc/service.proto	

server: 
	go build $(LDFLAGS) -o ./bin/pmd-server -v ./cmd/server

agent: 
	go build $(LDFLAGS) -o ./bin/pmd-agent -v ./cmd/agent

start: agent
	@./bin/pmd-agent

build: pb test server agent
	@true

clean:
	@rm -f bin/*
	@rm -f rpc/moviedownloader/*
	@rm -f rpc/service.swagger.json

fmt:
	# gofmt -l -w $(SRC)

test:
	# go test -short ./...

lint:
	# go vet ./...

test-all: lint test
	# go test -race ./...