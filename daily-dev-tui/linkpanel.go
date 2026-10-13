package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) openFailed(err error, p map[string]any, u string) model {
	m.linkURL, m.linkErr, m.linkPost, m.st, m.status = u, err.Error(), p, stLink, ""
	return m
}

func (m model) linkView() string {
	w := min(max(30, m.w-12), 100)
	str := func(k string) string { s, _ := m.linkPost[k].(string); return s }
	lines := []string{titleSt.Width(w).Render(fallback(str("title"), "(post)"))}
	if src, ok := m.linkPost["source"].(map[string]any); ok {
		if n, _ := src["name"].(string); n != "" {
			lines = append(lines, subSt.Render(n))
		}
	}
	if s := fallback(str("summary"), str("tldr")); s != "" {
		if r := []rune(s); len(r) > 400 {
			s = string(r[:400]) + "…"
		}
		lines = append(lines, lipgloss.NewStyle().Width(w).Render(s))
	}
	var stats []string
	for _, k := range []string{"numUpvotes", "numComments", "readTime", "publishedAt"} {
		if v, ok := m.linkPost[k]; ok && v != nil {
			stats = append(stats, fmt.Sprintf("%s %s", k, scalar(v)))
		}
	}
	if len(stats) > 0 {
		lines = append(lines, metaSt.Width(w).Render(strings.Join(stats, "  •  ")))
	}
	lines = append(lines, "", errSt.Width(w).Render("Could not open the link: "+m.linkErr), "",
		"Copy this link and open it manually:", linkSt.Width(w).Render(m.linkURL), "",
		metaSt.Render("on Linux install xdg-utils or set $BROWSER to open links automatically"),
		metaSt.Render("press any key to close"))
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(col("primary")).Padding(1, 2).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, box)
}

func fallback(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
