package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
		if err := config.GenerateTemplate(*templateFile); err != nil {
			log.Fatalf("Error generating template: %v", err)
		}
		fmt.Printf("Template generated successfully: %s\n", *templateFile)
		return
	}

	// Check if template exists, if not create it
	if _, err := os.Stat(*templateFile); os.IsNotExist(err) {
		fmt.Println("Template file not found. Generating default template...")
		if err := config.GenerateTemplate(*templateFile); err != nil {
			log.Fatalf("Error generating template: %v", err)
		}
		fmt.Printf("Template generated: %s\n", *templateFile)
		fmt.Println("Please edit the template file and run again.")
		return
	}

	// Load the template configuration
	cfg, err := config.LoadConfig(*templateFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
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

	// Parse Excel file
	fmt.Printf("Parsing Excel file: %s\n", *excelFile)
	excelData, err := parser.ParseExcel(*excelFile, cfg.ExcelSheet)
	if err != nil {
		log.Fatalf("Error parsing Excel: %v", err)
	}
	fmt.Printf("Excel parsed successfully: %d rows\n", len(excelData.Rows))

	// Merge and write output
	fmt.Printf("Merging data and writing to: %s\n", *outputFile)
	if err := merger.MergeAndWrite(csvData, excelData, *outputFile, cfg); err != nil {
		log.Fatalf("Error merging data: %v", err)
	}

	fmt.Println("Done!")
}
