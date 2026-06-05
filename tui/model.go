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
	stepDone
)

type Model struct {
	step step

	categories []string
	options    []string
	cursor     int
	selected   map[int]struct{}

	selectedCategory    string
	selectedOptionsList []string
	err                 error
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func Run() error {
	categories, err := templates.GetCategories()
	if err != nil {
		return err
	}

	categories = filterCategories(categories)
	sort.Strings(categories)
	if len(categories) == 0 {
		return errors.New("no template categories found")
	}

	m := &Model{
		step:       stepCategory,
		categories: categories,
		selected:   make(map[int]struct{}),
	}

	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
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
