# === Configurable Variables ===
BIN_PATH := bin
BIN_NAME := rinzlerlabs-sbc-hwmonitor
ENTRY_POINT := module.go
VERSION_PATH := utils/version.go
# Target platform. These drive both the compiler and the `viam module upload`
# --platform flag, so they must stay in agreement.
#
# Viam cloud build sets VIAM_TARGET_OS, and VIAM_BUILD_OS / VIAM_BUILD_ARCH,
# per target platform. Honor those first so a cloud build produces the artifact
# it was actually asked for, and fall back to linux/arm64 for local builds.
# Either can still be overridden on the command line to cross-build the other
# architecture declared in meta.json, ex: make build GOARCH=amd64
VIAM_BUILD_OS ?= $(VIAM_TARGET_OS)
GOOS ?= $(or $(VIAM_BUILD_OS),linux)
GOARCH ?= $(or $(VIAM_BUILD_ARCH),arm64)
PLATFORM := $(GOOS)/$(GOARCH)
PLATFORM_MONIKER := $(GOOS)-$(GOARCH)

# -trimpath drops local filesystem paths from the binary, so builds are
# reproducible and don't leak the build machine's directory layout.
# -s -w strip the symbol table and DWARF debug info at link time. Doing this
# in the linker rather than with strip(1) keeps cross-compilation working
# without needing binutils for the target architecture on the build host.
GO_LDFLAGS := -s -w
GO_BUILD_FLAGS := -trimpath -ldflags "$(GO_LDFLAGS)"

# Build without cgo so the binary is statically linked and carries no glibc
# version floor. A cgo-linked build inherits the build host's glibc symbol
# versions: building on a modern host stamps in GLIBC_2.34 (the release that
# folded libpthread/libresolv into libc), which then refuses to start on
# Ubuntu 20.04, Debian 11 or Raspberry Pi OS bullseye, all glibc 2.31. Nothing
# here needs cgo -- gopsutil is pure Go on Linux and the sensors read sysfs --
# and cross-compiled targets already disable it implicitly, so this makes every
# architecture behave the same way.
GO_BUILD_ENV := CGO_ENABLED=0

BIN := $(BIN_PATH)/$(BIN_NAME)
PACKAGE_DIR := package
PACKAGE_NAME := $(BIN_NAME).tar.gz

# === Dynamic Variables ===
VERSION := $(shell grep 'Version' $(VERSION_PATH) | sed -E 's/.*Version\s*=\s*"([^"]+)".*/\1/')
GIT_VERSION := $(shell git describe --tags --abbrev=0 | sed 's/^v//')
GOPSUTIL_VERSION := $(shell grep 'shirou/gopsutil' go.mod | sed -E 's/.*v([0-9]+\.[0-9]+\.[0-9]+).*/\1/')

# Warn on version mismatch
ifneq ($(VERSION),$(GIT_VERSION))
$(warning VERSION ($(VERSION)) and GIT_VERSION ($(GIT_VERSION)) do not match)
endif

build:
	@echo "Building $(BIN_NAME) for $(PLATFORM)..."
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_BUILD_ENV) \
		go build $(GO_BUILD_FLAGS) -o $(BIN) $(ENTRY_POINT)

package: build download-license
	@echo "Packaging $(BIN_NAME) for $(PLATFORM)..."
	@mkdir -p $(PACKAGE_DIR)
	@tar -czf $(PACKAGE_DIR)/$(PACKAGE_NAME) \
		$(BIN) meta.json gopsutil_LICENSE

# === Public Targets ===
.PHONY: build package upload \
        clean clean-package download-license

all: build

test: build
	@echo "Running tests..."
	@go test -race -cover -coverprofile=coverage.out -v ./...

upload: package
	@if [ "$(VERSION)" != "$(GIT_VERSION)" ]; then \
        echo "❌ VERSION ($(VERSION)) and GIT_VERSION ($(GIT_VERSION)) do not match."; \
        exit 1; \
    fi
	@if ! git describe --exact-match --tags HEAD >/dev/null 2>&1; then \
        echo "❌ HEAD is not tagged with $(VERSION). You must tag the latest commit before uploading."; \
        exit 1; \
    fi
	@if ! git diff --quiet || ! git diff --cached --quiet; then \
        echo "❌ Working directory has uncommitted changes. Please commit or stash them before uploading."; \
        exit 1; \
    fi
	@echo "✅ Git checks passed. Uploading..."
	@viam module update
	@viam module upload --version=$(VERSION) --platform=$(PLATFORM) $(PACKAGE_DIR)/$(PACKAGE_NAME)
	@echo "✅ Upload complete."

# License downloader
download-license:
	@echo "Downloading gopsutil LICENSE..."
	@curl -fsSL -o gopsutil_LICENSE "https://raw.githubusercontent.com/shirou/gopsutil/refs/tags/v$(GOPSUTIL_VERSION)/LICENSE"

# Cleanups
clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_PATH) $(PACKAGE_DIR) gopsutil_LICENSE

clean-package:
	@echo "Cleaning up package directory..."
	@rm -rf $(PACKAGE_DIR)
	