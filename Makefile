.PHONY: install build doc fmt gen lint test vet bench

PKG_NAME=$(shell basename `pwd`)
TARGET_OS="linux"
VERSION_VAR=build.Version
TIMESTAMP_VAR=build.Timestamp
VERSION=$(shell git describe --always --dirty --tags)
TIMESTAMP=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
GOBUILD_VERSION_ARGS := -ldflags "-X $(VERSION_VAR)=$(VERSION) -X $(TIMESTAMP_VAR)=$(TIMESTAMP)"

default: install

install:
	go get -x $(GOBUILD_VERSION_ARGS) -t -v ./...

build:
	go build -x $(GOBUILD_VERSION_ARGS) -v -o ./bin/$(PKG_NAME)

clean:
	rm -dRf ./bin

doc:
	godoc -http=:6060

fmt:
	go fmt ./...

gen:
	go generate

# https://golangci.com/
# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0
lint:
	golangci-lint run --timeout=300s

test:
	go test ./...

# Runs benchmarks
bench:
	go test ./... -bench=.

# https://godoc.org/golang.org/x/tools/cmd/vet
vet:
	go vet ./...
