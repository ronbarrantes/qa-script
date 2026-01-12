# QA Script - Project Summary

This document summarizes everything that was built in the QA Script project, including tools used, patterns implemented, and architecture decisions.

## Project Overview

**QA Script** is a location grouping and Excel generation tool with both CLI and GUI implementations. It processes CSV files containing location codes, groups them according to configurable rules, and outputs formatted Excel files. Optional Excel input can be used to highlight specific locations (QA_HOLD_PICKING status).

## Features Implemented

### Core Functionality
- **CSV Parsing**: Reads CSV files with a required "Location" column
- **Excel Parsing**: Reads Excel files with "Container Tag" and "Current Location" columns
- **Location Grouping**: Groups locations by configurable rules (prefix matching and exact matching)
- **Excel Output Generation**: Creates formatted Excel files with grouped locations in columns
- **Highlighting**: Highlights locations with QA_HOLD_PICKING status in yellow
- **Column Spillover**: Automatically creates new columns when groups exceed size limit
- **Template Generation**: Auto-generates YAML configuration templates from CSV data
- **Validation**: Validates that Excel locations are a subset of CSV locations

### CLI Tool
- Command-line interface with flags for CSV, Excel, output, and template files
- Template generation mode (`-generate-template`)
- Auto-template creation if template doesn't exist
- Timestamped output files

### GUI Tool
- Cross-platform desktop application built with Wails
- Drag-and-drop file interface
- Real-time file validation with visual feedback
- Browse button for file selection
- Dark theme UI with animations
- Self-contained (all processing logic built-in)

## Architecture

### Project Structure

```
qa-script/
├── main.go                    # CLI entry point
├── config/                    # Configuration management
│   ├── config.go             # YAML config loading/saving
│   ├── config_test.go        # Unit tests
│   └── example.yaml          # Example configuration
├── parser/                    # File parsing
│   ├── csv.go                # CSV parsing
│   └── excel.go              # Excel parsing
├── merger/                    # Grouping and Excel output
│   └── merger.go             # Grouping logic and Excel writing
├── rules/                     # Location parsing and matching rules
│   ├── rules.go              # Location parsing, sorting, matching
│   └── rules_test.go         # Unit tests
└── gui/                       # GUI application
    ├── main.go               # Wails entry point
    ├── app.go                # Backend application logic
    ├── wails.json            # Wails configuration
    ├── go.mod                # GUI module (depends on parent)
    ├── Makefile              # Build commands
    ├── frontend/
    │   ├── index.html        # UI structure
    │   ├── style.css         # Styling (dark theme)
    │   └── app.js            # Frontend logic
    └── build/                # Build artifacts
```

### Package Organization

