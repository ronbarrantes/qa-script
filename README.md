# QA Script

A Go executable that parses CSV and Excel files and combines them into a single output Excel file based on a configurable YAML template.

## Features

- Parse CSV files with configurable delimiters
- Parse Excel files with sheet selection
- Combine data from both sources into a single Excel output
- YAML-based template for defining the output format
- Automatic template generation on first run

## Installation

```bash
go build -o qa-script .
```

## Usage

### Generate Template

On first run (or explicitly), the tool generates a YAML template file:

```bash
# Explicit template generation
./qa-script -generate-template

# Or just run without a template - it will create one automatically
./qa-script
```

### Process Files

After configuring the template:

```bash
./qa-script -csv input.csv -excel input.xlsx -output output.xlsx
```

### Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-csv` | Path to the CSV file | (required) |
| `-excel` | Path to the Excel file | (required) |
| `-output` | Path for the output Excel file | `output.xlsx` |
| `-template` | Path to the YAML template file | `template.yaml` |
| `-generate-template` | Generate a YAML template file | `false` |

## Template Configuration

The YAML template defines how data is combined. Example:

```yaml
version: "1.0"

csv:
  columns: []      # Empty = include all columns
  skip_rows: 0     # Rows to skip at start
  delimiter: ','   # CSV delimiter

excel:
  columns: []      # Empty = include all columns
  sheet: ""        # Empty = use first sheet
  skip_rows: 0     # Rows to skip at start

output:
  sheet_name: Combined
  column_mapping:
    - output_name: ID
      source: csv
      source_column: id
    - output_name: Name
      source: excel
      source_column: name
      default: N/A
    - output_name: Value
      source: csv
      source_column: value
      default: "0"

excel_sheet: ""    # Sheet to read from input Excel
```

### Column Mapping

Each entry in `column_mapping` defines:
- `output_name`: Column header in the output file
- `source`: Where to get data (`csv` or `excel`)
- `source_column`: Column name in the source file
- `default`: Optional default value if source is empty

## Project Structure

```
qa-script/
├── main.go           # Entry point
├── config/
│   └── config.go     # YAML template handling
├── parser/
│   ├── csv.go        # CSV parsing
│   └── excel.go      # Excel parsing
├── merger/
│   └── merger.go     # Data merging and output
├── go.mod
├── go.sum
└── README.md
```

## Dependencies

- [excelize](https://github.com/xuri/excelize) - Excel file handling
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing
