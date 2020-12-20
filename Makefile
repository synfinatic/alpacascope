DIST_DIR ?= dist/
GOOS ?= $(shell uname -s | tr "[:upper:]" "[:lower:]")
ARCH ?= $(shell uname -m)
ifeq ($(ARCH),x86_64)
GOARCH             := arm64
else
GOARCH             := $(ARCH)  # no idea if this works for other platforms....
endif
BUILDINFOSDET ?=
PROGRAM_ARGS ?=

PROJECT_VERSION           := 0.0.1
DOCKER_REPO               := synfinatic
PROJECT_NAME              := alpaca-gateway
PROJECT_TAG               := $(shell git describe --tags 2>/dev/null $(git rev-list --tags --max-count=1))
ifeq ($(PROJECT_TAG),)
PROJECT_TAG               := NO-TAG
endif
PROJECT_COMMIT            := $(shell git rev-parse HEAD)
ifeq ($(PROJECT_COMMIT),)
PROJECT_COMMIT            := NO-CommitID
endif
VERSION_PKG               := $(shell echo $(PROJECT_VERSION) | sed 's/^v//g')
LICENSE                   := GPLv3
URL                       := https://github.com/$(DOCKER_REPO)/$(PROJECT_NAME)
DESCRIPTION               := Alpaca Gateway: Alpaca to Telescope Protocol Gateway for SkySafari
BUILDINFOS                := $(shell date +%FT%T%z)$(BUILDINFOSDET)
HOSTNAME                  := $(shell hostname)
LDFLAGS                   := -X "main.Version=$(PROJECT_VERSION)" -X "main.Buildinfos=$(BUILDINFOS)" -X "main.Tag=$(PROJECT_TAG)" -X "main.CommitID=$(PROJECT_COMMIT)"
OUTPUT_NAME               := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-$(GOOS)-$(GOARCH)  # default for current platform
# supported platforms for `make release`
WINDOWS_BIN               := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-windows-amd64
WINDOWS32_BIN             := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-windows-386
LINUX_BIN                 := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-linux-amd64
LINUXARM64_BIN            := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-linux-arm64
DARWIN_BIN                := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-darwin-arm64



ALL: $(OUTPUT_NAME) ## Build binary.  Needs to be a supported plaform as defined above

include help.mk  # place after ALL target and before all other targets

# release: linux-static freebsd mips64-static arm64-static $(OUTPUT_NAME) ## Build our release binaries

release: windows windows32 linux linuxarm64 darwin ## Build our release binaries

.PHONY: run
run: cmd/*.go  ## build and run cria using $PROGRAM_ARGS
	go run cmd/*.go $(PROGRAM_ARGS)

clean-all: clean ## clean _everything_

clean: ## Remove all binaries in dist
	rm -f dist/*

clean-go: ## Clean Go cache
	go clean -i -r -cache -modcache

go-get:  ## Get our go modules
	go get -v all

.PHONY: build-race
build-race: .prepare ## Build race detection binary
	go build -race -ldflags='$(LDFLAGS)' -o $(OUTPUT_NAME) cmd/*.go

debug: .prepare ## Run debug in dlv
	dlv debug cmd/*.go

.PHONY: unittest
unittest: ## Run go unit tests
	go test ./...

.PHONY: test-race
test-race: ## Run `go test -race` on the code
	@echo checking code for races...
	go test -race ./...

.PHONY: vet
vet: ## Run `go vet` on the code
	@echo checking code is vetted...
	go vet $(shell go list ./...)

test: vet unittest ## Run all tests

.prepare: $(DIST_DIR)

$(DIST_DIR):
	@if test -d $(DIST_DIR); then mkdir -p $(DIST_DIR) ; fi

.PHONY: fmt
fmt: ## Format Go code
	@go fmt cmd

.PHONY: test-fmt
test-fmt: fmt ## Test to make sure code if formatted correctly
	@if test `git diff cmd | wc -l` -gt 0; then \
	    echo "Code changes detected when running 'go fmt':" ; \
	    git diff -Xfiles ; \
	    exit -1 ; \
	fi

.PHONY: test-tidy
test-tidy:  ## Test to make sure go.mod is tidy
	@go mod tidy
	@if test `git diff go.mod | wc -l` -gt 0; then \
	    echo "Need to run 'go mod tidy' to clean up go.mod" ; \
	    exit -1 ; \
	fi

precheck: test test-fmt test-tidy  ## Run all tests that happen in a PR 


# Build targets for our supported plaforms
windows: $(WINDOWS_BIN)

$(WINDOWS_BIN): cmd/*.go alpaca/*.go .prepare
	GOARCH=amd64 GOOS=windows go build -ldflags='$(LDFLAGS)' -o $(WINDOWS_BIN) cmd/*.go
	@echo "Created: $(WINDOWS_BIN)"

windows32: $(WINDOWS32_BIN)

$(WINDOWS32_BIN):cmd/*.go alpaca/*.go .prepare
	GOARCH=386 GOOS=windows go build -ldflags='$(LDFLAGS)' -o $(WINDOWS32_BIN) cmd/*.go
	@echo "Created: $(WINDOWS32_BIN)"

linux: $(LINUX_BIN)

$(LINUX_BIN): cmd/*.go alpaca/*.go .prepare
	GOARCH=amd64 GOOS=linux go build -ldflags='$(LDFLAGS)' -o $(LINUX_BIN) cmd/*.go
	@echo "Created: $(LINUX_BIN)"

linuxarm64: $(LINUXARM64_BIN)

$(LINUXARM64_BIN): cmd/*.go alpaca/*.go .prepare
	GOARCH=arm64 GOOS=linux go build -ldflags='$(LDFLAGS)' -o $(LINUXARM64_BIN) cmd/*.go
	@echo "Created: $(LINUXARM64_BIN)"

darwin: $(DARWIN_BIN)

$(DARWIN_BIN): cmd/*.go alpaca/*.go .prepare
	GOARCH=amd64 GOOS=darwin go build -ldflags='$(LDFLAGS)' -o $(DARWIN_BIN) cmd/*.go
	@echo "Created: $(DARWIN_BIN)"

