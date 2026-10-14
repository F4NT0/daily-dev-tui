package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) openThemes() model {
	m.themeOrig, m.themeSel, m.st = themeName, 0, stThemes
	for i, n := range themeNames() {
		if n == themeName {
			m.themeSel = i
		}
	}
	return m
}

func (m model) updateThemes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	names := themeNames()
	switch msg.String() {
	case "up", "k":
		m.themeSel = (m.themeSel + len(names) - 1) % len(names)
		setTheme(names[m.themeSel])
	case "down", "j":
		m.themeSel = (m.themeSel + 1) % len(names)
		setTheme(names[m.themeSel])
	case "enter":
		saveTheme(names[m.themeSel])
		m.st, m.status = stMenu, "theme: "+names[m.themeSel]
	case "esc", "q", "t":
		setTheme(m.themeOrig)
		m.st = stMenu
	}
	return m, nil
}

func (m model) themesView() string {
	lines := []string{titleSt.Render("Themes"), ""}
	for i, n := range themeNames() {
		if i == m.themeSel {
			lines = append(lines, titleSt.Render("▸ "+n))
		} else {
			lines = append(lines, metaSt.Render("  "+n))
		}
	}
	lines = append(lines, "", metaSt.Render("↑/↓: preview • enter: apply • esc: cancel"))
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(col("primary")).Padding(1, 2).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, box)
}