**config/** - Configuration Management
- Handles YAML configuration loading and saving
- Custom YAML unmarshaling for flexible value formats (comma-separated strings or arrays)
- Template generation functionality

**parser/** - File Parsing
- CSV parsing using Go's `encoding/csv` package
- Excel parsing using `excelize` library
- Common interface: `CSVData` and `ExcelData` structs with similar methods
- Column lookup and value extraction utilities

**rules/** - Location Parsing and Matching
- Location code parsing with regex patterns
- Handles format: `PREFIX:CODE` (e.g., "SS4:GF225.B")
- Separates prefix (ignored) from code (used for matching)
- Matching rules:
  - Single letter: prefix match (e.g., "G" matches "GF", "GA")
  - Multi-letter: exact match (e.g., "GFT" matches only "GFT")
- Sorting logic (alphabetical by letters, numerical by numbers, then by suffix)

**merger/** - Grouping and Excel Output
- Groups locations according to configuration rules
- Excel output generation with formatting
- Column spillover logic (creates new columns when size limit exceeded)
- Highlighting support (yellow fill for QA_HOLD_PICKING locations)
- Unassigned locations handling (separate column at end)

**gui/** - Desktop Application
- Wails v2 framework for cross-platform desktop apps
- Go backend with context management
- Vanilla JavaScript frontend (no frameworks)
- Embedded frontend assets using `//go:embed`
- File validation with real-time feedback
- Drag-and-drop support with native file dialogs

## Tools and Technologies

### Programming Languages
- **Go 1.21+** (CLI) / **Go 1.22+** (GUI) - Backend logic
- **JavaScript** (Vanilla) - Frontend logic
- **HTML/CSS** - Frontend UI

### Libraries and Frameworks

#### CLI Dependencies
- `github.com/xuri/excelize/v2` (v2.8.1) - Excel file handling
- `gopkg.in/yaml.v3` (v3.0.1) - YAML configuration parsing

#### GUI Dependencies
- `github.com/wailsapp/wails/v2` (v2.11.0) - Cross-platform desktop framework
- `github.com/xuri/excelize/v2` (v2.8.1) - Excel file handling
- `qa-script` (local module) - Shared CLI logic via Go module replace

### Build Tools
- **Go Modules** - Dependency management
- **Wails CLI** - GUI build tool
- **Makefile** (GUI) - Build automation with platform detection

## Design Patterns and Architecture Decisions

### 1. Package Structure
- **Separation of Concerns**: Clear package boundaries (config, parser, rules, merger)
- **Domain-Driven Design**: Rules package encapsulates location parsing logic
- **Shared Code Reuse**: GUI uses CLI packages via Go module replace

### 2. Configuration Pattern
- **YAML Configuration**: Human-readable, version-controllable configuration
- **Flexible Input Format**: Supports both comma-separated strings and YAML arrays
- **Custom Unmarshaling**: `StringList` type with custom YAML unmarshaling
- **Template Generation**: Auto-generate templates from data

### 3. Parsing Pattern
- **Common Interface**: Similar methods for CSVData and ExcelData (`GetColumnIndex`, `GetColumnValues`)
- **Header-Based Access**: Column lookup by name rather than position
- **Error Handling**: Early validation with descriptive error messages

### 4. Location Matching Pattern
- **Parser-Separated Logic**: Location parsing isolated in rules package
- **Type-Based Matching**: Distinguishes between 2-letter (standard) and 3-letter (matching) groups
- **Flexible Matching**: Supports both prefix and exact matching

### 5. Excel Generation Pattern
- **Style Reuse**: Pre-defined styles (header, yellow highlight)
- **Column-Based Layout**: Groups become columns with spillover
- **Separator Columns**: Blank columns between groups for readability

### 6. GUI Pattern
- **Wails Architecture**: Go backend + web frontend
- **Context-Based State**: App struct stores context and file paths
- **Validation Pattern**: Frontend calls backend validation, displays results
- **Event-Driven UI**: Drag-and-drop events trigger validation
- **State Management**: Simple JavaScript state object for UI state

### 7. Testing Pattern
- **Table-Driven Tests**: Comprehensive test cases in test files
- **Temp File Testing**: Uses `t.TempDir()` for file-based tests
- **Edge Case Coverage**: Tests for various input formats and edge cases

## Configuration Format

### YAML Structure

```yaml
groups:
  - name: <group name>
    values: <comma-separated codes or YAML array>
  - name: <another group>
    values: <codes>

size: <max rows per column>
```

### Value Matching Rules

- **Single letter** (e.g., `a`): Matches any location code starting with that letter
  - Example: `a` matches `AB215`, `AC100`
- **Multi-letter** (e.g., `lud`, `gft`): Exact match for full letter prefix
  - Example: `lud` matches `LUD86` but not `LU123`
  - Example: `gft` matches `GFT31` but not `GF225`

### Location Format

Locations follow format: `PREFIX:CODE`
- `PREFIX:` (before colon) - Ignored for grouping/sorting
- `CODE` (after colon) - Used for grouping and sorting
- Example: `SS4:GF225.B` where `SS4:` is prefix, `GF225.B` is code

## Testing Strategy

### Unit Tests
- **config/config_test.go**: Tests YAML loading/saving, custom unmarshaling
- **rules/rules_test.go**: Tests location parsing, matching, sorting
- **Table-driven tests**: Multiple test cases per function
- **Edge case coverage**: Empty inputs, invalid formats, boundary conditions

### Test Patterns Used
- Temp directory for file operations
- Table-driven test structure
- Sub-tests for each test case
- Helper functions for complex comparisons

## Build System

### CLI Build
```bash
go build -o qa-script .
```

### GUI Build
- **Makefile** with platform detection
- **Platform-specific builds**: Linux, Windows, macOS
- **Development mode**: Hot-reload with `make dev`
- **Cross-compilation**: Windows from Linux/Mac

### Build Targets
- `make build` - Auto-detect platform
- `make build-linux` - Linux build (with webkit2gtk-4.1 tag)
- `make build-windows` - Windows build
- `make build-mac` - macOS build
- `make dev` - Development mode
- `make clean` - Clean artifacts

## GUI Design

### UI Theme
- **Dark Theme**: Dark background (#1b2636) with card-based layout
- **Color Scheme**:
  - Primary: Indigo (#4f46e5)
  - Success: Green (#10b981)
  - Error: Red (#ef4444)
  - Background cards: Dark blue-gray (#243447)

### UI Components
- **Drop Zones**: Two-column grid layout with drag-and-drop support
- **Validation Overlay**: Loading spinner during validation
- **File Info Display**: Shows filename, validation status, row/column counts
- **Buttons**: Cancel and OK with disabled states
- **Animations**: Shake on error, fade-in on success, hover effects

### Frontend Patterns
- **Vanilla JavaScript**: No frameworks, direct DOM manipulation
- **Event Listeners**: Drag-and-drop, click handlers
- **State Management**: Simple state object tracking CSV/Excel validation
- **Async/Await**: For backend method calls
- **Error Handling**: Try-catch with user-friendly error messages

### Drag-and-Drop Implementation
- Native HTML5 drag-and-drop API
- URI list handling for file paths
- Windows path normalization (file:// protocol handling)
- Visual feedback: drag-over states, validation overlays

## File Processing Flow

### CLI Flow
1. Parse command-line flags
2. Parse CSV file (extract locations)
3. If Excel provided: Parse Excel, validate locations, extract QA_HOLD_PICKING highlights
4. Load or generate template.yaml
5. Group locations according to config
6. Write Excel output with formatting and highlighting
7. Save timestamped output file

### GUI Flow
1. User drops/browses files
2. Frontend triggers validation (async call to backend)
3. Backend validates file structure and required columns
4. Validation result returned to frontend
5. Frontend updates UI (success/error states)
6. User clicks OK button
7. Backend processes files (same as CLI logic)
8. Success message displayed, app quits

## Key Algorithms

### Location Matching Algorithm
1. Parse location: Split on colon, extract code portion
2. Parse code: Extract letters (prefix), numbers, suffix using regex
3. Match against group value:
   - If single letter: Check if location letters start with that letter
   - If multi-letter: Check exact match with location letters

### Sorting Algorithm
1. Parse each location into LocationCode struct
2. Compare by letters (alphabetical)
3. If letters match, compare by numbers (numerical)
4. If numbers match, compare by suffix (alphabetical)
5. Preserve original location string (with prefix) in output

### Column Spillover Algorithm
1. Group locations according to config
2. For each group, calculate columns needed: `ceil(count / size)`
3. Write headers for each column
4. Distribute items across columns: `column = start + (index / size)`, `row = 2 + (index % size)`
5. Add blank separator column after group

## Error Handling

### Validation Strategy
- **Early Validation**: Validate file structure before processing
- **Descriptive Errors**: Clear error messages indicating what's wrong
- **Location Subset Check**: Excel locations must be subset of CSV locations
- **Column Existence Check**: Required columns must exist

### Error Patterns
- Log fatal errors and exit (CLI)
- Return error results to frontend (GUI)
- User-friendly error messages
- Invalid location lists for debugging

## Output Format

### Excel Structure
- **Sheet Name**: "Groups"
- **Headers**: Group names (bold, gray background) in row 1
- **Data**: Locations in rows 2+, distributed across columns
- **Spillover**: New columns created when size limit exceeded
- **Separators**: Blank columns between groups
- **Highlighting**: Yellow fill for QA_HOLD_PICKING locations
- **Unassigned**: Final column for unmatched locations

### Output File Naming
- CLI: `locations_<timestamp>.xlsx` (format: `060102_150405`)
- GUI: `locations_<timestamp>.xlsx` (same format, in CSV directory)

## Module Dependencies

### CLI Module (qa-script)
- Direct dependencies: excelize, yaml.v3
- No internal dependencies

### GUI Module (qa-script-gui)
- Wails framework
- excelize (same version as CLI)
- qa-script (parent module) via `replace` directive
- All CLI functionality available to GUI

## Platform Support

### CLI
- **All platforms**: Works on any platform with Go installed

### GUI
- **Windows**: Primary target, WebView2 runtime
- **macOS**: Native support, Xcode Command Line Tools required
- **Linux**: Ubuntu/Debian with webkit2gtk-4.1 (tested on Ubuntu 22.04+)

## Development Workflow

1. **CLI Development**: Standard Go development workflow
2. **GUI Development**: Wails dev mode with hot-reload
3. **Testing**: `go test ./...` for unit tests
4. **Building**: Makefile targets for different platforms
5. **Cross-compilation**: Possible for Windows from other platforms

## Code Quality Patterns

- **Error Wrapping**: `fmt.Errorf("...: %w", err)` for error context
- **Named Returns**: Used sparingly, mostly for clarity
- **Struct Methods**: OOP-style methods on data types
- **Package-Private**: Unexported functions for internal use
- **Documentation**: Package-level and function-level comments
- **Consistent Naming**: Clear, descriptive names

## Future Considerations (From Original TODO)

- GUI for Windows (✅ Completed)
- A good place to save the YAML configuration (✅ Auto-generated in CSV directory)
- Instructions on how to use (✅ GUI provides visual interface, README provides docs)

## Summary

This project demonstrates:
- **Dual Interface Design**: Both CLI and GUI for different use cases
- **Code Reuse**: GUI leverages CLI packages
- **Configuration-Driven**: Flexible YAML-based configuration
- **Cross-Platform**: Works on Windows, macOS, Linux
- **Modern Go Practices**: Modules, testing, error wrapping
- **Desktop App Development**: Wails framework usage
- **File Processing**: CSV/Excel parsing and generation
- **Complex Business Logic**: Location parsing, matching, and grouping rules
