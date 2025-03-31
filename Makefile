# === Configurable Variables ===
BIN_PATH := bin
BIN_NAME := rinzlerlabs-sbc-hwmonitor
ENTRY_POINT := module.go
VERSION_PATH := utils/version.go
PLATFORM := $(shell go env GOOS)/$(shell go env GOARCH)
PLATFORM_MONIKER := $(shell go env GOOS)-$(shell go env GOARCH)
GOOS=linux
GOARCH=arm64

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
	@GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -o $(BIN) $(ENTRY_POINT)

package: build
	@echo "Packaging $(BIN_NAME) for $(PLATFORM)..."
	@mkdir -p $(PACKAGE_DIR)
	@tar -czf $(PACKAGE_DIR)/$(PACKAGE_NAME) \
		$(BIN) meta.json gopsutil_LICENSE

# === Public Targets ===
.PHONY: build package upload \
        clean clean-package download-license

all: build

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
	