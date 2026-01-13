# QA Location Grouper

A desktop tool that groups warehouse location codes and generates formatted Excel reports.

## Quick Start (GUI)

1. **Download** the app for your platform from [Releases](../../releases)
2. **Launch** the application
3. **Drop your files** into the two zones:
   - **Locations CSV** — Your CSV file with a `Location` column
   - **Priorities XLSX** — Excel file with `Container Tag` and `Current Location` columns
4. **Click OK** to generate the output

The output file (`locations_output.xlsx`) is saved in the same folder as your CSV file.

![GUI Screenshot](cmd/gui/build/appicon.png)

## What It Does

- **Groups locations** by configurable letter prefixes (e.g., all "A" locations together)
- **Highlights priority items** (QA_HOLD_PICKING) in yellow
- **Handles spillover** — columns wrap when they exceed the size limit
- **Sorts intelligently** — alphabetically by letters, then numerically

## Configuration

On first run, a `rules.yaml` file is created in your CSV folder:

```yaml
groups:
  - title: pallets
    values: [a, b, c, d]
  - title: rack
    values: [e, f, g, gft]
  - title: bulk
    values: [h, lud]

size: 20  # Max rows per column before spillover
gap: 1    # Empty columns between groups
```

### Matching Rules

| Value | Matches |
|-------|---------|
| `a` | Any code starting with A (AB215, AC100) |
| `gft` | Only exact GFT codes (GFT31, GFT245) |
| `lud` | Only exact LUD codes (LUD86) |

## CLI Usage

For automation or scripting:

```bash
qa-cli -locations data.csv -priorities priorities.xlsx
```

Options:
- `-csv output.csv` — CSV output path (default: output.csv)
- `-xlsx output.xlsx` — Excel output path (default: output.xlsx)
- `-rules-dir .` — Directory for rules.yaml
- `-verbose` — Show detailed output

## Building from Source

### Requirements
- Go 1.21+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) (for GUI)

### Linux Dependencies (Ubuntu/Debian)

On Linux, you need GTK3 and WebKit2GTK development libraries:

```bash
# Ubuntu 24.04+ (uses webkit2gtk-4.1)
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev

# Ubuntu 22.04 or older (uses webkit2gtk-4.0)
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev
```

### Build Commands

```bash
# CLI only
make build-cli

# GUI (current platform)
make build-gui

# GUI for Linux specifically (includes webkit2gtk-4.1 support)
make build-gui-linux

# Run tests
make test
```

### Running the GUI

After building, the binary is located at:
```bash
./cmd/gui/build/bin/qa-gui
```

## License

MIT
