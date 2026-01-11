package merger

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"

	"qa-script/config"
)

var locationCodeRe = regexp.MustCompile(`:((?:[ABCEFGHJKLMNST][A-Z])|PRM|LUD|SLP|MEZ|GFT|HVC|HWK)\d{1,3}`)

func normalizeLocation(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// ExtractLocationCode returns the location code portion used for grouping (e.g. "HVC" from ":HVC12").
func ExtractLocationCode(location string) (string, bool) {
	m := locationCodeRe.FindStringSubmatch(strings.ToUpper(location))
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

// BuildDefaultGroupsFromLocations creates one group per discovered code, useful for template generation.
func BuildDefaultGroupsFromLocations(locations []string) []config.Group {
	seen := map[string]struct{}{}
	var codes []string
	for _, loc := range locations {
		if code, ok := ExtractLocationCode(loc); ok {
			if _, exists := seen[code]; !exists {
				seen[code] = struct{}{}
				codes = append(codes, code)
			}
		}
	}
	sort.Strings(codes)

	groups := make([]config.Group, 0, len(codes))
	for _, code := range codes {
		groups = append(groups, config.Group{Name: strings.ToLower(code), Values: []string{code}})
	}
	return groups
}

// WriteGroupedExcel writes a grouped output Excel:
// - locations are placed into columns by group, spilling after cfg.Size rows
// - any written cell that matches highlightLocations (case-insensitive) is filled yellow
func WriteGroupedExcel(outputPath string, cfg *config.Config, locations []string, highlightLocations map[string]struct{}) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	size := cfg.Size
	if size <= 0 {
		size = 20
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := cfg.OutputSheet
	if sheetName == "" {
		sheetName = "Groups"
	}
	f.SetSheetName("Sheet1", sheetName)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#EDEDED"}, Pattern: 1},
	})
	yellowStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFFF00"}, Pattern: 1},
	})

	// Precompute group membership values (supports both "A" prefix match and exact code match).
	groupValues := make([][]string, len(cfg.Groups))
	for i, g := range cfg.Groups {
		out := make([]string, 0, len(g.Values))
		for _, v := range g.Values {
			vn := strings.ToUpper(strings.TrimSpace(v))
			if vn != "" {
				out = append(out, vn)
			}
		}
		groupValues[i] = out
	}

	matchesGroup := func(groupIdx int, code string) bool {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			return false
		}
		for _, gv := range groupValues[groupIdx] {
			// Single-letter group value means "prefix match" (e.g. "A" matches "AB", "AN", etc).
			if len(gv) == 1 && gv[0] >= 'A' && gv[0] <= 'Z' {
				if strings.HasPrefix(code, gv) {
					return true
				}
				continue
			}
			// Otherwise, match exact extracted code (e.g. "HVC").
			if code == gv {
				return true
			}
		}
		return false
	}

	used := make([]bool, len(locations))

	colStart := 1
	for gi, g := range cfg.Groups {
		var groupLocs []string
		for i, loc := range locations {
			code, ok := ExtractLocationCode(loc)
			if !ok {
				continue
			}
			if matchesGroup(gi, code) {
				groupLocs = append(groupLocs, loc)
				used[i] = true
			}
		}

		colsUsed := 1
		if len(groupLocs) > 0 {
			colsUsed = int(math.Ceil(float64(len(groupLocs)) / float64(size)))
			if colsUsed < 1 {
				colsUsed = 1
			}
		}

		// Headers (repeat per spilled column).
		for c := 0; c < colsUsed; c++ {
			cell, _ := excelize.CoordinatesToCellName(colStart+c, 1)
			f.SetCellValue(sheetName, cell, g.Name)
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}

		// Values.
		for i, loc := range groupLocs {
			c := colStart + (i / size)
			r := 2 + (i % size)
			cell, _ := excelize.CoordinatesToCellName(c, r)
			f.SetCellValue(sheetName, cell, loc)
			if _, ok := highlightLocations[normalizeLocation(loc)]; ok {
				f.SetCellStyle(sheetName, cell, cell, yellowStyle)
			}
		}

		// Add a blank separator column after each group.
		colStart += colsUsed + 1
	}

	// Append any locations not covered by configured groups.
	var unmatched []string
	for i, loc := range locations {
		if used[i] {
			continue
		}
		unmatched = append(unmatched, loc)
	}
	if len(unmatched) > 0 {
		colsUsed := int(math.Ceil(float64(len(unmatched)) / float64(size)))
		if colsUsed < 1 {
			colsUsed = 1
		}
		for c := 0; c < colsUsed; c++ {
			cell, _ := excelize.CoordinatesToCellName(colStart+c, 1)
			f.SetCellValue(sheetName, cell, "unmatched")
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
		for i, loc := range unmatched {
			c := colStart + (i / size)
			r := 2 + (i % size)
			cell, _ := excelize.CoordinatesToCellName(c, r)
			f.SetCellValue(sheetName, cell, loc)
			if _, ok := highlightLocations[normalizeLocation(loc)]; ok {
				f.SetCellStyle(sheetName, cell, cell, yellowStyle)
			}
		}
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save output file: %w", err)
	}
	return nil
}
