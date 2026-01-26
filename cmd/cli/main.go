package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"qa-script/config"
	"qa-script/output"
	"qa-script/processor"
)

func main() {
	// Define flags
	locationsFile := flag.String("locations", "", "Path to locations CSV file (required)")
	prioritiesFile := flag.String("priorities", "", "Path to priorities XLSX file (required)")
	outputCSV := flag.String("csv", "output.csv", "Path for CSV output")
	outputXLSX := flag.String("xlsx", "output.xlsx", "Path for XLSX output")
	outputPreview := flag.String("preview", "", "Path for HTML preview output (optional; default is derived from --xlsx)")
	outputPNG := flag.String("png", "", "Path for PNG preview output (optional; default is derived from --xlsx)")
	noPreview := flag.Bool("no-preview", false, "Disable writing the HTML preview next to the XLSX")
	noPNG := flag.Bool("no-png", false, "Disable writing the PNG preview next to the XLSX")
	rulesDir := flag.String("rules-dir", ".", "Directory containing qa_loc_rules.yaml (will create default if not exists)")
	verbose := flag.Bool("verbose", false, "Show detailed output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "QA Script - Location Grouping Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -locations <file.csv> -priorities <file.xlsx> [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validate required flags
	if *locationsFile == "" || *prioritiesFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Check input files exist
	if _, err := os.Stat(*locationsFile); os.IsNotExist(err) {
		log.Fatalf("Locations file not found: %s", *locationsFile)
	}
	if _, err := os.Stat(*prioritiesFile); os.IsNotExist(err) {
		log.Fatalf("Priorities file not found: %s", *prioritiesFile)
	}

	// Ensure qa_loc_rules.yaml exists (create default if not)
	rulesPath, err := config.EnsureRulesFile(*rulesDir)
	if err != nil {
		log.Fatalf("Failed to ensure rules file: %v", err)
	}

	// Process the data
	result, err := processor.Process(*locationsFile, *prioritiesFile, rulesPath)
	if err != nil {
		log.Fatalf("Processing failed: %v", err)
	}

	if *verbose {
		printVerboseOutput(result)
	}

	// Create output data
	outputData := output.NewOutputData(
		result.TitleOrder,
		result.TitleGroupedLocations,
		result.PriorityLocations,
		result.ColumnGap,
		result.MaxRows,
	)

	// Write CSV output
	if err := output.WriteCSV(*outputCSV, outputData); err != nil {
		log.Fatalf("Failed to write CSV: %v", err)
	}
	absCSV, _ := filepath.Abs(*outputCSV)
	fmt.Printf("✓ CSV written to: %s\n", absCSV)

	// Write XLSX output
	if err := output.WriteXLSX(*outputXLSX, outputData); err != nil {
		log.Fatalf("Failed to write XLSX: %v", err)
	}
	absXLSX, _ := filepath.Abs(*outputXLSX)
	fmt.Printf("✓ XLSX written to: %s\n", absXLSX)

	// Write PNG preview output (best effort; requires Chrome/Chromium)
	if !*noPNG {
		pngPath := *outputPNG
		if pngPath == "" {
			pngPath = output.DefaultPNGPreviewPath(*outputXLSX)
		}
		if err := output.WritePNGPreview(pngPath, outputData); err != nil {
			if output.IsBrowserUnavailable(err) {
				fmt.Fprintf(os.Stderr, "! PNG preview not generated (%v). HTML preview may still be created.\n", err)
			} else {
				log.Fatalf("Failed to write PNG preview: %v", err)
			}
		} else {
			absPNG, _ := filepath.Abs(pngPath)
			fmt.Printf("✓ PNG preview written to: %s\n", absPNG)
		}
	}

	// Write HTML preview output (open in a browser, no Excel needed)
	if !*noPreview {
		previewPath := *outputPreview
		if previewPath == "" {
			previewPath = output.DefaultHTMLPreviewPath(*outputXLSX)
		}
		if err := output.WriteHTMLPreview(previewPath, outputData); err != nil {
			log.Fatalf("Failed to write HTML preview: %v", err)
		}
		absPreview, _ := filepath.Abs(previewPath)
		fmt.Printf("✓ Preview written to: %s\n", absPreview)
	}

	fmt.Printf("\nProcessed %d unique locations, %d priority locations\n",
		len(result.Locations), len(result.PriorityLocations))
}

func printVerboseOutput(result *processor.Result) {
	fmt.Println("=== Unique Locations ===")
	for i, loc := range result.Locations {
		fmt.Printf("%d: %s\n", i+1, loc)
	}
	fmt.Printf("\nTotal: %d unique locations (from %d CSV rows)\n\n",
		len(result.Locations), result.TotalCSVRows)

	fmt.Println("=== Priority Locations (QA_HOLD_PICKING) ===")
	for i, loc := range result.PriorityLocations {
		fmt.Printf("%d: %s\n", i+1, loc)
	}
	fmt.Printf("\nTotal: %d priority locations (from %d Excel rows)\n\n",
		len(result.PriorityLocations), result.TotalExcelRows)

	fmt.Println("=== Grouped by Title ===")
	for _, title := range result.TitleOrder {
		locs := result.TitleGroupedLocations[title]
		if len(locs) > 0 {
			fmt.Printf("[%s] %d locations\n", title, len(locs))
		}
	}
	if locs := result.TitleGroupedLocations["unassigned"]; len(locs) > 0 {
		fmt.Printf("[unassigned] %d locations\n", len(locs))
	}
	fmt.Println()
}
