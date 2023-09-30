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

install: agent
	@cp ./bin/pmd-agent ${GOPATH}/bin/

serve: server
	@./bin/pmd-server

pb:
	@docker run --rm \
		--platform linux/amd64 \
		-v ${PWD}/:/app \
		-e INPUT_PATH=./proto/midgarco \
		-e OUTPUT_PATH=./ \
		-e PROTO_FILE=api/v1/service.proto \
		-e GEN_GO=true \
		-e GEN_GO_GRPC=true \
		-e GEN_GRPC_GATEWAY=true \
		-e GRPC_API_CONFIGURATION=./proto/midgarco/api/v1/service.yaml \
		proto:dev -v --output-prefix-path /rpc/

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
