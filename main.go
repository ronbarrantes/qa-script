package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"qa-script/config"
	"qa-script/merger"
	"qa-script/parser"
)

func main() {
	// Define command-line flags
	csvFile := flag.String("csv", "", "Path to the CSV file")
	excelFile := flag.String("excel", "", "Path to the Excel file")
	outputFile := flag.String("output", "output.xlsx", "Path for the output Excel file")
	templateFile := flag.String("template", "template.yaml", "Path to the YAML template file")
	generateTemplate := flag.Bool("generate-template", false, "Generate a YAML template file")

	flag.Parse()

	// If generate-template flag is set, create the template and exit
	if *generateTemplate {
		// If a CSV is provided, generate a template seeded with discovered location codes.
		if *csvFile != "" {
			csvData, err := parser.ParseCSV(*csvFile)
			if err != nil {
				log.Fatalf("Error parsing CSV for template generation: %v", err)
			}
			locations, err := csvData.GetUniqueColumnValues("Location")
			if err != nil {
				log.Fatalf("Error reading CSV Location column for template generation: %v", err)
			}
			cfg := &config.Config{
				Size:        20,
				OutputSheet: "Groups",
				Groups:      merger.BuildDefaultGroupsFromLocations(locations),
			}
			if err := config.SaveConfig(cfg, *templateFile); err != nil {
				log.Fatalf("Error writing template: %v", err)
			}
			fmt.Printf("Template generated successfully from CSV: %s\n", *templateFile)
			return
		}

		if err := config.GenerateTemplate(*templateFile); err != nil {
			log.Fatalf("Error generating template: %v", err)
		}
		fmt.Printf("Template generated successfully: %s\n", *templateFile)
		return
	}

	// Validate required files
	if *csvFile == "" || *excelFile == "" {
		fmt.Println("Usage: qa-script -csv <csv_file> -excel <excel_file> [-output <output_file>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse CSV file
	fmt.Printf("Parsing CSV file: %s\n", *csvFile)
	csvData, err := parser.ParseCSV(*csvFile)
	if err != nil {
		log.Fatalf("Error parsing CSV: %v", err)
	}
	fmt.Printf("CSV parsed successfully: %d rows\n", len(csvData.Rows))

	locations, err := csvData.GetUniqueColumnValues("Location")
	if err != nil {
		log.Fatalf("Error reading CSV Location column: %v", err)
	}
	fmt.Printf("CSV unique locations: %d\n", len(locations))

	// Check if template exists, if not create it seeded from CSV.
	if _, err := os.Stat(*templateFile); os.IsNotExist(err) {
		fmt.Println("Template file not found. Generating template from CSV location codes...")
		genCfg := &config.Config{
			Size:        20,
			OutputSheet: "Groups",
			Groups:      merger.BuildDefaultGroupsFromLocations(locations),
		}
		if err := config.SaveConfig(genCfg, *templateFile); err != nil {
			log.Fatalf("Error generating template: %v", err)
		}
		fmt.Printf("Template generated: %s\n", *templateFile)
	}

	// Load the template configuration
	cfg, err := config.LoadConfig(*templateFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Parse Excel file
	fmt.Printf("Parsing Excel file: %s\n", *excelFile)
	excelData, err := parser.ParseExcel(*excelFile, cfg.ExcelSheet)
	if err != nil {
		log.Fatalf("Error parsing Excel: %v", err)
	}
	fmt.Printf("Excel parsed successfully: %d rows\n", len(excelData.Rows))

	// Build highlight set from Excel:
	// - filter rows where `Container Tag` == `QA_HOLD_PICKING`
	// - collect `Current Location`
	tagIdx := excelData.GetColumnIndex("Container Tag")
	locIdx := excelData.GetColumnIndex("Current Location")
	if tagIdx == -1 {
		log.Fatalf("Excel column %q not found", "Container Tag")
	}
	if locIdx == -1 {
		log.Fatalf("Excel column %q not found", "Current Location")
	}

	highlight := make(map[string]struct{})
	for _, row := range excelData.Rows {
		if tagIdx >= len(row) {
			continue
		}
		if strings.TrimSpace(row[tagIdx]) != "QA_HOLD_PICKING" {
			continue
		}
		if locIdx < len(row) {
			v := strings.ToUpper(strings.TrimSpace(row[locIdx]))
			if v != "" {
				highlight[v] = struct{}{}
			}
		}
	}
	fmt.Printf("Excel highlight locations (QA_HOLD_PICKING): %d\n", len(highlight))

	// Write grouped output
	fmt.Printf("Writing grouped output to: %s\n", *outputFile)
	if err := merger.WriteGroupedExcel(*outputFile, cfg, locations, highlight); err != nil {
		log.Fatalf("Error writing grouped Excel: %v", err)
	}

	fmt.Println("Done!")
}
