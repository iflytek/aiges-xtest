TOP_DIR := $(shell pwd)
.PHONY: build-linux
export GO111MODULE=on
export GOPROXY=https://goproxy.cn
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) unpack
BINARY_NAME=xtest
BINARY_LINUX=$(BINARY_NAME)-linux
GORELEASER_BIN = $(shell pwd)/bin/goreleaser
SHELL = /usr/bin/env bash -o pipefail

.SHELLFLAGS = -ec

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

install-goreleaser: ## check license if not exist install go-lint tools
	#goimports -l -w cmd
	#goimports -l -w pkg
	$(call go-get-tool,$(GORELEASER_BIN),github.com/goreleaser/goreleaser@v1.6.3)

build:
	mkdir -p bin
	$(GOBUILD) -v -o ./bin/xtest ./cmd
	cp ./xtest.toml ./bin

clean:
	rm -rf bin dist

dist: build pack


build-pack: SHELL:=/bin/bash
build-pack: install-goreleaser  ## build binaries by default
	@echo "build xtest bin"
	$(GORELEASER_BIN) build --snapshot --rm-dist  --timeout=1h

build-release: install-goreleaser  ## build binaries by default
	@echo "build xtest bin"
	$(GORELEASER_BIN) release --timeout=1h  --release-notes=hack/release/Note.md --debug  --rm-dist
haha: 
	echo $(LDFLAGS)