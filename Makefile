DIST_DIR ?= dist
GOOS ?= $(shell uname -s | tr "[:upper:]" "[:lower:]")
ARCH ?= $(shell uname -m)
ifeq ($(ARCH),x86_64)
GOARCH             := amd64
else
GOARCH             := $(ARCH)  # no idea if this works for other platforms....
endif
BUILDINFOSDET ?=
PROGRAM_ARGS ?=

PROJECT_VERSION           := 2.2.2
BUILD_ID                  := 1
DOCKER_REPO               := synfinatic
PROJECT_NAME              := alpacascope
PROJECT_TAG               := $(shell git describe --tags 2>/dev/null $(git rev-list --tags --max-count=1))
ifeq ($(PROJECT_TAG),)
PROJECT_TAG               := NO-TAG
endif
PROJECT_COMMIT            := $(shell git rev-parse HEAD)
ifeq ($(PROJECT_COMMIT),)
PROJECT_COMMIT            := NO-CommitID
endif
PROJECT_DELTA             := $(shell DELTA_LINES=$$(git diff | wc -l); if [ $${DELTA_LINES} -ne 0 ]; then echo $${DELTA_LINES} ; else echo "''" ; fi)
VERSION_PKG               := $(shell echo $(PROJECT_VERSION) | sed 's/^v//g')
LICENSE                   := GPLv3
URL                       := https://github.com/$(DOCKER_REPO)/$(PROJECT_NAME)
DESCRIPTION               := AlpacaScope: Alpaca to Telescope Protocol Proxy
BUILDINFOS                := $(shell date +%FT%T%z)$(BUILDINFOSDET)
HOSTNAME                  := $(shell hostname)
LDFLAGS                   := -X "main.Version=$(PROJECT_VERSION)" -X "main.Delta=$(PROJECT_DELTA)" -X "main.Buildinfos=$(BUILDINFOS)" -X "main.Tag=$(PROJECT_TAG)" -X "main.CommitID=$(PROJECT_COMMIT)"
OUTPUT_NAME               := $(DIST_DIR)/$(PROJECT_NAME)-$(PROJECT_VERSION)-$(GOOS)-$(GOARCH)  # default for current platform
# supported platforms for `make release`
LINUX_BIN                 := $(DIST_DIR)/$(PROJECT_NAME)-$(PROJECT_VERSION)-linux-amd64
LINUXARM64_BIN            := $(DIST_DIR)/$(PROJECT_NAME)-$(PROJECT_VERSION)-linux-arm64
LINUXARM32_BIN            := $(DIST_DIR)/$(PROJECT_NAME)-$(PROJECT_VERSION)-linux-arm32
LINUX_GUI                 := $(DIST_DIR)/$(PROJECT_NAME)-gui-$(PROJECT_VERSION)-linux-amd64
DARWIN_BIN                := $(DIST_DIR)/$(PROJECT_NAME)-$(PROJECT_VERSION)-darwin-amd64
DARWIN_RELEASE_GUI        := $(DIST_DIR)/AlpacaScope.app
DARWIN_RELEASE_ZIP        := $(DIST_DIR)/AlpacaScope-$(PROJECT_VERSION).app.zip
DARWIN_GUI                := $(DIST_DIR)/$(PROJECT_NAME)-gui-$(PROJECT_VERSION)-darwin-amd64
WINDOWS_RELEASE           := $(DIST_DIR)/AlpacaScope.exe
WINDOWS_CLI               := $(DIST_DIR)/AlpacaScope-CLI-$(PROJECT_VERSION).exe
WINDOWS                   := $(DIST_DIR)/AlpacaScope-Debug-$(PROJECT_VERSION).exe

GO_FILES = $(shell find . -type f -name '*.go' | grep -v _test.go) Makefile

ALL: $(OUTPUT_NAME) ## Build binary.  Needs to be a supported plaform as defined above

include help.mk  # place after ALL target and before all other targets

release: clean .build-release ## Build CLI release files
	@echo "You still need to build 'make windows-release' and 'make sign-relase'"

.PHONY: sign-release
sign-release: $(DIST_DIR)/release.sig ## Sign release

.PHONY:
.verify_windows:
	@if test ! -f $(WINDOWS_RELEASE); then echo "Missing Windows release binary"; exit 1; fi

$(DIST_DIR)/release.sig: .build-release $(DARWIN_RELEASE_ZIP) .verify_windows
	cd dist && shasum -a 256 * | gpg --clear-sign >release.sig

# This target builds anywhere
.build-release: $(LINUX_BIN) $(LINUXARM64_BIN) $(LINUXARM32_BIN) $(DARWIN_BIN) $(DARWIN_GUI) $(DARWIN_RELEASE_GUI) $(WINDOWS_CLI)

# this targets only build on MacOS
build-gui: darwin-gui darwin-release-gui windows linux-gui ## Build GUI binaries

.build-gui-check:
	@if test $(GOOS) != "darwin" ; then echo "$(MAKECMDGOALS) requires building on MacOS" ; exit 1 ; fi

.build-windows-check:
	@if test -z "`echo $(GOOS) | grep 'mingw64'`" ; then echo "$(MAKECMDGOALS) requires building on Windows/MINGW64" ; exit 1 ; fi


# Install fyne binary in $GOPATh/bin
.PHONY: .fyne .fyne-cross
.fyne:
	@if test -z "`which fyne`"; then echo "Please install fyne: go get fyne.io/fyne/v2/cmd/fyne" ; exit 1 ; fi

.fyne-cross:
	@if test -z "`which fyne-cross`"; then echo "Please install fyne-cross: go get github.com/fyne-io/fyne-cross" ; exit 1 ; fi

