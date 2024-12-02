BIN_PATH=bin
BIN_NAME=rinzlerlabs-sbc-hwmonitor
ENTRY_POINT=module.go
VERSION_PATH=utils/version.go
PLATFORM=linux/arm64

BIN=$(BIN_PATH)/$(BIN_NAME)
PACKAGE_NAME=$(BIN_NAME).tar.gz

VERSION=$(shell grep 'Version' $(VERSION_PATH) | sed -E 's/.*Version\s*=\s*"([^"]+)".*/\1/')
GIT_VERSION=$(shell git describe --tags --abbrev=0 | sed 's/^v//')
GOPSUTIL_VERSION=$(shell grep 'shirou/gopsutil' go.mod | sed -E 's/.*v([0-9]+\.[0-9]+\.[0-9]+).*/\1/')

ifneq ($(VERSION),$(GIT_VERSION))
$(warning VERSION ($(VERSION)) and GIT_VERSION ($(GIT_VERSION)) do not match)
endif

build:
	go mod tidy
	go build -o $(BIN) $(ENTRY_POINT)

clean-package:
	rm -rf $(PACKAGE_NAME)

clean-bin:
	rm -rf $(BIN_PATH)

download-license:
	curl -L -o gopsutil_LICENSE "https://raw.githubusercontent.com/shirou/gopsutil/refs/tags/v$(GOPSUTIL_VERSION)/LICENSE"

package: clean-package build download-license
	tar -czf $(PACKAGE_NAME) $(BIN) meta.json gopsutil_LICENSE

upload: package
	@if [ "$(VERSION)" != "$(GIT_VERSION)" ]; then \
        echo "VERSION ($(VERSION)) and GIT_VERSION ($(GIT_VERSION)) do not match"; \
        exit 1; \
    fi
	viam module update
	viam module upload --version=$(VERSION) --platform=$(PLATFORM) $(PACKAGE_NAME)
