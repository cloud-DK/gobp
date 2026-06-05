package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloud-dk/gobp/generator"
	"github.com/cloud-dk/gobp/tui"
)

func main() {
	result, err := tui.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(result.Selections) == 0 {
		return
	}

	outputDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if result.OutputDir != "" {
		outputDir = filepath.Join(outputDir, result.OutputDir)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	needsModInit := false
	for _, sel := range result.Selections {
		if sel.Category != "ui" {
			needsModInit = true
		}
		if err := generator.Generate(generator.Config{
			ModuleName: result.ModuleName,
			Category:   sel.Category,
			Option:     sel.Option,
			Variant:    sel.Variant,
			OutputDir:  outputDir,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "error generating %s/%s: %v\n", sel.Category, sel.Option, err)
			os.Exit(1)
		}
	}

	if needsModInit && result.ModuleName != "" {
		if err := generator.GoModInit(result.ModuleName, outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "error running go mod init: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Project generated in %s\n", outputDir)
}
