.PHONY: build build-cli build-gui run-cli run-gui test clean dev

# Default target
all: build

# Build both CLI and GUI
build: build-cli build-gui

# Build CLI only
build-cli:
	@echo "Building CLI..."
	@mkdir -p bin
	go build -o bin/qa-cli ./cmd/cli/
	@echo "✓ CLI built: bin/qa-cli"

# Build GUI only (requires wails)
build-gui:
	@echo "Building GUI..."
	cd cmd/gui && wails build
	@echo "✓ GUI built: cmd/gui/build/bin/"

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
