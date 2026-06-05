package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cloud-dk/gobp/templates"
)

const (
	selectedIcon  = ">"
	checkedIcon   = "✓"
	uncheckedIcon = " "
)

func (m *Model) View() tea.View {
	var sb strings.Builder
	sb.WriteString(BrandStyle.Render(ASCIILogo))
	sb.WriteString("\n")

	if m.err != nil {
		sb.WriteString(fmt.Sprintf("\n  Error: %v\n", m.err))
		return tea.NewView(sb.String())
	}

	switch m.step {
	case stepCategory:
		sb.WriteString("\n  Step 1: Choose a category\n\n")
		writeChoiceList(&sb, m.categories, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  ↑/↓: navigate   enter/space: select   ctrl+c/q: quit\n")

	case stepOption:
		sb.WriteString(fmt.Sprintf("\n  Category: %s\n\n", m.selectedCategory))
		sb.WriteString("  Step 2: Choose an option\n\n")
		writeChoiceList(&sb, m.options, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  ↑/↓: navigate   enter/space: select   esc: back   ctrl+c/q: quit\n")

	case stepVariant:
		sb.WriteString(fmt.Sprintf("\n  Category: %s   Option: %s\n\n", m.selectedCategory, m.selectedOption))
		sb.WriteString("  Step 3: Choose a dialect\n\n")
		writeChoiceList(&sb, m.variants, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  ↑/↓: navigate   enter/space: select   esc: back   ctrl+c/q: quit\n")

	case stepModule:
		header := fmt.Sprintf("  Category: %s   Option: %s", m.selectedCategory, m.selectedOption)
		if m.selectedVariant != "" {
			header += fmt.Sprintf("   Dialect: %s", m.selectedVariant)
		}
		sb.WriteString("\n" + header + "\n\n")
		sb.WriteString("  Enter module name\n\n")
		sb.WriteString(fmt.Sprintf("  > %s|\n", m.moduleInput))
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  type module path   enter: confirm   esc: back   ctrl+c/q: quit\n")

	case stepDone:
		sb.WriteString("\n  Done\n\n")
		sb.WriteString(fmt.Sprintf("  Category: %s\n", m.selectedCategory))
		sb.WriteString(fmt.Sprintf("  Option:   %s\n", m.selectedOption))
		if m.selectedVariant != "" {
			sb.WriteString(fmt.Sprintf("  Dialect:  %s\n", m.selectedVariant))
		}
		if m.moduleInput != "" {
			sb.WriteString(fmt.Sprintf("  Module:   %s\n", m.moduleInput))
		}
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  enter: generate   esc: back   ctrl+c/q: quit\n")
	}

	return tea.NewView(sb.String())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()

		if m.step == stepModule {
			switch key {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.step = m.prevFromModule()
				m.moduleInput = ""
				return m, nil
			case "enter":
				if m.moduleInput != "" {
					m.step = stepDone
				}
				return m, nil
			case "backspace", "ctrl+h":
				if len(m.moduleInput) > 0 {
					m.moduleInput = m.moduleInput[:len(m.moduleInput)-1]
				}
				return m, nil
			default:
				if len(key) == 1 {
					m.moduleInput += key
				}
				return m, nil
			}
		}

		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			m.goBack()
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			limit := m.currentListLen() - 1
			if m.cursor < limit {
				m.cursor++
			}
		case "enter", "space":
			return m.selectCurrent()
		}
	}

	return m, nil
}

func (m *Model) currentListLen() int {
	switch m.step {
	case stepCategory:
		return len(m.categories)
	case stepOption:
		return len(m.options)
	case stepVariant:
		return len(m.variants)
	}
	return 0
}

func (m *Model) goBack() {
	switch m.step {
	case stepOption:
		m.step = stepCategory
		m.cursor = 0
		m.options = nil
		m.selected = make(map[int]struct{})
		m.selectedCategory = ""
	case stepVariant:
		m.step = stepOption
		m.cursor = 0
		m.variants = nil
		m.selectedOption = ""
	case stepDone:
		if m.moduleInput != "" {
			m.step = stepModule
		} else if m.selectedVariant != "" {
			m.step = stepVariant
			m.cursor = 0
			m.selectedVariant = ""
		} else {
			m.step = stepOption
			m.cursor = 0
		}
	}
}

func (m *Model) prevFromModule() step {
	if m.selectedVariant != "" {
		return stepVariant
	}
	return stepOption
}

func (m *Model) selectCurrent() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepCategory:
		if len(m.categories) == 0 {
			return m, nil
		}
		m.selectedCategory = m.categories[m.cursor]
		options, err := templates.GetOptions(m.selectedCategory)
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
		m.selectedOption = m.options[m.cursor]
		m.cursor = 0
		if templates.HasVariants(m.selectedCategory, m.selectedOption) {
			variants, err := templates.GetVariants(m.selectedCategory, m.selectedOption)
			if err != nil {
				m.err = err
				return m, nil
			}
			sort.Strings(variants)
			m.variants = variants
			m.step = stepVariant
		} else {
			m.step = nextGoStep(m.selectedCategory)
		}

	case stepVariant:
		if len(m.variants) == 0 {
			return m, nil
		}
		m.selectedVariant = m.variants[m.cursor]
		m.cursor = 0
		m.step = nextGoStep(m.selectedCategory)

	case stepDone:
		return m, tea.Quit
	}

	return m, nil
}

// nextGoStep returns the step after variant/option selection.
// UI skips the module name step since it runs an external command.
func nextGoStep(category string) step {
	if category == "ui" {
		return stepDone
	}
	return stepModule
}

func writeChoiceList(sb *strings.Builder, items []string, cursor int) {
	if len(items) == 0 {
		sb.WriteString("  (no options found)\n")
		return
	}
	for i, item := range items {
		prefix := "    "
		if i == cursor {
			prefix = fmt.Sprintf("  %s ", selectedIcon)
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", prefix, item))
	}
}
