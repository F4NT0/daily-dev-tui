package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Glyphs use the "ANSI Shadow" figlet font, same as the clidocs project.
var glyphs = map[rune][6]string{
	'D': {"██████╗ ", "██╔══██╗", "██║  ██║", "██║  ██║", "██████╔╝", "╚═════╝ "},
	'A': {" █████╗ ", "██╔══██╗", "███████║", "██╔══██║", "██║  ██║", "╚═╝  ╚═╝"},
	'I': {"██╗", "██║", "██║", "██║", "██║", "╚═╝"},
	'L': {"██╗     ", "██║     ", "██║     ", "██║     ", "███████╗", "╚══════╝"},
	'Y': {"██╗   ██╗", "╚██╗ ██╔╝", " ╚████╔╝ ", "  ╚██╔╝  ", "   ██║   ", "   ╚═╝   "},
	'.': {"   ", "   ", "   ", "   ", "██╗", "╚═╝"},
	'E': {"███████╗", "██╔════╝", "█████╗  ", "██╔══╝  ", "███████╗", "╚══════╝"},
	'V': {"██╗   ██╗", "██║   ██║", "██║   ██║", "╚██╗ ██╔╝", " ╚████╔╝ ", "  ╚═══╝  "},
	'P': {"██████╗ ", "██╔══██╗", "██████╔╝", "██╔═══╝ ", "██║     ", "╚═╝     "},
	' ': {"  ", "  ", "  ", "  ", "  ", "  "},
}

func banner(text string) string {
	rows := make([]string, 6)
	for _, r := range text {
		for i := range rows {
			rows[i] += glyphs[r][i]
		}
	}
	return strings.Join(rows, "\n")
}

func (m model) tokenView() string {
	green := lipgloss.Color("135")
	art := lipgloss.NewStyle().Foreground(green).Bold(true).Render(banner("DAILY.DEV API"))
	if lipgloss.Width(art) > m.w {
		art = lipgloss.NewStyle().Foreground(green).Bold(true).Render(banner("DAILY.DEV") + "\n\n" + banner("API"))
	}
	lineW := min(60, m.w-4)
	line := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", lineW))
	body := strings.Join([]string{
		art, "",
		lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true).Render("Daily.dev public API client"),
		metaSt.Render("Browse feeds, posts, bookmarks and more through the Daily.dev REST API"),
		metaSt.Render("(authenticated with a personal Bearer token)"), "",
		line, "",
		lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("Enter your API token (or set DAILY_DEV_TOKEN):"), "",
		m.input.View(), "",
		metaSt.Render("enter: confirm   ctrl+c: quit"),
	}, "\n")
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, body)
}
