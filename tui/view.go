package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cloud-dk/gobp/templates"
)

var selectedIcon = ">"
var checkedIcon = "✓"
var uncheckedIcon = " "

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
		sb.WriteString("\n  Step 1 of 2: Choose a category\n\n")
		writeChoiceList(&sb, m.categories, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  ↑/↓: navigate   enter/space: select   ctrl+c/q: quit\n")

	case stepOption:
		sb.WriteString(fmt.Sprintf("\n  Category: %s\n\n", m.selectedCategory))
		sb.WriteString("  Step 2 of 2: Choose template options\n\n")
		writeToggleList(&sb, m.options, m.selected, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  ↑/↓: navigate   space: toggle   enter: confirm   ctrl+c/q: quit\n")

	case stepDone:
		sb.WriteString("\n  Selection complete\n\n")
		sb.WriteString(fmt.Sprintf("  Category: %s\n", m.selectedCategory))
		sb.WriteString(fmt.Sprintf("  Options:  %s\n", strings.Join(m.selectedOptionsList, ", ")))
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  enter/q: quit\n")
	}

	return tea.NewView(sb.String())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			m.moveCursorDown()
		case "enter":
			return m.selectCurrent()
		case "space":
			if m.step == stepOption {
				m.toggleCurrentOption()
				return m, nil
			}
			return m.selectCurrent()
		}
	}

	return m, nil
}

func (m *Model) moveCursorDown() {
	switch m.step {
	case stepCategory:
		if m.cursor < len(m.categories)-1 {
			m.cursor++
		}
	case stepOption:
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	}
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
		return m, nil

	case stepOption:
		if len(m.options) == 0 {
			m.selectedOptionsList = nil
			m.step = stepDone
			return m, nil
		}
		m.selectedOptionsList = m.selectedOptionsList[:0]
		for i, option := range m.options {
			if _, ok := m.selected[i]; ok {
				m.selectedOptionsList = append(m.selectedOptionsList, option)
			}
		}
		if len(m.selectedOptionsList) == 0 {
			m.selectedOptionsList = append(m.selectedOptionsList, m.options[m.cursor])
		}
		m.step = stepDone
		m.cursor = 0
		return m, nil

	case stepDone:
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) toggleCurrentOption() {
	if m.step != stepOption || len(m.options) == 0 {
		return
	}
	if _, ok := m.selected[m.cursor]; ok {
		delete(m.selected, m.cursor)
		return
	}
	m.selected[m.cursor] = struct{}{}
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

func writeToggleList(sb *strings.Builder, items []string, selected map[int]struct{}, cursor int) {
	if len(items) == 0 {
		sb.WriteString("  (no options found)\n")
		return
	}

	for i, item := range items {
		pointer := "   "
		if i == cursor {
			pointer = fmt.Sprintf(" %s ", selectedIcon)
		}
		check := uncheckedIcon
		if _, ok := selected[i]; ok {
			check = checkedIcon
		}
		sb.WriteString(fmt.Sprintf("%s[%s] %s\n", pointer, check, item))
	}
}
