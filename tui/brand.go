package tui

// Holds the brand information for the TUI application.
// Including ASCII art, colors, and other branding elements.

import "charm.land/lipgloss/v2"

var (
	BrandName     = "GoBP"
	BrandLongName = "Go BluePrint"

	darkMode lipgloss.LightDarkFunc
	darkBG   bool
)

const ASCIILogo = `
╔══════════════════════════╗
║   ____      ____  ____   ║
║  / ___| ___| __ )|  _ \  ║
║ | |  _ / _ \  _ \| |_) | ║
║ | |_| | (_) | |_) |  __/ ║
║  \____|\___/|____/|_|    ║
║      GO BLUEPRINT        ║
╚══════════════════════════╝
`

var BrandStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	Padding(1, 2)

func RenderLogo() string {
	return BrandStyle.Render(ASCIILogo)
}