package output

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

// DefaultHTMLPreviewPath returns a sibling .html path for an .xlsx path.
// If xlsxPath doesn't end with an extension, it appends ".html".
func DefaultHTMLPreviewPath(xlsxPath string) string {
	ext := filepath.Ext(xlsxPath)
	if ext == "" {
		return xlsxPath + ".html"
	}
	return strings.TrimSuffix(xlsxPath, ext) + ".html"
}

// WriteHTMLPreview writes an HTML "print screen" style preview of the same final layout as the XLSX.
// It does not require Excel to view: open the generated .html in any browser.
func WriteHTMLPreview(filePath string, data *OutputData) (err error) {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create preview file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close preview file: %w", closeErr)
		}
	}()

	// Build group titles (only including groups that have items)
	groupTitles := make([]string, 0, len(data.TitleOrder)+1)
	for _, title := range data.TitleOrder {
		if len(data.Grouped[title]) > 0 {
			groupTitles = append(groupTitles, title)
		}
	}
	if len(data.Grouped["unassigned"]) > 0 {
		groupTitles = append(groupTitles, "unassigned")
	}

	// Calculate columns needed for each group and max rows (matches XLSX writer)
	groupColumns := make([]int, len(groupTitles))
	maxRows := 0
	groupWidthCh := make([]int, len(groupTitles))
	for i, title := range groupTitles {
		locs := data.Grouped[title]
		cols := data.ColumnsNeeded(len(locs))
		groupColumns[i] = cols

		// With spillover, max rows is capped at MaxRows (or actual count if less)
		rowsForGroup := len(locs)
		if data.MaxRows > 0 && rowsForGroup > data.MaxRows {
			rowsForGroup = data.MaxRows
		}
		if rowsForGroup > maxRows {
			maxRows = rowsForGroup
		}

		// Match XLSX column width logic (character-based, with padding, min 15)
		maxDataWidth := len(title)
		for _, loc := range locs {
			if len(loc) > maxDataWidth {
				maxDataWidth = len(loc)
			}
		}
		width := maxDataWidth + 2
		if width < 15 {
			width = 15
		}
		groupWidthCh[i] = width
	}

	var b bytes.Buffer
	b.WriteString("<!doctype html>\n")
	b.WriteString("<html lang=\"en\">\n<head>\n")
	b.WriteString("  <meta charset=\"utf-8\" />\n")
	b.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />\n")
	b.WriteString("  <title>Location Priorities Preview</title>\n")
	b.WriteString("  <style>\n")
	b.WriteString("    :root { --grid:#d0d0d0; --priority:#ffff00; }\n")
	b.WriteString("    body { margin: 16px; font-family: Calibri, Arial, sans-serif; color: #111; }\n")
	b.WriteString("    .legend { margin: 0 0 12px 0; font-size: 13px; }\n")
	b.WriteString("    .legend .swatch { display:inline-block; width: 14px; height: 14px; background: var(--priority); border: 1px solid var(--grid); vertical-align: -2px; margin-right: 6px; }\n")
	b.WriteString("    table.sheet { border-collapse: collapse; border-spacing: 0; }\n")
	b.WriteString("    table.sheet th, table.sheet td { border: 1px solid var(--grid); padding: 4px 6px; font-size: 12px; line-height: 1.2; white-space: nowrap; }\n")
	b.WriteString("    table.sheet th { font-weight: 700; text-align: center; }\n")
	b.WriteString("    td.priority { background: var(--priority); }\n")
	b.WriteString("    td.gap, th.gap { border: 0; padding: 0; }\n")
	b.WriteString("    @media print { body { margin: 0.5in; } }\n")
	b.WriteString("  </style>\n")
	b.WriteString("</head>\n<body>\n")
	b.WriteString("  <div class=\"legend\"><span class=\"swatch\"></span>Priority location (QA_HOLD_PICKING)</div>\n")

	b.WriteString("  <table class=\"sheet\">\n")
	// Colgroup for consistent widths
	b.WriteString("    <colgroup>\n")
	for i := range groupTitles {
		for c := 0; c < groupColumns[i]; c++ {
			fmt.Fprintf(&b, "      <col style=\"width:%dch\" />\n", groupWidthCh[i])
		}
		if i < len(groupTitles)-1 {
			for g := 0; g < data.ColumnGap; g++ {
				b.WriteString("      <col style=\"width:2ch\" />\n")
			}
		}
	}
	b.WriteString("    </colgroup>\n")

	// Header row
	b.WriteString("    <thead>\n")
	b.WriteString("      <tr>\n")
	for i, title := range groupTitles {
		fmt.Fprintf(&b, "        <th colspan=\"%d\">%s</th>\n", groupColumns[i], html.EscapeString(title))
		if i < len(groupTitles)-1 && data.ColumnGap > 0 {
			fmt.Fprintf(&b, "        <th class=\"gap\" colspan=\"%d\"></th>\n", data.ColumnGap)
		}
	}
	b.WriteString("      </tr>\n")
	b.WriteString("    </thead>\n")

	// Body rows
	b.WriteString("    <tbody>\n")
	for row := 0; row < maxRows; row++ {
		b.WriteString("      <tr>\n")
		for i, title := range groupTitles {
			locs := data.Grouped[title]
			cols := groupColumns[i]

			for c := 0; c < cols; c++ {
				idx := row
				if data.MaxRows > 0 {
					idx = c*data.MaxRows + row
				}

				val := ""
				if idx < len(locs) {
					val = locs[idx]
				}

				classAttr := ""
				if val != "" && data.IsPriority(val) {
					classAttr = " class=\"priority\""
				}

				fmt.Fprintf(&b, "        <td%s>%s</td>\n", classAttr, html.EscapeString(val))
			}

			if i < len(groupTitles)-1 {
				for g := 0; g < data.ColumnGap; g++ {
					b.WriteString("        <td class=\"gap\"></td>\n")
				}
			}
		}
		b.WriteString("      </tr>\n")
	}
	b.WriteString("    </tbody>\n")

	b.WriteString("  </table>\n")
	b.WriteString("</body>\n</html>\n")

	if _, err := f.Write(b.Bytes()); err != nil {
		return fmt.Errorf("failed to write preview file: %w", err)
	}
	return nil
}
