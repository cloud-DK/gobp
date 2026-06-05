package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

const selectedIcon = ">"

func (m *Model) View() tea.View {
	var sb strings.Builder
	sb.WriteString(BrandStyle.Render(ASCIILogo))
	sb.WriteString("\n")

	if m.err != nil {
		sb.WriteString(fmt.Sprintf("\n  Error: %v\n", m.err))
		return tea.NewView(sb.String())
	}

	switch m.step {
	case stepCategories:
		sb.WriteString("\n  Choose categories\n\n")
		writeCategoryList(&sb, m.categories, m.categoryChecks, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  ↑/↓: navigate   space: toggle   enter: confirm   ctrl+c/q: quit\n")

	case stepOption:
		total := len(m.selections) + 1 + len(m.pendingCategories)
		current := len(m.selections) + 1
		sb.WriteString(fmt.Sprintf("\n  Configuring: %s  (%d/%d)\n\n", m.currentCategory, current, total))
		sb.WriteString("  Choose an option\n\n")
		writeChoiceList(&sb, m.options, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  ↑/↓: navigate   enter/space: select   esc: back   ctrl+c/q: quit\n")

	case stepVariant:
		total := len(m.selections) + 1 + len(m.pendingCategories)
		current := len(m.selections) + 1
		sb.WriteString(fmt.Sprintf("\n  Configuring: %s > %s  (%d/%d)\n\n", m.currentCategory, m.currentOption, current, total))
		sb.WriteString("  Choose a dialect\n\n")
		writeChoiceList(&sb, m.variants, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
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
		sb.WriteString(fmt.Sprintf("  > %s|\n", m.moduleInput))
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  type module path   enter: confirm   esc: back   ctrl+c/q: quit\n")

	case stepOutputDir:
		projectName := projectNameFromModule(m.moduleInput)
		sb.WriteString("\n  Where to generate?\n\n")
		opts := []string{
			"current directory (.)",
			fmt.Sprintf("subdirectory: %s", projectName),
		}
		writeChoiceList(&sb, opts, m.cursor)
		sb.WriteString("\n  ────────────────────────────────────\n")
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
		if m.outputDirChoice == 1 {
			sb.WriteString(fmt.Sprintf("  %-12s ./%s/\n", "output:", projectNameFromModule(m.moduleInput)))
		} else {
			sb.WriteString(fmt.Sprintf("  %-12s current directory\n", "output:"))
		}
		sb.WriteString("\n  ────────────────────────────────────\n")
		sb.WriteString("  enter: generate   esc: back   ctrl+c/q: quit\n")
	}

	return tea.NewView(sb.String())
}

func writeCategoryList(sb *strings.Builder, items []string, checked map[int]bool, cursor int) {
	for i, item := range items {
		box := "[ ]"
		if checked[i] {
			box = "[x]"
		}
		if i == cursor {
			sb.WriteString(fmt.Sprintf("  %s %s %s\n", selectedIcon, box, item))
		} else {
			sb.WriteString(fmt.Sprintf("      %s %s\n", box, item))
		}
	}
}

func writeChoiceList(sb *strings.Builder, items []string, cursor int) {
	if len(items) == 0 {
		sb.WriteString("  (no options found)\n")
		return
	}
	for i, item := range items {
		if i == cursor {
			sb.WriteString(fmt.Sprintf("  %s %s\n", selectedIcon, item))
		} else {
			sb.WriteString(fmt.Sprintf("    %s\n", item))
		}
	}
}
