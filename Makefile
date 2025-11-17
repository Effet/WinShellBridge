APP_NAME := WinShellBridge
DIST_DIR := dist
GOOS ?= windows
GOARCH ?= amd64
BIN := $(DIST_DIR)/$(APP_NAME).exe
LDFLAGS ?= -s -w

# For systray on Windows we need CGO. On macOS cross-compiling, install mingw-w64 and ensure CC points to it.
CGO_ENABLED ?= 1
CC ?= x86_64-w64-mingw32-gcc

.PHONY: all tidy clean build-windows

all: build-windows

tidy:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go mod tidy

build-windows:
	@mkdir -p $(DIST_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) \
		go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) .

clean:
	rm -rf $(DIST_DIR)
