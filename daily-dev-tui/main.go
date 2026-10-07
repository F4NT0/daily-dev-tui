package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stToken state = iota
	stMenu
	stForm
	stLoading
)

type resultMsg struct {
	res any
	err error
}

type model struct {
	st         state
	token      string
	client     *http.Client
	cur        int
	input      textinput.Model
	inputs     []textinput.Model
	focus      int
	vals       map[string]string
	active     *Endpoint
	cursor     string
	vp         viewport.Model
	sp         spinner.Model
	rightFocus bool
	w, h       int
	status     string
}

const leftW = 36

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Bearer token"
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()
	m := model{client: newClient(), input: ti, sp: spinner.New(spinner.WithSpinner(spinner.Dot)), vp: viewport.New(10, 10)}
	m.vp.SetContent(metaSt.Render("Select a request on the left and press Enter."))
	if t := strings.TrimSpace(os.Getenv("DAILY_DEV_TOKEN")); t != "" {
		m.token, m.st = t, stMenu
	}
	return m
}

func (m model) Init() tea.Cmd { return m.sp.Tick }

func (m model) fetch(e Endpoint, vals map[string]string) tea.Cmd {
	return func() tea.Msg {
		r, err := doGet(m.client, m.token, e, vals)
		return resultMsg{r, err}
	}
}

func (m *model) layout() {
	m.vp.Width = m.w - leftW - 6
	m.vp.Height = m.h - 3
}

func (m *model) run(e *Endpoint, vals map[string]string) tea.Cmd {
	m.active, m.vals, m.st, m.cursor = e, vals, stLoading, ""
	return m.fetch(*e, vals)
}

func (m *model) startForm(e *Endpoint) tea.Cmd {
	m.active, m.focus = e, 0
	m.inputs = nil
	for _, f := range e.Fields {
		ti := textinput.New()
		ti.Prompt = ""
		ti.Placeholder = f.Label
		m.inputs = append(m.inputs, ti)
	}
	m.inputs[0].Focus()
	m.st = stForm
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.layout()
		return m, nil
	case spinner.TickMsg:
		var c tea.Cmd
		m.sp, c = m.sp.Update(msg)
		return m, c
	case resultMsg:
		m.st = stMenu
		if msg.err != nil {
			m.vp.SetContent(errSt.Width(m.vp.Width).Render("Error: " + msg.err.Error()))
			m.status = "request failed"
		} else {
			m.cursor = nextCursor(msg.res)
			m.vp.SetContent(renderResult(msg.res, m.vp.Width))
			m.status = "ok"
			if m.cursor != "" {
				m.status = "ok - press n for next page"
			}
		}
		m.vp.GotoTop()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.st {
		case stToken:
			if msg.String() == "enter" {
				if t := strings.TrimSpace(m.input.Value()); t != "" {
					m.token, m.st = t, stMenu
				}
				return m, nil
			}
			var c tea.Cmd
			m.input, c = m.input.Update(msg)
			return m, c
		case stForm:
			return m.updateForm(msg)
		case stLoading:
			return m, nil
		}
		return m.updateMenu(msg)
	}
	return m, nil
}

func (m model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.st = stMenu
		return m, nil
	case "tab", "down", "shift+tab", "up":
		m.inputs[m.focus].Blur()
		d := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			d = -1
		}
		m.focus = (m.focus + d + len(m.inputs)) % len(m.inputs)
		return m, m.inputs[m.focus].Focus()
	case "enter":
		vals := map[string]string{}
		for i, f := range m.active.Fields {
			v := strings.TrimSpace(m.inputs[i].Value())
			if f.Req && v == "" {
				m.inputs[m.focus].Blur()
				m.focus = i
				m.status = f.Label + " is required"
				return m, m.inputs[i].Focus()
			}
			vals[f.Key] = v
		}
		return m, m.run(m.active, vals)
	}
	var c tea.Cmd
	m.inputs[m.focus], c = m.inputs[m.focus].Update(msg)
	return m, c
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	switch k {
	case "q":
		return m, tea.Quit
	case "tab":
		m.rightFocus = !m.rightFocus
		return m, nil
	case "n":
		if m.cursor != "" && m.active != nil {
			v := map[string]string{}
			for a, b := range m.vals {
				v[a] = b
			}
			v["cursor"] = m.cursor
			return m, m.run(m.active, v)
		}
	}
	if m.rightFocus {
		var c tea.Cmd
		m.vp, c = m.vp.Update(msg)
		return m, c
	}
	switch k {
	case "up", "k":
		if m.cur > 0 {
			m.cur--
		}
	case "down", "j":
		if m.cur < len(endpoints)-1 {
			m.cur++
		}
	case "enter":
		e := &endpoints[m.cur]
		if len(e.Fields) == 0 {
			return m, m.run(e, map[string]string{})
		}
		return m, m.startForm(e)
	default:
		var c tea.Cmd
		m.vp, c = m.vp.Update(msg)
		return m, c
	}
	return m, nil
}

func (m model) leftView() string {
	var b strings.Builder
	last := ""
	for i, e := range endpoints {
		if e.Group != last {
			last = e.Group
			b.WriteString("\n" + keySt.Bold(true).Render(e.Group) + "\n")
		}
		line := "  " + e.Name
		if i == m.cur {
			line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("135")).Render("▸ " + e.Name)
		}
		b.WriteString(line + "\n")
	}
	lines := strings.Split(b.String(), "\n")
	// keep the selected item visible
	idx := 0
	for n, l := range lines {
		if strings.Contains(l, "▸") {
			idx = n
		}
	}
	vh := m.h - 4
	start := 0
	if idx >= vh {
		start = idx - vh + 1
	}
	end := start + vh
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n")
}

func (m model) View() string {
	if m.w == 0 {
		return "loading..."
	}
	if m.st == stToken {
		return m.tokenView()
	}
	bc := lipgloss.Color("240")
	lb, rb := lipgloss.Color("135"), bc
	if m.rightFocus {
		lb, rb = bc, lipgloss.Color("135")
	}
	left := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lb).Width(leftW).Height(m.h-3).Padding(0, 1).Render(titleSt.Render("Daily.dev TUI") + m.leftView())

	var right string
	switch m.st {
	case stForm:
		var b strings.Builder
		b.WriteString(titleSt.Render(m.active.Name) + "  " + metaSt.Render("GET "+m.active.Path) + "\n\n")
		for i, f := range m.active.Fields {
			label := f.Label
			if f.Req {
				label += " *"
			}
			b.WriteString(keySt.Render(label) + "\n" + m.inputs[i].View() + "\n\n")
		}
		right = b.String()
	case stLoading:
		right = m.sp.View() + " Loading " + m.active.Path + "..."
	default:
		right = m.vp.View()
	}
	rw := m.w - leftW - 4
	rightBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(rb).Width(rw).Height(m.h-3).Padding(0, 1).Render(right)
	help := "↑/↓: select • enter: run • tab: focus results • ↑/↓/pgup/pgdn: scroll • n: next page • q: quit"
	if m.st == stForm {
		help = "tab/↑/↓: field • enter: run • esc: cancel"
	}
	st := ""
	if m.status != "" {
		st = "  [" + m.status + "]"
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, rightBox) + "\n" + metaSt.Render(help+st)
}

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
