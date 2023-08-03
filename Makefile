.PHONY: install build clean doc fmt lint test vet

PKG_NAME := telegram-uploader-bot
PKG := github.com/3cky/$(PKG_NAME)

VERSION_VAR := $(PKG)/build.Version
TIMESTAMP_VAR := $(PKG)/build.Timestamp

VERSION ?= $(shell git describe --always --dirty --tags)
TIMESTAMP := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

GOBUILD_LDFLAGS := -ldflags "-s -w -X $(VERSION_VAR)=$(VERSION) -X $(TIMESTAMP_VAR)=$(TIMESTAMP)"

default: install

install:
	go get -x $(GOBUILD_LDFLAGS) -t -v ./...

build:
	go build -x $(GOBUILD_LDFLAGS) -v -o ./bin/$(PKG_NAME)

clean:
	rm -dRf ./bin

doc:
	godoc -http=:6060

fmt:
	go fmt ./...

# https://golangci.com/
# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0
lint:
	golangci-lint run --timeout=300s

test:
	go test ./...

# https://godoc.org/golang.org/x/tools/cmd/vet
vet:
	go vet ./...
