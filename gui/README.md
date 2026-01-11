# QA Script GUI

A cross-platform desktop GUI for the QA Script file processor, built with [Wails](https://wails.io/).

## Features

- **Drag & Drop Interface**: Simply drag your files onto the designated areas
- **File Validation**: Automatically validates that files contain required columns
- **Cross-Platform**: Works on Windows (primary), macOS, and Linux

## Screenshots

The application provides two drop zones:

1. **Current Locations** (CSV) - Drop your CSV file here
   - Required column: `Location`
   
2. **Priorities** (Excel) - Drop your Excel file here
   - Required columns: `Container Tag`, `Current Location`

## Building

### Prerequisites

1. **Go 1.21+** - https://go.dev/dl/
2. **Wails CLI** - Install with: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
3. **Platform-specific dependencies**:

#### Windows
- WebView2 Runtime (usually pre-installed on Windows 10/11)

#### macOS
- Xcode Command Line Tools: `xcode-select --install`

#### Linux (Ubuntu/Debian)
```bash
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

### Build Commands

```bash
# Build for current platform
make build

# Platform-specific builds
make build-windows   # Windows
make build-linux     # Linux
make build-mac       # macOS

# Development mode with hot-reload
make dev
```

### Manual Build

```bash
# Windows
wails build

# Linux (Ubuntu 22.04+)
wails build -tags webkit2_41

# macOS
wails build
```

## Usage

1. Run the built executable (`qa-script-gui` or `qa-script-gui.exe`)
2. Drag your CSV file onto the "Current Locations" drop zone
3. Drag your Excel file onto the "Priorities" drop zone (optional)
4. Click **OK** to proceed or **Cancel** to exit

The application validates files before accepting them:
- CSV files must contain a `Location` column
- Excel files must contain `Container Tag` and `Current Location` columns

## Project Structure

```
gui/
├── app.go           # Go backend with file validation
├── main.go          # Wails application entry point
├── wails.json       # Wails configuration
├── go.mod           # Go module dependencies
├── Makefile         # Build commands
├── frontend/
│   ├── index.html   # UI structure
│   ├── style.css    # Styling
│   └── app.js       # Frontend logic
└── build/
    └── bin/         # Built executables
```

## Development

For development with hot-reload:

```bash
make dev
```

This will start the application with live reloading of frontend changes.
