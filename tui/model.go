package tui

import tea "charm.land/bubbletea/v2"

type Model struct {
	// Available categories and options
	Categories []string
	Options    []string
	// SelectedCategory maps category index to selected category name
	SelectedCategory map[int]bool
	SelectedOption   map[int]bool

	// Cursor position while navigating different stuff.
	Cursor int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func Run() error {
	m := &Model{
		SelectedCategory: make(map[int]bool),
		SelectedOption:   make(map[int]bool),
	}
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
