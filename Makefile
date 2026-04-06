APP_NAME := hoi4treesnap
DIST_DIR := dist
LINUX_BINARY := $(DIST_DIR)/$(APP_NAME)-linux-amd64
WINDOWS_BINARY := $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe
GO ?= go

.PHONY: build-linux build-windows test clean

build-linux:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="-s -w" -o $(LINUX_BINARY) .

build-windows:
	mkdir -p $(DIST_DIR)
	@if [ "$$(go env GOHOSTOS)" != "windows" ]; then \
		echo "build-windows is intended for a native Windows environment or the GitHub Actions windows-latest runner."; \
		exit 1; \
	fi
	GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags="-s -w" -o $(WINDOWS_BINARY) .

test:
	$(GO) test ./...

clean:
	rm -rf $(DIST_DIR)