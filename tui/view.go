package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

const selectedIcon = "›"

const divider = "  ────────────────────────────────────\n"

func (m *Model) View() tea.View {
	var sb strings.Builder
	sb.WriteString(BrandStyle.Render(ASCIILogo))
	sb.WriteString("\n")

	switch m.step {
	case stepError:
		sb.WriteString("\n  Error\n\n")
		sb.WriteString(fmt.Sprintf("  %v\n", m.err))
		sb.WriteString("\n" + divider)
		sb.WriteString("  esc: go back   ctrl+c/q: quit\n")

	case stepCategories:
		sb.WriteString("\n  Choose categories\n\n")
		writeCategoryList(&sb, m.categories, m.cursor)
		sb.WriteString("\n" + divider)
		sb.WriteString("  ↑/↓: navigate   space: toggle   enter: confirm   ctrl+c/q: quit\n")

	case stepOption:
		total := len(m.selections) + 1 + len(m.pendingCategories)
		current := len(m.selections) + 1
		sb.WriteString(fmt.Sprintf("\n  Configuring: %s  (%d/%d)\n\n", m.currentCategory.name, current, total))
		writeItemList(&sb, m.options, m.cursor)
		sb.WriteString("\n" + divider)
		sb.WriteString("  ↑/↓: navigate   enter/space: select   esc: back   ctrl+c/q: quit\n")

	case stepVariant:
		total := len(m.selections) + 1 + len(m.pendingCategories)
		current := len(m.selections) + 1
		sb.WriteString(fmt.Sprintf("\n  Configuring: %s › %s  (%d/%d)\n\n", m.currentCategory.name, m.currentOption.name, current, total))
		writeItemList(&sb, m.variants, m.cursor)
		sb.WriteString("\n" + divider)
		sb.WriteString("  ↑/↓: navigate   enter/space: select   esc: back   ctrl+c/q: quit\n")

	case stepModule:
		sb.WriteString("\n  ")
		parts := make([]string, len(m.selections))
		for i, s := range m.selections {
			if s.Variant != "" {
				parts[i] = fmt.Sprintf("%s/%s/%s", s.Category, s.Option, s.Variant)
			} else {
				parts[i] = fmt.Sprintf("%s/%s", s.Category, s.Option)
			}
		}
		sb.WriteString(strings.Join(parts, "  ·  "))
		sb.WriteString("\n\n  Enter module name\n\n")
		sb.WriteString(fmt.Sprintf("  › %s|\n", m.moduleInput))
		sb.WriteString("\n" + divider)
		sb.WriteString("  type module path   enter: confirm   esc: back   ctrl+c/q: quit\n")

	case stepOutputDir:
		projectName := projectNameFromModule(m.moduleInput)
		sb.WriteString("\n  Where to generate?\n\n")
		opts := []item{
			{name: "current directory (.)"},
			{name: fmt.Sprintf("subdirectory: %s/", projectName)},
		}
		writeItemList(&sb, opts, m.cursor)
		sb.WriteString("\n" + divider)
		sb.WriteString("  ↑/↓: navigate   enter: confirm   esc: back   ctrl+c/q: quit\n")

	case stepDone:
		sb.WriteString("\n  Summary\n\n")
		for _, s := range m.selections {
			if s.Variant != "" {
				sb.WriteString(fmt.Sprintf("  %-12s %s (%s)\n", s.Category+":", s.Option, s.Variant))
			} else {
				sb.WriteString(fmt.Sprintf("  %-12s %s\n", s.Category+":", s.Option))
			}
		}
		if m.moduleInput != "" {
			sb.WriteString(fmt.Sprintf("  %-12s %s\n", "module:", m.moduleInput))
		}
		if m.outputInSubdir {
			sb.WriteString(fmt.Sprintf("  %-12s ./%s/\n", "output:", projectNameFromModule(m.moduleInput)))
		} else {
			sb.WriteString(fmt.Sprintf("  %-12s current directory\n", "output:"))
		}
		sb.WriteString("\n" + divider)
		sb.WriteString("  enter: generate   esc: back   ctrl+c/q: quit\n")
	}

	return tea.NewView(sb.String())
}

func writeCategoryList(sb *strings.Builder, items []item, cursor int) {
	for i, it := range items {
		box := "[ ]"
		if it.checked {
			box = "[x]"
		}
		if i == cursor {
			sb.WriteString(fmt.Sprintf("  %s %s %s\n", selectedIcon, box, it.name))
		} else {
			sb.WriteString(fmt.Sprintf("      %s %s\n", box, it.name))
		}
	}
}

func writeItemList(sb *strings.Builder, items []item, cursor int) {
	if len(items) == 0 {
		sb.WriteString("  (no options found)\n")
		return
	}
	for i, it := range items {
		if i == cursor {
			sb.WriteString(fmt.Sprintf("  %s %-14s", selectedIcon, it.name))
			if it.desc != "" {
				sb.WriteString(fmt.Sprintf("  %s", it.desc))
			}
			sb.WriteString("\n")
		} else {
			sb.WriteString(fmt.Sprintf("    %-14s", it.name))
			if it.desc != "" {
				sb.WriteString(fmt.Sprintf("  %s", it.desc))
			}
			sb.WriteString("\n")
		}
	}
}
