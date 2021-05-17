SHELL = /bin/bash
BINARY ?= agones-minecraft-api
IMAGE ?= saulmaldonado/$(BINARY)
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell set -o pipefail; git describe --exact-match --tags HEAD 2> /dev/null | cut -c 2- || echo ${COMMIT})
BUILD_FLAGS ?= -v
ARCH ?= amd64

 -include .env
 export

.PHONY: build

install:
	go mod download

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o build/$(BINARY) $(BUILD_FLAGS) .

clean:
	@rm -rf build

lint:
	golangci-lint run ./...
