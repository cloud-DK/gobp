package main

import (
	"fmt"
	"os"

	"github.com/cloud-dk/gobp/generator"
	"github.com/cloud-dk/gobp/tui"
)

func main() {
	result, err := tui.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if result.Option == "" {
		return
	}

	outputDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := generator.Generate(generator.Config{
		ModuleName: result.ModuleName,
		Category:   result.Category,
		Option:     result.Option,
		Variant:    result.Variant,
		OutputDir:  outputDir,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error generating project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Project generated in %s\n", outputDir)
}
