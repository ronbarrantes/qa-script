# QA Script

A Go tool that reads location data from a CSV file, groups locations based on configurable rules, and outputs a formatted Excel file. Optionally, locations matching `QA_HOLD_PICKING` status from an input Excel file can be highlighted in yellow.

## Download

**For most users:** Download the **GUI version** (`qa-script-gui`) - it's a self-contained desktop application with all processing built-in. Just download, run, and drag-and-drop your files. See [gui/README.md](gui/README.md) for details.

**For advanced users/automation:** The CLI version (`qa-script`) is available for command-line usage and scripting.

## Features

- Parse CSV files to extract location codes
- Group locations by configurable letter prefixes or exact codes
- Automatic column spillover when groups exceed a size limit
- **Optional** Excel input to highlight locations with `QA_HOLD_PICKING` container tags
- Validation: Excel locations must be a subset of CSV locations
- Auto-generate YAML templates from CSV data
- Timestamped output files (`p1_<timestamp>.xlsx`)

## Installation

```bash
go build -o qa-script .
```

## Usage

### Basic Usage (CSV only)

```bash
./qa-script -csv locations.csv
```

### With Excel Highlighting

```bash
./qa-script -csv locations.csv -excel containers.xlsx
```

Output will be saved as `p1_20260111_143022.xlsx` (with current timestamp).

### Generate Template

Generate a YAML template seeded with location codes from your CSV:

```bash
./qa-script -generate-template -csv locations.csv
```

Or generate an empty template:

```bash
./qa-script -generate-template
```

### Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-csv` | Path to the CSV file (must have a "Location" column) | (required) |
| `-excel` | Path to the Excel file (optional, for highlighting) | (none) |
| `-output` | Path for the output Excel file | `p1_<timestamp>.xlsx` |
| `-template` | Path to the YAML template file | `template.yaml` |
| `-generate-template` | Generate a YAML template file | `false` |

## Configuration

The YAML configuration defines how locations are grouped in the output Excel.

### Example Configuration

```yaml
# Groups define how locations appear in the output
groups:
  - name: pallets
    values: a, b, c, lud, prm, slp

  - name: e - g
    values: e, f, g, gft, hwk, hvc

  - name: h - k
    values: h, j, k

  - name: lm
    values: l, m

  - name: n - t
    values: n, s, t, mez

# Maximum rows per column before spilling to next column
size: 20
```

### Configuration Options

| Field | Description | Default |
|-------|-------------|---------|
| `groups` | List of group definitions | (required) |
| `groups[].name` | Header name shown in output Excel | - |
| `groups[].values` | Comma-separated location codes or single letters | - |
| `size` | Max rows per column before creating a new column | `20` |

### Location Matching Rules

Location codes follow the format `PREFIX:CODE` (e.g., `SS4:GF225.B`). The `PREFIX:` portion is ignored for grouping; only the `CODE` part is used.

**Value matching:**

- **Single letter** (e.g., `a`): Matches any location code starting with that letter
  - `a` matches `AB215`, `AC100`, etc.
- **Multi-letter** (e.g., `lud`, `gft`): Exact match for the full letter prefix
  - `lud` matches `LUD86` but not `LU123`
  - `gft` matches `GFT31` but not `GF225`

### Output Format

- Each group becomes one or more columns in the output Excel
- If a group has more items than `size`, it spills into additional columns
- Groups are separated by blank columns
- Headers repeat for each spilled column
- Locations matching `QA_HOLD_PICKING` from the input Excel are highlighted yellow (if Excel provided)
- Unassigned locations (not matching any group) appear in an "unassigned" column at the end

**Example output with `size: 3`:**

```
| pallets |   | e - g | e - g |   | h - k |
|---------|---|-------|-------|---|-------|
| AB215   |   | EN15  | GF253 |   | HA101 |
| LUD86   |   | EN333 | GC111 |   | JB202 |
| CS121   |   | GFT31 |       |   |       |
```

## Input File Requirements

### CSV File (Required)

Must contain a `Location` column with location codes to be grouped.

### Excel File (Optional)

If provided, must contain:
- `Container Tag` column - used to identify `QA_HOLD_PICKING` items
- `Current Location` column - locations with `QA_HOLD_PICKING` tags will be highlighted

**Validation:** All `Current Location` values in the Excel file must exist in the CSV's `Location` column. If any Excel location is not found in the CSV, the tool will exit with an error listing the invalid locations.

The Excel file can be:
- Omitted entirely (just use `-csv`)
- Empty (has headers but no data rows)

In either case, the tool will still group and sort locations from the CSV, just without any highlighting.

## Project Structure

```
qa-script/
├── main.go           # Entry point and CLI handling
├── config/
│   ├── config.go     # YAML configuration loading/saving
│   └── config_test.go
├── parser/
│   ├── csv.go        # CSV file parsing
│   └── excel.go      # Excel file parsing
├── merger/
│   └── merger.go     # Grouping logic and Excel output
├── rules/
│   ├── rules.go      # Location parsing and matching rules
│   └── rules_test.go
├── go.mod
├── go.sum
└── README.md
```

## Dependencies

- [excelize](https://github.com/xuri/excelize) - Excel file handling
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing
