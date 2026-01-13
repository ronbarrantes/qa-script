// Package main is a development entry point for testing.
// For production use, build the CLI or GUI:
//   - CLI: go build -o bin/qa-cli ./cmd/cli/
//   - GUI: cd cmd/gui && wails build
package main

import (
	"log"

	"qa-script/config"
	"qa-script/output"
	"qa-script/processor"
)

func main() {
	rulesPath, err := config.EnsureRulesFile(".")
	if err != nil {
		log.Fatal(err)
	}

	result, err := processor.Process(
		"mock_data/TEST_Locations.csv",
		"mock_data/TEST_Priorities.xlsx",
		rulesPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	outputData := output.NewOutputData(
		result.TitleOrder,
		result.TitleGroupedLocations,
		result.PriorityLocations,
		result.Gap,
		result.Size,
	)

	if err := output.WriteCSV("output.csv", outputData); err != nil {
		log.Fatal(err)
	}
	if err := output.WriteXLSX("output.xlsx", outputData); err != nil {
		log.Fatal(err)
	}
}
