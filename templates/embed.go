package templates

import "embed"

//go:embed all:_shared server cli ui
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
