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

const version = "3.0.0"

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
	acc := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	hint := "enter: confirm   ctrl+c: quit"
	var panel []string
	switch m.st {
	case stToken:
		panel = []string{acc.Render("Enter your API token (or set DAILY_DEV_TOKEN):"), "", m.input.View()}
		hint = "enter: confirm   esc: back   ctrl+c: quit"
	case stOAuthForm:
		panel = []string{acc.Render("OAuth app credentials (or set DAILY_DEV_CLIENT_ID / DAILY_DEV_CLIENT_SECRET):"), "", m.oi[0].View(), m.oi[1].View(), "",
			metaSt.Render("Redirect URI to register: " + redirectURI())}
		hint = "enter: sign in   tab: next field   esc: back   ctrl+c: quit"
	case stOAuthWait:
		panel = []string{acc.Render("Waiting for you to approve access in the browser..."), ""}
		if m.sess != nil {
			panel = append(panel, metaSt.Render("If it did not open, visit:"), linkSt.Width(max(20, m.w-8)).Render(m.sess.URL))
		}
		hint = "esc: cancel   ctrl+c: quit"
	default:
		panel = []string{acc.Render("Choose how to sign in:"), ""}
		for i, o := range loginOptions {
			if i == m.loginSel {
				panel = append(panel, titleSt.Render("▸ "+o))
			} else {
				panel = append(panel, metaSt.Render("  "+o))
			}
		}
		hint = "↑/↓: select   enter: confirm   ctrl+c: quit"
	}
	if m.loginErr != "" {
		panel = append(panel, "", errSt.Width(max(20, m.w-8)).Render(m.loginErr))
	}
	body := strings.Join(append(append([]string{
		art, "",
		lipgloss.NewStyle().Width(lipgloss.Width(art)).Align(lipgloss.Center).Foreground(lipgloss.Color("214")).Bold(true).Render("v" + version), "",
		lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true).Render("Daily.dev public API client"),
		metaSt.Render("Browse feeds, posts, bookmarks and more through the Daily.dev REST API"),
		metaSt.Render("(sign in with a personal Bearer token or with OAuth)"), "",
		line, "",
	}, panel...), "", metaSt.Render(hint)), "\n")
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, body)
}
