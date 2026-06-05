package tui

import (
	"errors"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cloud-dk/gobp/templates"
)

type step int

const (
	stepCategory step = iota
	stepOption
	stepVariant
	stepModule
	stepDone
)

// Result holds the user's selections returned by Run.
type Result struct {
	Category   string
	Option     string
	Variant    string
	ModuleName string
}

type Model struct {
	step step

	categories []string
	options    []string
	variants   []string
	cursor     int
	selected   map[int]struct{}

	selectedCategory string
	selectedOption   string
	selectedVariant  string
	moduleInput      string
	err              error
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func Run() (*Result, error) {
	categories, err := templates.GetCategories()
	if err != nil {
		return nil, err
	}

	categories = filterCategories(categories)
	sort.Strings(categories)
	if len(categories) == 0 {
		return nil, errors.New("no template categories found")
	}

	m := &Model{
		step:       stepCategory,
		categories: categories,
		selected:   make(map[int]struct{}),
	}

	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := final.(*Model)
	return &Result{
		Category:   fm.selectedCategory,
		Option:     fm.selectedOption,
		Variant:    fm.selectedVariant,
		ModuleName: fm.moduleInput,
	}, nil
}

func filterCategories(categories []string) []string {
	filtered := make([]string, 0, len(categories))
	for _, category := range categories {
		if strings.HasPrefix(category, "_") {
			continue
		}
		filtered = append(filtered, category)
	}
	return filtered
}
