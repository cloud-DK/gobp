package tui

import tea "charm.land/bubbletea/v2"

var selectedIcon = ">"
var checkedIcon = "✓"
var uncheckedIcon = " "

func (m *Model) View() tea.View {
	return tea.NewView(BrandStyle.Render(ASCIILogo) + "\n")
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Move cursor up
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
			// Move cursor down
		case "down", "j":
			if m.Cursor < len(m.Options)-1 {
				m.Cursor++
			}
			// Key enter or space toggles the selected option
		case "enter", " ":
			// Toggle the selected option
			m.SelectedOption[m.Cursor] = !m.SelectedOption[m.Cursor]
			if m.SelectedOption[m.Cursor] {
				showChecked(true)
			}

		case "ctrl+c", "q":
			return m, tea.Quit
		}

	}

	return m, nil
}

func showChecked(checked bool) string {
	if checked {
		return checkedIcon
	}
	return uncheckedIcon
}
