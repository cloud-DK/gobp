package tui

import (
	"errors"
	"sort"
	"strings"
	"unicode/utf8"

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
	stepError
)

// item is a selectable list entry with an optional description and checkbox state.
type item struct {
	name    string
	desc    string
	checked bool
}

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

	categories []item
	cursor     int

	pendingCategories []item
	currentCategory   item
	currentOption     item
	options           []item
	variants          []item

	selections []Selection

	moduleInput    string
	outputInSubdir bool

	err     error
	errPrev step // step to return to on esc from error
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func Run() (*Result, error) {
	rawCategories, err := templates.GetCategories()
	if err != nil {
		return nil, err
	}
	rawCategories = filterCategories(rawCategories)
	sort.Strings(rawCategories)
	if len(rawCategories) == 0 {
		return nil, errors.New("no template categories found")
	}
	categories := make([]item, len(rawCategories))
	for i, c := range rawCategories {
		categories[i] = item{name: c}
	}
	m := &Model{
		step:       stepCategories,
		categories: categories,
	}
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm, ok := final.(*Model)
	if !ok {
		return nil, errors.New("unexpected model type returned from TUI")
	}
	outDir := ""
	if fm.outputInSubdir {
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
		if utf8.RuneCountInString(key) == 1 {
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
		if m.step == stepError {
			m.step = m.errPrev
			m.err = nil
		} else {
			m.goBack()
		}
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
			m.categories[m.cursor].checked = !m.categories[m.cursor].checked
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
		m.pendingCategories = nil
		m.currentCategory = item{}
		m.currentOption = item{}
		m.options = nil
		m.cursor = 0
	case stepVariant:
		m.step = stepOption
		m.currentOption = item{}
		m.variants = nil
		m.cursor = 0
	case stepModule:
		m.step = stepCategories
		m.pendingCategories = nil
		m.currentCategory = item{}
		m.currentOption = item{}
		m.options = nil
		m.variants = nil
		m.moduleInput = ""
		m.cursor = 0
	case stepOutputDir:
		m.step = stepModule
		m.cursor = 0
	case stepDone:
		m.step = stepOutputDir
		if m.outputInSubdir {
			m.cursor = 1
		} else {
			m.cursor = 0
		}
	}
}

func (m *Model) setError(err error, returnTo step) (tea.Model, tea.Cmd) {
	m.err = err
	m.errPrev = returnTo
	m.step = stepError
	return m, nil
}

func (m *Model) selectCurrent() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepCategories:
		m.selections = nil
		var checked []item
		for _, cat := range m.categories {
			if cat.checked {
				checked = append(checked, cat)
			}
		}
		if len(checked) == 0 {
			return m, nil
		}
		m.pendingCategories = checked[1:]
		m.currentCategory = checked[0]
		options, err := loadOptions(m.currentCategory.name)
		if err != nil {
			return m.setError(err, stepCategories)
		}
		m.options = options
		m.cursor = 0
		m.step = stepOption

	case stepOption:
		if len(m.options) == 0 {
			return m, nil
		}
		opt := m.options[m.cursor]
		if templates.HasVariants(m.currentCategory.name, opt.name) {
			variants, err := loadVariants(m.currentCategory.name, opt.name)
			if err != nil {
				return m.setError(err, stepOption)
			}
			m.variants = variants
			m.currentOption = opt
			m.cursor = 0
			m.step = stepVariant
		} else {
			m.selections = append(m.selections, Selection{m.currentCategory.name, opt.name, ""})
			m.step = m.advanceFromSelections()
		}

	case stepVariant:
		if len(m.variants) == 0 {
			return m, nil
		}
		v := m.variants[m.cursor]
		m.selections = append(m.selections, Selection{m.currentCategory.name, m.currentOption.name, v.name})
		m.step = m.advanceFromSelections()

	case stepOutputDir:
		m.outputInSubdir = m.cursor == 1
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
		options, err := loadOptions(m.currentCategory.name)
		if err != nil {
			m.err = err
			m.errPrev = stepOption
			return stepError
		}
		m.options = options
		m.currentOption = item{}
		m.variants = nil
		m.cursor = 0
		return stepOption
	}
	m.cursor = 0
	return stepModule
}

// loadOptions fetches options for a category and enriches them with descriptions from meta.json.
func loadOptions(category string) ([]item, error) {
	names, err := templates.GetOptions(category)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	items := make([]item, len(names))
	for i, name := range names {
		meta, err := templates.GetMeta(category, name)
		if err != nil {
			items[i] = item{name: name}
			continue
		}
		items[i] = item{name: name, desc: meta.Description}
	}
	return items, nil
}

// loadVariants fetches dialect variants with their descriptions.
func loadVariants(category, option string) ([]item, error) {
	dialects, err := templates.GetVariants(category, option)
	if err != nil {
		return nil, err
	}
	items := make([]item, len(dialects))
	for i, d := range dialects {
		items[i] = item{name: d.Name, desc: d.Description}
	}
	return items, nil
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
