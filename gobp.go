package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloud-dk/gobp/generator"
	"github.com/cloud-dk/gobp/templates"
	"github.com/cloud-dk/gobp/tui"
)

// Set at build time: -ldflags "-X main.version=1.2.3"
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

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

	// createdOutputDir is the directory gobp created; removed on any generation error.
	var createdOutputDir string
	if result.OutputDir != "" {
		outputDir = filepath.Join(outputDir, result.OutputDir)
		if _, statErr := os.Stat(outputDir); os.IsNotExist(statErr) {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "error creating output directory: %v\n", err)
				os.Exit(1)
			}
			createdOutputDir = outputDir
		} else if statErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", statErr)
			os.Exit(1)
		}
	}

	// failf prints the error and removes any directory gobp created before exiting.
	failf := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		if createdOutputDir != "" {
			os.RemoveAll(createdOutputDir)
		}
		os.Exit(1)
	}

	needsModInit := false
	for _, sel := range result.Selections {
		meta, _ := templates.GetMeta(sel.Category, sel.Option)
		if meta == nil || !meta.IsCmd {
			needsModInit = true
			break
		}
	}

	if needsModInit {
		if err := generator.WriteShared(result.ModuleName, outputDir); err != nil {
			failf("error writing shared templates: %v", err)
		}
	}

	for _, sel := range result.Selections {
		if err := generator.Generate(generator.Config{
			ModuleName: result.ModuleName,
			Category:   sel.Category,
			Option:     sel.Option,
			Variant:    sel.Variant,
			OutputDir:  outputDir,
		}); err != nil {
			failf("error generating %s/%s: %v", sel.Category, sel.Option, err)
		}
	}

	if needsModInit && result.ModuleName != "" {
		if err := generator.GoModInit(result.ModuleName, outputDir); err != nil {
			failf("error running go mod init: %v", err)
		}
	}

	fmt.Printf("Project generated in %s\n", outputDir)
}
