.PHONY: build build-cli build-gui build-all-cli test clean dev \
        build-cli-windows build-cli-linux build-cli-mac \
        build-gui-windows build-gui-linux build-gui-mac

# Default target - build for current platform
all: build

# Build both CLI and GUI for current platform
build: build-cli build-gui

# =============================================================================
# CLI Builds
# =============================================================================

# Build CLI for current platform
build-cli:
	@echo "Building CLI for current platform..."
	@mkdir -p bin
	go build -o bin/qa-cli ./cmd/cli/
	@echo "✓ CLI built: bin/qa-cli"

# Build CLI for all platforms
build-all-cli: build-cli-windows build-cli-linux build-cli-mac
	@echo "✓ All CLI builds complete"

# Build CLI for Windows (amd64)
build-cli-windows:
	@echo "Building CLI for Windows..."
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -o bin/qa-cli-windows-amd64.exe ./cmd/cli/
	@echo "✓ bin/qa-cli-windows-amd64.exe"

# Build CLI for Linux (amd64)
build-cli-linux:
	@echo "Building CLI for Linux..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/qa-cli-linux-amd64 ./cmd/cli/
	@echo "✓ bin/qa-cli-linux-amd64"

# Build CLI for macOS (arm64 - Apple Silicon)
build-cli-mac:
	@echo "Building CLI for macOS..."
	@mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -o bin/qa-cli-darwin-arm64 ./cmd/cli/
	GOOS=darwin GOARCH=amd64 go build -o bin/qa-cli-darwin-amd64 ./cmd/cli/
	@echo "✓ bin/qa-cli-darwin-arm64"
	@echo "✓ bin/qa-cli-darwin-amd64"

# =============================================================================
# GUI Builds (requires Wails CLI installed)
# =============================================================================

# Build GUI for current platform
# Note: On Linux with Ubuntu 24.04+, use build-gui-linux target for webkit2gtk-4.1 support
build-gui:
	@echo "Building GUI for current platform..."
ifeq ($(shell uname),Linux)
	cd cmd/gui && wails build -tags webkit2_41
else
	cd cmd/gui && wails build
endif
	@echo "✓ GUI built: cmd/gui/build/bin/"

# Build GUI for Windows
build-gui-windows:
	@echo "Building GUI for Windows..."
	cd cmd/gui && wails build -platform windows/amd64
	@echo "✓ GUI built for Windows"

# Build GUI for Linux
# Note: Uses webkit2_41 tag for Ubuntu 24.04+ which ships with webkit2gtk-4.1
build-gui-linux:
	@echo "Building GUI for Linux..."
	cd cmd/gui && wails build -platform linux/amd64 -tags webkit2_41
	@echo "✓ GUI built for Linux"

# Build GUI for macOS
build-gui-mac:
	@echo "Building GUI for macOS..."
	cd cmd/gui && wails build -platform darwin/universal
	@echo "✓ GUI built for macOS (universal)"

# =============================================================================
# Development & Testing
# =============================================================================

# Run CLI with test data
run-cli: build-cli
	./bin/qa-cli \
		-locations mock_data/TEST_Locations.csv \
		-priorities mock_data/TEST_Priorities.xlsx \
		-verbose

# Run GUI in dev mode
run-gui:
	cd cmd/gui && wails dev

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf cmd/gui/build/
	rm -f output.csv output.xlsx

# Development mode (using Air)
dev:
	air

# =============================================================================
# Help
# =============================================================================

help:
	@echo "QA Script Build System"
	@echo ""
	@echo "CLI Targets:"
	@echo "  build-cli           Build CLI for current platform"
	@echo "  build-all-cli       Build CLI for Windows, Linux, and macOS"
	@echo "  build-cli-windows   Build CLI for Windows (amd64)"
	@echo "  build-cli-linux     Build CLI for Linux (amd64)"
	@echo "  build-cli-mac       Build CLI for macOS (arm64 + amd64)"
	@echo ""
	@echo "GUI Targets:"
	@echo "  build-gui           Build GUI for current platform"
	@echo "  build-gui-windows   Build GUI for Windows"
	@echo "  build-gui-linux     Build GUI for Linux"
	@echo "  build-gui-mac       Build GUI for macOS (universal)"
	@echo ""
	@echo "Other:"
	@echo "  test                Run all tests"
	@echo "  clean               Remove build artifacts"
	@echo "  run-cli             Build and run CLI with test data"
	@echo "  run-gui             Run GUI in development mode"
