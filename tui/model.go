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
	stepCategories step = iota
	stepOption
	stepVariant
	stepModule
	stepOutputDir
	stepDone
)

type Selection struct {
	Category string
	Option   string
	Variant  string
}

type Result struct {
	Selections []Selection
	ModuleName string
	OutputDir  string // "" = cwd, else subdir name to create
}

type Model struct {
	step step

	categories     []string
	categoryChecks map[int]bool
	cursor         int

	pendingCategories []string
	currentCategory   string
	currentOption     string
	options           []string
	variants          []string

	selections []Selection

	moduleInput     string
	outputDirChoice int // 0 = cwd, 1 = subdir

	err error
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
		step:           stepCategories,
		categories:     categories,
		categoryChecks: make(map[int]bool),
	}
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := final.(*Model)
	outDir := ""
	if fm.outputDirChoice == 1 {
		outDir = projectNameFromModule(fm.moduleInput)
	}
	return &Result{
		Selections: fm.selections,
		ModuleName: fm.moduleInput,
		OutputDir:  outDir,
	}, nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.step == stepModule {
			return m.handleModuleInput(key)
		}
		return m.handleNavInput(key)
	}
	return m, nil
}

func (m *Model) handleModuleInput(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.goBack()
		m.moduleInput = ""
	case "enter":
		if m.moduleInput != "" {
			m.cursor = 0
			m.step = stepOutputDir
		}
	case "backspace", "ctrl+h":
		if len(m.moduleInput) > 0 {
			m.moduleInput = m.moduleInput[:len(m.moduleInput)-1]
		}
	default:
		if len(key) == 1 {
			m.moduleInput += key
		}
	}
	return m, nil
}

func (m *Model) handleNavInput(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "esc":
		m.goBack()
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		limit := m.currentListLen() - 1
		if m.cursor < limit {
			m.cursor++
		}
	case "space":
		if m.step == stepCategories {
			m.categoryChecks[m.cursor] = !m.categoryChecks[m.cursor]
		} else {
			return m.selectCurrent()
		}
	case "enter":
		return m.selectCurrent()
	}
	return m, nil
}

func (m *Model) currentListLen() int {
	switch m.step {
	case stepCategories:
		return len(m.categories)
	case stepOption:
		return len(m.options)
	case stepVariant:
		return len(m.variants)
	case stepOutputDir:
		return 2
	}
	return 0
}

func (m *Model) goBack() {
	switch m.step {
	case stepOption:
		m.step = stepCategories
		m.selections = nil
		m.pendingCategories = nil
		m.currentCategory = ""
		m.currentOption = ""
		m.options = nil
		m.cursor = 0
	case stepVariant:
		m.step = stepOption
		m.currentOption = ""
		m.variants = nil
		m.cursor = 0
	case stepModule:
		m.step = stepCategories
		m.selections = nil
		m.pendingCategories = nil
		m.currentCategory = ""
		m.currentOption = ""
		m.options = nil
		m.variants = nil
		m.moduleInput = ""
		m.cursor = 0
	case stepOutputDir:
		m.step = stepModule
		m.cursor = 0
	case stepDone:
		m.step = stepOutputDir
		m.cursor = m.outputDirChoice
	}
}

func (m *Model) selectCurrent() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepCategories:
		var checked []string
		for i, cat := range m.categories {
			if m.categoryChecks[i] {
				checked = append(checked, cat)
			}
		}
		if len(checked) == 0 {
			return m, nil
		}
		m.pendingCategories = checked[1:]
		m.currentCategory = checked[0]
		options, err := templates.GetOptions(m.currentCategory)
		if err != nil {
			m.err = err
			return m, nil
		}
		sort.Strings(options)
		m.options = options
		m.cursor = 0
		m.step = stepOption

	case stepOption:
		if len(m.options) == 0 {
			return m, nil
		}
		option := m.options[m.cursor]
		if templates.HasVariants(m.currentCategory, option) {
			variants, err := templates.GetVariants(m.currentCategory, option)
			if err != nil {
				m.err = err
				return m, nil
			}
			sort.Strings(variants)
			m.variants = variants
			m.currentOption = option
			m.cursor = 0
			m.step = stepVariant
		} else {
			m.selections = append(m.selections, Selection{m.currentCategory, option, ""})
			m.step = m.advanceFromSelections()
		}

	case stepVariant:
		if len(m.variants) == 0 {
			return m, nil
		}
		variant := m.variants[m.cursor]
		m.selections = append(m.selections, Selection{m.currentCategory, m.currentOption, variant})
		m.step = m.advanceFromSelections()

	case stepOutputDir:
		m.outputDirChoice = m.cursor
		m.step = stepDone

	case stepDone:
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) advanceFromSelections() step {
	if len(m.pendingCategories) > 0 {
		m.currentCategory = m.pendingCategories[0]
		m.pendingCategories = m.pendingCategories[1:]
		options, err := templates.GetOptions(m.currentCategory)
		if err != nil {
			m.err = err
			return stepOption
		}
		sort.Strings(options)
		m.options = options
		m.currentOption = ""
		m.variants = nil
		m.cursor = 0
		return stepOption
	}
	m.cursor = 0
	return stepModule
}

func filterCategories(categories []string) []string {
	filtered := make([]string, 0, len(categories))
	for _, c := range categories {
		if !strings.HasPrefix(c, "_") {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func projectNameFromModule(moduleName string) string {
	parts := strings.Split(moduleName, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return moduleName
}
