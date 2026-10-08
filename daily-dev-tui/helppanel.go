package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpTab struct {
	name string
	keys [][2]string
}

var helpTabs = []helpTab{
	{"Options panel", [][2]string{
		{"↑/↓", "select option"},
		{"tab", "change panel"},
		{"enter", "run command"},
		{"ctrl+h", "show this help"},
		{"q", "quit"},
	}},
	{"Posts panel", [][2]string{
		{"↑/↓ or j/k", "select post"},
		{"tab", "change panel"},
		{"b", "bookmark"},
		{"enter", "open in Daily.dev"},
		{"o", "open post link"},
		{"c", "show comments"},
		{"n", "next page"},
		{"ctrl+h", "show this help"},
		{"q", "quit"},
	}},
}

func (m model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+h", "enter":
		m.st = m.helpPrev
	case "tab", "right", "l", "left", "h", "shift+tab":
		m.helpTab = (m.helpTab + 1) % len(helpTabs)
	}
	return m, nil
}

func (m model) helpView() string {
	var tabs []string
	for i, t := range helpTabs {
		st := metaSt.Padding(0, 1)
		if i == m.helpTab {
			st = tagSt
		}
		tabs = append(tabs, st.Render(t.name))
	}
	lines := []string{titleSt.Render("Help") + "  " + strings.Join(tabs, " "), ""}
	for _, k := range helpTabs[m.helpTab].keys {
		lines = append(lines, keySt.Render(lipgloss.NewStyle().Width(20).Render(k[0]))+k[1])
	}
	lines = append(lines, "", metaSt.Render("tab: switch tab • esc/q/ctrl+h: close"))
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("135")).Padding(1, 2).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, box)
}
