package templates

import (
	"embed"
	"strings"
)

//go:embed all:_shared server cli ui database
var TemplateFS embed.FS

func GetCategories() ([]string, error) {
	entries, err := TemplateFS.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var categories []string
	for _, entry := range entries {
		if entry.IsDir() {
			categories = append(categories, entry.Name())
		}
	}
	return categories, nil
}

func GetOptions(category string) ([]string, error) {
	entries, err := TemplateFS.ReadDir(category)
	if err != nil {
		return nil, err
	}
	var options []string
	for _, entry := range entries {
		if entry.IsDir() {
			options = append(options, entry.Name())
		}
	}
	return options, nil
}

func HasVariants(category, option string) bool {
	_, err := TemplateFS.Open(category + "/" + option + "/dialects.txt")
	return err == nil
}

func GetVariants(category, option string) ([]string, error) {
	data, err := TemplateFS.ReadFile(category + "/" + option + "/dialects.txt")
	if err != nil {
		return nil, err
	}
	var variants []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			variants = append(variants, line)
		}
	}
	return variants, nil
}