# used by our github action to test building the release binaries + GUI on Linux
.build-test-binaries: $(LINUX_BIN) $(DARWIN_BIN) $(WINDOWS)

.PHONY: run
run: cmd/*.go  ## build and run cria using $PROGRAM_ARGS
	go run ./cmd/... $(PROGRAM_ARGS)

clean-all: clean ## clean _everything_

clean: ## Remove all binaries in dist
	rm -rf dist/*

clean-go: ## Clean Go cache
	go clean -i -r -cache -modcache

go-get:  ## Get our go modules
	go get -v all

.PHONY: build-race
build-race: .prepare ## Build race detection binary
	go build -race -ldflags='$(LDFLAGS)' -o $(OUTPUT_NAME) ./cmd/...
	go build -race -ldflags='$(LDFLAGS)' -o $(OUTPUT_NAME) ./gui/...

debug: .prepare ## Run debug in dlv
	dlv debug ./cmd/...

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
	@if test ! -d $(DIST_DIR); then mkdir -p $(DIST_DIR) ; fi

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
linux: $(LINUX_BIN)  ## Build Linux/x86_64 CLI

$(LINUX_BIN): $(GO_FILES) | .prepare
	GOARCH=amd64 GOOS=linux go build -ldflags='$(LDFLAGS)' -o $(LINUX_BIN) ./cmd/...
	@echo "Created: $(LINUX_BIN)"

linux-arm64: $(LINUXARM64_BIN)  ## Build Linux/arm64 CLI

$(LINUXARM64_BIN): $(GO_FILES) | .prepare
	GOARCH=arm64 GOOS=linux go build -ldflags='$(LDFLAGS)' -o $(LINUXARM64_BIN) ./cmd/...
	@echo "Created: $(LINUXARM64_BIN)"

linux-arm32: $(LINUXARM32_BIN)  ## Build Linux/arm64 CLI

$(LINUXARM32_BIN): $(GO_FILES) | .prepare
	GOARCH=arm GOOS=linux go build -ldflags='$(LDFLAGS)' -o $(LINUXARM32_BIN) ./cmd/...
	@echo "Created: $(LINUXARM32_BIN)"

darwin: $(DARWIN_BIN)  ## Build MacOS/x86_64 CLI

$(DARWIN_BIN): $(GO_FILES) | .prepare
	GOARCH=amd64 GOOS=darwin go build -ldflags='$(LDFLAGS)' -o $(DARWIN_BIN) ./cmd/...
	@echo "Created: $(DARWIN_BIN)"

darwin-gui: $(DARWIN_GUI)  ## Build MacOS/x86_64 GUI
darwin-release-gui: $(DARWIN_RELEASE_GUI)  ## Build MacOS/x86_64 Release GUI

$(DARWIN_RELEASE_GUI): $(GO_FILES) | .build-gui-check .prepare .fyne
	@fyne package -appID net.synfin.alpacascope -name AlpacaScope \
		-appVersion $(PROJECT_VERSION) -appBuild $(BUILD_ID) \
		-os darwin -sourceDir gui && \
		rm -rf $(DARWIN_RELEASE_GUI) && mv AlpacaScope.app $(DARWIN_RELEASE_GUI)

$(DARWIN_RELEASE_ZIP): $(DARWIN_RELEASE_GUI)
	@zip -mr $(DARWIN_RELEASE_ZIP) $(DARWIN_RELEASE_GUI)


$(DARWIN_GUI): $(GO_FILES) | .build-gui-check .prepare
	go build -ldflags='$(LDFLAGS)' -o $(DARWIN_GUI) ./gui/...

windows: $(WINDOWS)  ## Build Windows/x86_64 GUI

$(WINDOWS): $(GO_FILES) | .fyne-cross .prepare
	@fyne-cross windows -app-id net.synfin.alpacascope -developer "Aaron Turner" \
		-app-version $(PROJECT_VERSION) -ldflags '$(LDFLAGS)' \
		-icon $(shell pwd)/gui/Icon.png \
		-name AlpacaScope.exe $(shell pwd)/gui && \
		mv fyne-cross/bin/windows-amd64/AlpacaScope.exe $(WINDOWS)

windows-release: $(WINDOWS_RELEASE)  ## Build Windows/x86_64 release GUI

$(WINDOWS_RELEASE): $(GO_FILES) | .build-windows-check .prepare .fyne
	@rm -f dist/AlpacaScope-$(PROJECT_VERSION).exe && \
	fyne package -appID net.synfin.AlpacaScope -name net.synfin.AlpacaScope \
		-appVersion $(PROJECT_VERSION) -appBuild $(BUILD_ID) -os windows -release \
		-sourceDir gui && \
		mv gui/gui.exe $(WINDOWS_RELEASE)

windows-cli: $(WINDOWS_CLI)  ## Build Windows/amd64 CLI

$(WINDOWS_CLI): $(GO_FILES) | .prepare
	GOARCH=amd64 GOOS=windows go build -ldflags='$(LDFLAGS)' -o $(WINDOWS_CLI) ./cmd/...
	@echo "Created: $(WINDOWS_CLI)"

linux-gui: $(LINUX_GUI)  ## Build Linux/x86_64 GUI

$(LINUX_GUI): $(GO_FILES) | .prepare .fyne-cross
	@fyne-cross linux -app-id net.synfin.alpacascope \
		-app-version $(PROJECT_VERSION) -ldflags '$(LDFLAGS)' \
		-icon $(shell pwd)/gui/Icon.png \
		-name alpacascope $(shell pwd)/gui && \
		mv fyne-cross/bin/linux-amd64/alpacascope $(LINUX_GUI)
