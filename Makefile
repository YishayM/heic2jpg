# heic2jpg - Cross-platform build configuration
# Note: This project uses CGO (goheif depends on libde265), which complicates cross-compilation

VERSION := 1.0.0
BINARY_NAME := heic2jpg
DIST_DIR := dist
CMD_PATH := ./cmd/heic2jpg

# Build flags
LDFLAGS := -s -w
BUILD_FLAGS := -trimpath

# Detect current platform
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

.PHONY: all build build-all clean test help

# Default target
all: build

# Build for current platform
build:
	@echo "Building for current platform ($(GOOS)/$(GOARCH))..."
	@mkdir -p $(DIST_DIR)
	go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,) $(CMD_PATH)
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,)"

# Build for all platforms
# Note: Cross-compilation with CGO requires platform-specific C toolchains
# This target attempts to build for all platforms but may fail for non-native platforms
# For production releases, build natively on each target platform
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-windows-amd64
	@echo ""
	@echo "✓ Build complete! Binaries in $(DIST_DIR)/"
	@echo ""
	@echo "Note: Due to CGO dependencies, cross-compilation may fail."
	@echo "For production releases, build natively on each target platform."
	@echo ""
	@ls -lh $(DIST_DIR)/

# macOS Intel (amd64)
build-darwin-amd64:
	@echo "Building for darwin/amd64..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH) || \
		echo "⚠ Warning: darwin/amd64 build failed (CGO cross-compilation may require additional setup)"

# macOS Apple Silicon (arm64)
build-darwin-arm64:
	@echo "Building for darwin/arm64..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH) || \
		echo "⚠ Warning: darwin/arm64 build failed (CGO cross-compilation may require additional setup)"

# Linux amd64
build-linux-amd64:
	@echo "Building for linux/amd64..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH) || \
		echo "⚠ Warning: linux/amd64 build failed (CGO cross-compilation may require additional setup)"

# Windows amd64
build-windows-amd64:
	@echo "Building for windows/amd64..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH) || \
		echo "⚠ Warning: windows/amd64 build failed (CGO cross-compilation may require additional setup)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(DIST_DIR)
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Show help
help:
	@echo "heic2jpg - Makefile targets"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build for current platform"
	@echo "  build-all      Build for all platforms (darwin/amd64, darwin/arm64, linux/amd64, windows/amd64)"
	@echo "  clean          Remove build artifacts"
	@echo "  test           Run tests"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Note: This project uses CGO (goheif/libde265). Cross-compilation may require"
	@echo "      platform-specific C toolchains. Native builds should work without issues."

