.PHONY: install build build-static clean doc fmt lint test vet

PKG_NAME := telegram-uploader-bot
PKG := github.com/3cky/$(PKG_NAME)

ARCH := amd64 arm

BINDIR = bin

VERSION_VAR := $(PKG)/build.Version
TIMESTAMP_VAR := $(PKG)/build.Timestamp

VERSION ?= $(shell git describe --always --dirty --tags)
TIMESTAMP := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

GOBUILD_LDFLAGS := -ldflags "-s -w -X $(VERSION_VAR)=$(VERSION) -X $(TIMESTAMP_VAR)=$(TIMESTAMP)"
GOBUILD_LDFLAGS_STATIC := -ldflags "-linkmode external -extldflags \"-static\" -s -w -X $(VERSION_VAR)=$(VERSION) -X $(TIMESTAMP_VAR)=$(TIMESTAMP)"

default: install

install:
	go install -x $(GOBUILD_LDFLAGS) -v

build:
	go build -x $(GOBUILD_LDFLAGS) -v -o $(BINDIR)/$(PKG_NAME)

build-static: $(addprefix build-static-, $(ARCH))

build-static-amd64:
	env CGO_ENABLED=1 CC=musl-gcc GOOS=linux GOARCH=amd64 go build -a -installsuffix "static" $(GOBUILD_LDFLAGS_STATIC) -o $(BINDIR)/$(PKG_NAME).amd64

build-static-arm:
	env CGO_ENABLED=1 CC=arm-linux-musleabi-gcc GOOS=linux GOARCH=arm go build -a -installsuffix "static" $(GOBUILD_LDFLAGS_STATIC) -o $(BINDIR)/$(PKG_NAME).arm

shasum:
	cd $(BINDIR) && for file in $(ARCH) ; do sha256sum ./$(PKG_NAME).$${file} > ./$(PKG_NAME).$${file}.sha256.txt; done

clean:
	rm -dRf $(BINDIR)

dist: clean build-static shasum

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
