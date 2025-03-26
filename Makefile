# === Configurable Variables ===
BIN_PATH := bin
BIN_NAME := rinzlerlabs-sbc-hwmonitor
ENTRY_POINT := module.go
VERSION_PATH := utils/version.go
PLATFORM := linux/arm64  # Default platform for single-platform targets
BUILD_TAGS ?=

BIN := $(BIN_PATH)/$(BIN_NAME)
PACKAGE_DIR := package
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# === Dynamic Variables ===
VERSION := $(shell grep 'Version' $(VERSION_PATH) | sed -E 's/.*Version\s*=\s*"([^"]+)".*/\1/')
GIT_VERSION := $(shell git describe --tags --abbrev=0 | sed 's/^v//')
GOPSUTIL_VERSION := $(shell grep 'shirou/gopsutil' go.mod | sed -E 's/.*v([0-9]+\.[0-9]+\.[0-9]+).*/\1/')

# Warn on version mismatch
ifneq ($(VERSION),$(GIT_VERSION))
$(warning VERSION ($(VERSION)) and GIT_VERSION ($(GIT_VERSION)) do not match)
endif

# === Package Name Generator ===
define PACKAGE_NAME
$(BIN_NAME)_$(subst /,_,$(1)).tar.gz
endef

# === Build Template ===
define BUILD_TEMPLATE
build_$(subst /,_,$(1)):
	@echo "Building $(BIN_NAME) for $(1)..."
	@GOOS=$(word 1,$(subst /, ,$(1))) GOARCH=$(word 2,$(subst /, ,$(1))) \
		CGO_ENABLED=1 \
		go build -tags=$(BUILD_TAGS) -o $(BIN) $(ENTRY_POINT)
endef

# === Package Template ===
define PACKAGE_TEMPLATE
package_$(subst /,_,$(1)): download-license build_$(subst /,_,$(1))
	@echo "Packaging $(BIN_NAME) for $(1)..."
	@mkdir -p $(PACKAGE_DIR)
	@tar -czf $(PACKAGE_DIR)/$(call PACKAGE_NAME,$(1)) \
		$(BIN) meta.json gopsutil_LICENSE
endef

# === Upload Template ===
define UPLOAD_TEMPLATE
upload_$(subst /,_,$(1)): _package_$(subst /,_,$(1))
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
	@echo "✅ Git checks passed. Uploading $(call PACKAGE_NAME,$(1)) for platform $(1)..."
	@viam module update
	@viam module upload --version=$(VERSION) --platform=$(1) $(PACKAGE_DIR)/$(call PACKAGE_NAME,$(1))
	@echo "✅ Upload complete."
endef


# === Evaluate Templates ===
$(foreach platform,$(PLATFORMS),$(eval $(call BUILD_TEMPLATE,$(platform))))
$(foreach platform,$(PLATFORMS),$(eval $(call PACKAGE_TEMPLATE,$(platform))))
$(foreach platform,$(PLATFORMS),$(eval $(call UPLOAD_TEMPLATE,$(platform))))

# === Public Targets ===
.PHONY: all build-all package-all upload-all \
        build package upload \
        clean clean-package download-license

all: build

build-all: $(foreach platform,$(PLATFORMS),build_$(subst /,_,$(platform)))
package-all: $(foreach platform,$(PLATFORMS),package_$(subst /,_,$(platform)))
upload-all: $(foreach platform,$(PLATFORMS),upload_$(subst /,_,$(platform)))

# Single-platform targets (use PLATFORM=...)
build:
	$(MAKE) build_$(subst /,_,$(PLATFORM))

package:
	$(MAKE) package_$(subst /,_,$(PLATFORM))

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
	@viam module upload --version=$(VERSION) --platform=$(PLATFORM) $(PACKAGE_DIR)/$(call PACKAGE_NAME,$(PLATFORM))
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
	