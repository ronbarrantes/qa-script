.PHONY: build build-linux build-windows build-mac dev clean

# Default build (detect platform)
build:
	@echo "Building for current platform..."
	@case "$$(uname -s)" in \
		Linux*)  $(MAKE) build-linux ;; \
		Darwin*) $(MAKE) build-mac ;; \
		MINGW*|CYGWIN*|MSYS*) $(MAKE) build-windows ;; \
		*) echo "Unknown platform, trying generic build..." && wails build ;; \
	esac

# Linux build (uses webkit2gtk-4.1 for Ubuntu 22.04+)
build-linux:
	@echo "Building for Linux..."
	wails build -tags webkit2_41

# Windows build (primary target platform)
build-windows:
	@echo "Building for Windows..."
	wails build -platform windows/amd64

# macOS build
build-mac:
	@echo "Building for macOS..."
	wails build

# Development mode
dev:
	@echo "Starting development mode..."
	@case "$$(uname -s)" in \
		Linux*) wails dev -tags webkit2_41 ;; \
		*) wails dev ;; \
	esac

# Clean build artifacts
clean:
	rm -rf build/bin/*
	@echo "Cleaned build artifacts"

# Cross-compile for Windows from Linux/Mac
cross-windows:
	@echo "Cross-compiling for Windows..."
	GOOS=windows GOARCH=amd64 wails build -platform windows/amd64 -skipbindings

# Help
help:
	@echo "QA Script GUI - Build Commands"
	@echo ""
	@echo "  make build         - Build for current platform"
	@echo "  make build-linux   - Build for Linux"
	@echo "  make build-windows - Build for Windows"
	@echo "  make build-mac     - Build for macOS"
	@echo "  make dev           - Run in development mode"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make cross-windows - Cross-compile for Windows"
