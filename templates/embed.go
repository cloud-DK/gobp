package templates

import (
	"embed"
	"encoding/json"
)

//go:embed all:_shared server cli ui database
var TemplateFS embed.FS

type Meta struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	IsCmd       bool          `json:"isCmd,omitempty"` // true for categories that run an external command (not go mod init)
	Dialects    []DialectMeta `json:"dialects,omitempty"`
}

type DialectMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

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

// GetMeta returns the metadata for a template option. If meta.json is absent
// or unreadable it returns a best-effort struct so callers never get nil.
func GetMeta(category, option string) (*Meta, error) {
	data, err := TemplateFS.ReadFile(category + "/" + option + "/meta.json")
	if err != nil {
		return &Meta{Name: option}, nil
	}
	var m Meta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Name == "" {
		m.Name = option
	}
	return &m, nil
}

func HasVariants(category, option string) (bool, error) {
	m, err := GetMeta(category, option)
	if err != nil {
		return false, err
	}
	return len(m.Dialects) > 0, nil
}

func GetVariants(category, option string) ([]DialectMeta, error) {
	m, err := GetMeta(category, option)
	if err != nil {
		return nil, err
	}
	return m.Dialects, nil
}
