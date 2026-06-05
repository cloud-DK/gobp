package tui

import "charm.land/lipgloss/v2"

const ASCIILogo = `
╔════════════════════════╗
║    __       ___  ___   ║
║   / _| ___ | _ )| _ \  ║
║  | |_ / _ \| _ \|  _/  ║
║   \__|\___/|___/|_|    ║
║                        ║
║      go blueprint      ║
╚════════════════════════╝
`

var BrandStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Padding(0, 2)
