package tui

import "charm.land/lipgloss/v2"

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
	Padding(1, 2)
