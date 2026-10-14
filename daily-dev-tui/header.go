package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type userMsg struct{ name string }

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

const noticeTTL = 6 * time.Second

type rateInfo struct {
	known       bool
	limit, left int
	reset       time.Time
}

var (
	rateMu sync.Mutex
	rate   rateInfo
)

func recordRate(h http.Header) {
	l, err1 := strconv.Atoi(strings.TrimSpace(h.Get("x-ratelimit-limit")))
	r, err2 := strconv.Atoi(strings.TrimSpace(h.Get("x-ratelimit-remaining")))
	if err1 != nil || err2 != nil {
		return
	}
	ri := rateInfo{known: true, limit: l, left: r}
	if f, err := strconv.ParseFloat(strings.TrimSpace(h.Get("x-ratelimit-reset")), 64); err == nil {
		switch {
		case f > 1e12:
			ri.reset = time.UnixMilli(int64(f))
		case f > 1e9:
			ri.reset = time.Unix(int64(f), 0)
		default:
			ri.reset = time.Now().Add(time.Duration(f * float64(time.Second)))
		}
	}
	rateMu.Lock()
	rate = ri
	rateMu.Unlock()
}

func fmtDur(d time.Duration) string {
	s := int(d.Round(time.Second) / time.Second)
	if s >= 60 {
		return fmt.Sprintf("%dm %02ds", s/60, s%60)
	}
	return fmt.Sprintf("%ds", s)
}

func userFromProfile(res any) string {
	m, _ := res.(map[string]any)
	if d, ok := m["data"].(map[string]any); ok {
		m = d
	}
	for _, k := range []string{"username", "name"} {
		if s, _ := m[k].(string); s != "" {
			return s
		}
	}
	return ""
}

func (m model) fetchUser() tea.Cmd {
	return func() tea.Msg {
		r, err := doGet(m.client, m.bearer(), Endpoint{"Profile", "My profile", "/profile/", nil}, nil)
		if err != nil {
			return userMsg{"unknown"}
		}
		if n := userFromProfile(r); n != "" {
			return userMsg{n}
		}
		return userMsg{"unknown"}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch t := msg.(type) {
	case tickMsg:
		if m.status != "" && time.Since(m.statusAt) > noticeTTL {
			m.status = ""
		}
		return m, tickCmd()
	case userMsg:
		m.user = t.name
		return m, nil
	}
	prev := m.status
	nm, cmd := m.step(msg)
	mm := nm.(model)
	switch msg.(type) {
	case resultMsg, actionMsg:
		mm.statusAt = time.Now()
	default:
		if mm.status != "" && mm.status != prev {
			mm.statusAt = time.Now()
		}
	}
	if (mm.token != "" || mm.oauth != nil) && !mm.userReq {
		mm.userReq = true
		cmd = tea.Batch(cmd, mm.fetchUser())
	}
	return mm, cmd
}

func isFailure(s string) bool {
	s = strings.ToLower(s)
	for _, w := range []string{"fail", "error", "required", "denied", "no browser", "invalid"} {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

func seg(w int, a lipgloss.Position, s string) string {
	return lipgloss.NewStyle().Inline(true).Width(w).MaxWidth(w).Align(a).Render(s)
}

func (m model) infoText() string {
	rateMu.Lock()
	r := rate
	rateMu.Unlock()
	user := m.user
	if user == "" {
		user = "..."
	}
	parts := []string{metaSt.Render("User: ") + titleSt.Render(user)}
	if !r.known {
		parts = append(parts, metaSt.Render("Requests left: -"), metaSt.Render("Limit: -"), metaSt.Render("Resets in: -"))
		return strings.Join(parts, metaSt.Render("  │  "))
	}
	left, reset := r.left, "-"
	if !r.reset.IsZero() {
		if d := time.Until(r.reset); d > 0 {
			reset = fmtDur(d)
		} else {
			reset, left = "now", r.limit
		}
	}
	lc := lipgloss.NewStyle().Foreground(col("ok")).Bold(true)
	switch {
	case left == 0:
		lc = lipgloss.NewStyle().Foreground(col("error")).Bold(true)
	case r.limit > 0 && left*4 <= r.limit:
		lc = lipgloss.NewStyle().Foreground(col("bookmark")).Bold(true)
	}
	parts = append(parts,
		metaSt.Render("Requests left: ")+lc.Render(strconv.Itoa(left)),
		metaSt.Render("Limit: ")+keySt.Render(strconv.Itoa(r.limit)+"/min"),
		metaSt.Render("Resets in: ")+keySt.Render(reset))
	return strings.Join(parts, metaSt.Render("  │  "))
}

func (m model) headerView() string {
	info := m.infoText()
	nw := min(max((m.w-lipgloss.Width(info))/2, 20), 40)
	notice := ""
	if m.status != "" {
		c := col("ok")
		if isFailure(m.status) {
			c = col("error")
		}
		notice = lipgloss.NewStyle().Foreground(c).Bold(true).Render("● " + m.status)
	}
	return seg(nw, lipgloss.Left, "") + seg(m.w-2*nw, lipgloss.Center, info) + seg(nw, lipgloss.Right, notice)
}
